package server

import (
	"fmt"
	"net"

	pb "github.com/OpenSlides/openslides-manage-service/management"
	"google.golang.org/grpc"
)

// Run starts the manage server.
func Run(cfg *Config) error {
	lis, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("listen on addr %s: %w", cfg.Addr(), err)
	}

	s := grpc.NewServer()
	pb.RegisterManageServer(s, &manageServer{cfg: cfg})

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("running service: %w", err)
	}
	return nil
}
