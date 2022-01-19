package client

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/tunnel"
	"github.com/spf13/cobra"
)

// RunClient is the entrypoint for the client tool of this service. It starts the root command.
func RunClient() int {
	err := RootCmd().Execute()

	if err == nil {
		return 0
	}

	code := 1
	var errExit interface {
		ExitCode() int
	}
	if errors.As(err, &errExit) {
		code = errExit.ExitCode()
		if code <= 0 {
			code = 1
			err = fmt.Errorf("wrong error code for error: %w", err)
		}
	}
	fmt.Printf("Error: %v\n", err)
	return code
}

// RootHelp is the main help text for the client tool.
const RootHelp = `openslides is an admin tool to setup an OpenSlides instance and perform manager actions on it.`

// RootCmd returns the root cobra command.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "openslides",
		Short:         "Swiss army knife for OpenSlides admins",
		Long:          RootHelp,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		setup.Cmd(),
		config.Cmd(),
		config.CmdCreateDefault(),
		unaryConnection(initialdata.Cmd),
		unaryConnection(setpassword.Cmd),
		unaryConnection(createuser.Cmd),
		streamConnection(tunnel.Cmd),
	)

	return cmd
}

type connectionParams struct {
	addr         *string
	passwordFile *string
	timeout      *time.Duration
	noSSL        *bool
}

func (c *connectionParams) Addr() string {
	return *c.addr
}

func (c *connectionParams) PasswordFile() string {
	return *c.passwordFile
}

func (c *connectionParams) Timeout() time.Duration {
	return *c.timeout
}

func (c *connectionParams) NoSSL() bool {
	return *c.noSSL
}

func unaryConnection(fn func(cmd *cobra.Command, cp connection.Params) *cobra.Command) *cobra.Command {
	cmd, cp := newCmdWithParams()
	cp.timeout = cmd.Flags().DurationP("timeout", "t", connection.DefaultTimeout, "time to wait for the command's response")
	return fn(cmd, cp)
}

func streamConnection(fn func(cmd *cobra.Command, cp connection.Params) *cobra.Command) *cobra.Command {
	return fn(newCmdWithParams())
}

func newCmdWithParams() (*cobra.Command, *connectionParams) {
	cmd := &cobra.Command{}
	addr := cmd.Flags().StringP("address", "a", connection.DefaultAddr, "address of the OpenSlides manage service")
	defaultPasswordFile := path.Join(".", setup.SecretsDirName, setup.ManageAuthPasswordFileName)
	passwordFile := cmd.Flags().String("password-file", defaultPasswordFile, "file with password for authorization to manage service, not usable in development mode")
	noSSL := cmd.Flags().Bool("no-ssl", false, "use an unencrypted connection to manage service")

	cp := &connectionParams{
		addr:         addr,
		passwordFile: passwordFile,
		noSSL:        noSSL,
	}
	return cmd, cp
}
