package connection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context, address string) (proto.ManageClient, func() error, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}

	return proto.NewManageClient(conn), conn.Close, nil
}
