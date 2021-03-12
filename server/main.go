package main

import (
	"log"
	"os"

	"github.com/OpenSlides/openslides-manage-service/server/manageServer"
)

func main() {
	if err := manageServer.Run(os.Args); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
