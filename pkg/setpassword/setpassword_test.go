package setpassword_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Client tests.

type mockSetpasswordClient struct {
	expected struct {
		UserID   int64
		Password string
	}
	called bool
}

func (m *mockSetpasswordClient) SetPassword(ctx context.Context, in *proto.SetPasswordRequest, opts ...grpc.CallOption) (*proto.SetPasswordResponse, error) {
	m.called = true
	if m.expected.UserID != in.UserID {
		return nil, fmt.Errorf("wrong user ID, expected %d, got %d", m.expected.UserID, in.UserID)
	}
	if m.expected.Password != in.Password {
		return nil, fmt.Errorf("wrong password, expected %q, got %q", m.expected.Password, in.Password)
	}
	return &proto.SetPasswordResponse{}, nil
}

func TestSetPassword(t *testing.T) {
	t.Run("set password of some user", func(t *testing.T) {
		mc := new(mockSetpasswordClient)
		mc.expected.UserID = int64(6268132) // some random user ID
		mc.expected.Password = "my_expected_password_vie2aiFoo3"
		ctx := context.Background()
		if err := setpassword.Run(ctx, mc, mc.expected.UserID, mc.expected.Password); err != nil {
			t.Fatalf("running setpassword.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
	})
}

// Server tests

type mockAuth struct{}

func (m *mockAuth) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

type mockDatastore struct {
	content map[string]json.RawMessage
}

func (m *mockDatastore) Set(fqfield string, value json.RawMessage) error {
	if m.content == nil {
		m.content = make(map[string]json.RawMessage)
	}
	m.content[fqfield] = value
	return nil
}

func TestSetPasswordServerAll(t *testing.T) {
	md := new(mockDatastore)
	ma := new(mockAuth)
	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: "my_password",
	}
	if _, err := setpassword.SetPassword(in, md, ma); err != nil {
		t.Fatalf("running SetPassword() failed: %v", err)
	}
	key := "user/1/password"
	expected := fmt.Sprintf("%q", "hash:my_password")
	got := string(md.content[key])
	if expected != got {
		t.Fatalf("wrong (mock) datastore key %s, expected %q, got %q ", key, expected, got)
	}
}
