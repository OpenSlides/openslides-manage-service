package set_test

import (
	"context"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/set"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing set.Cmd()", func(t *testing.T) {
		// cmd := set.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing set subcommand: %v", err)
		// }
	})
}

// Client tests

type mockSetClient struct{}

func (m *mockSetClient) Set(ctx context.Context, in *proto.SetRequest, opts ...grpc.CallOption) (*proto.SetResponse, error) {
	return &proto.SetResponse{}, nil
}

func TestSet(t *testing.T) {
	payload := `---\nkey: test_string_boe7ahthu0Fie1Eghai4}`

	t.Run("set organization settings", func(t *testing.T) {
		mc := new(mockSetClient)
		ctx := context.Background()
		if err := set.Run(ctx, mc, "organization", []byte(payload)); err != nil {
			t.Fatalf("running set.Run() failed with error: %v", err)
		}
	})

	t.Run("set with unknown action", func(t *testing.T) {
		mc := new(mockSetClient)
		ctx := context.Background()

		hasErrMsg := `unknown action "unknown action 7f79hefvvdfget"`
		err := set.Run(ctx, mc, "unknown action 7f79hefvvdfget", []byte(payload))
		if err == nil {
			t.Fatalf("running set.Run() with unknown action should return error but it does not")
		}
		if !strings.Contains(err.Error(), hasErrMsg) {
			t.Fatalf("running set.Run() with unknown action, got error message %q, expected %q", err.Error(), hasErrMsg)
		}
	})
}

// Server tests

// No tests here because the code does to do anything interesting.
