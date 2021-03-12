package main

import (
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
)

func main() {
	if err := client.Execute(); err != nil {
		os.Exit(1)
	}
}
