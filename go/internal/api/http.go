package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"aigcpanel/go/internal/app"
	"aigcpanel/go/internal/domain"
)

type Server struct{ svc *app.Service }

func NewServer(svc *app.Service) *Server { return &Server{svc: svc} }

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/users", s.handleListUsers)
	mux.HandleFunc("POST /api/v1/users", s.handleCreateUser)
	mux.HandleFunc("GET /api/v1/servers", s.handleListServers)
	mux.HandleFunc("POST /api/v1/servers", s.handleCreateServer)
	mux.HandleFunc("PATCH /api/v1/servers/", s.handlePatchServer)
	mux.HandleFunc("GET /api/v1/voice-profiles", s.handleListVoiceProfiles)
	mux.HandleFunc("POST /api/v1/voice-profiles", s.handleCreateVoiceProfile)
	mux.HandleFunc("GET /api/v1/video-templates", s.handleListVideoTemplates)
	mux.HandleFunc("POST /api/v1/video-templates", s.handleCreateVideoTemplate)
	mux.HandleFunc("GET /api/v1/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/v1/tasks", s.handleCreateTask)
	mux.HandleFunc("PATCH /api/v1/tasks/", s.handlePatchTask)
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(r *http.Request, dst any) error { return json.NewDecoder(r.Body).Decode(dst) }

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.Health())
}
func (s *Server) handleListUsers(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListUsers())
}
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var in domain.User
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateUser(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) handleListServers(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListServers())
}
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var in domain.ModelServer
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateServer(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (s *Server) handlePatchServer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/servers/")
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.UpdateServerStatus(id, in.Status)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListVoiceProfiles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListVoiceProfiles())
}
func (s *Server) handleCreateVoiceProfile(w http.ResponseWriter, r *http.Request) {
	var in domain.VoiceProfile
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateVoiceProfile(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) handleListVideoTemplates(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListVideoTemplates())
}
func (s *Server) handleCreateVideoTemplate(w http.ResponseWriter, r *http.Request) {
	var in domain.VideoTemplate
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateVideoTemplate(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) handleListTasks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListTasks())
}
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var in domain.Task
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateTask(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (s *Server) handlePatchTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.UpdateTaskStatus(id, in.Status)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}
