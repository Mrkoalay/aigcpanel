package service

import (
	"os"
	"sync"

	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/store"
)

type taskService struct{}

var Task = new(taskService)

var (
	taskStoreOnce sync.Once
	taskStore     *store.SQLiteStore
	taskStoreErr  error
)

func (s *taskService) store() (*store.SQLiteStore, error) {
	taskStoreOnce.Do(func() {
		dsn := getenv("AIGCPANEL_SQLITE_DSN", "data/aigcpanel.db")
		taskStore, taskStoreErr = store.NewSQLiteStore(dsn)
	})
	return taskStore, taskStoreErr
}

func (s *taskService) CreateTask(task domain.AppTask) (domain.AppTask, error) {
	st, err := s.store()
	if err != nil {
		return domain.AppTask{}, err
	}
	return st.CreateTask(task)
}

func (s *taskService) GetTask(id int64) (domain.AppTask, error) {
	st, err := s.store()
	if err != nil {
		return domain.AppTask{}, err
	}
	return st.GetTask(id)
}

func (s *taskService) ListTasks(filters store.TaskFilters) ([]domain.AppTask, error) {
	st, err := s.store()
	if err != nil {
		return nil, err
	}
	return st.ListTasks(filters)
}

func (s *taskService) UpdateTask(id int64, updates map[string]any) (domain.AppTask, error) {
	st, err := s.store()
	if err != nil {
		return domain.AppTask{}, err
	}
	return st.UpdateTask(id, updates)
}

func (s *taskService) DeleteTask(id int64) error {
	st, err := s.store()
	if err != nil {
		return err
	}
	return st.DeleteTask(id)
}

func getenv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}
