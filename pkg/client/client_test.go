package client_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/get"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/migrations"
	"github.com/OpenSlides/openslides-manage-service/pkg/set"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/version"
)

func TestRunClient(t *testing.T) {
	if code := client.RunClient(); code != 0 {
		t.Fatal("running RunClient() failed")
	}
}

func TestCmdHelpTexts(t *testing.T) {
	cmd := client.RootCmd()
	cmdTests := []struct {
		name             string
		input            []string
		outputStartsWith []byte
	}{
		{
			name:             "root command",
			input:            []string{},
			outputStartsWith: []byte(client.RootHelp),
		},

		{
			name:             "setup command",
			input:            []string{"setup", "--help"},
			outputStartsWith: []byte(setup.SetupHelp),
		},

		{
			name:             "config command",
			input:            []string{"config", "--help"},
			outputStartsWith: []byte(config.ConfigHelp),
		},

		{
			name:             "config create default",
			input:            []string{"config-create-default", "--help"},
			outputStartsWith: []byte(config.ConfigCreateDefaultHelp),
		},

		{
			name:             "initial-data command",
			input:            []string{"initial-data", "--help"},
			outputStartsWith: []byte(initialdata.InitialDataHelp),
		},

		{
			name:             "migrations command",
			input:            []string{"migrations", "--help"},
			outputStartsWith: []byte(migrations.MigrationsHelp),
		},

		{
			name:             "createuser command",
			input:            []string{"create-user", "--help"},
			outputStartsWith: []byte(createuser.CreateUserHelp),
		},

		{
			name:             "set-password command",
			input:            []string{"set-password", "--help"},
			outputStartsWith: []byte(setpassword.SetPasswordHelp),
		},

		{
			name:             "get command",
			input:            []string{"get", "--help"},
			outputStartsWith: []byte(get.GetHelp),
		},

		{
			name:             "set command",
			input:            []string{"set", "--help"},
			outputStartsWith: []byte(set.SetHelp),
		},

		{
			name:             "version",
			input:            []string{"version", "--help"},
			outputStartsWith: []byte(version.VersionHelp),
		},
	}

	for _, tt := range cmdTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.input)

			cmd.Execute()

			got, err := ioutil.ReadAll(buf)
			if err != nil {
				t.Fatalf("reading output from command execution: %v", err)
			}
			gotStartsWith := got[:len(tt.outputStartsWith)]
			if !bytes.Equal(tt.outputStartsWith, gotStartsWith) {
				t.Fatalf("wrong cobra command output, expected %q, got %q", tt.outputStartsWith, gotStartsWith)
			}

		})
	}
}
