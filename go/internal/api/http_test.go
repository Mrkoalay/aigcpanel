package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestResolveLocalModelConfig(t *testing.T) {
	h := newTestServer(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"name":"demo","version":"1.0.0","title":"Demo","platformName":"linux","platformArch":"amd64","functions":["tts"]}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"configPath": configPath})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/app/local-models/resolve", bytes.NewReader(body))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["name"] != "demo" {
		t.Fatalf("expected name demo, got %v", out["name"])
	}
}
