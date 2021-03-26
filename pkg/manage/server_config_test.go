package manage_test

import (
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

func TestConfigFromEnv(t *testing.T) {
	lookup := func(key string) (string, bool) {
		defaults := map[string]string{
			"AUTH_HOST": "test-auth",
		}
		v, ok := defaults[key]
		return v, ok
	}
	cfg := manage.ServerConfigFromEnv(lookup)
	if cfg.AuthHost != "test-auth" {
		t.Errorf("config.AuthHost == `%s`, expected `test-auth`", cfg.AuthHost)
	}
	if cfg.Port != "9008" {
		t.Errorf("config.Port == `%s`, expected `9008`", cfg.Port)
	}
}

type LookupEnverStub struct {
	data map[string]string
}

func (l LookupEnverStub) LookupEnv(key string) (string, bool) {
	v, ok := l.data[key]
	return v, ok
}
