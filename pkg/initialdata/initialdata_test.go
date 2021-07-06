package initialdata_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Run("executing initialdata.Cmd() without flags so using default initial data", func(t *testing.T) {
		cmd := initialdata.Cmd()
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing initial-data subcommand: %v", err)
		}
	})
}

type MockInitialdataClient struct {
	expected []byte
}

func (m *MockInitialdataClient) InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error) {
	if bytes.Compare(m.expected, in.Data) != 0 {
		return nil, fmt.Errorf("wrong initial data, expected %q, got %q", m.expected, in.Data)
	}
	return &proto.InitialDataResponse{Initialized: true}, nil
}

func TestInitialdata(t *testing.T) {
	t.Run("default initial data", func(t *testing.T) {
		mc := new(MockInitialdataClient)
		mc.expected = []byte(initialdata.DefaultInitialData)
		ctx := context.Background()
		if err := initialdata.Initialdata(ctx, mc, ""); err != nil {
			t.Fatalf("running Initialdata() failed with error: %v", err)
		}
	})
	t.Run("custom initial data", func(t *testing.T) {
		customIniD := "foo:bar"

		f, err := os.CreateTemp("", "openslides-initial-data.json")
		if err != nil {
			t.Fatalf("creating temporary file for initial data: %v", err)
		}
		defer os.Remove(f.Name())
		f.WriteString(customIniD)
		if err := f.Close(); err != nil {
			t.Fatalf("closing temporary file for initial data: %v", err)
		}

		mc := new(MockInitialdataClient)
		mc.expected = []byte(customIniD)
		ctx := context.Background()
		if err := initialdata.Initialdata(ctx, mc, f.Name()); err != nil {
			t.Fatalf("running Initialdata() failed with error: %v", err)
		}
	})
}
