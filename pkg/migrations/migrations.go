package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// MigrationsHelp contains the short help text for the command.
	MigrationsHelp = "Wrapper to the OpenSlides backend migration tool for applying migrations to the datastore."

	// MigrationsHelpExtra contains the long help text for the command without
	// the headline.
	MigrationsHelpExtra = `See help text for the repective commands for more information.`

	defaultInterval  = 1 * time.Second
	withIntervalFlag = true
	migrationRunning = "migration_running"

	maxCallRecvMsgSize = 1073741824 // corresponds to 1 GB
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrations",
		Short: MigrationsHelp,
		Long:  MigrationsHelp + "\n\n" + MigrationsHelpExtra,
	}

	// TODO: Verbose flag

	cmd.AddCommand(
		migrateCmd(),
		finalizeCmd(),
		resetCmd(),
		clearCollectionfieldTablesCmd(),
		statsCmd(),
		progressCmd(),
	)

	return cmd
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Prepare migrations but do not apply them to the datastore.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, withIntervalFlag)
}

func finalizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize",
		Short: "Prepare migrations and apply them to the datastore.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, withIntervalFlag)
}

func resetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset unapplied migrations.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, !withIntervalFlag)
}

func clearCollectionfieldTablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-collectionfield-tables",
		Short: "Clear all data from auxillary tables. Can be done to clean up diskspace, but only when OpenSlides is offline.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, !withIntervalFlag)
}

func statsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Print some statistics about the current migration state.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, !withIntervalFlag)
}

func progressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "progress",
		Short: "Query the progress of a currently running migration command.",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, !withIntervalFlag)
}

func setupMigrationCmd(cmd *cobra.Command, withInterval bool) *cobra.Command {
	cp := connection.Unary(cmd)

	var interval *time.Duration
	if withInterval {
		intervalHelpText := "interval of progress calls on running migrations, " +
			"set 0 to disable progress and let the command return immediately"
		interval = cmd.Flags().Duration("interval", defaultInterval, intervalHelpText)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		dialCtx, cancel := context.WithTimeout(ctx, *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(dialCtx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, cmd.Use, interval, cp.Timeout); err != nil {
			return fmt.Errorf("running migrations command %q: %w", cmd.Use, err)
		}
		return nil
	}

	return cmd
}

// Client

type gRPCClient interface {
	Migrations(ctx context.Context, in *proto.MigrationsRequest, opts ...grpc.CallOption) (*proto.MigrationsResponse, error)
}

// Run calls respective procedure to run migrations command via given gRPC client.
func Run(ctx context.Context, gc gRPCClient, command string, intervalFlag *time.Duration, timeoutFlag *time.Duration) error {
	mR, err := runMigrationsCmd(ctx, gc, command, *timeoutFlag)
	if err != nil {
		return fmt.Errorf("running migrations command: %w", err)
	}

	var interval time.Duration
	if intervalFlag != nil {
		interval = *intervalFlag
	}
	if interval == 0 || !mR.Running() {
		var mRText string
		var err error
		// when calling any command but 'stats' the fields besides 'output' only clutter our output.
		if command == "stats" {
			mRText, err = mR.GetStats()
		} else {
			mRText, err = mR.GetOutput()
		}
		if err != nil {
			return fmt.Errorf("parsing migrations response: %w", err)
		}
		fmt.Print(mRText)
		return nil
	}

	outCount := 0
	fmt.Print("Progress:\n")
	for {
		time.Sleep(interval)
		mR, err := runMigrationsCmd(ctx, gc, "progress", *timeoutFlag)
		if err != nil {
			return fmt.Errorf("running migrations command: %w", err)
		}

		if mR.Faulty() {
			out, err := mR.GetOutput()
			if err != nil {
				return fmt.Errorf("parsing migrations response: %w", err)
			}
			fmt.Print(out)
		} else {
			out, c := mR.OutputSince(outCount)
			fmt.Print(out)
			outCount = c
		}

		if !mR.Running() {
			break
		}
	}

	return nil
}

// MigrationResponse handles the JSON response from the backend when calling
// migrations commands.
type MigrationResponse struct {
	Success   bool            `json:"success"`
	Status    string          `json:"status"`
	Output    string          `json:"output"`
	Exception string          `json:"exception"`
	Stats     json.RawMessage `json:"stats"`
}

// GetOutput parses and returns the output field of the migrations commands
// reponse for proper display.
// If the reponse conveys an error happened, all fields are returned.
func (mR MigrationResponse) GetOutput() (string, error) {
	if mR.Faulty() {
		return mR.Yaml()
	}
	return mR.Output, nil
}

// GetStats parses and returns the stats field of the migrations commands
// reponse for proper display.
// If the reponse conveys an error happened, all fields are returned.
func (mR MigrationResponse) GetStats() (string, error) {
	if mR.Faulty() {
		return mR.Yaml()
	}
	y, err := yaml.Marshal(mR.Stats)
	if err != nil {
		return "", fmt.Errorf("marshalling to YAML: %w", err)
	}
	return string(y), nil
}

// Yaml returns the migrations command reponse formatted in YAML for proper
// display.
func (mR MigrationResponse) Yaml() (string, error) {
	y, err := yaml.Marshal(mR)
	if err != nil {
		return "", fmt.Errorf("marshalling to YAML: %w", err)
	}
	return string(y), nil
}

// Faulty returns True if the migration command returns success false or any
// exception string.
func (mR MigrationResponse) Faulty() bool {
	return !mR.Success || mR.Exception != ""
}

// Running returns True if the migration command returns status
// "migration_running".
func (mR MigrationResponse) Running() bool {
	return mR.Status == migrationRunning
}

// OutputSince provides the content of the migrations response output. The
// number of given lines where omitted and we also get the number of lines of
// the rest of the content.
func (mR MigrationResponse) OutputSince(lines int) (string, int) {
	s := strings.Split(strings.TrimSpace(mR.Output), "\n")
	out := strings.Join(s[lines:], "\n")
	if lines < len(s) {
		out += "\n"
	}
	return out, len(s)
}

func runMigrationsCmd(ctx context.Context, gc gRPCClient, command string, timeout time.Duration) (MigrationResponse, error) {
	migrCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req := &proto.MigrationsRequest{
		Command: command,
	}

	resp, err := gc.Migrations(
		migrCtx,
		req,
		grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
	)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return MigrationResponse{}, fmt.Errorf("calling manage service (running migrations command): %s", s.Message())
	}

	var mR MigrationResponse
	if err := json.Unmarshal(resp.Response, &mR); err != nil {
		return MigrationResponse{}, fmt.Errorf("unmarshalling migration response %q: %w", string(resp.Response), err)
	}
	return mR, nil
}

// Server

type backendAction interface {
	Migrations(ctx context.Context, command string) (json.RawMessage, error)
}

// Migrations runs a migrations command.
func Migrations(ctx context.Context, in *proto.MigrationsRequest, ba backendAction) (*proto.MigrationsResponse, error) {
	result, err := ba.Migrations(ctx, in.Command)
	if err != nil {
		return nil, fmt.Errorf("requesting backend migrations command %q: %w", in.Command, err)
	}
	return &proto.MigrationsResponse{Response: result}, nil

}
