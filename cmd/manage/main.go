package main

import (
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

func main() {
	if err := manage.RunClient(); err != nil {
		os.Exit(1)
	}
}
