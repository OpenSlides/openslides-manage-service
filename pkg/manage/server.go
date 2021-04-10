package manage

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"time"

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

	srv := grpc.NewServer()
	proto.RegisterManageServer(srv, newServer(cfg))

	go func() {
		waitForShutdown()
		srv.GracefulStop()
	}()

	fmt.Printf("Running manage service on %s\n", addr)
	if err := srv.Serve(lis); err != nil {
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

	DatastoreReaderProtocol string `env:"DATASTORE_READER_PROTOCOL,http"`
	DatastoreReaderHost     string `env:"DATASTORE_READER_HOST,datastore-reader"`
	DatastoreReaderPort     string `env:"DATASTORE_READER_PORT,9010"`

	DatastoreWriterProtocol string `env:"DATASTORE_WRITER_PROTOCOL,http"`
	DatastoreWriterHost     string `env:"DATASTORE_WRITER_HOST,datastore-writer"`
	DatastoreWriterPort     string `env:"DATASTORE_WRITER_PORT,9011"`
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
func (c *ServerConfig) AuthURL() *url.URL {
	u := url.URL{
		Scheme: c.AuthProtocol,
		Host:   c.AuthHost + ":" + c.AuthPort,
	}
	return &u
}

// DatastoreReaderURL returns an URL object to the datastore reader service with empty path.
func (c *ServerConfig) DatastoreReaderURL() *url.URL {
	u := url.URL{
		Scheme: c.DatastoreReaderProtocol,
		Host:   c.DatastoreReaderHost + ":" + c.DatastoreReaderPort,
		Path:   "/internal/datastore/reader",
	}
	return &u
}

// DatastoreWriterURL returns an URL object to the datastore writer service with empty path.
func (c *ServerConfig) DatastoreWriterURL() *url.URL {
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

// waitForService checks that all services at the given addresses are available.
//
// Blocks until a connection to any of the services can be established or the
// context is canceled.
func waitForService(ctx context.Context, addrs ...string) {
	d := net.Dialer{}

	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			con, err := d.DialContext(ctx, "tcp", addr)
			for err != nil {
				if ctx.Err() != nil {
					// Time is up, dont try again.
					return
				}

				time.Sleep(100 * time.Millisecond)
				con, err = d.DialContext(ctx, "tcp", addr)
			}
			con.Close()

		}(addr)
	}
	wg.Wait()
}
