package store

import "errors"

var ErrNotFound = errors.New("object not found")

// ObjectStore defines the interface for object storage operations.
// Implementations handle how objects are persisted and deduplicated.
type ObjectStore interface {
	Put(bucket, objectID string, data []byte) error
	Get(bucket, objectID string) ([]byte, error)
	Delete(bucket, objectID string) error
}
