package server

import (
	"context"
	"errors"

	pb "github.com/OpenSlides/openslides-manage-service/management"
)

type manageServer struct {
	pb.UnimplementedManageServer
}

func (m *manageServer) ResetPassword(ctx context.Context, in *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	return new(pb.ResetPasswordResponse), nil
}

func (m *manageServer) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
