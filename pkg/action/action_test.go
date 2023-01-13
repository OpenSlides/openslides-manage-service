package action_test

import (
	"context"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/action"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing set.Cmd()", func(t *testing.T) {
		// cmd := action.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing action subcommand: %v", err)
		// }
	})
}

// Client tests

type mockSetClient struct{}

func (m *mockSetClient) Action(ctx context.Context, in *proto.ActionRequest, opts ...grpc.CallOption) (*proto.ActionResponse, error) {
	return &proto.ActionResponse{}, nil
}

func TestSet(t *testing.T) {
	payload := `---\nkey: test_string_boe7ahthu0Fie1Eghai4}`

	t.Run("set organization settings", func(t *testing.T) {
		mc := new(mockSetClient)
		ctx := context.Background()
		if err := action.Run(ctx, mc, "organization", []byte(payload)); err != nil {
			t.Fatalf("running action.Run() failed with error: %v", err)
		}
	})

}

// Server tests

// No tests here because the code does to do anything interesting.
