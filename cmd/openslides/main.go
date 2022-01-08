package main

import (
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
)

func main() {
	code := client.RunClient()
	if code != 0 {
		os.Exit(code)
	}
}
