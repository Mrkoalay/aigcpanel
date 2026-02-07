package domain

type DataTaskModel struct {
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

func (DataTaskModel) TableName() string {
	return "data_task"
}
