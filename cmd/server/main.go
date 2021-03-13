package main

import (
	"log"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/server"
)

func main() {
	// TODO: Read env
	if err := server.Run(":8001"); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
