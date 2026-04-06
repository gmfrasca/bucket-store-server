package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gfrasca/rhobs-challenge/internal/store"
)

// handlePut stores an object in the given bucket.
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	objectID := r.PathValue("objectID")
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
		}
		return
	}

	if err := s.store.Put(bucket, objectID, data); err != nil {
		http.Error(w, "failed to store object", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": objectID})
}

// handleGet retrieves an object from the given bucket.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	objectID := r.PathValue("objectID")

	data, err := s.store.Get(bucket, objectID)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to retrieve object", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// handleDelete removes an object from the given bucket.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	objectID := r.PathValue("objectID")

	err := s.store.Delete(bucket, objectID)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to delete object", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
