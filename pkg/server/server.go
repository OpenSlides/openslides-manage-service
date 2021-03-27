package server

import (
	"fmt"
	"net"

	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/util"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Run starts the manage server.
func Run(cfg *util.ServerConfig) error {
	addr := cfg.Addr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on addr %s: %w", addr, err)
	}

	s := grpc.NewServer()
	proto.RegisterManageServer(s, newServer(cfg))

	fmt.Printf("Running manage service on %s", addr)

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("running service: %w", err)
	}
	return nil
}

// Server implements the manage methods on server side.
type Server struct {
	setpassword.ServiceSetPassword
	createuser.ServerCreateUser
}

func newServer(cfg *util.ServerConfig) Server {
	return Server{
		ServiceSetPassword: setpassword.ServiceSetPassword{Config: cfg},
	}
}
