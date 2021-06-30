package client_test

import (
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
)

func TestRunClient(t *testing.T) {
	if err := client.RunClient(); err != nil {
		t.Errorf("running RunClient() failed: %v", err)
	}
}
