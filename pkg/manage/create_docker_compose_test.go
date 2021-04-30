package manage_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

type TestWriteCloser struct {
	Buffer *bytes.Buffer
}

func NewTestWriteCloser() *TestWriteCloser {
	t := new(TestWriteCloser)
	t.Buffer = new(bytes.Buffer)
	return t
}

func (t TestWriteCloser) Write(p []byte) (n int, err error) {
	return t.Buffer.Write(p)
}

func (t TestWriteCloser) Close() error {
	return nil
}

func TestCreateDockerComposeYML(t *testing.T) {
	cases := []bool{
		true,
		false,
	}

	for _, remote := range cases {
		wc := NewTestWriteCloser()
		creator := func(name string) (io.WriteCloser, error) {
			return wc, nil
		}

		if err := manage.CreateDockerComposeYML(context.Background(), creator, "", remote); err != nil {
			t.Errorf("CreateDockerComposeYML should not return an error, but returns error: %s", err)
		}

		if strings.Contains(wc.Buffer.String(), "ghcr.io") != remote {
			t.Error(
				"CreateDockerComposeYML with remote true should and with remote ",
				"false should not write GitHub Container Registry URIs to the file")
		}
	}
}
