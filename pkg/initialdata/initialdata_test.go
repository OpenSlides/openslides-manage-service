package initialdata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing initialdata.Cmd() without flags so using default initial data", func(t *testing.T) {
		// cmd := initialdata.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing initial-data subcommand: %v", err)
		// }
	})
}

// Client tests

type mockInitialdataClient struct {
	expected []byte
	called   bool
}

func (m *mockInitialdataClient) InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error) {
	m.called = true
	if !bytes.Equal(m.expected, in.Data) {
		return nil, fmt.Errorf("wrong initial data, expected %q, got %q", m.expected, in.Data)
	}
	return &proto.InitialDataResponse{Initialized: true}, nil
}

func TestInitialdata(t *testing.T) {
	t.Run("default initial data", func(t *testing.T) {
		mc := new(mockInitialdataClient)
		mc.expected = []byte("")
		ctx := context.Background()
		if err := initialdata.Run(ctx, mc, nil); err != nil {
			t.Fatalf("running initialdata.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
	})
	t.Run("custom initial data", func(t *testing.T) {
		customIniD := `{"key":"test_string_phiC0ChaibieSoo9aezaigaiyof9ieVu"}`
		mc := new(mockInitialdataClient)
		mc.expected = []byte(customIniD)
		ctx := context.Background()
		if err := initialdata.Run(ctx, mc, []byte(customIniD)); err != nil {
			t.Fatalf("running initialdata.Run() failed with error: %v", err)
		}
	})
}

// Server tests

type mockAction struct {
	called map[string][]json.RawMessage
}

func newMockAction() *mockAction {
	ma := new(mockAction)
	ma.called = make(map[string][]json.RawMessage)
	return ma
}

func (m *mockAction) Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error) {
	switch name {
	case "organization.initial_import", "user.set_password":
		m.called[name] = append(m.called[name], data)
	default:
		return nil, fmt.Errorf("action %q is not defined here", name)
	}
	return nil, nil
}

func TestInitialDataServerAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ma := newMockAction()
	in := &proto.InitialDataRequest{}

	// Prepare superadmin secret
	testDir, err := os.MkdirTemp("", "openslides-manage-service-run-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)
	secDir := path.Join(testDir, setup.SecretsDirName)
	if err := os.Mkdir(secDir, os.ModePerm); err != nil {
		t.Fatalf("generating temporary subdirectory failed: %v", err)
	}
	superadminPassword := "my_superadmin_password_aijooP4EeC"
	f, err := os.Create(path.Join(secDir, setup.SuperadminFileName))
	if err != nil {
		t.Fatalf("creating temporary file for superadmin password: %v", err)
	}
	f.WriteString(superadminPassword)
	if err := f.Close(); err != nil {
		t.Fatalf("closing temporary file for superadmin password: %v", err)
	}

	// Run tests
	t.Run("running the first time", func(t *testing.T) {
		p := path.Join(testDir, setup.SecretsDirName, setup.SuperadminFileName)
		resp, err := initialdata.InitialData(ctx, in, p, ma)
		if err != nil {
			t.Fatalf("running InitialData() failed: %v", err)
		}
		if !resp.Initialized {
			t.Fatalf("running InitialData() should return a truthy result, got falsy")
		}

		expected := json.RawMessage(fmt.Sprintf(`[{"id":1,"password":"%s"}]`, superadminPassword))
		got := ma.called["user.set_password"][0]

		if !bytes.Equal(expected, got) {
			t.Fatalf("wrong superadmin password, expected %q, got %q", expected, got)
		}
	})
}
