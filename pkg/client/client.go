package client

import (
	"errors"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/action"
	"github.com/OpenSlides/openslides-manage-service/pkg/checkserver"
	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/get"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/migrations"
	"github.com/OpenSlides/openslides-manage-service/pkg/set"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/version"
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
		Use:               "openslides",
		Short:             "Swiss army knife for OpenSlides admins",
		Long:              RootHelp,
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	cmd.AddCommand(
		setup.Cmd(),
		config.Cmd(),
		config.CmdCreateDefault(),
		checkserver.Cmd(),
		initialdata.Cmd(),
		migrations.Cmd(),
		createuser.Cmd(),
		setpassword.Cmd(),
		get.Cmd(),
		set.Cmd(),
		action.Cmd(),
		version.Cmd(),
	)

	return cmd
}
