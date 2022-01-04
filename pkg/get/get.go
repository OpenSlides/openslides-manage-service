package get

import (
    "context"
    "fmt"

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
Use options to narrow down output.`
)

// Cmd returns the get subcommand.
func Cmd(cmd *cobra.Command, cfg connection.Params) *cobra.Command {
    cmd.Use = "get"
    cmd.Short = GetHelp
    cmd.Long = GetHelp + "\n\n" + GetHelpExtra
    cmd.Args = cobra.ExactArgs(1)

    existsHelpText := "check only for existance (requires --filter)"
    exists := cmd.Flags().Bool("exists", false, existsHelpText)

    filterHelpText := "provide a filter based on a collection field"
    filter := cmd.Flags().StringToString("filter", nil, filterHelpText)

    fieldsHelpText := "only include the provided fields in output"
    fields := cmd.Flags().StringSlice("fields", nil, fieldsHelpText)

    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        // validate flags
        if *exists {
            if len(*filter) == 0 {
                return fmt.Errorf("filter missing, needed to check existance of a model")
            }
        }
        if len(*filter) > 1 {
            return fmt.Errorf("only one filter is allowed")
        }

        ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
        defer cancel()

        cl, close, err := connection.Dial(ctx, cfg.Addr(), cfg.PasswordFile(), !cfg.NoSSL())
        if err != nil {
            return fmt.Errorf("connecting to gRPC server: %w", err)
        }
        defer close()

        collection := args[0]
        if err := Run(ctx, cl, collection, *exists, *filter, *fields); err != nil {
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
func Run(ctx context.Context, gc gRPCClient, collection string, exists bool, filter map[string]string, fields []string) error {
    in := &proto.GetRequest{}
    in.Collection = collection
    in.Exists = exists
    in.Filter = filter
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
    Exists(ctx context.Context, collection string, filter map[string]string) (bool, error)
    Filter(ctx context.Context, collection string, filter map[string]string, fields []string) (string, error)
    GetAll(ctx context.Context, collection string, fields []string) (string, error)
}

// This function is the server side entrypoint for this package.
func Get(ctx context.Context, in *proto.GetRequest, ds datastorereader) (*proto.GetResponse, error) {
    //resp := &proto.GetResponse{}

		// if --exists was provided do an /exists request
    if in.Exists {
        res, err := ds.Exists(ctx, in.Collection, in.Filter)
        if err != nil {
            return nil, fmt.Errorf("requesting datastore/exists: %w", err)
        }
        return &proto.GetResponse{Value: fmt.Sprintf("%v", res)}, nil
    }
		// if --filter was provided do an /filter request
    if len(in.Filter) == 1 {
        res, err := ds.Filter(ctx, in.Collection, in.Filter, in.Fields)
        if err != nil {
            return nil, fmt.Errorf("requesting datastore/filter: %w", err)
        }
        //return &proto.GetResponse{Value: fmt.Sprintf("%v", string(res[:]))}, nil
        return &proto.GetResponse{Value: fmt.Sprintf("%s", res)}, nil
    }
		// else do an /get_all request
    res, err := ds.GetAll(ctx, in.Collection, in.Fields)
    if err != nil {
        return nil, fmt.Errorf("requesting datastore/get_all: %w", err)
    }
    //return &proto.GetResponse{Value: fmt.Sprintf("%v", string(res[:]))}, nil
    return &proto.GetResponse{Value: fmt.Sprintf("%s", res)}, nil
}
