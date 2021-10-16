package server

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/action"
	"github.com/OpenSlides/openslides-manage-service/pkg/auth"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/pkg/tunnel"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const runDir = "/run"

// Run starts the manage server.
func Run(cfg *Config) error {
	addr := ":" + cfg.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on address %q: %w", addr, err)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc.UnaryServerInterceptor(authUnaryInterceptor),
		),
		grpc.ChainStreamInterceptor(
			grpc.StreamServerInterceptor(authStreamInterceptor),
		),
	)
	manageSrv, err := newServer(cfg)
	if err != nil {
		return fmt.Errorf("creating server object: %w", err)
	}
	proto.RegisterManageServer(grpcSrv, manageSrv)

	go func() {
		waitForShutdown()
		grpcSrv.GracefulStop()
	}()

	log.Printf("Running manage service on %s\n", addr)
	if err := grpcSrv.Serve(lis); err != nil {
		return fmt.Errorf("running manage service: %w", err)
	}

	return nil
}

// srv implements the manage methods on server side.
type srv struct {
	config     *Config
	authSecret []byte
}

func newServer(cfg *Config) (*srv, error) {
	sec, err := shared.ServerAuthSecret(cfg.PasswordFile, cfg.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting server auth secret: %w", err)
	}
	s := &srv{
		config:     cfg,
		authSecret: sec,
	}
	return s, nil
}

func (s *srv) CheckServer(context.Context, *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {
	return nil, fmt.Errorf("currently not implemented")
}

func (s *srv) InitialData(ctx context.Context, in *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	ds := datastore.New(s.config.datastoreReaderURL(), s.config.datastoreWriterURL())
	auth := auth.New(s.config.authURL())
	return initialdata.InitialData(ctx, in, runDir, ds, auth)

}

func (s *srv) CreateUser(ctx context.Context, in *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	pw, err := s.serverAuthPassword()
	if err != nil {
		return nil, fmt.Errorf("getting manage auth password: %w", err)
	}
	if err := connection.CheckAuthFromContext(ctx, pw); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}
	a := action.New(s.config.actionURL())
	return createuser.CreateUser(ctx, in, a)
}

func (s *srv) SetPassword(ctx context.Context, in *proto.SetPasswordRequest) (*proto.SetPasswordResponse, error) {
	ds := datastore.New(s.config.datastoreReaderURL(), s.config.datastoreWriterURL())
	auth := auth.New(s.config.authURL())
	return setpassword.SetPassword(ctx, in, ds, auth)
}

func (s *srv) Tunnel(ts proto.Manage_TunnelServer) error {
	return tunnel.Tunnel(ts)
}

func authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := info.Server.(*srv).serverAuth(ctx); err != nil {
		return nil, fmt.Errorf("server authentication: %w", err)
	}
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("calling handler: %w", err)
	}
	return resp, nil
}

func authStreamInterceptor(serv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := serv.(*srv).serverAuth(ss.Context()); err != nil {
		return fmt.Errorf("server authentication: %w", err)
	}
	if err := handler(serv, ss); err != nil {
		return fmt.Errorf("calling handler: %w", err)
	}
	return nil
}

func (s *srv) serverAuth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("getting metadata from context: failed")
	}
	a := md.Get("authorization")
	if len(a) == 0 {
		return fmt.Errorf("no authorization header found")
	}
	password, err := base64.StdEncoding.DecodeString(a[0])
	if err != nil {
		return fmt.Errorf("decoding password (base64): %w", err)
	}

	if subtle.ConstantTimeCompare(password, s.authSecret) != 1 {
		return fmt.Errorf("password does not match")
	}

	return nil
}

// Config holds config data for the server.
type Config struct {
	// The struct tag `env` is used to populate the values from environment
	// variables. The first value is the name of the environment variable. After
	// a comma the default value can be given. If no default value is given, then
	// an empty string is used. The type of a env field has to be string.
	Port         string `env:"MANAGE_PORT,9008"`
	PasswordFile string `env:"MANAGE_AUTH_PASSWORD_FILE,/run/secrets/manage_auth_password"`

	ActionProtocol string `env:"ACTION_PROTOCOL,http"`
	ActionHost     string `env:"ACTION_HOST,backend"`
	ActionPort     string `env:"ACTION_PORT,9002"`

	AuthProtocol string `env:"AUTH_PROTOCOL,http"`
	AuthHost     string `env:"AUTH_HOST,auth"`
	AuthPort     string `env:"AUTH_PORT,9004"`

	DatastoreReaderProtocol string `env:"DATASTORE_READER_PROTOCOL,http"`
	DatastoreReaderHost     string `env:"DATASTORE_READER_HOST,datastore-reader"`
	DatastoreReaderPort     string `env:"DATASTORE_READER_PORT,9010"`

	DatastoreWriterProtocol string `env:"DATASTORE_WRITER_PROTOCOL,http"`
	DatastoreWriterHost     string `env:"DATASTORE_WRITER_HOST,datastore-writer"`
	DatastoreWriterPort     string `env:"DATASTORE_WRITER_PORT,9011"`

	OpenSlidesDevelopment string `env:"OPENSLIDES_DEVELOPMENT,0"`
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

// actionURL returns an URL object to the backend action service.
func (c *Config) actionURL() *url.URL {
	u := url.URL{
		Scheme: c.ActionProtocol,
		Host:   c.ActionHost + ":" + c.ActionPort,
		Path:   "/internal/handle_request",
	}
	return &u
}

// authURL returns an URL object to the auth service with empty path.
func (c *Config) authURL() *url.URL {
	u := url.URL{
		Scheme: c.AuthProtocol,
		Host:   c.AuthHost + ":" + c.AuthPort,
	}
	return &u
}

// datastoreReaderURL returns an URL object to the datastore reader service.
func (c *Config) datastoreReaderURL() *url.URL {
	u := url.URL{
		Scheme: c.DatastoreReaderProtocol,
		Host:   c.DatastoreReaderHost + ":" + c.DatastoreReaderPort,
		Path:   "/internal/datastore/reader",
	}
	return &u
}

// datastoreWriterURL returns an URL object to the datastore writer service.
func (c *Config) datastoreWriterURL() *url.URL {
	u := url.URL{
		Scheme: c.DatastoreWriterProtocol,
		Host:   c.DatastoreWriterHost + ":" + c.DatastoreWriterPort,
		Path:   "/internal/datastore/writer",
	}
	return &u
}

// waitForShutdown blocks until the service exits.
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

// // waitForService checks that all services at the given addresses are available.
// //
// // Blocks until a connection to every service can be established or the
// // context is canceled.
// func waitForService(ctx context.Context, addrs ...string) {
// 	d := net.Dialer{}

// 	var wg sync.WaitGroup
// 	for _, addr := range addrs {
// 		wg.Add(1)
// 		go func(addr string) {
// 			defer wg.Done()

// 			con, err := d.DialContext(ctx, "tcp", addr)
// 			for err != nil {
// 				if ctx.Err() != nil {
// 					// Time is up, dont try again.
// 					return
// 				}

// 				time.Sleep(100 * time.Millisecond)
// 				con, err = d.DialContext(ctx, "tcp", addr)
// 			}
// 			con.Close()

// 		}(addr)
// 	}
// 	wg.Wait()
// }
