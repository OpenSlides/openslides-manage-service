package initialdata_test

import (
	"context"
	"fmt"
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
	validator func(in *proto.InitialDataRequest) error
}

func (m *MockInitialdataClient) InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error) {
	return nil, m.validator(in)
}

func TestInitialdata(t *testing.T) {
	iniD := "foo: bar"

	mc := new(MockInitialdataClient)
	mc.validator = func(in *proto.InitialDataRequest) error {
		got := string(in.Data)
		if iniD != got {
			x := string(in.Data)
			return fmt.Errorf("wrong initial data, expected %q, got %q", iniD, x)
		}
		return nil
	}
	ctx := context.Background()
	if err := initialdata.Initialdata(ctx, mc); err != nil {
		t.Fatalf("running Initialdata() failed with error: %v", err)
	}
}
