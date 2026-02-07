package sqllite

import (
	"aigcpanel/go/internal/component/log"
	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/utils"
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStore struct {
	db *gorm.DB
}

func Init() {

	_, taskStoreErr := NewSQLiteStore(utils.SQLiteFile)
	if taskStoreErr != nil {
		log.Logger.Error(taskStoreErr.Error())
	}

}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return nil
}

func (s *SQLiteStore) migrate() error {
	return s.db.AutoMigrate(&appTaskModel{})
}

type TaskFilters struct {
	Biz    string
	Status []string
	Type   *int
}

func (s *SQLiteStore) CreateTask(task domain.AppTask) (domain.AppTask, error) {
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
	model := appTaskModelFromDomain(task)
	if err := s.db.Create(&model).Error; err != nil {
		return domain.AppTask{}, err
	}
	return model.toDomain(), nil
}

func (s *SQLiteStore) GetTask(id int64) (domain.AppTask, error) {
	var model appTaskModel
	if err := s.db.First(&model, id).Error; err != nil {
		return domain.AppTask{}, err
	}
	return model.toDomain(), nil
}

func (s *SQLiteStore) ListTasks(filters TaskFilters) ([]domain.AppTask, error) {
	query := s.db.Model(&appTaskModel{})
	if filters.Biz != "" {
		query = query.Where("biz = ?", filters.Biz)
	}
	if len(filters.Status) > 0 {
		statuses := make([]string, 0, len(filters.Status))
		for _, status := range filters.Status {
			if status == "" {
				continue
			}
			statuses = append(statuses, status)
		}
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
	}
	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}
	var models []appTaskModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	tasks := make([]domain.AppTask, 0, len(models))
	for _, model := range models {
		tasks = append(tasks, model.toDomain())
	}
	return tasks, nil
}

func (s *SQLiteStore) UpdateTask(id int64, updates map[string]any) (domain.AppTask, error) {
	if len(updates) == 0 {
		return s.GetTask(id)
	}
	updates["updatedAt"] = time.Now().UnixMilli()
	if err := s.db.Model(&appTaskModel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return domain.AppTask{}, err
	}
	return s.GetTask(id)
}

func (s *SQLiteStore) DeleteTask(id int64) error {
	return s.db.Delete(&appTaskModel{}, id).Error
}

type appTaskModel struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt     int64  `gorm:"column:createdAt;not null"`
	UpdatedAt     int64  `gorm:"column:updatedAt;not null"`
	Biz           string `gorm:"column:biz;index:idx_data_task_biz"`
	Type          int    `gorm:"column:type;default:1;index:idx_data_task_type"`
	Title         string `gorm:"column:title"`
	Status        string `gorm:"column:status;index:idx_data_task_status"`
	StatusMsg     string `gorm:"column:statusMsg"`
	StartTime     int64  `gorm:"column:startTime"`
	EndTime       int64  `gorm:"column:endTime"`
	ServerName    string `gorm:"column:serverName"`
	ServerTitle   string `gorm:"column:serverTitle"`
	ServerVersion string `gorm:"column:serverVersion"`
	Param         string `gorm:"column:param"`
	JobResult     string `gorm:"column:jobResult"`
	ModelConfig   string `gorm:"column:modelConfig"`
	Result        string `gorm:"column:result"`
}

func (appTaskModel) TableName() string {
	return "data_task"
}

func appTaskModelFromDomain(task domain.AppTask) appTaskModel {
	return appTaskModel{
		ID:            task.ID,
		Biz:           task.Biz,
		Type:          task.Type,
		Title:         task.Title,
		Status:        task.Status,
		StatusMsg:     task.StatusMsg,
		StartTime:     task.StartTime,
		EndTime:       task.EndTime,
		ServerName:    task.ServerName,
		ServerTitle:   task.ServerTitle,
		ServerVersion: task.ServerVersion,
		Param:         task.Param,
		JobResult:     task.JobResult,
		ModelConfig:   task.ModelConfig,
		Result:        task.Result,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}

func (m appTaskModel) toDomain() domain.AppTask {
	return domain.AppTask{
		ID:            m.ID,
		Biz:           m.Biz,
		Type:          m.Type,
		Title:         m.Title,
		Status:        m.Status,
		StatusMsg:     m.StatusMsg,
		StartTime:     m.StartTime,
		EndTime:       m.EndTime,
		ServerName:    m.ServerName,
		ServerTitle:   m.ServerTitle,
		ServerVersion: m.ServerVersion,
		Param:         m.Param,
		JobResult:     m.JobResult,
		ModelConfig:   m.ModelConfig,
		Result:        m.Result,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
