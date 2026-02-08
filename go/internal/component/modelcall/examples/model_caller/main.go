package main

import (
	"encoding/json"
	"fmt"
	"xiacutai-server/internal/component/modelcall"
	"xiacutai-server/internal/component/modelcall/easyserver"

	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// main 主函数，演示如何使用 aigcpanel-go 包
func main() {
	// 检查命令行参数
	if len(os.Args) < 4 {
		fmt.Println("使用方法: ./model_caller <模型路径> <配置文件路径> <功能名称> [参数...]")
		fmt.Println("示例: ./model_caller /path/to/model /path/to/config.json soundTts text=你好，世界 speaker=中文女 speed=1.0")
		fmt.Println("\n支持的功能:")
		fmt.Println("  soundTts   - 语音合成")
		fmt.Println("  soundClone - 语音克隆")
		fmt.Println("  videoGen   - 视频生成")
		fmt.Println("  asr        - 语音识别")
		return
	}

	// 获取命令行参数
	modelPath := os.Args[1]
	configPath := os.Args[2]
	functionName := os.Args[3]
	params := os.Args[4:]

	// 检查模型路径是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Fatalf("模型路径不存在: %s", modelPath)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("配置文件不存在: %s", configPath)
	}

	// 调用指定功能
	result, err := callModelFunction(modelPath, configPath, functionName, params)
	if err != nil {
		log.Fatalf("调用模型功能失败: %v", err)
	}

	// 输出结果
	printCallResult(result)
}

// callModelFunction 调用模型功能
func callModelFunction(modelPath, configPath, functionName string, params []string) (*easyserver.TaskResult, error) {
	// 加载配置
	config, err := modelcall.LoadConfigFromJSON(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 创建 EasyServer 实例
	server := createCallerEasyServer(modelPath, config)

	// 启动服务器
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	// 解析参数
	paramMap := make(map[string]interface{})
	for _, param := range params {
		parts := parseCallParam(param)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// 尝试转换为数字
			if floatVal, parseErr := parseFloat(value); parseErr == nil {
				paramMap[key] = floatVal
			} else {
				paramMap[key] = value
			}
		}
	}

	// 构建功能数据
	functionData := &easyserver.ServerFunctionDataType{
		ID:    fmt.Sprintf("call-%d", time.Now().Unix()),
		Param: paramMap,
	}

	// 根据功能类型设置特定参数
	switch functionName {
	case "soundTts":
		if text, ok := paramMap["text"].(string); ok {
			functionData.Text = text
		}
	case "soundClone":
		if text, ok := paramMap["text"].(string); ok {
			functionData.Text = text
		}
		if promptAudio, ok := paramMap["promptAudio"].(string); ok {
			functionData.PromptAudio = promptAudio
		}
		if promptText, ok := paramMap["promptText"].(string); ok {
			functionData.PromptText = promptText
		}
	case "videoGen":
		if video, ok := paramMap["video"].(string); ok {
			functionData.Video = video
		}
		if audio, ok := paramMap["audio"].(string); ok {
			functionData.Audio = audio
		}
	case "asr":
		if audio, ok := paramMap["audio"].(string); ok {
			functionData.Audio = audio
		}
	}

	// 调用功能
	var result *easyserver.TaskResult

	switch functionName {
	case "soundTts":
		result, err = server.SoundTts(*functionData)
	case "soundClone":
		result, err = server.SoundClone(*functionData)
	case "videoGen":
		result, err = server.VideoGen(*functionData)
	case "asr":
		result, err = server.Asr(*functionData)
	default:
		return nil, fmt.Errorf("不支持的功能: %s", functionName)
	}

	if err != nil {
		return nil, fmt.Errorf("调用功能失败: %v", err)
	}

	return result, nil
}

// loadConfigFromJSON 从 JSON 文件加载配置
func loadConfigFromJSON(configPath string) (*easyserver.ServerConfig, error) {
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
	entry := configJSON.Entry
	var entryArgs []string
	var envs []string
	var content string

	if configJSON.EasyServer.Entry != "" {
		entry = configJSON.EasyServer.Entry
		entryArgs = configJSON.EasyServer.EntryArgs
		envs = configJSON.EasyServer.Envs
		content = configJSON.EasyServer.Content
	} else if configJSON.Launcher.Entry != "" {
		entry = configJSON.Launcher.Entry
		entryArgs = configJSON.Launcher.EntryArgs
		envs = configJSON.Launcher.Envs
	}

	// 转换 functions
	functions := make([]easyserver.ServerFunction, len(configJSON.Functions))
	for i, f := range configJSON.Functions {
		functions[i] = easyserver.ServerFunction(f)
	}

	// 转换 settings
	settings := make([]struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		Default     string `json:"default"`
		Placeholder string `json:"placeholder"`
	}, len(configJSON.Settings))
	for i, s := range configJSON.Settings {
		settings[i] = s
	}

	// 构建 ServerConfig
	config := &easyserver.ServerConfig{
		Name:          configJSON.Name,
		Version:       configJSON.Version,
		Title:         configJSON.Title,
		Description:   configJSON.Description,
		PlatformName:  configJSON.PlatformName,
		PlatformArch:  configJSON.PlatformArch,
		ServerRequire: configJSON.ServerRequire,
		Entry:         entry,
		Functions:     functions,
		Settings:      settings,
	}

	// 设置 EasyServer 配置
	if entry != configJSON.Entry || len(entryArgs) > 0 || len(envs) > 0 || content != "" {
		config.EasyServer = &struct {
			Entry     string   `json:"entry"`
			EntryArgs []string `json:"entryArgs"`
			Envs      []string `json:"envs"`
			Content   string   `json:"content"`
		}{
			Entry:     entry,
			EntryArgs: entryArgs,
			Envs:      envs,
			Content:   content,
		}
	}

	return config, nil
}

// createCallerEasyServer 创建用于调用的 EasyServer 实例
func createCallerEasyServer(modelPath string, config *easyserver.ServerConfig) *easyserver.EasyServer {
	return &easyserver.EasyServer{
		ServerConfig: *config,
		ServerInfo: &easyserver.ServerInfo{
			LocalPath: modelPath,
			Name:      config.Name,
			Version:   config.Version,
			Config:    *config,
		},
	}
}

// parseCallParam 解析调用参数
func parseCallParam(param string) []string {
	parts := strings.SplitN(param, "=", 2)
	if len(parts) != 2 {
		return []string{param}
	}
	return parts
}

// parseFloat 尝试将字符串转换为浮点数
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// printCallResult 打印调用结果
func printCallResult(result *easyserver.TaskResult) {
	fmt.Printf("调用结果:\n")
	fmt.Printf("  状态码: %d\n", result.Code)
	fmt.Printf("  消息: %s\n", result.Msg)

	if result.Data != nil {
		fmt.Printf("  数据: %+v\n", result.Data)
	}

	fmt.Println("\n调用完成!")
}
