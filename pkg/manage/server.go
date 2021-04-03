package manage

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

// RunServer starts the manage server.
func RunServer(cfg *ServerConfig) error {
	addr := ":" + cfg.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on addr %s: %w", addr, err)
	}

	s := grpc.NewServer()
	proto.RegisterManageServer(s, newServer(cfg))

	fmt.Printf("Running manage service on %s\n", addr)

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("running service: %w", err)
	}
	return nil
}

// Server implements the manage methods on server side.
type Server struct {
	config *ServerConfig
}

func newServer(cfg *ServerConfig) *Server {
	return &Server{
		config: cfg,
	}
}

// ServerConfig holds config data for the server.
type ServerConfig struct {
	// The struct tag `env` is used to populate the values from environment
	// variables. The first value is the name of the environment variable. After
	// a comma the default value can be given. If no default value is given, then
	// an empty string is used. The type of a env field has to be string.
	Port string `env:"MANAGE_PORT,9008"`

	AuthProtocol string `env:"AUTH_PROTOCOL,http"`
	AuthHost     string `env:"AUTH_HOST,auth"`
	AuthPort     string `env:"AUTH_PORT,9004"`

	DatastoreWriterProtocol string `env:"DATASTORE_WRITER_PROTOCOL,http"`
	DatastoreWriterHost     string `env:"DATASTORE_WRITER_HOST,datastore-writer"`
	DatastoreWriterPort     string `env:"DATASTORE_WRITER_PORT,9011"`

	DatastoreReaderProtocol string `env:"DATASTORE_READER_PROTOCOL,http"`
	DatastoreReaderHost     string `env:"DATASTORE_READER_HOST,datastore-reader"`
	DatastoreReaderPort     string `env:"DATASTORE_READER_PORT,9010"`
}

// ServerConfigFromEnv creates a Config object where the values are populated from the
// environment.
//
// Example:
// cfg := ServerConfigFromEnv(os.LookupEnv)
func ServerConfigFromEnv(loockup func(string) (string, bool)) *ServerConfig {
	c := ServerConfig{}
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

// AuthURL returns an URL object to the auth service with empty path.
func (c *ServerConfig) AuthURL() url.URL {
	u := url.URL{
		Scheme: c.AuthProtocol,
		Host:   c.AuthHost + ":" + c.AuthPort,
	}
	return u
}

// DatastoreWriterURL returns an URL object to the datastore writer service with empty path.
func (c *ServerConfig) DatastoreWriterURL() url.URL {
	u := url.URL{
		Scheme: c.DatastoreWriterProtocol,
		Host:   c.DatastoreWriterHost + ":" + c.DatastoreWriterPort,
	}
	return u
}
