package domain

type DataStorageModel struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt int64  `gorm:"column:createdAt;not null"`
	UpdatedAt int64  `gorm:"column:updatedAt;not null"`
	Sort      int64  `gorm:"column:sort;not null"`
	Biz       string `gorm:"column:biz;index:idx_data_storage_biz"`
	Title     string `gorm:"column:title"`
	Content   string `gorm:"column:content"`
}

func (DataStorageModel) TableName() string {
	return "data_storage"
}
