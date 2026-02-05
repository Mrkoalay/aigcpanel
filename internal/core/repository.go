package core

import (
	"context"
	"errors"
	"time"

	"aigcpanel/internal/platform/db"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *db.FileDB
}

func NewRepository(database *db.FileDB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) CreateUser(_ context.Context, name, email string) (User, error) {
	now := time.Now()
	u := User{Name: name, Email: email, CreatedAt: now}
	err := r.db.Update(func(s *db.State) error {
		s.Seq["users"]++
		u.ID = s.Seq["users"]
		s.Users = append([]map[string]any{toUserMap(u)}, s.Users...)
		return nil
	})
	return u, err
}

func (r *Repository) ListUsers(_ context.Context) ([]User, error) {
	s := r.db.Snapshot()
	out := make([]User, 0, len(s.Users))
	for _, item := range s.Users {
		out = append(out, fromUserMap(item))
	}
	return out, nil
}

func (r *Repository) CreateServer(_ context.Context, s Server) (Server, error) {
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	err := r.db.Update(func(st *db.State) error {
		st.Seq["servers"]++
		s.ID = st.Seq["servers"]
		st.Servers = append([]map[string]any{toServerMap(s)}, st.Servers...)
		return nil
	})
	return s, err
}

func (r *Repository) ListServers(_ context.Context) ([]Server, error) {
	s := r.db.Snapshot()
	out := make([]Server, 0, len(s.Servers))
	for _, item := range s.Servers {
		out = append(out, fromServerMap(item))
	}
	return out, nil
}

func (r *Repository) UpdateServerStatus(_ context.Context, id int64, status string) error {
	found := false
	return r.db.Update(func(st *db.State) error {
		for i, raw := range st.Servers {
			s := fromServerMap(raw)
			if s.ID == id {
				s.Status = status
				s.UpdatedAt = time.Now()
				st.Servers[i] = toServerMap(s)
				found = true
				break
			}
		}
		if !found {
			return ErrNotFound
		}
		return nil
	})
}

func (r *Repository) CreateTask(_ context.Context, t Task) (Task, error) {
	now := time.Now()
	t.Status, t.Result, t.CreatedAt, t.UpdatedAt = TaskPending, "", now, now
	err := r.db.Update(func(st *db.State) error {
		st.Seq["tasks"]++
		t.ID = st.Seq["tasks"]
		st.Tasks = append([]map[string]any{toTaskMap(t)}, st.Tasks...)
		return nil
	})
	return t, err
}

func (r *Repository) ListTasks(_ context.Context) ([]Task, error) {
	s := r.db.Snapshot()
	out := make([]Task, 0, len(s.Tasks))
	for _, item := range s.Tasks {
		out = append(out, fromTaskMap(item))
	}
	return out, nil
}

func (r *Repository) UpdateTaskStatus(_ context.Context, id int64, status TaskStatus, result string) error {
	found := false
	return r.db.Update(func(st *db.State) error {
		for i, raw := range st.Tasks {
			t := fromTaskMap(raw)
			if t.ID == id {
				t.Status = status
				t.Result = result
				t.UpdatedAt = time.Now()
				st.Tasks[i] = toTaskMap(t)
				found = true
				break
			}
		}
		if !found {
			return ErrNotFound
		}
		return nil
	})
}

func toUserMap(u User) map[string]any {
	return map[string]any{"id": u.ID, "name": u.Name, "email": u.Email, "createdAt": u.CreatedAt.Format(time.RFC3339Nano)}
}
func fromUserMap(m map[string]any) User {
	return User{ID: toI64(m["id"]), Name: toS(m["name"]), Email: toS(m["email"]), CreatedAt: toTime(m["createdAt"])}
}
func toServerMap(s Server) map[string]any {
	return map[string]any{"id": s.ID, "name": s.Name, "type": s.Type, "endpoint": s.Endpoint, "status": s.Status, "createdAt": s.CreatedAt.Format(time.RFC3339Nano), "updatedAt": s.UpdatedAt.Format(time.RFC3339Nano)}
}
func fromServerMap(m map[string]any) Server {
	return Server{ID: toI64(m["id"]), Name: toS(m["name"]), Type: toS(m["type"]), Endpoint: toS(m["endpoint"]), Status: toS(m["status"]), CreatedAt: toTime(m["createdAt"]), UpdatedAt: toTime(m["updatedAt"])}
}
func toTaskMap(t Task) map[string]any {
	return map[string]any{"id": t.ID, "name": t.Name, "kind": t.Kind, "serverId": t.ServerID, "payload": t.Payload, "status": string(t.Status), "result": t.Result, "createdAt": t.CreatedAt.Format(time.RFC3339Nano), "updatedAt": t.UpdatedAt.Format(time.RFC3339Nano)}
}
func fromTaskMap(m map[string]any) Task {
	return Task{ID: toI64(m["id"]), Name: toS(m["name"]), Kind: toS(m["kind"]), ServerID: toI64(m["serverId"]), Payload: toS(m["payload"]), Status: TaskStatus(toS(m["status"])), Result: toS(m["result"]), CreatedAt: toTime(m["createdAt"]), UpdatedAt: toTime(m["updatedAt"])}
}
func toI64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	}
	return 0
}
func toS(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func toTime(v any) time.Time { t, _ := time.Parse(time.RFC3339Nano, toS(v)); return t }
