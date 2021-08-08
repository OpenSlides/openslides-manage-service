package initialdata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing initialdata.Cmd() without flags so using default initial data", func(t *testing.T) {
		cmd := initialdata.Cmd()
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing initial-data subcommand: %v", err)
		}
	})
}

// Client tests

type mockInitialdataClient struct {
	expected []byte
	called   bool
}

func (m *mockInitialdataClient) InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error) {
	m.called = true
	if bytes.Compare(m.expected, in.Data) != 0 {
		return nil, fmt.Errorf("wrong initial data, expected %q, got %q", m.expected, in.Data)
	}
	return &proto.InitialDataResponse{Initialized: true}, nil
}

func TestInitialdata(t *testing.T) {
	t.Run("default initial data", func(t *testing.T) {
		mc := new(mockInitialdataClient)
		mc.expected = []byte(initialdata.DefaultInitialData)
		ctx := context.Background()
		if err := initialdata.Run(ctx, mc, ""); err != nil {
			t.Fatalf("running initialdata.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
	})
	t.Run("custom initial data", func(t *testing.T) {
		customIniD := `{"key":"test_string_phiC0ChaibieSoo9aezaigaiyof9ieVu"}`
		f, err := os.CreateTemp("", "openslides-initial-data.json")
		if err != nil {
			t.Fatalf("creating temporary file for initial data: %v", err)
		}
		defer os.Remove(f.Name())
		f.WriteString(customIniD)
		if err := f.Close(); err != nil {
			t.Fatalf("closing temporary file for initial data: %v", err)
		}

		mc := new(mockInitialdataClient)
		mc.expected = []byte(customIniD)
		ctx := context.Background()
		if err := initialdata.Run(ctx, mc, f.Name()); err != nil {
			t.Fatalf("running initialdata.Run() failed with error: %v", err)
		}
	})
}

// Server tests

type mockDatastore struct {
	content map[string]json.RawMessage
}

func (m *mockDatastore) Exists(ctx context.Context, collection string, id int) (bool, error) {
	k := fmt.Sprintf("%s/%d/id", collection, id)
	_, ok := m.content[k]
	return ok, nil
}

func (m *mockDatastore) Create(ctx context.Context, fqid string, fields map[string]json.RawMessage) error {
	ss := strings.Split(fqid, "/")
	collection := ss[0]
	id, _ := strconv.Atoi(ss[1])

	if exists, _ := m.Exists(ctx, collection, id); exists {
		return fmt.Errorf("object %q already exists", fqid)
	}
	if m.content == nil {
		m.content = make(map[string]json.RawMessage)
	}
	for field, value := range fields {
		m.content[fqid+"/"+field] = value
	}
	return nil
}

func (m *mockDatastore) Set(ctx context.Context, fqfield string, value json.RawMessage) error {
	if m.content == nil {
		m.content = make(map[string]json.RawMessage)
	}
	m.content[fqfield] = value
	return nil
}

type mockAuth struct{}

func (m *mockAuth) Hash(ctx context.Context, password string) (string, error) {
	return "hash:" + password, nil
}

func TestInitialDataServerAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	md := new(mockDatastore)
	ma := new(mockAuth)
	in := &proto.InitialDataRequest{
		Data: initialdata.DefaultInitialData,
	}

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
		resp, err := initialdata.InitialData(ctx, in, testDir, md, ma)
		if err != nil {
			t.Fatalf("running InitialData() failed: %v", err)
		}
		if !resp.Initialized {
			t.Fatalf("running InitialData() should return a truthy result, got falsy")
		}

		expected := fmt.Sprintf("%q", "hash:"+superadminPassword)
		got := string(md.content["user/1/password"])
		if expected != got {
			t.Fatalf("wrong superadmin password, expected %q, got %q", expected, got)
		}
	})
	t.Run("running the second time", func(t *testing.T) {
		resp, err := initialdata.InitialData(ctx, in, testDir, md, ma)
		if err != nil {
			t.Fatalf("running InitialData() failed: %v", err)
		}
		if resp.Initialized {
			t.Fatalf("running InitialData() should return a falsy result, got truthy")
		}

		expected := fmt.Sprintf("%q", "hash:"+superadminPassword)
		got := string(md.content["user/1/password"])
		if expected != got {
			t.Fatalf("wrong superadmin password, expected %q, got %q", expected, got)
		}
	})
}

func TestInitialDataServer(t *testing.T) {
	md := new(mockDatastore)

	t.Run("checking datastore existance", func(t *testing.T) {
		exists, err := initialdata.CheckDatastore(context.Background(), md)
		if err != nil {
			t.Fatalf("checking if data in datastore exist failed: %v", err)
		}
		if exists {
			t.Fatal("(fake) database should be empty, but is not")
		}
	})

	t.Run("adding initial data", func(t *testing.T) {
		if err := initialdata.InsertIntoDatastore(context.Background(), md, initialdata.DefaultInitialData); err != nil {
			t.Fatalf("inserting initial data into datastore failed: %v", err)
		}
	})

	t.Run("checking datastore again", func(t *testing.T) {
		exists, err := initialdata.CheckDatastore(context.Background(), md)
		if err != nil {
			t.Fatalf("checking if data in datastore exist failed: %v", err)
		}
		if !exists {
			t.Fatal("(fake) database should not be empty, but is")
		}
	})

	t.Run("setting superadmin password", func(t *testing.T) {
		ma := new(mockAuth)

		superadminPassword := "my_superadmin_password_Do7aeRaing"
		f, err := os.CreateTemp("", "openslides-superadmin-secret")
		if err != nil {
			t.Fatalf("creating temporary file for superadmin password: %v", err)
		}
		defer os.Remove(f.Name())
		f.WriteString(superadminPassword)
		if err := f.Close(); err != nil {
			t.Fatalf("closing temporary file for superadmin password: %v", err)
		}

		if err := initialdata.SetSuperadminPassword(context.Background(), f.Name(), md, ma); err != nil {
			t.Fatalf("setting superadmin password failed: %v", err)
		}
		key := "user/1/password"
		expected := fmt.Sprintf("%q", "hash:"+superadminPassword)
		got := string(md.content[key])
		if expected != got {
			t.Fatalf("wrong superadmin password, expected %q, got %q", expected, got)
		}
	})
}
