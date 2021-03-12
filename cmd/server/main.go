package main

import (
	"log"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/server"
)

func main() {
	if err := server.Run(os.Args); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
