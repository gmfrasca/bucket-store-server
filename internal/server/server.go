package server

import (
	"log"
	"net/http"

	"github.com/gfrasca/bucket-store-server/internal/store"
)

// Server handles HTTP requests for the object storage API.
type Server struct {
	store store.ObjectStore
	addr  string
}

// New creates a Server backed by the given ObjectStore, listening on addr.
func New(s store.ObjectStore, addr string) *Server {
	return &Server{store: s, addr: addr}
}

// Handler returns the HTTP handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /objects/{bucket}/{objectID}", s.handlePut)
	mux.HandleFunc("GET /objects/{bucket}/{objectID}", s.handleGet)
	mux.HandleFunc("DELETE /objects/{bucket}/{objectID}", s.handleDelete)
	return mux
}

// Run starts the HTTP server. Blocks until the server exits.
func (s *Server) Run() error {
	log.Printf("starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.Handler())
}
