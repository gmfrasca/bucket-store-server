package store

import (
	"errors"
	"testing"
)

// newTestStore creates a new DiskStore with a temporary directory for testing.
func newTestStore(t *testing.T) *DiskStore {
	t.Helper()
	s, err := NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}
	return s
}

// Test the Put and Get methods of the DiskStore.
func TestPutAndGet(t *testing.T) {
	store := newTestStore(t)

	if err := store.Put("bucket1", "object1", []byte("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, err := store.Get("bucket1", "object1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Get = %q, want %q", got, "hello")
	}
}

// Test the Get method of the DiskStore when the object is not found.
func TestGetNotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Get("bucket1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get missing = %v, want ErrNotFound", err)
	}
}

// Test the Delete method of the DiskStore when the object is successfully deleted.
func TestDeleteSuccess(t *testing.T) {
	store := newTestStore(t)
	store.Put("bucket1", "object1", []byte("data"))

	if err := store.Delete("bucket1", "object1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := store.Get("bucket1", "object1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
}

// Test the Delete method of the DiskStore when the object is not found.
func TestDeleteNotFound(t *testing.T) {
	store := newTestStore(t)

	err := store.Delete("bucket1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete missing = %v, want ErrNotFound", err)
	}
}

// Test the Put method of the DiskStore when the object already exists.
func TestPutOverwrite(t *testing.T) {
	store := newTestStore(t)

	store.Put("bucket1", "object1", []byte("v1"))
	store.Put("bucket1", "object1", []byte("v2"))

	got, _ := store.Get("bucket1", "object1")
	if string(got) != "v2" {
		t.Errorf("Get after overwrite = %q, want %q", got, "v2")
	}
}

// Dedup is key-based: same content under different objectIDs produces
// independent objects. If content-based dedup is added, this test
// should be updated to reflect shared storage.
func TestDuplicateContentIndependentKeys(t *testing.T) {
	s := newTestStore(t)
	content := []byte("identical-payload")

	s.Put("bucket1", "obj-a", content)
	s.Put("bucket1", "obj-b", content)

	gotA, _ := s.Get("bucket1", "obj-a")
	gotB, _ := s.Get("bucket1", "obj-b")

	if string(gotA) != string(content) {
		t.Errorf("obj-a = %q, want %q", gotA, content)
	}
	if string(gotB) != string(content) {
		t.Errorf("obj-b = %q, want %q", gotB, content)
	}

	// Deleting one must not affect the other.
	s.Delete("bucket1", "obj-a")

	gotB, err := s.Get("bucket1", "obj-b")
	if err != nil {
		t.Fatalf("Get obj-b after deleting obj-a: %v", err)
	}
	if string(gotB) != string(content) {
		t.Errorf("obj-b after obj-a deleted = %q, want %q", gotB, content)
	}
}

// Test the bucket isolation of the DiskStore.
func TestBucketIsolation(t *testing.T) {
	store := newTestStore(t)

	store.Put("bucket1", "object1", []byte("from-b1"))
	store.Put("bucket2", "object1", []byte("from-b2"))

	got1, _ := store.Get("bucket1", "object1")
	got2, _ := store.Get("bucket2", "object1")

	if string(got1) != "from-b1" {
		t.Errorf("bucket1/object1 = %q, want %q", got1, "from-b1")
	}
	if string(got2) != "from-b2" {
		t.Errorf("bucket2/object1 = %q, want %q", got2, "from-b2")
	}
}
