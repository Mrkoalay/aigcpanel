package service

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/log"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/utils"

	"gorm.io/gorm"
)

type model struct{}

var Model = new(model)

////////////////////////////////////////////////////////////
// 公共解析
////////////////////////////////////////////////////////////

func parseConfigToInfo(cfg map[string]any, parent string) domain.LocalModelConfigInfo {

	toString := func(v any) string {
		if x, ok := v.(string); ok {
			return x
		}
		return ""
	}

	functions := make([]string, 0)
	if arr, ok := cfg["functions"].([]any); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				functions = append(functions, str)
			}
		}
	}

	settings := make([]any, 0)
	if arr, ok := cfg["settings"].([]any); ok {
		settings = arr
	}

	setting := map[string]any{}
	if m, ok := cfg["setting"].(map[string]any); ok {
		setting = m
	}

	return domain.LocalModelConfigInfo{
		Type:              "LOCAL_DIR",
		Name:              toString(cfg["name"]),
		Version:           toString(cfg["version"]),
		ServerRequire:     firstNonEmpty(toString(cfg["serverRequire"]), "*"),
		Title:             toString(cfg["title"]),
		Description:       toString(cfg["description"]),
		DeviceDescription: toString(cfg["deviceDescription"]),
		Path:              parent,
		PlatformName:      toString(cfg["platformName"]),
		PlatformArch:      toString(cfg["platformArch"]),
		Entry:             toString(cfg["entry"]),
		Functions:         functions,
		Settings:          settings,
		Setting:           setting,
		Config:            cfg,
		Status:            "3",
	}
}

func firstNonEmpty(v string, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

////////////////////////////////////////////////////////////
// ModelAdd
////////////////////////////////////////////////////////////

func (s *model) ModelAdd(configPath string) (domain.LocalModelConfigInfo, error) {

	log.Info("开始添加模型", zap.String("config", configPath))

	if strings.TrimSpace(configPath) == "" {
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}

	if filepath.Base(configPath) != "config.json" {
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}

	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("读取配置失败", zap.Error(err))
		return domain.LocalModelConfigInfo{}, err
	}

	var cfg map[string]any
	if err := json.Unmarshal(buf, &cfg); err != nil {
		log.Error("json解析失败", zap.Error(err))
		return domain.LocalModelConfigInfo{}, err
	}

	parent := filepath.Dir(configPath)
	info := parseConfigToInfo(cfg, parent)

	log.Info("准备注册模型",
		zap.String("name", info.Name),
		zap.String("version", info.Version),
		zap.String("path", info.Path),
	)

	if err := registerModel(info); err != nil {
		return domain.LocalModelConfigInfo{}, err
	}

	return info, nil
}

////////////////////////////////////////////////////////////
// ModelList
////////////////////////////////////////////////////////////

func (s *model) ModelList(functionName string) ([]domain.LocalModelConfigInfo, error) {

	reg, err := loadRegistry()
	if err != nil {
		return nil, err
	}

	list := make([]domain.LocalModelConfigInfo, 0)

	for _, r := range reg.Records {

		cfg, ok := resolveModelConfig(r)
		if !ok {
			continue
		}

		modelConfigInfo := parseConfigToInfo(cfg, r.LocalPath)
		functions := modelConfigInfo.Functions
		if functionName != "" {
			if functions != nil && utils.Contains(functions, functionName) {
				list = append(list, modelConfigInfo)
			}
			continue
		}

		list = append(list, modelConfigInfo)
	}

	log.Info("返回模型列表", zap.Int("count", len(list)))

	return list, nil
}

func (s *model) ModelUpdateSetting(name, version string, newSetting map[string]any) error {

	key := name + "|" + version

	log.Info("更新模型设置",
		zap.String("key", key),
		zap.Any("setting", newSetting),
	)

	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	for i, r := range reg.Records {

		if r.Key != key {
			continue
		}

		// ---------- 校验参数合法性 ----------
		validKeys := map[string]bool{}
		for _, s := range r.Settings {
			if m, ok := s.(map[string]any); ok {
				if n, ok := m["name"].(string); ok {
					validKeys[n] = true
				}
			}
		}

		for k := range newSetting {
			if !validKeys[k] {
				log.Warn("非法设置字段",
					zap.String("key", key),
					zap.String("field", k),
				)
				return errs.New("存在未定义的设置字段: " + k)
			}
		}

		// ---------- 合并设置 ----------
		if r.Setting == nil {
			r.Setting = map[string]any{}
		}

		for k, v := range newSetting {
			r.Setting[k] = v
		}

		reg.Records[i].Setting = r.Setting

		log.Info("模型设置更新成功", zap.String("key", key))
		return saveRegistry(reg)
	}

	return errs.New("模型不存在")
}
func (s *model) ModelUpdateStatus(key string, status int) error {
	db := sqllite.GetSession()
	return db.Model(&domain.LocalModelRegistryModel{}).
		Where("key = ?", key).
		Updates(map[string]any{
			"status": status,
		}).Error

}

func (s *model) ModelDelete(name string, version string) error {

	key := name + "|" + version

	log.Info("请求删除模型", zap.String("key", key))

	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	index := -1

	for i, r := range reg.Records {
		if r.Key == key {
			index = i
			break
		}
	}

	if index == -1 {
		log.Warn("删除失败，模型不存在", zap.String("key", key))
		return errs.New("模型不存在")
	}

	removed := reg.Records[index]

	// 从数组移除
	reg.Records = append(reg.Records[:index], reg.Records[index+1:]...)

	log.Info("模型已从注册表移除",
		zap.String("key", removed.Key),
		zap.String("path", removed.LocalPath),
	)

	return saveRegistry(reg)
}
func (s *model) Get(modelKey string) (*domain.LocalModelConfigInfo, error) {
	localModelConfigInfo := &domain.LocalModelConfigInfo{}
	reg, err := loadRegistry()
	if err != nil {
		return localModelConfigInfo, err
	}

	index := -1

	for i, r := range reg.Records {
		if r.Key == modelKey {
			index = i

			break
		}
	}

	if index == -1 {
		log.Warn("模型不存在", zap.String("key", modelKey))
		return localModelConfigInfo, errs.New("模型不存在")
	}

	record := reg.Records[index]
	cfg, ok := resolveModelConfig(record)
	if !ok {
		return localModelConfigInfo, errs.New("模型配置损坏")
	}
	modelConfigInfo := parseConfigToInfo(cfg, record.LocalPath)

	return &modelConfigInfo, nil
}

func resolveModelConfig(r ModelRecord) (map[string]any, bool) {
	if len(r.Config) > 0 {
		return r.Config, true
	}

	configPath := filepath.Join(r.LocalPath, "config.json")
	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Warn("模型配置丢失", zap.String("path", configPath))
		return nil, false
	}

	var cfg map[string]any
	if err := json.Unmarshal(buf, &cfg); err != nil {
		log.Warn("模型配置损坏", zap.String("path", configPath))
		return nil, false
	}

	return cfg, true
}
func (s *model) GetByDB(modelKey string) (*domain.LocalModelRegistryModel, error) {
	LocalModelRegistryModel := &domain.LocalModelRegistryModel{}
	if err := sqllite.GetSession().Where("key=?", modelKey).Order("id ASC").Find(&LocalModelRegistryModel).Error; err != nil {
		return nil, err
	}

	return LocalModelRegistryModel, nil
}

type ModelRecord struct {
	Key       string         `json:"key"`
	Name      string         `json:"name"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Version   string         `json:"version"`
	Type      string         `json:"type"`
	AutoStart bool           `json:"autoStart"`
	Functions []string       `json:"functions"`
	LocalPath string         `json:"localPath"`
	Settings  []any          `json:"settings"`
	Setting   map[string]any `json:"setting"`
	Config    map[string]any `json:"config"`
}

type ModelRegistry struct {
	Records []ModelRecord `json:"records"`
}

////////////////////////////////////////////////////////////
// Registry 操作
////////////////////////////////////////////////////////////

func loadRegistry() (*ModelRegistry, error) {
	rows := make([]domain.LocalModelRegistryModel, 0)
	if err := sqllite.GetSession().Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	records := make([]ModelRecord, 0, len(rows))
	for _, row := range rows {
		r, err := convertDBToRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return &ModelRegistry{Records: records}, nil
}

func saveRegistry(reg *ModelRegistry) error {
	return sqllite.GetSession().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&domain.LocalModelRegistryModel{}).Error; err != nil {
			return err
		}

		for _, r := range reg.Records {
			row, err := convertRecordToDB(r)
			if err != nil {
				return err
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func convertRecordToDB(r ModelRecord) (domain.LocalModelRegistryModel, error) {
	functionsRaw, err := json.Marshal(r.Functions)
	if err != nil {
		return domain.LocalModelRegistryModel{}, err
	}
	settingsRaw, err := json.Marshal(r.Settings)
	if err != nil {
		return domain.LocalModelRegistryModel{}, err
	}
	settingRaw, err := json.Marshal(r.Setting)
	if err != nil {
		return domain.LocalModelRegistryModel{}, err
	}
	configRaw, err := json.Marshal(r.Config)
	if err != nil {
		return domain.LocalModelRegistryModel{}, err
	}

	return domain.LocalModelRegistryModel{
		Key:       r.Key,
		Name:      r.Name,
		Title:     r.Title,
		Version:   r.Version,
		Type:      r.Type,
		Status:    r.Status,
		AutoStart: r.AutoStart,
		Functions: string(functionsRaw),
		LocalPath: r.LocalPath,
		Settings:  string(settingsRaw),
		Setting:   string(settingRaw),
		Config:    string(configRaw),
	}, nil
}

func convertDBToRecord(row domain.LocalModelRegistryModel) (ModelRecord, error) {
	r := ModelRecord{
		Key:       row.Key,
		Name:      row.Name,
		Title:     row.Title,
		Version:   row.Version,
		Type:      row.Type,
		AutoStart: row.AutoStart,
		LocalPath: row.LocalPath,
	}

	if row.Functions != "" {
		if err := json.Unmarshal([]byte(row.Functions), &r.Functions); err != nil {
			return ModelRecord{}, err
		}
	}

	if row.Settings != "" {
		if err := json.Unmarshal([]byte(row.Settings), &r.Settings); err != nil {
			return ModelRecord{}, err
		}
	}

	if row.Setting != "" {
		if err := json.Unmarshal([]byte(row.Setting), &r.Setting); err != nil {
			return ModelRecord{}, err
		}
	}

	if row.Config != "" {
		if err := json.Unmarshal([]byte(row.Config), &r.Config); err != nil {
			return ModelRecord{}, err
		}
	}

	return r, nil
}

////////////////////////////////////////////////////////////
// 清理失效模型
////////////////////////////////////////////////////////////

func cleanInvalid(reg *ModelRegistry) {

	valid := make([]ModelRecord, 0)

	for _, r := range reg.Records {
		if _, err := os.Stat(r.LocalPath); err == nil {
			valid = append(valid, r)
		} else {
			log.Warn("移除失效模型",
				zap.String("key", r.Key),
				zap.String("path", r.LocalPath),
			)
		}
	}

	reg.Records = valid
}

////////////////////////////////////////////////////////////
// 注册模型
////////////////////////////////////////////////////////////

func registerModel(info domain.LocalModelConfigInfo) error {
	key := info.Name + "|" + info.Version
	path := filepath.Clean(info.Path)

	db := sqllite.GetSession()

	// 只查这一条，不再 load 全表
	var row domain.LocalModelRegistryModel
	err := db.Where("key = ?", key).First(&row).Error

	if err == nil {
		// ---- 保留你原来的判断逻辑 ----
		if filepath.Clean(row.LocalPath) == path {
			log.Warn("重复注册模型", zap.String("key", key))
			return errs.New("模型已存在")
		}

		log.Warn("模型路径更新",
			zap.String("old", row.LocalPath),
			zap.String("new", path),
		)

		// 你原来只更新 LocalPath + Config（语义保持一致）
		configRaw, jerr := json.Marshal(info.Config)
		if jerr != nil {
			return jerr
		}

		return db.Model(&domain.LocalModelRegistryModel{}).
			Where("key = ?", key).
			Updates(map[string]any{
				"local_path": path,
				"config":     string(configRaw),
			}).Error
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 不存在：新增（只插这一条）
	rec := ModelRecord{
		Key:       key,
		Name:      info.Name,
		Title:     info.Title,
		Version:   info.Version,
		Type:      "localDir",
		AutoStart: true,
		Functions: info.Functions,
		LocalPath: path,
		Settings:  info.Settings,
		Setting:   info.Setting,
		Config:    info.Config,
		Status:    info.Status,
	}

	newRow, err := convertRecordToDB(rec)
	if err != nil {
		return err
	}

	return db.Create(&newRow).Error
}
