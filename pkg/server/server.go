package server

import (
	"context"
	"errors"

	"github.com/OpenSlides/openslides-manage-service/pkg/server/serverutil"
	"github.com/OpenSlides/openslides-manage-service/pkg/set_password"
	pb "github.com/OpenSlides/openslides-manage-service/proto"
)

type Server struct {
	set_password.SetPassworder
}

func newServer(cfg *serverutil.Config) Server {
	return Server{
		SetPassworder: set_password.SetPassworder{Config: cfg},
	}
}

func (s Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
