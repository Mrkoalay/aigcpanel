package task

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Task struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateTaskRequest struct {
	Name string `json:"name"`
}

type Store struct {
	mu    sync.RWMutex
	seq   int
	tasks map[string]Task
}

func NewStore() *Store {
	return &Store{tasks: map[string]Task{}}
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		res = append(res, t)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.After(res[j].CreatedAt)
	})
	return res
}

func (s *Store) Create(name string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	id := time.Now().Format("20060102150405") + "-" + strconv.Itoa(s.seq)
	t := Task{
		ID:        id,
		Name:      name,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	s.tasks[t.ID] = t
	return t
}

func (s *Store) UpdateStatus(id string, status Status) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	t.Status = status
	s.tasks[id] = t
	return t, nil
}
