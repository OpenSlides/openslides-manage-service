package server

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/OpenSlides/openslides-manage-service/management"
)

type manageServer struct {
	pb.UnimplementedManageServer
}

func (m *manageServer) ResetPassword(ctx context.Context, in *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	hash, err := hashPassword(ctx, in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := setPassword(ctx, int(in.UserID), hash); err != nil {
		return nil, fmt.Errorf("set password: %w", err)
	}
	return new(pb.ResetPasswordResponse), nil
}

func (m *manageServer) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
