package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiskStore implements ObjectStore with on-disk persistence.
// Objects are stored as files at <dataDir>/<bucket>/<objectID>.
type DiskStore struct {
	dataDir string
}

// NewDiskStore creates a DiskStore rooted at dataDir, creating it if needed.
func NewDiskStore(dataDir string) (*DiskStore, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}
	return &DiskStore{dataDir: dataDir}, nil
}

// Put stores data, creating the bucket directory if it doesn't exist.
// Overwrites any existing object with the same objectID.
func (s *DiskStore) Put(bucket, objectID string, data []byte) error {
	dir := filepath.Join(s.dataDir, bucket)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating bucket directory: %w", err)
	}

	path := filepath.Join(dir, objectID)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing object: %w", err)
	}
	return nil
}

// Get returns the object data, or ErrNotFound if it doesn't exist.
func (s *DiskStore) Get(bucket, objectID string) ([]byte, error) {
	path := filepath.Join(s.dataDir, bucket, objectID)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("reading object: %w", err)
	}
	return data, nil
}

// Delete removes the object, or returns ErrNotFound if it doesn't exist.
func (s *DiskStore) Delete(bucket, objectID string) error {
	path := filepath.Join(s.dataDir, bucket, objectID)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}
