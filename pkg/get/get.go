package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// GetHelp contains the short help text for the command.
	GetHelp = "Get models from the datastore"

	// GetHelpExtra contains the long help text for the command without
	// the headline.
	GetHelpExtra = `Provide a collection to list contained models.
Use options to narrow down output.

Examples:
  openslides get user --fields first_name,last_name --filter is_active=false
  openslides get agenda_item --exists --filter meeting_id=1,closed=true`
)

// Cmd returns the get subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get collection",
		Short: GetHelp,
		Long:  GetHelp + "\n\n" + GetHelpExtra,
		Args:  cobra.ExactArgs(1),
	}
	cp := connection.Unary(cmd)

	existsHelpText := "check only for existance (requires --filter)"
	exists := cmd.Flags().Bool("exists", false, existsHelpText)

	fieldsHelpText := "only include the provided fields in output"
	fields := cmd.Flags().StringSlice("fields", nil, fieldsHelpText)

	filterHelpText := "provide a simple filter using the '=' operator, multiple filters are AND'ed"
	filter := cmd.Flags().StringToString("filter", nil, filterHelpText)

	filterRawHelpText := "provide a filter in raw JSON, enabling all operators and much more complex queries"
	filterRaw := cmd.Flags().String("filter-raw", "", filterRawHelpText)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// validate flags
		if *filter != nil && *filterRaw != "" {
			return fmt.Errorf("simple and raw filter provided, only either is allowed")
		}
		if *exists {
			if *filter == nil && *filterRaw == "" {
				return fmt.Errorf("filter missing, needed to check existance of a model")
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		collection := args[0]
		if err := Run(ctx, cl, collection, *exists, *filter, *filterRaw, *fields); err != nil {
			return fmt.Errorf("getting collection %s: %w", collection, err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Get(ctx context.Context, in *proto.GetRequest, opts ...grpc.CallOption) (*proto.GetResponse, error)
}

// Run calls respective procedure to get a model
func Run(ctx context.Context, gc gRPCClient, collection string, exists bool, filter map[string]string, filterRaw string, fields []string) error {
	in := &proto.GetRequest{}
	in.Collection = collection
	in.Exists = exists
	in.Filter = filter
	in.FilterRaw = filterRaw
	in.Fields = fields

	resp, err := gc.Get(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service: %s", s.Message())
	}

	fmt.Printf("%s\n", resp.Value)
	return nil
}

// Server

type datastorereader interface {
	Exists(ctx context.Context, collection string, filter string) (bool, error)
	Filter(ctx context.Context, collection string, filter string, fields string) (string, error)
	GetAll(ctx context.Context, collection string, fields string) (string, error)
}

// Get queries the datastore-reader for requested models
// This function is the server side entrypoint for this package.
func Get(ctx context.Context, in *proto.GetRequest, ds datastorereader) (*proto.GetResponse, error) {
	var filter string = in.FilterRaw
	if filter == "" {
		filter = makeFilterString(in.Filter)
	}
	var fields = makeFieldsString(in.Fields)
	// if --exists was provided do a /exists request
	if in.Exists {
		res, err := ds.Exists(ctx, in.Collection, filter)
		if err != nil {
			return nil, fmt.Errorf("requesting datastore/exists: %w", err)
		}
		return &proto.GetResponse{Value: fmt.Sprintf("%v", res)}, nil
	}
	// if --filter or --filter-raw was provided do a /filter request
	if filter != "" {
		res, err := ds.Filter(ctx, in.Collection, filter, fields)
		if err != nil {
			return nil, fmt.Errorf("requesting datastore/filter: %w", err)
		}
		return &proto.GetResponse{Value: fmt.Sprintf("%s", res)}, nil
	}
	// else do a /get_all request
	res, err := ds.GetAll(ctx, in.Collection, fields)
	if err != nil {
		return nil, fmt.Errorf("requesting datastore/get_all: %w", err)
	}
	return &proto.GetResponse{Value: fmt.Sprintf("%s", res)}, nil
}

// constructs the filter string used in DS request from the filter map
func makeFilterString(filterMap map[string]string) string {
	parts := []string{}
	for k, v := range filterMap {
		parts = append(parts, fmt.Sprintf(
			`{ "field": "%s", "value": "%s", "operator": "=" }`, k, v,
		))
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return fmt.Sprintf(
		`{ "and_filter" : [ %s ] }`, strings.Join(parts, ", "),
	)
}

// constructs the fields string used in DS request from the fields array
func makeFieldsString(fieldsArray []string) string {
	str := ""
	if len(fieldsArray) > 0 {
		str = fmt.Sprintf(`[ "%s" ]`, strings.Join(fieldsArray, `", "`))
	}
	return str
}
