package server

import (
	"fmt"
	"net"

	pb "github.com/OpenSlides/openslides-manage-service/management"
	"google.golang.org/grpc"
)

// Run starts the manage server.
func Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on addr %s: %w", addr, err)
	}

	s := grpc.NewServer()
	pb.RegisterManageServer(s, &manageServer{})

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("running service: %w", err)
	}
	return nil
}
