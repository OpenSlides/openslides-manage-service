package server

import (
	"fmt"
	"net"

	pb "github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Run starts the manage server.
func Run(cfg *Config) error {
	addr := cfg.Addr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on addr %s: %w", addr, err)
	}

	s := grpc.NewServer()
	pb.RegisterManageServer(s, &manageServer{cfg: cfg})

	fmt.Printf("Running manage service on %s", addr)

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("running service: %w", err)
	}
	return nil
}
