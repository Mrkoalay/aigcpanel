package service

import (
	"time"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
)

type taskService struct{}

var DataTask = new(taskService)

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

func (s *taskService) GetTask(id int64) (domain.DataTaskModel, error) {
	session := sqllite.GetSession()
	var model domain.DataTaskModel
	if err := session.First(&model, id).Error; err != nil {
		return domain.DataTaskModel{}, err
	}
	return model, nil
}

func (s *taskService) ListTasks(filters sqllite.TaskFilters) ([]domain.DataTaskModel, error) {
	session := sqllite.GetSession()
	query := session.Model(&domain.DataTaskModel{})

	if filters.Biz != "" {
		query = query.Where("biz = ?", filters.Biz)
	}

	if len(filters.Status) > 0 {
		statuses := make([]string, 0, len(filters.Status))
		for _, status := range filters.Status {
			if status != "" {
				statuses = append(statuses, status)
			}
		}
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
	}

	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	if filters.Page > 0 {
		limit := filters.Size
		offset := (filters.Page - 1) * filters.Size
		query = query.Limit(limit).Offset(offset)
	}

	var models []domain.DataTaskModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *taskService) UpdateTask(id int64, updates map[string]any) (domain.DataTaskModel, error) {
	if len(updates) == 0 {
		return s.GetTask(id)
	}
	updates["updatedAt"] = time.Now().UnixMilli()

	session := sqllite.GetSession()
	if err := session.Model(&domain.DataTaskModel{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return domain.DataTaskModel{}, err
	}

	return s.GetTask(id)
}

func (s *taskService) DeleteTask(id int64) error {
	session := sqllite.GetSession()
	return session.Delete(&domain.DataTaskModel{}, id).Error
}
