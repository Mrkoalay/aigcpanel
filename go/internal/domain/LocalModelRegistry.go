package domain

type LocalModelRegistryModel struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Key       string `gorm:"column:key;type:text;uniqueIndex" json:"key"`
	Status    string `gorm:"column:status;type:text;" json:"status"` // 你可沿用：1未下载、2下载中、3已下载、4依赖安装中、5完成、-1失败
	Name      string `gorm:"column:name;type:text" json:"name"`
	Title     string `gorm:"column:title;type:text" json:"title"`
	URL       string `gorm:"column:url;type:text" json:"url"`
	Version   string `gorm:"column:version;type:text" json:"version"`
	Type      string `gorm:"column:type;type:text" json:"type"`
	AutoStart bool   `gorm:"column:autoStart" json:"autoStart"`
	Functions string `gorm:"column:functions;type:text" json:"functions"`
	LocalPath string `gorm:"column:localPath;type:text" json:"localPath"`
	Settings  string `gorm:"column:settings;type:text" json:"settings"`
	Setting   string `gorm:"column:setting;type:text" json:"setting"`
	Config    string `gorm:"column:config;type:text" json:"config"`
}

func (LocalModelRegistryModel) TableName() string {
	return "local_model_registry"
}
