package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"aigcpanel/go/internal/api"
	"aigcpanel/go/internal/app"
	"aigcpanel/go/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "db.json")
	st, err := store.NewJSONStore(dsn)
	if err != nil {
		t.Fatal(err)
	}
	return api.NewServer(app.NewService(st)).Routes()
}

func TestCreateAndListUsers(t *testing.T) {
	h := newTestServer(t)
	body := []byte(`{"name":"alice","email":"a@example.com"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var users []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestPatchTaskNotFound(t *testing.T) {
	h := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/not-exists", bytes.NewReader([]byte(`{"status":"done"}`)))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
