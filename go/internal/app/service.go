package app

import (
	"fmt"
	"slices"
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
