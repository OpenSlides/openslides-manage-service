package server

import (
	"net/url"
	"reflect"
	"strings"
)

// Config holds config data for the server.
type Config struct {
	// The struct tag `env` is used to populate the values from environment
	// variables. The first value is the name of the environment variable. After
	// a comma the default value can be given. If no default value is given, then
	// an empty string is used. The type of a env field has to be string.
	Host string `env:"MANAGE_HOST"`
	Port string `env:"MANAGE_PORT,8001"`

	AuthProtocol string `env:"AUTH_PROTOCOL,http"`
	AuthHost     string `env:"AUTH_HOST,auth"`
	AuthPort     string `env:"AUTH_PORT,9004"`

	DatastoreWriterProtocol string `env:"DATASTORE_WRITER_PROTOCOL,http"`
	DatastoreWriterHost     string `env:"DATASTORE_WRITER_HOST,datastore-writer"`
	DatastoreWriterPort     string `env:"DATASTORE_WRITER_PORT,9011"`
}

// ConfigFromEnv creates a Config object where the values are populated from the
// environment.
//
// Example:
// cfg := ConfigFromEnv(os.LookupEnv)
func ConfigFromEnv(loockup func(string) (string, bool)) *Config {
	c := Config{}
	v := reflect.ValueOf(&c).Elem()
	t := reflect.TypeOf(c)
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("env")
		if tag == "" {
			// No struct tag
			continue
		}

		parts := strings.SplitN(tag, ",", 2)

		envValue, ok := loockup(parts[0])
		if !ok && len(parts) == 2 {
			envValue = parts[1]
		}

		v.Field(i).SetString(envValue)
	}
	return &c
}

// Addr return the address of the manage service.
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

// AuthURL returns an URL object to the auth service with empty path.
func (c *Config) AuthURL() url.URL {
	u := url.URL{
		Scheme: c.AuthProtocol,
		Host:   c.AuthHost + ":" + c.AuthPort,
	}
	return u
}

// DatastoreWriterURL returns an URL object to the datastore writer service with empty path.
func (c *Config) DatastoreWriterURL() url.URL {
	u := url.URL{
		Scheme: c.DatastoreWriterProtocol,
		Host:   c.DatastoreWriterHost + ":" + c.DatastoreWriterPort,
	}
	return u
}
