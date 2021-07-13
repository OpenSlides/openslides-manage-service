package setpassword

import (
	"context"
	"encoding/json"

	"github.com/OpenSlides/openslides-manage-service/proto"
)

type datastore interface {
	Set(fqfield string, value json.RawMessage) error
}

// SetPassword gets the hash and sets the password for the given user.
func SetPassword(ctx context.Context, in *proto.SetPasswordRequest, ds datastore) (*proto.SetPasswordResponse, error) {
	// h, err := services.Auth.Hash(in.Password)
	// if err != nil {
	// 	return nil, fmt.Errorf("hashing passwort: %w", err)
	// }

	// _ = h

	return nil, nil
}
