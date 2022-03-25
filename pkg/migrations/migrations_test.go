package migrations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/migrations"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing migrations.Cmd() ...", func(t *testing.T) {
		// cmd := migrations.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing migrations subcommand: %v", err)
		// }
	})
}

// Client tests

type mockMigrationsClient struct {
	expected string
	called   bool
	response []byte
}

func (m *mockMigrationsClient) Migrations(ctx context.Context, in *proto.MigrationsRequest, opts ...grpc.CallOption) (*proto.MigrationsResponse, error) {
	m.called = true
	if m.expected != in.Command {
		return nil, fmt.Errorf("wrong command, expected %q, got %q", m.expected, in.Command)
	}
	return &proto.MigrationsResponse{Response: m.response}, nil
}

func TestMigrations(t *testing.T) {
	t.Run("one command", func(t *testing.T) {
		mc := new(mockMigrationsClient)
		mc.expected = "stats"
		mc.response = []byte(`{"success": true, "stats": {}}`)
		ctx := context.Background()
		if err := migrations.Run(ctx, mc, "stats", 0); err != nil {
			t.Fatalf("running migrations.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
	})
}
