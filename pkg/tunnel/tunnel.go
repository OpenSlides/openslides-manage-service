package tunnel

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// TunnelHelp contains the short help text for the command.
	TunnelHelp = "Creates tcp tunnels to the OpenSlides network"

	// TunnelHelpExtra contains the long help text for the command without the headline.
	TunnelHelpExtra = `Opens local ports to all services and creates
tunnels into the OpenSlides network to the services they belong to.

Without any argument the command creates tunnels to all services at their
default ports. To specify tunnels, the flag "-L" can be used or the name
of a known service as argument.

The syntax of the -L argument is the same as "ssh -L". The argument can be
used more then one time.

Example:

Open tunnels to all known services:
$ openslides tunnel

Open tunnels to the datastore and auth

$ openslides tunnel datastore-reader datastore-writer auth

Open a tunnel to auth on localhost:8080
$ openslides tunnel -L localhost:8080:auth:9004
`
)

// Cmd returns the set-password subcommand.
func Cmd() *cobra.Command {
	services := map[string]string{
		"backend-action":    ":9002:backend:9002",
		"backend-presenter": ":9003:backend:9003",
		"auth":              ":9004:auth:9004",
		"media":             ":9006:media:9006",
		"datastore-reader":  ":9010:datastore-reader:9010",
		"datastore-writer":  ":9011:datastore-writer:9011",
		"autoupdate":        ":9012:autoupdate:9012",
		"icc":               ":9013:icc:9013",
		"postgres":          ":5432:postgres:5432",
		"redis":             ":6379:redis:6379",
		// TODO: Add voting.
	}

	var serviceNames []string
	for service := range services {
		serviceNames = append(serviceNames, service)
	}

	cmd := &cobra.Command{
		Use:       "tunnel",
		Short:     TunnelHelp,
		Long:      TunnelHelp + "\n\n" + TunnelHelpExtra,
		Args:      cobra.OnlyValidArgs,
		ValidArgs: serviceNames,
	}

	bindLocal := cmd.Flags().StringArrayP("bind", "L", nil, "[bind_address:]port:host:hostport")
	addr := cmd.Flags().StringP("address", "a", connection.DefaultAddr, "address of the OpenSlides manage service")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := contextWithInterrupt(context.Background())
		defer cancel()

		cl, close, err := connection.Dial(ctx, *addr)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if len(*bindLocal) == 0 && len(args) == 0 {
			// No tunnel was specified. Use all services.
			args = serviceNames
		}

		for _, arg := range args {
			*bindLocal = append(*bindLocal, services[arg])
		}

		if err := Run(ctx, cl, *bindLocal); err != nil {
			return fmt.Errorf("opening tunnel: %w", err)
		}
		return nil
	}
	return cmd
}

// contextWithInterrupt works like signal.NotifyContext
//
// In only listens on os.Interrupt. If the signal is received two times,
// os.Exit(1) is called.
func contextWithInterrupt(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		select {
		case <-sigint:
		case <-ctx.Done():
			return
		}
		cancel()

		// If the signal was send for the second time, make a hard cut.
		select {
		case <-sigint:
		case <-ctx.Done():
			return
		}
		os.Exit(1)
	}()
	return ctx, cancel
}

// Client

type gRPCClient interface {
	Tunnel(ctx context.Context, opts ...grpc.CallOption) (proto.Manage_TunnelClient, error)
}

// Run calls respective procedure to open a tunnel to one or more services.
func Run(ctx context.Context, gc gRPCClient, bindLocal []string) error {
	addrs, err := tunnelParseArgument(bindLocal)
	if err != nil {
		return fmt.Errorf("parsing bind local: %w", err)
	}

	var wg sync.WaitGroup
	for local, remote := range addrs {
		wg.Add(1)
		go func(local, remote string) {
			defer wg.Done()

			if err := newTunnel(ctx, gc, local, remote); err != nil {
				log.Printf("Error connecting %s to %s: %v", local, remote, err)
				return
			}
		}(local, remote)
	}

	wg.Wait()
	return nil
}

// tunnelParseArgument parses the "-L" argument of the tunnel command.
//
// The argument is a list of the values of all the "-L" arguments.
//
// The keys of the returned map are the local addr (for example ":9002") and the
// values the remote addr (for example "backend:9002").
func tunnelParseArgument(args []string) (map[string]string, error) {
	m := make(map[string]string, len(args))
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		if len(parts) < 3 || len(parts) > 4 {
			return nil, fmt.Errorf("invalid argument %q: expected 2 or 3 colons, got %d", arg, len(parts)-1)
		}

		if len(parts) == 3 {
			parts = append([]string{""}, parts...)
		}
		m[parts[0]+":"+parts[1]] = parts[2] + ":" + parts[3]
	}
	return m, nil
}

// newTunnel creates a new tunnel via grpc to the manage service.
//
// Listens on the given localAddr, sends all data via grpc to the manage server
// and there redirect it to the remoteAddr.
//
// Blocks until the tunnel is closed.
func newTunnel(ctx context.Context, gc gRPCClient, localAddr string, remoteAddr string) error {
	// Listen on localAddr
	lst, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("start listening on %s: %w", localAddr, err)
	}
	defer lst.Close()
	log.Printf("Listen on %s", localAddr)

	// Waiting for connections
	for {
		conn, err := lst.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go func(ctx context.Context, conn net.Conn) {
			defer conn.Close()

			// Open Tunnel
			ctx = metadata.NewOutgoingContext(
				ctx,
				metadata.Pairs("addr", remoteAddr),
			)

			tunnel, err := gc.Tunnel(ctx)
			if err != nil {
				log.Printf("Error creating tunnel: %v", err)
				return
			}

			// Connect the local connection to the tunnel
			if err := copyStream(ctx, tunnel, conn); err != nil {
				log.Printf("Error tunneling data: %v", err)
				return
			}
		}(ctx, conn)
	}
}

// Server

// Tunnel redirects a package to a different service.
func Tunnel(ts proto.Manage_TunnelServer) error {
	md, ok := metadata.FromIncomingContext(ts.Context())
	if !ok {
		return fmt.Errorf("unable to get metadata from context")
	}
	addr := md.Get("addr")
	if len(addr) != 1 {
		return fmt.Errorf("expect one address (host:port) in the meta data")
	}

	conn, err := new(net.Dialer).DialContext(ts.Context(), "tcp", addr[0])
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", addr[0], err)
	}
	defer conn.Close()

	if err := copyStream(ts.Context(), ts, conn); err != nil {
		return fmt.Errorf("connection grpc to server: %w", err)
	}

	return nil
}
