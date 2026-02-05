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
	// 健康检查
	mux.HandleFunc("GET /health", s.handleHealth)
	// 用户列表
	mux.HandleFunc("GET /api/v1/users", s.handleListUsers)
	// 创建用户
	mux.HandleFunc("POST /api/v1/users", s.handleCreateUser)
	// 模型服务列表
	mux.HandleFunc("GET /api/v1/servers", s.handleListServers)
	// 创建模型服务
	mux.HandleFunc("POST /api/v1/servers", s.handleCreateServer)
	// 更新模型服务状态
	mux.HandleFunc("PATCH /api/v1/servers/", s.handlePatchServer)
	// 声音档案列表
	mux.HandleFunc("GET /api/v1/voice-profiles", s.handleListVoiceProfiles)
	// 创建声音档案
	mux.HandleFunc("POST /api/v1/voice-profiles", s.handleCreateVoiceProfile)
	// 视频模板列表
	mux.HandleFunc("GET /api/v1/video-templates", s.handleListVideoTemplates)
	// 创建视频模板
	mux.HandleFunc("POST /api/v1/video-templates", s.handleCreateVideoTemplate)
	// 平台任务列表
	mux.HandleFunc("GET /api/v1/tasks", s.handleListTasks)
	// 创建平台任务
	mux.HandleFunc("POST /api/v1/tasks", s.handleCreateTask)
	// 更新平台任务状态
	mux.HandleFunc("PATCH /api/v1/tasks/", s.handlePatchTask)

	// 应用任务列表（支持 biz/type/status 过滤）
	mux.HandleFunc("GET /api/v1/app/tasks", s.handleListAppTasks)
	// 创建应用任务
	mux.HandleFunc("POST /api/v1/app/tasks", s.handleCreateAppTask)
	// 获取应用任务详情
	mux.HandleFunc("GET /api/v1/app/tasks/", s.handleGetAppTask)
	// 更新应用任务
	mux.HandleFunc("PATCH /api/v1/app/tasks/", s.handlePatchAppTask)
	// 删除应用任务
	mux.HandleFunc("DELETE /api/v1/app/tasks/", s.handleDeleteAppTask)

	// 应用存储列表（按 biz）
	mux.HandleFunc("GET /api/v1/app/storages", s.handleListStorages)
	// 创建应用存储记录
	mux.HandleFunc("POST /api/v1/app/storages", s.handleCreateStorage)
	// 获取应用存储记录
	mux.HandleFunc("GET /api/v1/app/storages/", s.handleGetStorage)
	// 更新应用存储记录
	mux.HandleFunc("PATCH /api/v1/app/storages/", s.handlePatchStorage)
	// 删除应用存储记录
	mux.HandleFunc("DELETE /api/v1/app/storages/", s.handleDeleteStorage)
	// 清空某业务应用存储
	mux.HandleFunc("DELETE /api/v1/app/storages", s.handleClearStorage)

	// 应用模板列表
	mux.HandleFunc("GET /api/v1/app/templates", s.handleListTemplates)
	// 创建应用模板
	mux.HandleFunc("POST /api/v1/app/templates", s.handleCreateTemplate)
	// 获取应用模板（ID 或 Name）
	mux.HandleFunc("GET /api/v1/app/templates/", s.handleGetTemplate)
	// 更新应用模板
	mux.HandleFunc("PATCH /api/v1/app/templates/", s.handlePatchTemplate)
	// 删除应用模板
	mux.HandleFunc("DELETE /api/v1/app/templates/", s.handleDeleteTemplate)
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

func (s *Server) handleListAppTasks(w http.ResponseWriter, r *http.Request) {
	biz := r.URL.Query().Get("biz")
	if biz == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "biz is required"})
		return
	}
	if status := r.URL.Query().Get("status"); status != "" {
		writeJSON(w, http.StatusOK, s.svc.ListAppTasksByStatus(biz, strings.Split(status, ",")))
		return
	}
	typeVal := 1
	if r.URL.Query().Get("type") == "2" {
		typeVal = 2
	}
	writeJSON(w, http.StatusOK, s.svc.ListAppTasks(biz, typeVal))
}

func (s *Server) handleGetAppTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/tasks/")
	out, err := s.svc.GetAppTask(id)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateAppTask(w http.ResponseWriter, r *http.Request) {
	var in domain.AppTask
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateAppTask(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) handlePatchAppTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/tasks/")
	var in map[string]any
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.UpdateAppTask(id, in)
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

func (s *Server) handleDeleteAppTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/tasks/")
	err := s.svc.DeleteAppTask(id)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListStorages(w http.ResponseWriter, r *http.Request) {
	biz := r.URL.Query().Get("biz")
	if biz == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "biz is required"})
		return
	}
	writeJSON(w, http.StatusOK, s.svc.ListStorages(biz))
}
func (s *Server) handleGetStorage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/storages/")
	out, err := s.svc.GetStorage(id)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}
func (s *Server) handleCreateStorage(w http.ResponseWriter, r *http.Request) {
	var in domain.StorageRecord
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateStorage(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (s *Server) handlePatchStorage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/storages/")
	var in map[string]any
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.UpdateStorage(id, in)
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
func (s *Server) handleDeleteStorage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/storages/")
	err := s.svc.DeleteStorage(id)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
func (s *Server) handleClearStorage(w http.ResponseWriter, r *http.Request) {
	biz := r.URL.Query().Get("biz")
	if biz == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "biz is required"})
		return
	}
	if err := s.svc.ClearStorage(biz); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListTemplates(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.ListAppTemplates())
}
func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	idOrName := strings.TrimPrefix(r.URL.Path, "/api/v1/app/templates/")
	if t, err := s.svc.GetAppTemplateByID(idOrName); err == nil {
		writeJSON(w, http.StatusOK, t)
		return
	}
	out, err := s.svc.GetAppTemplateByName(idOrName)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}
func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var in domain.AppTemplate
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.CreateAppTemplate(in)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (s *Server) handlePatchTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/templates/")
	var in map[string]any
	if err := decode(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.svc.UpdateAppTemplate(id, in)
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
func (s *Server) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/app/templates/")
	err := s.svc.DeleteAppTemplate(id)
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
