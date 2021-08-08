package setpassword_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// Client tests.

type mockSetpasswordClient struct {
	givenUserID   int64
	givenPassword string
	err           error
}

func (m *mockSetpasswordClient) SetPassword(ctx context.Context, in *proto.SetPasswordRequest, opts ...grpc.CallOption) (*proto.SetPasswordResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	m.givenUserID = in.UserID
	m.givenPassword = in.Password
	return &proto.SetPasswordResponse{}, nil
}

func TestSetPassword(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("set password of some user", func(t *testing.T) {
		mc := new(mockSetpasswordClient)
		inUserID := int64(6268133) // some random user ID
		inPassword := "my_expected_password_vie2aiFoo3"

		if err := setpassword.Run(ctx, mc, inUserID, inPassword); err != nil {
			t.Fatalf("running setpassword.Run() failed with error: %v", err)
		}

		if mc.givenUserID != inUserID {
			t.Fatalf("gRPC client was called with %d, expected %d", mc.givenUserID, inUserID)
		}

		if mc.givenPassword != inPassword {
			t.Fatalf("gRPC client was called with %q, expected %q", mc.givenPassword, inPassword)
		}
	})

	t.Run("with error", func(t *testing.T) {
		myerror := errors.New("my error")
		mc := new(mockSetpasswordClient)
		mc.err = myerror

		err := setpassword.Run(ctx, mc, 62681321, "testvaluefoo jcigh")

		if !errors.Is(err, myerror) {
			t.Fatalf("setpassword.Run() returned '%v', expected error wrapping '%v'", err, myerror)
		}
	})
}

// Server tests

type mockAuth struct{}

func (m *mockAuth) Hash(ctx context.Context, password string) (string, error) {
	return "hash:" + password, nil
}

type mockDatastore struct {
	content map[string]json.RawMessage
}

func (m *mockDatastore) Set(ctx context.Context, fqfield string, value json.RawMessage) error {
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
	if _, err := setpassword.SetPassword(context.Background(), in, md, ma); err != nil {
		t.Fatalf("running SetPassword() failed: %v", err)
	}

	key := "user/1/password"
	expected := fmt.Sprintf("%q", "hash:my_password")
	got := string(md.content[key])
	if expected != got {
		t.Fatalf("wrong (mock) datastore key %s, expected %q, got %q ", key, expected, got)
	}
}
