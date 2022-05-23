package migrations_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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
		timeout := 1 * time.Second
		if err := migrations.Run(ctx, mc, "stats", nil, &timeout); err != nil {
			t.Fatalf("running migrations.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
	})
}

func TestMigrationsResponse(t *testing.T) {
	output := "First line\nSecond line\nThird line\n"
	mR := migrations.MigrationResponse{
		Output:  output,
		Success: true,
		Status:  "some status",
		Stats:   json.RawMessage([]byte(`{"some_key": "some value"}`)),
	}

	t.Run("method Running()", func(t *testing.T) {
		if mR.Running() {
			t.Fatalf("method Running() should return false, but it returns true")
		}
	})

	t.Run("method Yaml()", func(t *testing.T) {
		expected := `exception: ""
output: |
  First line
  Second line
  Third line
stats:
  some_key: some value
status: some status
success: true
`
		got, err := mR.Yaml()
		if err != nil {
			t.Fatalf("method Yaml() returned error: %v", err)
		}
		if got != expected {
			t.Fatalf("method Yaml(): expected %s, got %s", expected, got)
		}
	})

	t.Run("method GetOutput()", func(t *testing.T) {
		expected := `First line
Second line
Third line
`
		got, err := mR.GetOutput()
		if err != nil {
			t.Fatalf("method GetOutput() returned error: %v", err)
		}
		if got != expected {
			t.Fatalf("method GetOutput(): expected %s, got %s", expected, got)
		}
	})

	t.Run("method GetStats()", func(t *testing.T) {
		expected := `some_key: some value`
		got, err := mR.GetStats()
		if err != nil {
			t.Fatalf("method GetStats() returned error: %v", err)
		}
		if got != expected {
			t.Fatalf("method GetStats(): expected %s, got %s", expected, got)
		}
	})

	t.Run("method OutputSince()", func(t *testing.T) {
		got, next := mR.OutputSince(0)
		if got != output {
			t.Fatalf("method OutputSince(): expected output %s, returned %s", output, got)
		}
		if next != 3 {
			t.Fatalf("method OutputSince(): expected 3 lines, returned %d", next)
		}
	})
}
