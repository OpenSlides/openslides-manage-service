package server

import (
	"fmt"
	"reflect"
	"strings"
)

// Config holds config data fot the server.
type Config struct {
	// The structag `env` is used to polulate the values from environment
	// varialbes. The first value is the name of the environment variable. After
	// a `,` the default value can be given. If no default value is given, then
	// "" is used. The type of a env-field has to be string.
	Host string `env:"MANAGE_HOST"`
	Port string `env:"MANAGE_PORT,8001"`

	AuthProtocol string `env:"AUTH_PROTOCOL,http"`
	AuthHost     string `env:"AUTH_HOST,auth"`
	AuthPort     string `env:"AUTH_PORT,9004"`

	DSWriterProtocol string `env:"DATASTORE_WRITER_PROTOCOL,http"`
	DSWriterHost     string `env:"DATASTORE_WRITER_HOST,datastore-writer"`
	DSWriterPort     string `env:"DATASTORE_WRITER_PORT,9011"`
}

// ConfigFromEnv creates a Config-object where the values are polulated from the
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

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

func (c *Config) AuthAddr() string {
	return fmt.Sprintf("%s://%s:%s", c.AuthProtocol, c.AuthHost, c.AuthPort)
}

func (c *Config) DSWriterAddr() string {
	return fmt.Sprintf("%s://%s:%s", c.DSWriterProtocol, c.DSWriterHost, c.DSWriterPort)
}
