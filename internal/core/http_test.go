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

	uReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"Name":"alice","Email":"a@example.com"}`))
	uRec := httptest.NewRecorder()
	mux.ServeHTTP(uRec, uReq)
	if uRec.Code != http.StatusCreated {
		t.Fatalf("unexpected user create status: %d", uRec.Code)
	}

	sReq := httptest.NewRequest(http.MethodPost, "/api/v1/servers", bytes.NewBufferString(`{"Name":"gpu-server","Type":"python","Endpoint":"http://127.0.0.1:9000"}`))
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

	tReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"Name":"tts-job","Kind":"tts","ServerID":`+strconv.FormatInt(sData.Data.ID, 10)+`,"Payload":"{\"text\":\"hello\"}"}`))
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
