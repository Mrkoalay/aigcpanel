package service

import (
	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/errs"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

const registryFile = "data/models.json"

type model struct{}

var Model = new(model)

////////////////////////////////////////////////////////////
// ModelAdd 主入口
////////////////////////////////////////////////////////////

func (s *model) ModelAdd(configPath string) (domain.LocalModelConfigInfo, error) {

	errs.Info("开始添加模型", zap.String("config", configPath))

	if strings.TrimSpace(configPath) == "" {
		errs.Warn("configPath为空")
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}

	if filepath.Base(configPath) != "config.json" {
		errs.Warn("非法文件名", zap.String("file", configPath))
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}

	buf, err := os.ReadFile(configPath)
	if err != nil {
		errs.Error("读取配置失败", zap.Error(err))
		return domain.LocalModelConfigInfo{}, err
	}

	var cfg map[string]any
	if err := json.Unmarshal(buf, &cfg); err != nil {
		errs.Error("json解析失败", zap.Error(err))
		return domain.LocalModelConfigInfo{}, err
	}

	parent := filepath.Dir(configPath)

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

	info := domain.LocalModelConfigInfo{
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

	errs.Info("准备注册模型",
		zap.String("name", info.Name),
		zap.String("version", info.Version),
		zap.String("path", info.Path),
	)

	if err := registerModel(info); err != nil {
		errs.Error("注册模型失败", zap.Error(err))
		return domain.LocalModelConfigInfo{}, err
	}

	errs.Info("模型添加完成", zap.String("key", info.Name+"|"+info.Version))

	return info, nil
}

func firstNonEmpty(v string, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

////////////////////////////////////////////////////////////
// Registry 数据结构
////////////////////////////////////////////////////////////

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
// 读取 registry
////////////////////////////////////////////////////////////

func loadRegistry() (*ModelRegistry, error) {

	if _, err := os.Stat(registryFile); os.IsNotExist(err) {
		errs.Warn("模型索引不存在，创建新索引")
		return &ModelRegistry{Records: []ModelRecord{}}, nil
	}

	buf, err := os.ReadFile(registryFile)
	if err != nil {
		errs.Error("读取模型索引失败", zap.Error(err))
		return nil, err
	}

	var reg ModelRegistry
	if err := json.Unmarshal(buf, &reg); err != nil {
		errs.Error("解析模型索引失败", zap.Error(err))
		return nil, err
	}

	errs.Info("当前模型数量", zap.Int("count", len(reg.Records)))

	return &reg, nil
}

////////////////////////////////////////////////////////////
// 保存 registry
////////////////////////////////////////////////////////////

func saveRegistry(reg *ModelRegistry) error {

	errs.Info("保存模型索引", zap.Int("count", len(reg.Records)))

	os.MkdirAll(filepath.Dir(registryFile), 0755)

	f, err := os.Create(registryFile)
	if err != nil {
		errs.Error("创建索引文件失败", zap.Error(err))
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(reg); err != nil {
		errs.Error("写入索引失败", zap.Error(err))
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////
// 清理失效模型
////////////////////////////////////////////////////////////

func cleanInvalid(reg *ModelRegistry) {

	before := len(reg.Records)
	valid := make([]ModelRecord, 0)

	for _, r := range reg.Records {
		if _, err := os.Stat(r.LocalPath); err == nil {
			valid = append(valid, r)
		} else {
			errs.Warn("移除失效模型",
				zap.String("key", r.Key),
				zap.String("path", r.LocalPath),
			)
		}
	}

	reg.Records = valid

	if before != len(valid) {
		errs.Info("模型清理完成",
			zap.Int("before", before),
			zap.Int("after", len(valid)),
		)
	}
}

////////////////////////////////////////////////////////////
// 注册模型核心逻辑
////////////////////////////////////////////////////////////

func registerModel(info domain.LocalModelConfigInfo) error {

	key := info.Name + "|" + info.Version
	path := filepath.Clean(info.Path)

	errs.Info("注册模型",
		zap.String("key", key),
		zap.String("path", path),
	)

	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	cleanInvalid(reg)

	for i, r := range reg.Records {

		if r.Key == key {

			if filepath.Clean(r.LocalPath) == path {
				errs.Warn("重复注册模型", zap.String("key", key))
				return errs.New("模型已存在")
			}

			errs.Warn("模型路径变化",
				zap.String("old", r.LocalPath),
				zap.String("new", path),
			)

			reg.Records[i].LocalPath = path
			reg.Records[i].Config = info.Config
			return saveRegistry(reg)
		}
	}

	errs.Info("新增模型记录", zap.String("key", key))

	record := ModelRecord{
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
	}

	reg.Records = append(reg.Records, record)

	return saveRegistry(reg)
}
