// Package easyserver 提供了用于管理和运行本地 AI 模型服务器的核心功能
// 支持语音合成、语音克隆、视频生成、语音识别等多种 AI 功能
package easyserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ServerStatus 表示服务器的运行状态
type ServerStatus string

const (
	ServerStopped  ServerStatus = "stopped"  // 服务器已停止
	ServerStarting ServerStatus = "starting" // 服务器正在启动
	ServerRunning  ServerStatus = "running"  // 服务器正在运行
	ServerStopping ServerStatus = "stopping" // 服务器正在停止
	ServerError    ServerStatus = "error"    // 服务器出现错误
)

// ServerFunction 表示服务器支持的功能类型
type ServerFunction string

const (
	FunctionVideoGen   ServerFunction = "videoGen"   // 视频生成功能
	FunctionSoundTts   ServerFunction = "soundTts"   // 语音合成功能
	FunctionSoundClone ServerFunction = "soundClone" // 语音克隆功能
	FunctionAsr        ServerFunction = "asr"        // 语音识别功能
)

// ServerConfig 表示服务器的配置信息
type ServerConfig struct {
	Name          string           `json:"name"`          // 服务器名称
	Version       string           `json:"version"`       // 服务器版本
	Title         string           `json:"title"`         // 服务器标题
	Description   string           `json:"description"`   // 服务器描述
	PlatformName  string           `json:"platformName"`  // 平台名称
	PlatformArch  string           `json:"platformArch"`  // 平台架构
	ServerRequire string           `json:"serverRequire"` // 服务器要求
	Entry         string           `json:"entry"`         // 入口点
	Functions     []ServerFunction `json:"functions"`     // 支持的功能列表
	EasyServer    *struct {
		Entry     string   `json:"entry"`     // EasyServer 入口点
		EntryArgs []string `json:"entryArgs"` // EasyServer 入口参数
		Envs      []string `json:"envs"`      // 环境变量
		Content   string   `json:"content"`   // 内容
	} `json:"easyServer,omitempty"` // EasyServer 特定配置
	Settings []struct {
		Name        string `json:"name"`        // 设置名称
		Type        string `json:"type"`        // 设置类型
		Title       string `json:"title"`       // 设置标题
		Default     string `json:"default"`     // 默认值
		Placeholder string `json:"placeholder"` // 占位符
	} `json:"settings"` // 设置列表
}

// ServerInfo 表示服务器的运行时信息
type ServerInfo struct {
	LocalPath        string                 `json:"localPath"`        // 本地路径
	Name             string                 `json:"name"`             // 服务器名称
	Version          string                 `json:"version"`          // 服务器版本
	Setting          map[string]interface{} `json:"setting"`          // 设置
	LogFile          string                 `json:"logFile"`          // 日志文件路径
	EventChannelName string                 `json:"eventChannelName"` // 事件通道名称
	Config           ServerConfig           `json:"config"`           // 服务器配置
}

// ServerFunctionDataType 表示服务器功能的数据类型
type ServerFunctionDataType struct {
	ID          string                 `json:"id"`                    // 任务ID
	Result      map[string]interface{} `json:"result"`                // 结果数据
	Param       map[string]interface{} `json:"param,omitempty"`       // 参数
	Text        string                 `json:"text,omitempty"`        // 文本内容
	Video       string                 `json:"video,omitempty"`       // 视频文件路径
	Audio       string                 `json:"audio,omitempty"`       // 音频文件路径
	PromptAudio string                 `json:"promptAudio,omitempty"` // 提示音频
	PromptText  string                 `json:"promptText,omitempty"`  // 提示文本
}

// LauncherResultType 表示启动器的结果类型
type LauncherResultType struct {
	Result  map[string]interface{} `json:"result"`  // 结果数据
	EndTime *int64                 `json:"endTime"` // 结束时间
}

// TaskResult 表示任务的执行结果
type TaskResult struct {
	Code int         `json:"code"` // 状态码
	Msg  string      `json:"msg"`  // 消息
	Data interface{} `json:"data"` // 数据
}

// TaskResultData represents the data structure of a task result
type TaskResultData struct {
	Type  string      `json:"type"`
	Start int64       `json:"start"`
	End   int64       `json:"end"`
	Data  interface{} `json:"data"`
}

// ExtractResultFromLogs extracts result from logs
// This function mimics the behavior of extractResultFromLogs in the Electron project
func ExtractResultFromLogs(dataID string, logs string) (map[string]interface{}, error) {
	var result map[string]interface{}

	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		// Looking for pattern: AigcPanelRunResult[dataID][base64EncodedResult]
		pattern := fmt.Sprintf("AigcPanelRunResult[%s][", dataID)
		startIdx := strings.Index(line, pattern)
		if startIdx != -1 {
			// Find the closing bracket
			startIdx += len(pattern)
			endIdx := strings.Index(line[startIdx:], "]")
			if endIdx != -1 {
				encodedResult := line[startIdx : startIdx+endIdx]
				// Decode base64
				decodedBytes, err := base64.StdEncoding.DecodeString(encodedResult)
				if err != nil {
					return nil, fmt.Errorf("failed to decode base64 result: %v", err)
				}

				// Parse JSON
				err = json.Unmarshal(decodedBytes, &result)
				if err != nil {
					return nil, fmt.Errorf("failed to parse JSON result: %v", err)
				}
				return result, nil
			}
		}
	}

	return result, nil
}

// Sleep is a utility function to sleep for specified milliseconds
func Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
