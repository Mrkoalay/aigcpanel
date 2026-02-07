package service

import (
	"aigcpanel/go/internal/component/sqllite"
	"aigcpanel/go/internal/domain"
	"time"
)

type taskService struct{}

var Task = new(taskService)

func (s *taskService) CreateTask(task domain.DataTaskModel) (domain.DataTaskModel, error) {
	now := time.Now().UnixMilli()

	if task.CreatedAt == 0 {
		task.CreatedAt = now
	}
	if task.UpdatedAt == 0 {
		task.UpdatedAt = now
	}
	if task.Type == 0 {
		task.Type = 1
	}
	session := sqllite.GetSession()
	if err := session.Save(&task).Error; err != nil {
		return domain.DataTaskModel{}, err
	}

	return task, nil
}

/*
func (s *taskService) GetTask(id int64) (domain.DataTaskModel, error) {
	st, err := s.store()
	if err != nil {
		return domain.DataTaskModel{}, err
	}
	return st.GetTask(id)
}

func (s *taskService) ListTasks(filters sqllite.TaskFilters) ([]domain.DataTaskModel, error) {
	st, err := s.store()
	if err != nil {
		return nil, err
	}
	return st.ListTasks(filters)
}

func (s *taskService) UpdateTask(id int64, updates map[string]any) (domain.DataTaskModel, error) {
	st, err := s.store()
	if err != nil {
		return domain.DataTaskModel{}, err
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
*/
