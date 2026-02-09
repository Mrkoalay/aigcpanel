package service

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/log"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/utils"
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

		configPath := filepath.Join(r.LocalPath, "config.json")

		buf, err := os.ReadFile(configPath)
		if err != nil {
			log.Warn("模型配置丢失", zap.String("path", configPath))
			continue
		}

		var cfg map[string]any
		if err := json.Unmarshal(buf, &cfg); err != nil {
			log.Warn("模型配置损坏", zap.String("path", configPath))
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
	configPath := filepath.Join(record.LocalPath, "config.json")

	buf, err := os.ReadFile(configPath)
	if err != nil {
		return localModelConfigInfo, errs.New("模型配置丢失")
	}
	var cfg map[string]any
	if err := json.Unmarshal(buf, &cfg); err != nil {
		return localModelConfigInfo, errs.New("模型配置损坏")

	}
	modelConfigInfo := parseConfigToInfo(cfg, record.LocalPath)

	return &modelConfigInfo, nil
}

type ModelRecord struct {
	Key       string         `json:"key"`
	Name      string         `json:"name"`
	Title     string         `json:"title"`
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

	if _, err := os.Stat(utils.RegistryFile); os.IsNotExist(err) {
		return &ModelRegistry{Records: []ModelRecord{}}, nil
	}

	buf, err := os.ReadFile(utils.RegistryFile)
	if err != nil {
		return nil, err
	}

	var reg ModelRegistry
	if err := json.Unmarshal(buf, &reg); err != nil {
		return nil, err
	}

	return &reg, nil
}

func saveRegistry(reg *ModelRegistry) error {

	os.MkdirAll(filepath.Dir(utils.RegistryFile), 0755)

	f, err := os.Create(utils.RegistryFile)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(reg)
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

	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	cleanInvalid(reg)

	for i, r := range reg.Records {

		if r.Key == key {

			if filepath.Clean(r.LocalPath) == path {
				log.Warn("重复注册模型", zap.String("key", key))
				return errs.New("模型已存在")
			}

			log.Warn("模型路径更新",
				zap.String("old", r.LocalPath),
				zap.String("new", path),
			)

			reg.Records[i].LocalPath = path
			reg.Records[i].Config = info.Config
			return saveRegistry(reg)
		}
	}

	reg.Records = append(reg.Records, ModelRecord{
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
	})

	return saveRegistry(reg)
}

////////////////////////////////////////////////////////////
// 启动时调用
////////////////////////////////////////////////////////////

func InitModelRegistry() {
	reg, err := loadRegistry()
	if err != nil {
		return
	}
	cleanInvalid(reg)
	saveRegistry(reg)
}
