package checkserver_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/checkserver"
	"github.com/OpenSlides/openslides-manage-service/proto"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing checkserver.Cmd()", func(t *testing.T) {})
}

// Client tests

func TestCheckServerClient(t *testing.T) {
	t.Skip("this test is still missing")
}

// Server tests

type mockAction struct{}

func (m *mockAction) Health(ctx context.Context) (json.RawMessage, error) {
	return nil, nil // There is no response here.
}

func TestCheckServer(t *testing.T) {
	ma := new(mockAction)
	in := &proto.CheckServerRequest{}
	if resp := checkserver.CheckServer(context.Background(), in, ma); !resp.Ready {
		t.Fatalf("running CheckServer() should not return a falsy ready flag")
	}
}
