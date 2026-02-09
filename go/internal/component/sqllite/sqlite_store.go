package sqllite

import (
	"errors"
	"time"
	"xiacutai-server/internal/component/log"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	// ⭐ 关键：纯Go sqlite 驱动（无CGO）
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *gorm.DB
}

var DB *SQLiteStore

func Init() {
	store, err := NewSQLiteStore(utils.SQLiteFile)
	if err != nil {
		log.Logger.Error(err.Error())
		panic(err)
	}
	DB = store
}
func GetSession() *gorm.DB {

	return DB.db
}
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {

	// ⭐ 必须用 DriverName=sqlite
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        dsn,
	}, &gorm.Config{})

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
	return s.db.AutoMigrate(
		&domain.DataTaskModel{},
		&domain.DataStorageModel{}, // ⭐ 新表
	)
}

type TaskFilters struct {
	Biz    string
	Status []string
	Type   *int

	Page int `form:"page"`
	Size int `form:"size"`
}

func (s *SQLiteStore) CreateTask(task domain.DataTaskModel) (domain.DataTaskModel, error) {
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

	if err := s.db.Create(&task).Error; err != nil {
		return domain.DataTaskModel{}, err
	}

	return task, nil
}

func (s *SQLiteStore) GetTask(id int64) (domain.DataTaskModel, error) {
	var model domain.DataTaskModel
	if err := s.db.First(&model, id).Error; err != nil {
		return domain.DataTaskModel{}, err
	}
	return model, nil
}

func (s *SQLiteStore) ListTasks(filters TaskFilters) ([]domain.DataTaskModel, error) {

	query := s.db.Model(&domain.DataTaskModel{})

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

	var models []domain.DataTaskModel
	if err := query.Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

func (s *SQLiteStore) UpdateTask(id int64, updates map[string]any) (domain.DataTaskModel, error) {

	if len(updates) == 0 {
		return s.GetTask(id)
	}

	updates["updatedAt"] = time.Now().UnixMilli()

	if err := s.db.Model(&domain.DataTaskModel{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return domain.DataTaskModel{}, err
	}

	return s.GetTask(id)
}

func (s *SQLiteStore) DeleteTask(id int64) error {
	return s.db.Delete(&domain.DataTaskModel{}, id).Error
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
