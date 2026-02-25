package service

import (
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
)

type dataVideoTemplateService struct{}

var DataVideoTemplate = new(dataVideoTemplateService)

type DataVideoTemplateFilters struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

func (s *dataVideoTemplateService) Create(record domain.DataVideoTemplateModel) (domain.DataVideoTemplateModel, error) {
	session := sqllite.GetSession()
	if err := session.Create(&record).Error; err != nil {
		return domain.DataVideoTemplateModel{}, err
	}
	return record, nil
}

func (s *dataVideoTemplateService) Get(id int64) (domain.DataVideoTemplateModel, error) {
	session := sqllite.GetSession()
	var model domain.DataVideoTemplateModel
	if err := session.First(&model, id).Error; err != nil {
		return domain.DataVideoTemplateModel{}, err
	}
	return model, nil
}

func (s *dataVideoTemplateService) List(req DataVideoTemplateFilters) ([]domain.DataVideoTemplateModel, error) {
	session := sqllite.GetSession()
	query := session.Model(&domain.DataVideoTemplateModel{})
	if req.Page > 0 {
		limit := req.Size
		offset := (req.Page - 1) * req.Size
		query = query.Limit(limit).Offset(offset)
	}
	var models []domain.DataVideoTemplateModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *dataVideoTemplateService) Update(id int64, updates map[string]any) (domain.DataVideoTemplateModel, error) {
	if len(updates) == 0 {
		return s.Get(id)
	}

	session := sqllite.GetSession()
	if err := session.Model(&domain.DataVideoTemplateModel{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return domain.DataVideoTemplateModel{}, err
	}
	return s.Get(id)
}

func (s *dataVideoTemplateService) Delete(id int64) error {
	session := sqllite.GetSession()
	return session.Delete(&domain.DataVideoTemplateModel{}, id).Error
}
