package create_user

import (
	"context"
	"errors"

	"github.com/OpenSlides/openslides-manage-service/proto"
)

type CreateUserer struct {
}

func (s CreateUserer) CreateUser(ctx context.Context, in *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
