package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *DiskStore {
	t.Helper()
	store, err := NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}
	return store
}

// countBlobs returns the number of content blob files (excluding .refcount).
func countBlobs(t *testing.T, s *DiskStore) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(s.dataDir, "blobs"))
	if err != nil {
		t.Fatalf("reading blobs dir: %v", err)
	}
	n := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".refcount" {
			n++
		}
	}
	return n
}

func TestPutAndGet(t *testing.T) {
	store := newTestStore(t)
	initialBlobs := countBlobs(t, store)

	if err := store.Put("bucket1", "object1", []byte("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if n := countBlobs(t, store); n != initialBlobs+1 {
		t.Fatalf("blob count = %d, want %d", n, initialBlobs+1)
	}

	got, err := store.Get("bucket1", "object1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Get = %q, want %q", got, "hello")
	}
}

func TestGetNotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Get("bucket1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get missing = %v, want ErrNotFound", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	store := newTestStore(t)
	store.Put("bucket1", "object1", []byte("data"))

	if err := store.Delete("bucket1", "object1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if n := countBlobs(t, store); n != 0 {
		t.Fatalf("blob count = %d, want 0 (should be GC'd)", n)
	}

	_, err := store.Get("bucket1", "object1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	store := newTestStore(t)
	initialBlobs := countBlobs(t, store)

	err := store.Delete("bucket1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete missing = %v, want ErrNotFound", err)
	}

	if n := countBlobs(t, store); n != initialBlobs {
		t.Fatalf("blob count = %d, want %d", n, initialBlobs)
	}
}

func TestPutOverwrite(t *testing.T) {
	store := newTestStore(t)

	store.Put("bucket1", "object1", []byte("v1"))
	store.Put("bucket1", "object1", []byte("v2"))

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count = %d, want 1 (old blob should be GC'd)", n)
	}

	got, _ := store.Get("bucket1", "object1")
	if string(got) != "v2" {
		t.Errorf("Get after overwrite = %q, want %q", got, "v2")
	}
}

// Same content under different objectIDs shares a single blob on disk.
// Deleting one ref must not affect the other.
func TestDuplicateContentSharesBlob(t *testing.T) {
	store := newTestStore(t)
	content := []byte("identical-payload")

	store.Put("bucket1", "obj-a", content)
	store.Put("bucket1", "obj-b", content)

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count = %d, want 1 (content should be deduplicated)", n)
	}

	gotA, _ := store.Get("bucket1", "obj-a")
	gotB, _ := store.Get("bucket1", "obj-b")
	if string(gotA) != string(content) {
		t.Errorf("obj-a = %q, want %q", gotA, content)
	}
	if string(gotB) != string(content) {
		t.Errorf("obj-b = %q, want %q", gotB, content)
	}

	store.Delete("bucket1", "obj-a")

	gotB, err := store.Get("bucket1", "obj-b")
	if err != nil {
		t.Fatalf("Get obj-b after deleting obj-a: %v", err)
	}
	if string(gotB) != string(content) {
		t.Errorf("obj-b after obj-a deleted = %q, want %q", gotB, content)
	}

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count after one delete = %d, want 1 (still referenced)", n)
	}
}

// Same content across different buckets shares a single blob.
func TestCrossBucketDedup(t *testing.T) {
	store := newTestStore(t)
	content := []byte("shared-across-buckets")

	store.Put("bucket1", "obj1", content)
	store.Put("bucket2", "obj1", content)

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count = %d, want 1 (cross-bucket dedup)", n)
	}

	got1, _ := store.Get("bucket1", "obj1")
	got2, _ := store.Get("bucket2", "obj1")
	if string(got1) != string(content) || string(got2) != string(content) {
		t.Errorf("content mismatch after cross-bucket dedup")
	}
}

// PUT same content to the same key is idempotent — no extra blobs or refcount drift.
func TestIdempotentPut(t *testing.T) {
	store := newTestStore(t)
	content := []byte("same-data")

	store.Put("bucket1", "obj1", content)
	store.Put("bucket1", "obj1", content)

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count = %d, want 1 after idempotent put", n)
	}

	got, _ := store.Get("bucket1", "obj1")
	if string(got) != string(content) {
		t.Errorf("Get = %q, want %q", got, content)
	}

	// Refcount should be exactly 1 despite two puts, so delete should GC.
	store.Delete("bucket1", "obj1")
	if n := countBlobs(t, store); n != 0 {
		t.Fatalf("blob count after delete = %d, want 0 (should be GC'd)", n)
	}
}

// Overwriting a key with different content garbage-collects the old blob
// when no other refs point to it.
func TestOverwriteGarbageCollectsOldBlob(t *testing.T) {
	store := newTestStore(t)

	store.Put("bucket1", "obj1", []byte("v1"))
	store.Put("bucket1", "obj1", []byte("v2"))

	if n := countBlobs(t, store); n != 1 {
		t.Fatalf("blob count after overwrite = %d, want 1 (old blob should be GC'd)", n)
	}

	got, _ := store.Get("bucket1", "obj1")
	if string(got) != "v2" {
		t.Errorf("Get after overwrite = %q, want %q", got, "v2")
	}
}

func TestBucketIsolation(t *testing.T) {
	store := newTestStore(t)

	store.Put("bucket1", "object1", []byte("from-b1"))
	store.Put("bucket2", "object1", []byte("from-b2"))

	if n := countBlobs(t, store); n != 2 {
		t.Fatalf("blob count = %d, want 2 (different content, no dedup)", n)
	}

	got1, _ := store.Get("bucket1", "object1")
	got2, _ := store.Get("bucket2", "object1")
	if string(got1) != "from-b1" || string(got2) != "from-b2" {
		t.Errorf("content mismatch after bucket isolation")
	}
}
