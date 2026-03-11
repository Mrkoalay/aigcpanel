package service

import (
	"encoding/json"
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

type StorageFilters struct {
	Biz  string
	Page int `form:"page"`
	Size int `form:"size"`
}

type SoundMediaResp struct {
	ID        int64              `json:"ID"`
	CreatedAt int64              `json:"CreatedAt"`
	UpdatedAt int64              `json:"UpdatedAt"`
	Sort      int64              `json:"Sort"`
	Biz       string             `json:"Biz"`
	Title     string             `json:"Title"`
	Content   SoundPromptContent `json:"content"`
}
type SoundPromptContent struct {
	AsrStatus  string `json:"asrStatus"`
	PromptText string `json:"promptText"`
	URL        string `json:"url"`
}

func (s *storageService) ListStorages(req StorageFilters) ([]SoundMediaResp, error) {
	session := sqllite.GetSession()
	query := session.Model(&domain.DataStorageModel{})
	if req.Biz != "" {
		query = query.Where("biz = ?", req.Biz)
	}
	if req.Page > 0 {
		limit := req.Size
		offset := (req.Page - 1) * req.Size
		query = query.Limit(limit).Offset(offset)
	}
	var models []domain.DataStorageModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	var resp []SoundMediaResp

	for _, item := range models {
		var content SoundPromptContent
		if item.Content != "" {
			_ = json.Unmarshal([]byte(item.Content), &content)
		}

		resp = append(resp, SoundMediaResp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Sort:      item.Sort,
			Biz:       item.Biz,
			Title:     item.Title,
			Content:   content,
		})
	}

	return resp, nil
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
