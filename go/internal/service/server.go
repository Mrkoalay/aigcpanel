package service

import (
	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/errs"
	"aigcpanel/go/internal/store"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type server struct{ store *store.JSONStore }

var Server = new(server)

func (s *server) ServerAdd(configPath string) (domain.LocalModelConfigInfo, error) {
	if strings.TrimSpace(configPath) == "" {
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}
	if filepath.Base(configPath) != "config.json" {
		return domain.LocalModelConfigInfo{}, errs.New("参数异常")
	}
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return domain.LocalModelConfigInfo{}, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(buf, &cfg); err != nil {
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
	}, nil
}

func firstNonEmpty(v string, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
