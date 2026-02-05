package app

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/store"
)

type Service struct {
	store *store.JSONStore
}

func NewService(s *store.JSONStore) *Service { return &Service{store: s} }

func (s *Service) Health() map[string]string {
	return map[string]string{"status": "ok", "service": "aigcpanel-go"}
}

func id(prefix string) string { return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()) }

func nowMS() int64 { return time.Now().UnixMilli() }

func parseIntID(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

func (s *Service) ListUsers() []domain.User { return s.store.Snapshot().Users }
func (s *Service) CreateUser(in domain.User) (domain.User, error) {
	now := time.Now()
	in.ID, in.CreatedAt, in.UpdatedAt = id("usr"), now, now
	if in.Role == "" {
		in.Role = "user"
	}
	return in, s.store.Update(func(db *domain.Database) error {
		db.Users = append(db.Users, in)
		return nil
	})
}

func (s *Service) ListServers() []domain.ModelServer { return s.store.Snapshot().ModelServers }
func (s *Service) CreateServer(in domain.ModelServer) (domain.ModelServer, error) {
	now := time.Now()
	in.ID, in.CreatedAt, in.UpdatedAt = id("srv"), now, now
	if in.Status == "" {
		in.Status = "stopped"
	}
	if in.Config == nil {
		in.Config = map[string]string{}
	}
	return in, s.store.Update(func(db *domain.Database) error {
		db.ModelServers = append(db.ModelServers, in)
		return nil
	})
}
func (s *Service) UpdateServerStatus(serverID, status string) (domain.ModelServer, error) {
	var out domain.ModelServer
	err := s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.ModelServers, func(item domain.ModelServer) bool { return item.ID == serverID })
		if i < 0 {
			return ErrNotFound
		}
		db.ModelServers[i].Status = status
		db.ModelServers[i].UpdatedAt = time.Now()
		out = db.ModelServers[i]
		return nil
	})
	return out, err
}

func (s *Service) ListVoiceProfiles() []domain.VoiceProfile { return s.store.Snapshot().VoiceProfiles }
func (s *Service) CreateVoiceProfile(in domain.VoiceProfile) (domain.VoiceProfile, error) {
	now := time.Now()
	in.ID, in.CreatedAt, in.UpdatedAt = id("vcp"), now, now
	return in, s.store.Update(func(db *domain.Database) error {
		db.VoiceProfiles = append(db.VoiceProfiles, in)
		return nil
	})
}

func (s *Service) ListVideoTemplates() []domain.VideoTemplate {
	return s.store.Snapshot().VideoTemplates
}
func (s *Service) CreateVideoTemplate(in domain.VideoTemplate) (domain.VideoTemplate, error) {
	now := time.Now()
	in.ID, in.CreatedAt, in.UpdatedAt = id("tpl"), now, now
	return in, s.store.Update(func(db *domain.Database) error {
		db.VideoTemplates = append(db.VideoTemplates, in)
		return nil
	})
}

func (s *Service) ListTasks() []domain.Task { return s.store.Snapshot().Tasks }
func (s *Service) CreateTask(in domain.Task) (domain.Task, error) {
	now := time.Now()
	in.ID, in.CreatedAt, in.UpdatedAt = id("tsk"), now, now
	if in.Status == "" {
		in.Status = "pending"
	}
	if in.Input == nil {
		in.Input = map[string]string{}
	}
	if in.Output == nil {
		in.Output = map[string]string{}
	}
	return in, s.store.Update(func(db *domain.Database) error {
		db.Tasks = append(db.Tasks, in)
		return nil
	})
}
func (s *Service) UpdateTaskStatus(taskID, status string) (domain.Task, error) {
	var out domain.Task
	err := s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.Tasks, func(item domain.Task) bool { return item.ID == taskID })
		if i < 0 {
			return ErrNotFound
		}
		db.Tasks[i].Status = status
		db.Tasks[i].UpdatedAt = time.Now()
		out = db.Tasks[i]
		return nil
	})
	return out, err
}

func (s *Service) ListAppTasks(biz string, taskType int) []domain.AppTask {
	tasks := s.store.Snapshot().AppTasks
	out := make([]domain.AppTask, 0)
	for _, t := range tasks {
		if t.Biz == biz && t.Type == taskType {
			out = append(out, t)
		}
	}
	return out
}

func (s *Service) ListAppTasksByStatus(biz string, status []string) []domain.AppTask {
	allow := map[string]bool{}
	for _, st := range status {
		allow[st] = true
	}
	out := make([]domain.AppTask, 0)
	for _, t := range s.store.Snapshot().AppTasks {
		if t.Biz == biz && allow[t.Status] {
			out = append(out, t)
		}
	}
	return out
}

func (s *Service) GetAppTask(id string) (domain.AppTask, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.AppTask{}, ErrNotFound
	}
	for _, t := range s.store.Snapshot().AppTasks {
		if t.ID == idNum {
			return t, nil
		}
	}
	return domain.AppTask{}, ErrNotFound
}

func (s *Service) CreateAppTask(in domain.AppTask) (domain.AppTask, error) {
	now := nowMS()
	in.ID = time.Now().UnixNano()
	in.CreatedAt = now
	in.UpdatedAt = now
	return in, s.store.Update(func(db *domain.Database) error {
		db.AppTasks = append(db.AppTasks, in)
		return nil
	})
}

func (s *Service) UpdateAppTask(id string, patch map[string]any) (domain.AppTask, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.AppTask{}, ErrNotFound
	}
	var out domain.AppTask
	err = s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.AppTasks, func(item domain.AppTask) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		t := db.AppTasks[i]
		for k, v := range patch {
			switch k {
			case "biz":
				t.Biz, _ = v.(string)
			case "type":
				switch vv := v.(type) {
				case float64:
					t.Type = int(vv)
				case int:
					t.Type = vv
				}
			case "title":
				t.Title, _ = v.(string)
			case "status":
				t.Status, _ = v.(string)
			case "statusMsg":
				t.StatusMsg, _ = v.(string)
			case "startTime":
				switch vv := v.(type) {
				case float64:
					t.StartTime = int64(vv)
				case int64:
					t.StartTime = vv
				}
			case "endTime":
				switch vv := v.(type) {
				case float64:
					t.EndTime = int64(vv)
				case int64:
					t.EndTime = vv
				}
			case "serverName":
				t.ServerName, _ = v.(string)
			case "serverTitle":
				t.ServerTitle, _ = v.(string)
			case "serverVersion":
				t.ServerVersion, _ = v.(string)
			case "param":
				t.Param, _ = v.(string)
			case "modelConfig":
				t.ModelConfig, _ = v.(string)
			case "jobResult":
				t.JobResult, _ = v.(string)
			case "result":
				t.Result, _ = v.(string)
			}
		}
		t.UpdatedAt = nowMS()
		db.AppTasks[i] = t
		out = t
		return nil
	})
	return out, err
}

func (s *Service) DeleteAppTask(id string) error {
	idNum, err := parseIntID(id)
	if err != nil {
		return ErrNotFound
	}
	return s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.AppTasks, func(item domain.AppTask) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		db.AppTasks = append(db.AppTasks[:i], db.AppTasks[i+1:]...)
		return nil
	})
}

func (s *Service) ListStorages(biz string) []domain.StorageRecord {
	out := make([]domain.StorageRecord, 0)
	for _, item := range s.store.Snapshot().Storages {
		if item.Biz == biz {
			out = append(out, item)
		}
	}
	return out
}

func (s *Service) GetStorage(id string) (domain.StorageRecord, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.StorageRecord{}, ErrNotFound
	}
	for _, item := range s.store.Snapshot().Storages {
		if item.ID == idNum {
			return item, nil
		}
	}
	return domain.StorageRecord{}, ErrNotFound
}

func (s *Service) CreateStorage(in domain.StorageRecord) (domain.StorageRecord, error) {
	now := nowMS()
	in.ID = time.Now().UnixNano()
	in.CreatedAt = now
	in.UpdatedAt = now
	return in, s.store.Update(func(db *domain.Database) error {
		db.Storages = append(db.Storages, in)
		return nil
	})
}

func (s *Service) UpdateStorage(id string, patch map[string]any) (domain.StorageRecord, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.StorageRecord{}, ErrNotFound
	}
	var out domain.StorageRecord
	err = s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.Storages, func(item domain.StorageRecord) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		t := db.Storages[i]
		if v, ok := patch["biz"].(string); ok {
			t.Biz = v
		}
		if v, ok := patch["title"].(string); ok {
			t.Title = v
		}
		if v, ok := patch["content"].(string); ok {
			t.Content = v
		}
		if v, ok := patch["sort"].(float64); ok {
			t.Sort = int64(v)
		}
		t.UpdatedAt = nowMS()
		db.Storages[i] = t
		out = t
		return nil
	})
	return out, err
}

func (s *Service) DeleteStorage(id string) error {
	idNum, err := parseIntID(id)
	if err != nil {
		return ErrNotFound
	}
	return s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.Storages, func(item domain.StorageRecord) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		db.Storages = append(db.Storages[:i], db.Storages[i+1:]...)
		return nil
	})
}

func (s *Service) ClearStorage(biz string) error {
	return s.store.Update(func(db *domain.Database) error {
		items := make([]domain.StorageRecord, 0)
		for _, item := range db.Storages {
			if item.Biz != biz {
				items = append(items, item)
			}
		}
		db.Storages = items
		return nil
	})
}

func (s *Service) ListAppTemplates() []domain.AppTemplate {
	return s.store.Snapshot().AppTemplates
}

func (s *Service) GetAppTemplateByID(id string) (domain.AppTemplate, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.AppTemplate{}, ErrNotFound
	}
	for _, item := range s.store.Snapshot().AppTemplates {
		if item.ID == idNum {
			return item, nil
		}
	}
	return domain.AppTemplate{}, ErrNotFound
}

func (s *Service) GetAppTemplateByName(name string) (domain.AppTemplate, error) {
	for _, item := range s.store.Snapshot().AppTemplates {
		if item.Name == name {
			return item, nil
		}
	}
	return domain.AppTemplate{}, ErrNotFound
}

func (s *Service) CreateAppTemplate(in domain.AppTemplate) (domain.AppTemplate, error) {
	now := nowMS()
	in.ID = time.Now().UnixNano()
	in.CreatedAt = now
	in.UpdatedAt = now
	return in, s.store.Update(func(db *domain.Database) error {
		db.AppTemplates = append(db.AppTemplates, in)
		return nil
	})
}

func (s *Service) UpdateAppTemplate(id string, patch map[string]any) (domain.AppTemplate, error) {
	idNum, err := parseIntID(id)
	if err != nil {
		return domain.AppTemplate{}, ErrNotFound
	}
	var out domain.AppTemplate
	err = s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.AppTemplates, func(item domain.AppTemplate) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		t := db.AppTemplates[i]
		if v, ok := patch["name"].(string); ok {
			t.Name = v
		}
		if v, ok := patch["video"].(string); ok {
			t.Video = v
		}
		if v, ok := patch["info"].(string); ok {
			t.Info = v
		}
		t.UpdatedAt = nowMS()
		db.AppTemplates[i] = t
		out = t
		return nil
	})
	return out, err
}

func (s *Service) DeleteAppTemplate(id string) error {
	idNum, err := parseIntID(id)
	if err != nil {
		return ErrNotFound
	}
	return s.store.Update(func(db *domain.Database) error {
		i := slices.IndexFunc(db.AppTemplates, func(item domain.AppTemplate) bool { return item.ID == idNum })
		if i < 0 {
			return ErrNotFound
		}
		db.AppTemplates = append(db.AppTemplates[:i], db.AppTemplates[i+1:]...)
		return nil
	})
}
