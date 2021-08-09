package client_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/tunnel"
)

func TestRunClient(t *testing.T) {
	if err := client.RunClient(); err != nil {
		t.Fatalf("running RunClient() failed: %v", err)
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
			name:             "initial-data command",
			input:            []string{"initial-data", "--help"},
			outputStartsWith: []byte(initialdata.InitialDataHelp),
		},

		{
			name:             "set-password command",
			input:            []string{"set-password", "--help"},
			outputStartsWith: []byte(setpassword.SetPasswordHelp),
		},

		{
			name:             "tunnel command",
			input:            []string{"tunnel", "--help"},
			outputStartsWith: []byte(tunnel.TunnelHelp),
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
