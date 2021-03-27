package manage_test

import (
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

func TestConfigFromEnv(t *testing.T) {
	lookup := func(key string) (string, bool) {
		defaults := map[string]string{
			"MANAGE_HOST": "test-host",
		}
		v, ok := defaults[key]
		return v, ok
	}
	cfg := manage.ServerConfigFromEnv(lookup)
	if cfg.Host != "test-host" {
		t.Errorf("config.Host == `%s`, expected `test-host`", cfg.Host)
	}
	if cfg.Port != "8001" {
		t.Errorf("config.Port == `%s`, expected `8001`", cfg.Host)
	}
}

type LookupEnverStub struct {
	data map[string]string
}

func (l LookupEnverStub) LookupEnv(key string) (string, bool) {
	v, ok := l.data[key]
	return v, ok
}
