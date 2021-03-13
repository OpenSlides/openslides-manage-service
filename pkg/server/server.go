package server

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/OpenSlides/openslides-manage-service/proto"
)

type manageServer struct {
	pb.UnimplementedManageServer
	cfg *Config
}

func (m *manageServer) SetPassword(ctx context.Context, in *pb.SetPasswordRequest) (*pb.SetPasswordResponse, error) {
	hash, err := hashPassword(ctx, m.cfg, in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := setPassword(ctx, m.cfg, int(in.UserID), hash); err != nil {
		return nil, fmt.Errorf("set password: %w", err)
	}
	return new(pb.SetPasswordResponse), nil
}

func (m *manageServer) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
