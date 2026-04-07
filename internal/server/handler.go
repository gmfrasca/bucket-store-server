package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gfrasca/bucket-store-server/internal/store"
)

const maxBodySize = 100 * 1024 * 1024 // 100 MB

func isValidPathSegment(s string) bool {
	return s != "" && s != ".." && !strings.Contains(s, "/")
}

// validatePathParams checks bucket and objectID for path traversal attempts.
// Returns false and writes a 400 response if invalid.
func validatePathParams(w http.ResponseWriter, bucket, objectID string) bool {
	if !isValidPathSegment(bucket) || !isValidPathSegment(objectID) {
		http.Error(w, "invalid bucket or object ID", http.StatusBadRequest)
		return false
	}
	return true
}

// handlePut stores an object in the given bucket.
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	objectID := r.PathValue("objectID")
	defer r.Body.Close()

	if !validatePathParams(w, bucket, objectID) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
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

	if !validatePathParams(w, bucket, objectID) {
		return
	}

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

	if !validatePathParams(w, bucket, objectID) {
		return
	}

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
