package createuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
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
	CreateUserHelpExtra = `This command creates a new user with the given given YAML or JSON formatted user
data including default password and organization management level. Provide the
user data directly or use the --file flag with a file or use this flag with - to
read from stdin.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-user [user-data]",
		Short: CreateUserHelp,
		Long:  CreateUserHelp + "\n\n" + CreateUserHelpExtra,
		Args:  cobra.RangeArgs(0, 1),
	}
	cp := connection.Unary(cmd)

	userFileHelpText := "YAML or JSON file with user data " +
		"(required fields: username, default_password; " +
		"extra fields: first_name, last_name, email, organization_management_level); " +
		"you can use - to provide the payload via stdin"
	userFile := cmd.Flags().StringP("file", "f", "", userFileHelpText)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		args = append(args, "") // This is to ensure that the slice always has enough values.
		userData, err := shared.InputOrFileOrStdin(args[0], *userFile)
		if err != nil {
			return fmt.Errorf("reading user data from positional argument or file or stdin: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, userData); err != nil {
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
func Run(ctx context.Context, gc gRPCClient, userData []byte) error {

	in := &proto.CreateUserRequest{}
	if err := yaml.Unmarshal(userData, in); err != nil {
		return fmt.Errorf("unmarshalling user data: %w", err)
	}

	if in.Username == "" {
		return fmt.Errorf("missing username in user data")
	}
	if in.DefaultPassword == "" {
		return fmt.Errorf("missing default_password in user data")
	}
	if in.OrganizationManagementLevel != "" {
		if err := checkOrganizationManagementLevel(in.OrganizationManagementLevel); err != nil {
			return fmt.Errorf("wrong value for organization_management_level in user data: %w", err)
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

type backendAction interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// CreateUser creates the given user.
// This function is the server side entrypoint for this package.
func CreateUser(ctx context.Context, in *proto.CreateUserRequest, ba backendAction) (*proto.CreateUserResponse, error) {
	name := "user.create"
	payload := []*proto.CreateUserRequest{in}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling action data: %w", err)
	}
	result, err := ba.Single(ctx, name, transform(data))
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
