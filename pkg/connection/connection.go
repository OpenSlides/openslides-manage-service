package connection

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

const (
	// DefaultAddr holds the default host and port to be used for the gRPC connection that is established by some commands.
	DefaultAddr = "localhost:9008"

	// DefaultTimeout holds the default timeout for the gRPC connection that is established by some commands.
	DefaultTimeout = 5 * time.Second
)

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context, address string) (proto.ManageClient, func() error, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}

	return proto.NewManageClient(conn), conn.Close, nil
}
