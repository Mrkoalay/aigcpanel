package core

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	repo *Repository
}

func NewHTTPHandler(repo *Repository) *HTTPHandler { return &HTTPHandler{repo: repo} }

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/users", h.users)
	mux.HandleFunc("/api/v1/servers", h.servers)
	mux.HandleFunc("/api/v1/servers/", h.serverStatus)
	mux.HandleFunc("/api/v1/tasks", h.tasks)
	mux.HandleFunc("/api/v1/tasks/", h.taskStatus)
	// compatibility endpoint for current Electron frontend call path
	mux.HandleFunc("/api/app_manager/user_info", h.appManagerUserInfo)
}

func (h *HTTPHandler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTPHandler) appManagerUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		writeBizError(w, http.StatusInternalServerError, "internal error")
		return
	}
	var user any = map[string]any{}
	if len(users) > 0 {
		u := users[0]
		user = map[string]any{
			"id":         strconv.FormatInt(u.ID, 10),
			"name":       u.Name,
			"avatar":     "",
			"deviceCode": "",
		}
	}
	writeBizOK(w, map[string]any{
		"apiToken": "",
		"user":     user,
		"data":     map[string]any{},
		"basic":    map[string]any{},
	})
}

func (h *HTTPHandler) users(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		items, err := h.repo.ListUsers(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	case http.MethodPost:
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		item, err := h.repo.CreateUser(ctx, strings.TrimSpace(req.Name), strings.TrimSpace(req.Email))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) servers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		items, err := h.repo.ListServers(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	case http.MethodPost:
		var req struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Endpoint string `json:"endpoint"`
			Status   string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Endpoint) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if req.Status == "" {
			req.Status = "stopped"
		}
		item, err := h.repo.CreateServer(ctx, Server{Name: req.Name, Type: req.Type, Endpoint: req.Endpoint, Status: req.Status})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) serverStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, err := parseID(strings.TrimPrefix(r.URL.Path, "/api/v1/servers/"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.repo.UpdateServerStatus(r.Context(), id, strings.TrimSpace(req.Status)); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "status": req.Status}})
}

func (h *HTTPHandler) tasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		items, err := h.repo.ListTasks(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	case http.MethodPost:
		var req struct {
			Name     string `json:"name"`
			Kind     string `json:"kind"`
			Payload  string `json:"payload"`
			ServerID int64  `json:"serverId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Kind) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		item, err := h.repo.CreateTask(ctx, Task{Name: strings.TrimSpace(req.Name), Kind: strings.TrimSpace(req.Kind), ServerID: req.ServerID, Payload: req.Payload})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) taskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, err := parseID(strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Status TaskStatus `json:"status"`
		Result string     `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Status != TaskPending && req.Status != TaskRunning && req.Status != TaskSuccess && req.Status != TaskFailed {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}
	if err := h.repo.UpdateTaskStatus(r.Context(), id, req.Status, req.Result); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "status": req.Status, "result": req.Result}})
}

func parseID(s string) (int64, error) { return strconv.ParseInt(strings.TrimSpace(s), 10, 64) }

func writeBizOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{"code": 0, "msg": "", "data": data})
}

func writeBizError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"code": 10000, "msg": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
