package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gfrasca/bucket-store-server/internal/server"
	"github.com/gfrasca/bucket-store-server/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	testStore, err := store.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatalf("error creating new test server's disk store: %v", err)
	}
	return server.New(testStore, "").Handler()
}

func TestPutReturns201(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("PUT status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if !strings.Contains(rec.Body.String(), `"id":"object1"`) {
		t.Errorf("PUT body = %q, want JSON with id:object1", rec.Body.String())
	}
}

func TestGetReturnsStoredObject(t *testing.T) {
	server := newTestServer(t)

	put := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("hello"))
	server.ServeHTTP(httptest.NewRecorder(), put)

	req := httptest.NewRequest("GET", "/objects/bucket1/object1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("GET body = %q, want body:hello", rec.Body.String())
	}
}

func TestGetNotFoundReturns404(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest("GET", "/objects/bucket1/missing", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteReturns200(t *testing.T) {
	server := newTestServer(t)

	put := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("data"))
	server.ServeHTTP(httptest.NewRecorder(), put)

	req := httptest.NewRequest("DELETE", "/objects/bucket1/object1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("DELETE status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDeleteNotFoundReturns404(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest("DELETE", "/objects/bucket1/missing", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("DELETE missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetAfterDeleteReturns404(t *testing.T) {
	server := newTestServer(t)

	put := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("data"))
	server.ServeHTTP(httptest.NewRecorder(), put)

	del := httptest.NewRequest("DELETE", "/objects/bucket1/object1", nil)
	server.ServeHTTP(httptest.NewRecorder(), del)

	req := httptest.NewRequest("GET", "/objects/bucket1/object1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPutOverwriteReturnsUpdatedData(t *testing.T) {
	server := newTestServer(t)

	put1 := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("v1"))
	server.ServeHTTP(httptest.NewRecorder(), put1)

	put2 := httptest.NewRequest("PUT", "/objects/bucket1/object1", strings.NewReader("v2"))
	server.ServeHTTP(httptest.NewRecorder(), put2)

	req := httptest.NewRequest("GET", "/objects/bucket1/object1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Body.String() != "v2" {
		t.Errorf("GET after overwrite = %q, want %q", rec.Body.String(), "v2")
	}
}
