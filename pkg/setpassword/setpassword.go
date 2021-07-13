package setpassword

import (
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/proto"
)

type datastore interface {
	Set(fqfield string, value json.RawMessage) error
}

type auth interface {
	Hash(password string) (string, error)
}

// SetPassword gets the hash and sets the password for the given user.
func SetPassword(in *proto.SetPasswordRequest, ds datastore, auth auth) (*proto.SetPasswordResponse, error) {
	hash, err := auth.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing passwort: %w", err)
	}

	key := fmt.Sprintf("user/%d/password", in.UserID)
	value := []byte(`"` + hash + `"`)
	if err := ds.Set(key, value); err != nil {
		return nil, fmt.Errorf("writing key %q to datastore: %w", key, err)
	}

	return &proto.SetPasswordResponse{}, nil
}
