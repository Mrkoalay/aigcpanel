package service

import (
	"time"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
)

type storageService struct{}

var DataStorage = new(storageService)

func (s *storageService) CreateStorage(record domain.DataStorageModel) (domain.DataStorageModel, error) {
	now := time.Now().UnixMilli()
	if record.CreatedAt == 0 {
		record.CreatedAt = now
	}
	if record.UpdatedAt == 0 {
		record.UpdatedAt = now
	}

	session := sqllite.GetSession()
	if err := session.Create(&record).Error; err != nil {
		return domain.DataStorageModel{}, err
	}
	return record, nil
}

func (s *storageService) GetStorage(id int64) (domain.DataStorageModel, error) {
	session := sqllite.GetSession()
	var model domain.DataStorageModel
	if err := session.First(&model, id).Error; err != nil {
		return domain.DataStorageModel{}, err
	}
	return model, nil
}

func (s *storageService) ListStorages(biz string) ([]domain.DataStorageModel, error) {
	session := sqllite.GetSession()
	query := session.Model(&domain.DataStorageModel{})
	if biz != "" {
		query = query.Where("biz = ?", biz)
	}
	var models []domain.DataStorageModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *storageService) UpdateStorage(id int64, updates map[string]any) (domain.DataStorageModel, error) {
	if len(updates) == 0 {
		return s.GetStorage(id)
	}
	updates["updatedAt"] = time.Now().UnixMilli()

	session := sqllite.GetSession()
	if err := session.Model(&domain.DataStorageModel{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return domain.DataStorageModel{}, err
	}
	return s.GetStorage(id)
}

func (s *storageService) DeleteStorage(id int64) error {
	session := sqllite.GetSession()
	return session.Delete(&domain.DataStorageModel{}, id).Error
}

func (s *storageService) DeleteStoragesByBiz(biz string) error {
	session := sqllite.GetSession()
	return session.Where("biz = ?", biz).Delete(&domain.DataStorageModel{}).Error
}
