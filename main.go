package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gfrasca/bucket-store-server/internal/server"
	"github.com/gfrasca/bucket-store-server/internal/store"
)

var defaultDataDir = "./data"
var defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	s, err := store.NewDiskStore(dataDir)
	if err != nil {
		log.Fatalf("failed to create disk store: %v", err)
	}

	srv := server.New(s, fmt.Sprintf(":%s", port))
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
