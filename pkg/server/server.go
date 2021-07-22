package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/auth"
	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/tunnel"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
)

const runDir = "/run"

// Run starts the manage server.
func Run(cfg *Config) error {
	addr := ":" + cfg.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on address %q: %w", addr, err)
	}

	srv := grpc.NewServer()
	proto.RegisterManageServer(srv, newServer(cfg))

	go func() {
		waitForShutdown()
		srv.GracefulStop()
	}()

	log.Printf("Running manage service on %s\n", addr)
	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("running manage service: %w", err)
	}

	return nil
}

// srv implements the manage methods on server side.
type srv struct {
	config *Config
}

func newServer(cfg *Config) *srv {
	return &srv{
		config: cfg,
	}
}

func (s *srv) CheckServer(context.Context, *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {
	return nil, fmt.Errorf("currently not implemented")
}

func (s *srv) InitialData(ctx context.Context, in *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	ds := datastore.New(s.config.datastoreReaderURL(), s.config.datastoreWriterURL())
	auth := auth.New(s.config.authURL())
	return initialdata.InitialData(ctx, in, runDir, ds, auth)

}

func (s *srv) CreateUser(context.Context, *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	return nil, fmt.Errorf("currently not implemented")
}

func (s *srv) SetPassword(ctx context.Context, in *proto.SetPasswordRequest) (*proto.SetPasswordResponse, error) {
	ds := datastore.New(s.config.datastoreReaderURL(), s.config.datastoreWriterURL())
	auth := auth.New(s.config.authURL())
	return setpassword.SetPassword(ctx, in, ds, auth)
}

func (s *srv) Tunnel(ts proto.Manage_TunnelServer) error {
	return tunnel.Tunnel(ts)
}

// Config holds config data for the server.
type Config struct {
	// The struct tag `env` is used to populate the values from environment
	// variables. The first value is the name of the environment variable. After
	// a comma the default value can be given. If no default value is given, then
	// an empty string is used. The type of a env field has to be string.
	Port string `env:"MANAGE_PORT,9008"`

	AuthProtocol string `env:"AUTH_PROTOCOL,http"`
	AuthHost     string `env:"AUTH_HOST,auth"`
	AuthPort     string `env:"AUTH_PORT,9004"`

	DatastoreReaderProtocol string `env:"DATASTORE_READER_PROTOCOL,http"`
	DatastoreReaderHost     string `env:"DATASTORE_READER_HOST,datastore-reader"`
	DatastoreReaderPort     string `env:"DATASTORE_READER_PORT,9010"`

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

// authURL returns an URL object to the auth service with empty path.
func (c *Config) authURL() *url.URL {
	u := url.URL{
		Scheme: c.AuthProtocol,
		Host:   c.AuthHost + ":" + c.AuthPort,
	}
	return &u
}

// datastoreReaderURL returns an URL object to the datastore reader service with empty path.
func (c *Config) datastoreReaderURL() *url.URL {
	u := url.URL{
		Scheme: c.DatastoreReaderProtocol,
		Host:   c.DatastoreReaderHost + ":" + c.DatastoreReaderPort,
		Path:   "/internal/datastore/reader",
	}
	return &u
}

// datastoreWriterURL returns an URL object to the datastore writer service with empty path.
func (c *Config) datastoreWriterURL() *url.URL {
	u := url.URL{
		Scheme: c.DatastoreWriterProtocol,
		Host:   c.DatastoreWriterHost + ":" + c.DatastoreWriterPort,
		Path:   "/internal/datastore/writer",
	}
	return &u
}

// waitForShutdown blocks until the service exists.
//
// It listens on SIGINT and SIGTERM. If the signal is received for a second
// time, the process is killed with statuscode 1.
func waitForShutdown() {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt)
	<-sigint
	go func() {
		<-sigint
		os.Exit(1)
	}()
}
