package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/server"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc/status"
)

const (
	timeout = 1 * time.Second
)

func healthcheck() error {
	cfg := server.ConfigFromEnv(os.LookupEnv)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cl, close, err := connection.Dial(
		ctx,
		"localhost:"+cfg.Port,
		cfg.ManageAuthPasswordFile,
		false,
	)
	if err != nil {
		return fmt.Errorf("connecting to gRPC server: %w", err)
	}
	defer close()

	resp, err := cl.Health(ctx, &proto.HealthRequest{})
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
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
