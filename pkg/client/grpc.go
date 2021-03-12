package client

import (
	"context"
	"log"

	"github.com/OpenSlides/openslides-manage-service/management"
	"google.golang.org/grpc"
)

func connect(ctx context.Context, address string) management.ManageClient {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//TODO: Close conn?
	return management.NewManageClient(conn)
}
