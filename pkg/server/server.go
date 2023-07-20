package server

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/action"
	"github.com/OpenSlides/openslides-manage-service/pkg/backendaction"
	"github.com/OpenSlides/openslides-manage-service/pkg/checkserver"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/datastorereader"
	"github.com/OpenSlides/openslides-manage-service/pkg/get"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/migrations"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/pkg/version"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Run starts the manage server.
func Run(cfg *Config) error {
	logger, err := shared.NewLogger(cfg.OpenSlidesLoglevel)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	addr := ":" + cfg.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on address %q: %w", addr, err)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc.UnaryServerInterceptor(logUnaryInterceptor),
			grpc.UnaryServerInterceptor(authUnaryInterceptor),
		),
	)
	manageSrv, err := newServer(cfg, logger)
	if err != nil {
		return fmt.Errorf("creating server object: %w", err)
	}
	proto.RegisterManageServer(grpcSrv, manageSrv)

	go func() {
		waitForShutdown()
		grpcSrv.GracefulStop()
	}()

	logger.Infof("Manage service is listening on %s\n", addr)
	if err := grpcSrv.Serve(lis); err != nil {
		return fmt.Errorf("running manage service: %w", err)
	}

	return nil
}

// srv implements the manage methods on server side.
type srv struct {
	config *Config
	pw     []byte
	logger shared.Logger
}

func newServer(cfg *Config, logger shared.Logger) (*srv, error) {
	pw, err := shared.AuthSecret(cfg.ManageAuthPasswordFile, cfg.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting server auth secret: %w", err)
	}
	s := &srv{
		config: cfg,
		pw:     pw,
		logger: logger,
	}
	return s, nil
}

func (s *srv) CheckServer(ctx context.Context, in *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendHealthURL(), pw, backendaction.HealthRoute)
	return checkserver.CheckServer(ctx, in, a), nil // CheckServer does not return an error for better handling in the client.

}

func (s *srv) InitialData(ctx context.Context, in *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendActionURL(), pw, backendaction.ActionRoute)
	return initialdata.InitialData(ctx, in, s.config.SuperadminPasswordFile, a)

}

func (s *srv) Migrations(ctx context.Context, in *proto.MigrationsRequest) (*proto.MigrationsResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendMigrationsURL(), pw, backendaction.MigrationsRoute)
	return migrations.Migrations(ctx, in, a)

}

func (s *srv) CreateUser(ctx context.Context, in *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendActionURL(), pw, backendaction.ActionRoute)
	return createuser.CreateUser(ctx, in, a)
}

func (s *srv) SetPassword(ctx context.Context, in *proto.SetPasswordRequest) (*proto.SetPasswordResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendActionURL(), pw, backendaction.ActionRoute)
	return setpassword.SetPassword(ctx, in, a)
}

func (s *srv) Get(ctx context.Context, in *proto.GetRequest) (*proto.GetResponse, error) {
	ds := datastorereader.New(s.config.datastoreReaderURL())
	return get.Get(ctx, in, ds)
}

func (s *srv) Action(ctx context.Context, in *proto.ActionRequest) (*proto.ActionResponse, error) {
	pw, err := shared.AuthSecret(s.config.InternalAuthPasswordFile, s.config.OpenSlidesDevelopment)
	if err != nil {
		return nil, fmt.Errorf("getting internal auth password from file: %w", err)
	}
	a := backendaction.New(s.config.manageBackendActionURL(), pw, backendaction.ActionRoute)
	return action.Action(ctx, in, a)
}

func (s *srv) Version(ctx context.Context, in *proto.VersionRequest) (*proto.VersionResponse, error) {
	return version.Version(ctx, in, s.config.clientVersionURL())
}

func (s *srv) Health(ctx context.Context, in *proto.HealthRequest) (*proto.HealthResponse, error) {
	// Returns always true because the server is considered healthy if it is
	// able to return this response.
	return &proto.HealthResponse{Healthy: true}, nil
}

func logUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	info.Server.(*srv).logger.Debugf("Incomming unary RPC for %s: %v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("calling handler: %w", err)
	}
	return resp, nil
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

	if subtle.ConstantTimeCompare(password, s.pw) != 1 {
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
	Port                     string `env:"MANAGE_PORT,9008"`
	ManageAuthPasswordFile   string `env:"MANAGE_AUTH_PASSWORD_FILE,/run/secrets/manage_auth_password"`
	InternalAuthPasswordFile string `env:"INTERNAL_AUTH_PASSWORD_FILE,/run/secrets/internal_auth_password"`
	SuperadminPasswordFile   string `env:"SUPERADMIN_PASSWORD_FILE,/run/secrets/superadmin"`

	ManageActionProtocol string `env:"ACTION_PROTOCOL,http"`
	ManageActionHost     string `env:"ACTION_HOST,backendManage"`
	ManageActionPort     string `env:"ACTION_PORT,9002"`

	DatastoreReaderProtocol string `env:"DATASTORE_READER_PROTOCOL,http"`
	DatastoreReaderHost     string `env:"DATASTORE_READER_HOST,datastore-reader"`
	DatastoreReaderPort     string `env:"DATASTORE_READER_PORT,9010"`

	OpenSlidesDevelopment string `env:"OPENSLIDES_DEVELOPMENT,0"`
	OpenSlidesLoglevel    string `env:"OPENSLIDES_LOGLEVEL,info"`
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

// manageBackendActionURL returns an URL object to the backend action service
// with action route.
func (c *Config) manageBackendActionURL() *url.URL {
	u := url.URL{
		Scheme: c.ManageActionProtocol,
		Host:   c.ManageActionHost + ":" + c.ManageActionPort,
		Path:   "/internal/handle_request",
	}
	return &u
}

// manageBackendMigrationsURL returns an URL object to the backend action
// service with migrations route.
func (c *Config) manageBackendMigrationsURL() *url.URL {
	u := url.URL{
		Scheme: c.ManageActionProtocol,
		Host:   c.ManageActionHost + ":" + c.ManageActionPort,
		Path:   "/internal/migrations",
	}
	return &u
}

// manageBackendHealthURL returns an URL object to the backend action service
// with health route.
func (c *Config) manageBackendHealthURL() *url.URL {
	u := url.URL{
		Scheme: c.ManageActionProtocol,
		Host:   c.ManageActionHost + ":" + c.ManageActionPort,
		Path:   "/system/action/health",
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

// clientVersionURL returns an URL object to the client service.
func (c *Config) clientVersionURL() *url.URL {
	u := url.URL{ // TODO: Protocol, host and port should be retrieved from environment variables.
		Scheme: "http",
		Host:   "client:9001",
		Path:   "/assets/version.txt",
	}
	return &u
}

// waitForShutdown blocks until the service exits.
//
// It listens on SIGINT and SIGTERM. If the signal is received for a second
// time, the process is killed with statuscode 1.
func waitForShutdown() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, unix.SIGINT, unix.SIGTERM)
	<-sigs
	go func() {
		<-sigs
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
