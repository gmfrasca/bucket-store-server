package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// DiskStore implements ObjectStore using content-addressable storage (CAS).
// Identical data across any bucket is stored as a single blob identified by
// its SHA-256 hash. On-disk refcount files track how many references point
// to each blob so it can be garbage-collected when the last ref is removed.
//
// Layout:
//
//	dataDir/blobs/<hash>             : raw content
//	dataDir/blobs/<hash>.refcount    : integer reference count
//	dataDir/refs/<bucket>/<objectID> : text file containing the blob hash
type DiskStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewDiskStore creates a DiskStore rooted at dataDir, creating the blobs/
// and refs/ subdirectories if needed.
func NewDiskStore(dataDir string) (*DiskStore, error) {
	for _, sub := range []string{"blobs", "refs"} {
		if err := os.MkdirAll(filepath.Join(dataDir, sub), 0o755); err != nil {
			return nil, fmt.Errorf("creating %s directory: %w", sub, err)
		}
	}
	return &DiskStore{dataDir: dataDir}, nil
}

func (s *DiskStore) blobPath(hash string) string {
	return filepath.Join(s.dataDir, "blobs", hash)
}

func (s *DiskStore) refcountPath(hash string) string {
	return filepath.Join(s.dataDir, "blobs", hash+".refcount")
}

func (s *DiskStore) refPath(bucket, objectID string) string {
	return filepath.Join(s.dataDir, "refs", bucket, objectID)
}

func contentHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (s *DiskStore) readRef(bucket, objectID string) (string, error) {
	data, err := os.ReadFile(s.refPath(bucket, objectID))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading ref: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (s *DiskStore) readRefCount(hash string) (int, error) {
	data, err := os.ReadFile(s.refcountPath(hash))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reading refcount: %w", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parsing refcount: %w", err)
	}
	return n, nil
}

func (s *DiskStore) writeRefCount(hash string, count int) error {
	return os.WriteFile(s.refcountPath(hash), []byte(strconv.Itoa(count)), 0o644)
}

// removeBlob deletes the blob and its refcount file.
func (s *DiskStore) removeBlob(hash string) error {
	if err := os.Remove(s.blobPath(hash)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing blob: %w", err)
	}
	if err := os.Remove(s.refcountPath(hash)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing refcount: %w", err)
	}
	return nil
}

// decrRef decrements the refcount for hash and garbage-collects the blob
// if the count reaches zero.
func (s *DiskStore) decrRef(hash string) error {
	count, err := s.readRefCount(hash)
	if err != nil {
		return err
	}
	count--
	if count <= 0 {
		return s.removeBlob(hash)
	}
	return s.writeRefCount(hash, count)
}

// Put stores data using content-addressable storage. If the same content
// already exists (in any bucket), the blob is shared. Overwrites update
// the reference and garbage-collect the old blob if its refcount drops to zero.
func (s *DiskStore) Put(bucket, objectID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := contentHash(data)

	oldH, err := s.readRef(bucket, objectID)
	if err != nil {
		return err
	}
	if oldH == h {
		return nil
	}

	if _, err := os.Stat(s.blobPath(h)); os.IsNotExist(err) {
		if err := os.WriteFile(s.blobPath(h), data, 0o644); err != nil {
			return fmt.Errorf("writing blob: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("checking blob existence: %w", err)
	}

	count, err := s.readRefCount(h)
	if err != nil {
		return err
	}
	if err := s.writeRefCount(h, count+1); err != nil {
		return fmt.Errorf("updating refcount: %w", err)
	}

	refDir := filepath.Join(s.dataDir, "refs", bucket)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return fmt.Errorf("creating ref directory: %w", err)
	}
	if err := os.WriteFile(s.refPath(bucket, objectID), []byte(h), 0o644); err != nil {
		return fmt.Errorf("writing ref: %w", err)
	}

	if oldH != "" {
		if err := s.decrRef(oldH); err != nil {
			return fmt.Errorf("cleaning up old blob: %w", err)
		}
	}

	return nil
}

// Get returns the object data by resolving the ref to a blob hash,
// or ErrNotFound if the ref doesn't exist.
func (s *DiskStore) Get(bucket, objectID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h, err := s.readRef(bucket, objectID)
	if err != nil {
		return nil, err
	}
	if h == "" {
		return nil, ErrNotFound
	}

	data, err := os.ReadFile(s.blobPath(h))
	if err != nil {
		return nil, fmt.Errorf("reading blob: %w", err)
	}
	return data, nil
}

// Delete removes the ref and decrements the blob's refcount,
// garbage-collecting the blob if no refs remain.
// Returns ErrNotFound if the ref doesn't exist.
func (s *DiskStore) Delete(bucket, objectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	h, err := s.readRef(bucket, objectID)
	if err != nil {
		return err
	}
	if h == "" {
		return ErrNotFound
	}

	if err := os.Remove(s.refPath(bucket, objectID)); err != nil {
		return fmt.Errorf("removing ref: %w", err)
	}

	if err := s.decrRef(h); err != nil {
		return fmt.Errorf("decrementing refcount: %w", err)
	}

	return nil
}
