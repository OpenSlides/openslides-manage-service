package connection

import (
	"context"
	"encoding/base64"
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

type BasicAuth struct {
	password string
}

func (a BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	enc := base64.StdEncoding.EncodeToString([]byte(a.password))
	return map[string]string{
		"Authorization": "Basic " + enc,
	}, nil
}

func (a BasicAuth) RequireTransportSecurity() bool {
	return false
}

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context, address string) (proto.ManageClient, func() error, error) {
	creds := BasicAuth{
		password: "foo",
	}
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithPerRPCCredentials(creds))
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}
	return proto.NewManageClient(conn), conn.Close, nil
}
