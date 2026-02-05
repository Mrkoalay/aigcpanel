package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"aigcpanel/internal/platform/db"
)

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	dir := t.TempDir()
	database, err := db.OpenFileDB(filepath.Join(dir, "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHTTPHandler(NewRepository(database)).Register(mux)
	return mux
}

func TestAPIsWithFileDB(t *testing.T) {
	mux := newTestMux(t)

	uReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"name":"alice","email":"a@example.com"}`))
	uRec := httptest.NewRecorder()
	mux.ServeHTTP(uRec, uReq)
	if uRec.Code != http.StatusCreated {
		t.Fatalf("unexpected user create status: %d", uRec.Code)
	}

	sReq := httptest.NewRequest(http.MethodPost, "/api/v1/servers", bytes.NewBufferString(`{"name":"gpu-server","type":"python","endpoint":"http://127.0.0.1:9000"}`))
	sRec := httptest.NewRecorder()
	mux.ServeHTTP(sRec, sReq)
	if sRec.Code != http.StatusCreated {
		t.Fatalf("unexpected server create status: %d", sRec.Code)
	}
	var sData struct {
		Data Server `json:"data"`
	}
	if err := json.NewDecoder(sRec.Body).Decode(&sData); err != nil {
		t.Fatal(err)
	}

	tReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"name":"tts-job","kind":"tts","serverId":`+strconv.FormatInt(sData.Data.ID, 10)+`,"payload":"{\"text\":\"hello\"}"}`))
	tRec := httptest.NewRecorder()
	mux.ServeHTTP(tRec, tReq)
	if tRec.Code != http.StatusCreated {
		t.Fatalf("unexpected task create status: %d", tRec.Code)
	}
	var tData struct {
		Data Task `json:"data"`
	}
	if err := json.NewDecoder(tRec.Body).Decode(&tData); err != nil {
		t.Fatal(err)
	}

	patch := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+strconv.FormatInt(tData.Data.ID, 10), bytes.NewBufferString(`{"status":"success","result":"ok"}`))
	patchRec := httptest.NewRecorder()
	mux.ServeHTTP(patchRec, patch)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("unexpected task patch status: %d", patchRec.Code)
	}

	list := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, list)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listRec.Code)
	}
}

func TestElectronCompatibleUserInfoEndpoint(t *testing.T) {
	mux := newTestMux(t)

	create := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"name":"alice","email":"a@example.com"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, create)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d", createRec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/app_manager/user_info", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			User map[string]any `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Code != 0 || payload.Data.User["name"] != "alice" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestPatchNotFoundReturns404(t *testing.T) {
	mux := newTestMux(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/999", bytes.NewBufferString(`{"status":"failed","result":"x"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestFileDBWritesData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aigcpanel.json")
	database, err := db.OpenFileDB(path)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(database)
	if _, err := repo.CreateUser(t.Context(), "bob", "bob@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
