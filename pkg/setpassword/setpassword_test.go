package setpassword_test

import (
	"encoding/json"
	"testing"
)

type MockAuth struct{}

func (m *MockAuth) Hash(password string) (string, error) {
	return "hashed_password", nil
}

type MockDatastore struct{}

func (d *MockDatastore) Write(fqid string, fields map[string]json.RawMessage) error {
	return nil
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
