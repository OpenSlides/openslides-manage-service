package connection

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	// DefaultAddr holds the default host and port to be used for the gRPC connection that is established by some commands.
	DefaultAddr = "localhost:8000"

	// DefaultTimeout holds the default timeout for the gRPC connection that is established by some commands.
	DefaultTimeout = 5 * time.Second

	// AuthHeader is the name of the header that contains the basic authoriztation password.
	AuthHeader = "authorization"
)

// BasicAuth contains the password used in basic authorization process. The password has to be encoded in base64.
// The struct implements https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
type BasicAuth struct {
	password []byte
}

// GetRequestMetadata gets the current request metadata.
// See https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
func (a BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": base64.StdEncoding.EncodeToString(a.password),
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
// See https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
func (a BasicAuth) RequireTransportSecurity() bool {
	return false
}

// CheckAuthFromContext checks if the basic authorization header is present and contains the correct password.
func CheckAuthFromContext(ctx context.Context, secret []byte) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("unable to get metadata from context")
	}
	a := md.Get("authorization")
	password, err := base64.StdEncoding.DecodeString(a[0])
	if err != nil {
		return fmt.Errorf("decoding password (base64): %w", err)
	}

	if !bytes.Equal(password, secret) {
		return fmt.Errorf("password does not match")
	}
	return nil
}

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context, address, passwordFile string) (proto.ManageClient, func() error, error) {
	pw := []byte(shared.DevelopmentPassword)
	dev, _ := strconv.ParseBool(os.Getenv("OPENSLIDES_DEVELOPMENT"))
	// Error value does not matter here. In case of an error dev is false and
	// this is the expected behavior.
	if !dev {
		var err error
		filePW, err := os.ReadFile(passwordFile)
		if err != nil {
			return nil, nil, fmt.Errorf("reading manage auth password file %q: %w", passwordFile, err)
		}
		pw = filePW
	}
	creds := BasicAuth{
		password: pw,
	}

	// TODO: https://itnext.io/practical-guide-to-securing-grpc-connections-with-go-and-tls-part-1-f63058e9d6d1
	transCreds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(transCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(creds),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContext(): %w", err)
	}
	return proto.NewManageClient(conn), conn.Close, nil
}
