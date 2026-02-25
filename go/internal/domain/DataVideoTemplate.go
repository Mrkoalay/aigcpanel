package domain

// DataVideoTemplateModel 数字人形象模板
type DataVideoTemplateModel struct {
	ID    int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name  string `gorm:"column:name" json:"name"`
	Video string `gorm:"column:video" json:"video"`
	Info  string `gorm:"column:info" json:"info"`
}

func (DataVideoTemplateModel) TableName() string {
	return "data_video_template"
}
