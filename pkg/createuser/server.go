package createuser

import (
	"context"
	"errors"

	"github.com/OpenSlides/openslides-manage-service/proto"
)

// ServerCreateUser implements the command on server side.
type ServerCreateUser struct{}

// CreateUser TODO
func (s ServerCreateUser) CreateUser(ctx context.Context, in *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	return nil, errors.New("TODO")
}
