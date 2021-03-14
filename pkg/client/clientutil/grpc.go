package clientutil

import (
	"context"
	"log"

	pb "github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Connect creates a grpc connection to the server.
func Connect(ctx context.Context, address string) pb.ManageClient {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//TODO: Close conn?
	return pb.NewManageClient(conn)
}
