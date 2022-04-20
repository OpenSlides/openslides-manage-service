package version

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// VersionHelp contains the short help text for the command.
	VersionHelp = "Retrieves OpenSlides client version tag"

	// VersionHelpExtra contains the long help text for the command without
	// the headline.
	VersionHelpExtra = `The version tag is created during client image build.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: VersionHelp,
		Long:  VersionHelp + "\n\n" + VersionHelpExtra,
		Args:  cobra.NoArgs,
	}
	cp := connection.Unary(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl); err != nil {
			return fmt.Errorf("run version call: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Version(ctx context.Context, in *proto.VersionRequest, opts ...grpc.CallOption) (*proto.VersionResponse, error)
}

// Run calls respective procedure via given gRPC client.
func Run(ctx context.Context, gc gRPCClient) error {
	in := &proto.VersionRequest{}

	resp, err := gc.Version(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (retrieving version): %s", s.Message())
	}

	fmt.Println(strings.TrimSpace(resp.Version))
	return nil
}

// Server

// Version retrieves the version tag from the client container.
// This function is the server side entrypoint for this package.
func Version(ctx context.Context, in *proto.VersionRequest, clientVersionURL *url.URL) (*proto.VersionResponse, error) {
	addr := clientVersionURL.String()
	req, err := http.NewRequestWithContext(ctx, "GET", addr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request to client service: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to client service at %q: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return nil, fmt.Errorf("got response %q: %q", resp.Status, body)
	}

	encodedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &proto.VersionResponse{Version: string(encodedResp)}, nil
}
