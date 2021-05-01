package manage_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

func TestCreateDockerComposeYML(t *testing.T) {
	cases := []bool{
		true,
		false,
	}

	for _, remote := range cases {
		buf := new(bytes.Buffer)
		creator := func(name string) (io.Writer, error) {
			return buf, nil
		}

		if err := manage.CreateDockerComposeYML(context.Background(), creator, "", remote); err != nil {
			t.Errorf("CreateDockerComposeYML should not return an error, but returns error: %s", err)
		}

		if remote {
			if !strings.Contains(buf.String(), "ghcr.io") {
				t.Error(
					"CreateDockerComposeYML with remote true should write ",
					"GitHub Container Registry URIs to the file")
			}
			continue
		}
		if strings.Contains(buf.String(), "ghcr.io") {
			t.Error(
				"CreateDockerComposeYML with remote false should not write ",
				"GitHub Container Registry URIs to the file")
		}
	}
}
