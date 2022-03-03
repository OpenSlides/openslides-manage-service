package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	host    = "localhost"
	timeout = 1 * time.Second
)

func healthcheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	port, ok := os.LookupEnv("MANAGE_PORT")
	if !ok {
		port = "9008"
	}
	conn, err := grpc.DialContext(ctx, host+port,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("creating gRPC client connection with grpc.DialContext(): %w", err)
	}
	defer conn.Close()

	resp, err := proto.NewManageClient(conn).Health(ctx, &proto.HealthRequest{})
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service: %s", s.Message())
	}

	if !resp.Healthy {
		return fmt.Errorf("manage service is unhealthy")
	}
	return nil
}

func main() {
	if err := healthcheck(); err != nil {
		os.Exit(1)
	}
}
