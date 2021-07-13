package setpassword_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/proto"
)

const hashedPassword = "hashed_password"

type mockAuth struct{}

func (m *mockAuth) Hash(password string) (string, error) {
	return hashedPassword, nil
}

type mockDatastore struct {
	content map[string]json.RawMessage
}

func (m *mockDatastore) Set(fqfield string, value json.RawMessage) error {
	return nil
}

func TestSetPasswordServerAll(t *testing.T) {
	ma := new(mockAuth)
	_ = ma
	md := new(mockDatastore)
	ctx := context.Background()
	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: "password",
	}
	if _, err := setpassword.SetPassword(ctx, in, md); err != nil {
		t.Fatalf("running SetPassword() failed: %v", err)
	}
	key := "user/1/password"
	got := md.content[key]
	if hashedPassword != string(got) {
		t.Fatalf("wrong (mock) datastore key %s, expected %q, got %q ", key, hashedPassword, got)
	}
}

func TestSetPasswordServer(t *testing.T) {
	// ma := new(MockAuth)
	// md := new(MockDatastore)
	// services := &connection.Services{
	// 	Auth:      ma,
	// 	Datastore: md,
	// }

	// t.Run("get hash form auth server", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	in := &proto.SetPasswordRequest{
	// 		UserID:   1,
	// 		Password: "password",
	// 	}
	// 	resp, err := setpassword.SetPassword(ctx, in, services)
	// 	if err != nil {
	// 		t.Fatalf("setting password: %v", err)
	// 	}
	// 	resp
	// })

}
