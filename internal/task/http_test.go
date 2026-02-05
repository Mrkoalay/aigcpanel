package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskLifecycle(t *testing.T) {
	store := NewStore()
	h := NewHTTPHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"name":"demo"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", createRec.Code)
	}
	var createResp struct {
		Data Task `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response failed: %v", err)
	}
	if createResp.Data.ID == "" {
		t.Fatalf("task id should not be empty")
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+createResp.Data.ID, bytes.NewBufferString(`{"status":"running"}`))
	patchRec := httptest.NewRecorder()
	mux.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("unexpected patch status: %d", patchRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listRec.Code)
	}
}
