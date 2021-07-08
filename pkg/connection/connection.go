package connection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context) (proto.ManageClient, func() error, error) {
	address := "localhost:9006"

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}

	return proto.NewManageClient(conn), conn.Close, nil
}

// Services contains connections to several services like datastore and auth.s
type Services struct {
	Datastore datastore.Datastore
}
