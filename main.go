package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /objects/{bucket}/{objectID}", handlePut)
	mux.HandleFunc("GET /objects/{bucket}/{objectID}", handleGet)
	mux.HandleFunc("DELETE /objects/{bucket}/{objectID}", handleDelete)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": objectID})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}
