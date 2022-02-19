package createuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// CreateUserHelp contains the short help text for the command.
	CreateUserHelp = "Creates a new user in OpenSlides"

	// CreateUserHelpExtra contains the long help text for the command without
	// the headline.
	CreateUserHelpExtra = `This command creates a new user with the given user
data including default password and organization management level.`
)

// Cmd returns the subcommand.
func Cmd(cmd *cobra.Command, cfg connection.Params) *cobra.Command {
	cmd.Use = "create-user"
	cmd.Short = CreateUserHelp
	cmd.Long = CreateUserHelp + "\n\n" + CreateUserHelpExtra
	cmd.Args = cobra.NoArgs

	// TODO: Let this read from stdin and also accept positional argument with user data in YAML oder JSON to be consistent with set command.
	userFileHelpText := "custom YAML file with user data " +
		"(required fields: username, default_password; " +
		"extra fields: first_name, last_name, email, organization_management_level)"
	userFile := cmd.Flags().StringP("file", "f", "", userFileHelpText)
	cmd.MarkFlagRequired("file")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		cmd.Flags()

		cl, close, err := connection.Dial(ctx, cfg.Addr(), cfg.PasswordFile(), !cfg.NoSSL())
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, *userFile); err != nil {
			return fmt.Errorf("creating user: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	CreateUser(ctx context.Context, in *proto.CreateUserRequest, opts ...grpc.CallOption) (*proto.CreateUserResponse, error)
}

// Run calls respective procedure to set password of the given user.
func Run(ctx context.Context, gc gRPCClient, userFile string) error {
	userData, err := os.ReadFile(userFile)
	if err != nil {
		return fmt.Errorf("reading user file %q: %w", userFile, err)
	}

	in := &proto.CreateUserRequest{}
	if err := yaml.Unmarshal(userData, in); err != nil {
		return fmt.Errorf("unmarshalling user YAML file: %w", err)
	}

	if in.Username == "" {
		return fmt.Errorf("missing username in user YAML file")
	}
	if in.DefaultPassword == "" {
		return fmt.Errorf("missing default_password in user YAML file")
	}
	if in.OrganizationManagementLevel != "" {
		if err := checkOrganizationManagementLevel(in.OrganizationManagementLevel); err != nil {
			return fmt.Errorf("wrong value for organization_management_level in user YAML file: %w", err)
		}
	}

	resp, err := gc.CreateUser(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service: %s", s.Message())
	}
	fmt.Printf("User %d created successfully.\n", resp.UserID)

	return nil
}

func checkOrganizationManagementLevel(v string) error {
	enum := []string{"superadmin", "can_manage_organization", "can_manage_users"}
	for _, e := range enum {
		if v == e {
			return nil
		}
	}
	return fmt.Errorf("wrong value %q", v)
}

// Server

type action interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// CreateUser creates the given user.
// This function is the server side entrypoint for this package.
func CreateUser(ctx context.Context, in *proto.CreateUserRequest, a action) (*proto.CreateUserResponse, error) {
	name := "user.create"
	payload := []*proto.CreateUserRequest{in}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling action data: %w", err)
	}
	result, err := a.Single(ctx, name, transform(data))
	if err != nil {
		return nil, fmt.Errorf("requesting backend action %q: %w", name, err)
	}

	var ids []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(result, &ids); err != nil {
		return nil, fmt.Errorf("unmarshalling action result %q: %w", string(result), err)
	}
	if len(ids) != 1 {
		return nil, fmt.Errorf("wrong lenght of action result, expected 1 item, got %d", len(ids))
	}
	return &proto.CreateUserResponse{UserID: int64(ids[0].ID)}, nil
}

// transform changes some JSON keys so we can use OpenSlides' template fields.
func transform(b []byte) []byte {
	fields := map[string]string{
		"committee__management_level": "committee_$_management_level",
		"group__ids":                  "group_$_ids",
	}
	s := string(b)
	for old, new := range fields {
		s = strings.ReplaceAll(s,
			fmt.Sprintf(`"%s":`, old),
			fmt.Sprintf(`"%s":`, new),
		)

	}
	return []byte(s)
}
