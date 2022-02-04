package createuser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCmd(t *testing.T) {
	t.Skip("this test does not work because there is no (fake) server running")
	t.Run("executing createuser.Cmd() with some data", func(t *testing.T) {
		// cmd := createuser.Cmd()
		// if err := cmd.Execute(); err != nil {
		// 	t.Fatalf("executing create-user subcommand: %v", err)
		// }
	})
}

// Client tests

type mockCreateUserClient struct {
	expectedUsername string
	called           bool
}

func (m *mockCreateUserClient) CreateUser(ctx context.Context, in *proto.CreateUserRequest, opts ...grpc.CallOption) (*proto.CreateUserResponse, error) {
	m.called = true
	if m.expectedUsername != in.Username {
		return nil, fmt.Errorf("wrong user data, expected username %q, got %q", m.expectedUsername, in.Username)
	}
	return &proto.CreateUserResponse{}, nil
}

func TestCreateUser(t *testing.T) {
	t.Run("create a new user", func(t *testing.T) {
		username := "my_new_username_Xohqu1rai2"
		user := fmt.Sprintf(`---
username: %s
first_name: my_first_name_Tahz3raegh
last_name: my_last_name_Einohjee1E
is_active: true
default_password: my_new_password_weu6Aichaj
email: my_new_emailaddress_ooduN6fuh7@ohth1Osa1I.com
organization_management_level: superadmin
committee_$_management_level: {"can_manage": ["1"]}
group_$_ids: {"1": [1,2]}
`, username)
		testUserFile(t, username, user, "")
	})

	t.Run("create a new user with invalid YAML user file 1", func(t *testing.T) {
		username := "my_new_username_Quitee9Aek"
		user := fmt.Sprintf(`---
first_name: my_first_name_Mei6zai2Ie
`)
		hasErrMsg := "missing username"
		testUserFile(t, username, user, hasErrMsg)
	})

	t.Run("create a new user with invalid YAML user file 2", func(t *testing.T) {
		username := "my_new_username_oe4As0iege"
		user := fmt.Sprintf(`---
username: %s
`, username)
		hasErrMsg := "missing default_password"
		testUserFile(t, username, user, hasErrMsg)
	})

	t.Run("create a new user with valid YAML user file 3", func(t *testing.T) {
		username := "my_new_username_AekQuitee9"
		user := fmt.Sprintf(`---
username: %s
default_password: my_password_ab6ee8guYa
`, username)
		testUserFile(t, username, user, "")
	})

	t.Run("create a new user with invalid YAML user file 4", func(t *testing.T) {
		username := "my_username_Ish2aeb1Sh"
		user := fmt.Sprintf(`---
username: %s
default_password: my_password_Sei8quozac
organization_management_level: some_custom_invalid_string_fuu2Thaequ
`, username)
		hasErrMsg := "wrong value for organization_management_level"
		testUserFile(t, username, user, hasErrMsg)
	})
}

func testUserFile(t testing.TB, username, user, hasErrMsg string) {
	t.Helper()
	f, err := os.CreateTemp("", "user.yml")
	if err != nil {
		t.Fatalf("creating temporary file for user creation: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(user)
	if err := f.Close(); err != nil {
		t.Fatalf("closing temporary file for user creation: %v", err)
	}

	mc := new(mockCreateUserClient)
	mc.expectedUsername = username
	ctx := context.Background()
	err = createuser.Run(ctx, mc, f.Name())

	if hasErrMsg == "" {
		if err != nil {
			t.Fatalf("running createuser.Run() failed with error: %v", err)
		}
		if !mc.called {
			t.Fatalf("gRPC client was not called")
		}
		return
	}
	if err == nil {
		t.Fatalf("running createuser.Run() should fail with error but it didn't")
	}
	if mc.called {
		t.Fatalf("gRPC client was called")
	}
	if !strings.Contains(err.Error(), hasErrMsg) {
		t.Fatalf("running createuser.Run() with invalid YAML user file, got error message %q, expected %q", err.Error(), hasErrMsg)
	}
}

// Server tests

type mockAction struct {
	expUserID int64
}

func (m *mockAction) Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error) {
	switch name {
	case "user.create":
		return m.userCreate()
	default:
		return nil, fmt.Errorf("action %q is not defined here", name)
	}
}

func (m *mockAction) userCreate() (json.RawMessage, error) {
	r := []struct {
		UserID int64 `json:"id"`
	}{
		{UserID: m.expUserID},
	}
	encR, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshalling JSON: %w", err)
	}
	return encR, nil
}

func TestCreateUserServerAll(t *testing.T) {
	expUserID := int64(7483798057)
	ma := new(mockAction)
	ma.expUserID = expUserID
	t.Run("create a user", func(t *testing.T) {
		var ls []interface{}
		ls = append(ls, "1")
		l, _ := structpb.NewList(ls)
		in := &proto.CreateUserRequest{
			Username:                    "my_username_moh9Sep8Ae",
			DefaultPassword:             "my_password_nah4Aigahp",
			OrganizationManagementLevel: "superadmin",
			IsActive:                    true,
			Committee_ManagementLevel: map[string]*structpb.ListValue{
				"can_manage": l,
			},
		}
		res, err := createuser.CreateUser(context.Background(), in, ma)
		if err != nil {
			t.Fatalf("running CreateUser() failed: %v", err)
		}
		if res.UserID != expUserID {
			t.Fatalf("wrong user id, expected %d, got %d", expUserID, res.UserID)
		}
	})
}
