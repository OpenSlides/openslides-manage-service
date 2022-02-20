package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
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
		Short: "...TODO...",
		Args:  cobra.NoArgs,
	}
	return setupMigrationCmd(cmd, !withIntervalFlag)
}

func setupMigrationCmd(cmd *cobra.Command, withInterval bool) *cobra.Command {
	cp := connection.Unary(cmd)

	var interval time.Duration
	if withInterval {
		intervalHelpText := "interval of progress calls on running migrations, " +
			"set 0 to disable progress and let the command return immediately"
		interval = *cmd.Flags().Duration("interval", defaultInterval, intervalHelpText)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, cmd.Use, interval); err != nil {
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
func Run(ctx context.Context, gc gRPCClient, command string, interval time.Duration) error {
	mR, err := runMigrationsCmd(ctx, gc, command)
	if err != nil {
		return fmt.Errorf("running migrations command: %w", err)
	}
	fmt.Printf(mR.String())

	if interval == 0 || mR.Status != 1 {
		return nil
	}

	for {
		time.Sleep(interval)
		mR, err := runMigrationsCmd(ctx, gc, "progress")
		if err != nil {
			return fmt.Errorf("running migrations command: %w", err)
		}
		fmt.Printf(mR.String())
		if mR.Status != 1 {
			break
		}
	}

	return nil
}

type migrationResponse struct {
	Success   bool   `json:"success"`
	Status    int    `json:"status"`
	Output    string `json:"output"`
	Exception string `json:"exception"`
}

func (mR migrationResponse) String() string {
	return fmt.Sprintf(
		"Success: %t\nStatus: %d\nOutput: %s\nException: %s\n",
		mR.Success, mR.Status, mR.Output, mR.Exception)
}

func runMigrationsCmd(ctx context.Context, gc gRPCClient, command string) (migrationResponse, error) {
	req := &proto.MigrationsRequest{
		Command: command,
	}

	resp, err := gc.Migrations(ctx, req)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return migrationResponse{}, fmt.Errorf("calling manage service (running migrations command): %s", s.Message())
	}

	var mR migrationResponse
	if err := json.Unmarshal(resp.Response, &mR); err != nil {
		return migrationResponse{}, fmt.Errorf("decoding migration response %q: %w", string(resp.Response), err)
	}

	return mR, nil
}

// Server

type action interface {
	Migrations(ctx context.Context, command string) (json.RawMessage, error)
}

// Migrations runs a migrations command.
func Migrations(ctx context.Context, in *proto.MigrationsRequest, a action) (*proto.MigrationsResponse, error) {
	result, err := a.Migrations(ctx, in.Command)
	if err != nil {
		return nil, fmt.Errorf("requesting backend migrations command %q: %w", in.Command, err)
	}
	return &proto.MigrationsResponse{Response: result}, nil

}
