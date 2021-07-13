package setpassword_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/proto"
)

type mockAuth struct{}

func (m *mockAuth) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

type mockDatastore struct {
	content map[string]json.RawMessage
}

func (m *mockDatastore) Set(fqfield string, value json.RawMessage) error {
	if m.content == nil {
		m.content = make(map[string]json.RawMessage)
	}
	m.content[fqfield] = value
	return nil
}

func TestSetPasswordServerAll(t *testing.T) {
	md := new(mockDatastore)
	ma := new(mockAuth)
	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: "my_password",
	}
	if _, err := setpassword.SetPassword(in, md, ma); err != nil {
		t.Fatalf("running SetPassword() failed: %v", err)
	}
	key := "user/1/password"
	expected := fmt.Sprintf("%q", "hash:my_password")
	got := string(md.content[key])
	if expected != got {
		t.Fatalf("wrong (mock) datastore key %s, expected %q, got %q ", key, expected, got)
	}
}
