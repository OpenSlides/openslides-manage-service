package main

import (
	"log"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/manage"
)

func main() {
	cfg := manage.ServerConfigFromEnv(os.LookupEnv)
	if err := manage.RunServer(cfg); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
