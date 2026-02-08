// Package aigcpanel 提供了用于管理本地 AI 模型的工具和服务
// 包括语音合成、语音克隆、视频生成、语音识别等功能
package modelcall

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"xiacutai-server/internal/component/modelcall/easyserver"
)

// LoadConfigFromJSON 从 JSON 文件加载服务器配置
// 参数:
//   - configPath: 配置文件路径
//
// 返回:
//   - *easyserver.ServerConfig: 服务器配置对象
//   - error: 错误信息
func LoadConfigFromJSON(configPath string) (*easyserver.ServerConfig, error) {
	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 定义一个中间结构体来解析 JSON
	var configJSON struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		ServerRequire string `json:"serverRequire"`
		PlatformName  string `json:"platformName"`
		PlatformArch  string `json:"platformArch"`
		Entry         string `json:"entry"`
		EasyServer    struct {
			Entry     string   `json:"entry"`
			EntryArgs []string `json:"entryArgs"`
			Envs      []string `json:"envs"`
			Content   string   `json:"content"`
		} `json:"easyServer"`
		Launcher struct {
			Entry     string   `json:"entry"`
			EntryArgs []string `json:"entryArgs"`
			Envs      []string `json:"envs"`
		} `json:"launcher"`
		Functions []string `json:"functions"`
		Settings  []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Title       string `json:"title"`
			Default     string `json:"default"`
			Placeholder string `json:"placeholder"`
		} `json:"settings"`
	}

	// 解析 JSON
	if err := json.Unmarshal(data, &configJSON); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 确定使用哪个 entry
	entry := configJSON.EasyServer.Entry
	entryArgs := configJSON.EasyServer.EntryArgs
	envs := configJSON.EasyServer.Envs

	// 如果 easyServer.entry 是 "launcher"，则使用 launcher 配置
	if entry == "launcher" {
		entry = configJSON.Launcher.Entry
		entryArgs = configJSON.Launcher.EntryArgs
		envs = configJSON.Launcher.Envs
	}

	// 转换为 ServerConfig 结构体
	config := &easyserver.ServerConfig{
		Name:          configJSON.Name,
		Version:       configJSON.Version,
		Title:         configJSON.Title,
		Description:   configJSON.Description,
		ServerRequire: configJSON.ServerRequire,
		PlatformName:  configJSON.PlatformName,
		PlatformArch:  configJSON.PlatformArch,
		Entry:         configJSON.Entry,
		EasyServer: &struct {
			Entry     string   `json:"entry"`
			EntryArgs []string `json:"entryArgs"`
			Envs      []string `json:"envs"`
			Content   string   `json:"content"`
		}{
			Entry:     entry,
			EntryArgs: entryArgs,
			Envs:      envs,
			Content:   configJSON.EasyServer.Content,
		},
	}

	// 转换 functions
	for _, f := range configJSON.Functions {
		config.Functions = append(config.Functions, easyserver.ServerFunction(f))
	}

	// 复制 settings
	for _, s := range configJSON.Settings {
		config.Settings = append(config.Settings, s)
	}

	return config, nil
}
