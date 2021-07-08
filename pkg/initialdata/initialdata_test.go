package initialdata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
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
		if err := initialdata.Run(ctx, mc, ""); err != nil {
			t.Fatalf("running Initialdata() failed with error: %v", err)
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

		mc := new(MockInitialdataClient)
		mc.expected = []byte(customIniD)
		ctx := context.Background()
		if err := initialdata.Run(ctx, mc, f.Name()); err != nil {
			t.Fatalf("running Initialdata() failed with error: %v", err)
		}
	})
}

type MockDatastore struct {
	content map[string]json.RawMessage
}

func (m *MockDatastore) Exists(collection string, id int) (bool, error) {
	k := fmt.Sprintf("%s/%d/id", collection, id)
	_, ok := m.content[k]
	return ok, nil
}

func (m *MockDatastore) Create(fqid string, fields map[string]json.RawMessage) error {

	ss := strings.Split(fqid, "/")
	collection := ss[0]
	id, _ := strconv.Atoi(ss[1])

	if exists, _ := m.Exists(collection, id); exists {
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

func TestInitialdataServer(t *testing.T) {
	md := new(MockDatastore)

	t.Run("checking datastore", func(t *testing.T) {
		exists, err := initialdata.CheckDatastore(md)
		if err != nil {
			t.Fatalf("checking if data in datastore exist failed: %v", err)
		}
		if exists {
			t.Fatal("(fake) database should be empty, but is not")
		}
	})

	t.Run("adding initial data", func(t *testing.T) {
		if err := initialdata.InsertIntoDatastore(md, initialdata.DefaultInitialData); err != nil {
			t.Fatalf("inserting initial data into datastore failed: %v", err)
		}
	})

	t.Run("checking datastore again", func(t *testing.T) {
		exists, err := initialdata.CheckDatastore(md)
		if err != nil {
			t.Fatalf("checking if data in datastore exist failed: %v", err)
		}
		if !exists {
			t.Fatal("(fake) database should not be empty, but is")
		}
	})

	t.Run("setting admin password", func(t *testing.T) {
		t.Fail() // TODO: Implement this but implement common SetPassword first.
	})
}
