package manage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

const helpTunnel = `Opens local ports to all services and creates
tunnels into the OpenSlides network to the services they belong to.`

func cmdTunnel(cfg *ClientConfig) *cobra.Command {
	cmd := cobra.Command{
		Use:   "tunnel",
		Short: "Creates tcp tunnels to the OpenSlides network.",
		Long:  helpTunnel,
	}

	local := cmd.Flags().StringP("local", "L", ":1080", "Local address the tunnel will listen to")
	remote := cmd.Flags().StringP("remote", "R", "localhost:6379", "Remote address the tunnel will connect to")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Connect to manage server via grpc.
		service, close, err := Dial(context.Background(), cfg.Address)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := newTunnel(service, *local, *remote); err != nil {
			return fmt.Errorf("creating tunnel: %w", err)
		}
		return nil
	}
	return &cmd
}

// newTunnel creates a new tunnel via grpc to the manage service.
//
// Listens on the given localAddr, sends all data via grpc to the manage server
// and there redirect it to the remoteAddr.
//
// Blocks until the tunnel is closed.
func newTunnel(service proto.ManageClient, localAddr string, remoteAddr string) error {
	// Listen on localAddr
	addr := ":1080"
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start listening on %s: %v", addr, err)
	}
	defer lst.Close()
	log.Printf("Listen on %s", addr)

	// Waiting for connections
	for {
		conn, err := lst.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()

			// Open Tunnel
			ctx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("addr", remoteAddr),
			)
			tunnel, err := service.Tunnel(ctx)
			if err != nil {
				log.Printf("Error creating tunnel: %v", err)
				return
			}

			// Connecting the local connection to the tunnel
			if err := copyStream(tunnel, conn); err != nil {
				log.Printf("Error tunneling data: %v", err)
				return
			}
		}(conn)
	}
}

// Tunnel redirects a package to a different service.
func (s *Server) Tunnel(ts proto.Manage_TunnelServer) error {
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
		return fmt.Errorf("connecting to %s: %v", addr[0], err)
	}
	defer conn.Close()

	if err := copyStream(ts, conn); err != nil {
		return fmt.Errorf("connection grpc to server: %v", err)
	}

	return nil
}

// SendReceiver reads and writes from a grpc tunnel connection.
type SendReceiver interface {
	Recv() (*proto.TunnelData, error)
	Send(*proto.TunnelData) error
}

// copyStream connects the grcp connection with a io.ReadWriter.
//
// Blocks until one connection is closed.
func copyStream(sr SendReceiver, rw io.ReadWriter) error {
	fromGRPC := make(chan error, 1)
	fromRW := make(chan error, 1)

	// Message from grpc.
	go func() {

		defer close(fromGRPC)

		for {
			c, err := sr.Recv()
			if err != nil {
				if err != io.EOF {
					fromGRPC <- fmt.Errorf("receiving data: %w", err)
				}
				return
			}

			if _, err = rw.Write(c.Data); err != nil {
				fromGRPC <- fmt.Errorf("writing data data: %w", err)
				return
			}
		}
	}()

	// Message from ReadWriter.
	go func() {
		defer close(fromRW)
		buff := make([]byte, 1_000_000)
		for {
			bs, err := rw.Read(buff)
			if err != nil {
				if err != io.EOF {
					fromRW <- fmt.Errorf("receiving data: %w", err)
				}
				return
			}

			err = sr.Send(&proto.TunnelData{
				Data: buff[:bs],
			})

			if err != nil {
				fromRW <- fmt.Errorf("writing data data: %w", err)
				return
			}
		}
	}()

	// Wait for one side to finish.
	select {
	case err := <-fromGRPC:
		if err != nil {
			return fmt.Errorf("from grpc: %w", err)
		}
	case err := <-fromRW:
		if err != nil {
			return fmt.Errorf("from read writer: %w", err)
		}
	}
	return nil
}
