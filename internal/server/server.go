package server

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

// Run starts the HTTP server and blocks until a SIGINT or SIGTERM is received,
// then shuts down gracefully allowing in-flight requests to complete.
func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("Interrupt received. Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("starting server on %s", s.addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
