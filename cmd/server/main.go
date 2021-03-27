package main

import (
	"log"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/server"
	"github.com/OpenSlides/openslides-manage-service/pkg/util"
)

func main() {
	cfg := util.ServerConfigFromEnv(os.LookupEnv)
	if err := server.Run(cfg); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
