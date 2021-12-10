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

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing setpassword.Cmd()", func(t *testing.T) {
		// cmd := setpassword.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing set-password subcommand: %v", err)
		// }
	})
}

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

type mockAction struct{}

func (m *mockAction) Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error) {
	switch name {
	case "user.set_password":
		return m.setPassword()
	default:
		return nil, fmt.Errorf("action %q is not defined here", name)
	}
}

func (m *mockAction) setPassword() (json.RawMessage, error) {
	r := []struct {
		Foo string `json:"foo"`
	}{
		{Foo: "bar"},
	}
	encR, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshalling JSON: %w", err)
	}
	return encR, nil
}

func TestSetPasswordServerAll(t *testing.T) {
	ma := new(mockAction)
	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: "my_password",
	}
	if _, err := setpassword.SetPassword(context.Background(), in, ma); err != nil {
		t.Fatalf("running SetPassword() failed: %v", err)
	}
}
