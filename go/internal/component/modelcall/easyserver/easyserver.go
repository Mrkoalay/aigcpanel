package easyserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// EasyServer 表示一个 EasyServer 实例，用于管理本地 AI 模型服务
type EasyServer struct {
	ServerConfig  ServerConfig // 服务器配置
	IsRunning     bool         // 是否正在运行
	ServerInfo    *ServerInfo  // 服务器信息
	ServerRuntime struct {
		StartTime int64 // 启动时间
	}
	controller *exec.Cmd // 控制进程
}

// NewEasyServer 创建一个新的 EasyServer 实例
// 参数:
//   - config: 服务器配置
//
// 返回:
//   - *EasyServer: EasyServer 实例
func NewEasyServer(config ServerConfig) *EasyServer {
	return &EasyServer{
		ServerConfig: config,
		IsRunning:    false,
	}
}

// Init 初始化 EasyServer
// 返回:
//   - error: 错误信息
func (es *EasyServer) Init() error {
	return nil
}

// Config 获取服务器配置
// 返回:
//   - *TaskResult: 配置结果
//   - error: 错误信息
func (es *EasyServer) Config() (*TaskResult, error) {
	return &TaskResult{
		Code: 0,
		Msg:  "success",
		Data: es.ServerConfig,
	}, nil
}

// Start 启动 EasyServer
// 返回:
//   - error: 错误信息
func (es *EasyServer) Start() error {
	es.IsRunning = true
	es.ServerRuntime.StartTime = time.Now().Unix()
	return nil
}

// Ping 检查服务器是否响应
// 返回:
//   - bool: 是否响应
//   - error: 错误信息
func (es *EasyServer) Ping() (bool, error) {
	return es.IsRunning, nil
}

// Stop 停止 EasyServer
// 返回:
//   - error: 错误信息
func (es *EasyServer) Stop() error {
	if es.controller != nil {
		if err := es.controller.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
		es.controller.Wait()
		es.controller = nil
	}
	es.IsRunning = false
	return nil
}

// Cancel 取消当前操作
// 返回:
//   - error: 错误信息
func (es *EasyServer) Cancel() error {
	return es.Stop()
}

// CallFunc 调用服务器功能的通用方法
// 参数:
//   - data: 功能数据
//   - configCalculator: 配置计算函数
//   - resultDataCalculator: 结果数据计算函数
//
// 返回:
//   - *TaskResult: 任务结果
//   - error: 错误信息
func (es *EasyServer) CallFunc(
	data ServerFunctionDataType,
	configCalculator func(ServerFunctionDataType) (map[string]interface{}, error),
	resultDataCalculator func(ServerFunctionDataType, LauncherResultType) (map[string]interface{}, error),
) (*TaskResult, error) {

	_ = map[string]int{
		"timeout": 24 * 3600, // 24 hours default
	}

	resultData := map[string]interface{}{
		"type":  "success",
		"start": 0,
		"end":   0,
		"data":  map[string]interface{}{},
	}

	if !es.IsRunning {
		resultData["type"] = "retry"
		return &TaskResult{
			Code: 0,
			Msg:  "ok",
			Data: resultData,
		}, nil
	}

	es.IsRunning = true
	resultData["start"] = time.Now().Unix()

	defer func() {
		es.IsRunning = false
	}()

	// In a real implementation, you would send "taskRunning" event here
	fmt.Printf("Task running: %s\n", data.ID)

	// Calculate config data
	configData, err := configCalculator(data)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate config: %v", err)
	}

	// Add settings to config data
	configData["setting"] = es.ServerInfo.Setting

	// Prepare config JSON file
	configJsonPath, err := es.prepareConfigJson(configData)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare config JSON: %v", err)
	}

	// Clean up config file when done
	defer os.Remove(configJsonPath)

	// Prepare command
	command := []string{es.ServerConfig.EasyServer.Entry}
	if es.ServerConfig.EasyServer.EntryArgs != nil {
		command = append(command, es.ServerConfig.EasyServer.EntryArgs...)
	}

	// Replace placeholders
	for i := range command {
		command[i] = strings.ReplaceAll(command[i], "${CONFIG}", configJsonPath)
		command[i] = strings.ReplaceAll(command[i], "${ROOT}", es.ServerInfo.LocalPath)
	}

	// Prepare environment variables
	envMap := es.prepareEnvironment()

	// Prepare launcher result
	launcherResult := LauncherResultType{
		Result:  map[string]interface{}{},
		EndTime: nil,
	}

	// Execute command
	err = es.executeCommand(command, envMap, configJsonPath, data.ID, &launcherResult)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	// Calculate end time
	endTime := time.Now().Unix()
	resultData["end"] = endTime

	// Calculate result data
	resultDataFinal, err := resultDataCalculator(data, launcherResult)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate result data: %v", err)
	}

	resultData["data"] = resultDataFinal

	return &TaskResult{
		Code: 0,
		Msg:  "ok",
		Data: resultData,
	}, nil
}

// prepareConfigJson creates a temporary config JSON file
func (es *EasyServer) prepareConfigJson(configData map[string]interface{}) (string, error) {
	// Create a temporary file
	tmpFile, err := ioutil.TempFile("", "easyserver-config-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Marshal config data to JSON
	configBytes, err := json.Marshal(configData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config data: %v", err)
	}

	// Write to file
	_, err = tmpFile.Write(configBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write config data: %v", err)
	}

	return tmpFile.Name(), nil
}

// prepareEnvironment prepares environment variables
func (es *EasyServer) prepareEnvironment() map[string]string {
	envMap := make(map[string]string)

	// Copy system environment variables
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[strings.ToUpper(parts[0])] = parts[1]

		}
	}

	// Add custom paths
	path := envMap["PATH"]
	envMap["PATH"] = fmt.Sprintf("%s;%s;%s/binary", path, es.ServerInfo.LocalPath, es.ServerInfo.LocalPath)

	// 添加 _aienv 方便调试
	venv := es.ServerInfo.LocalPath + string(os.PathSeparator) + "_aienv"
	python := venv + string(os.PathSeparator) + "Scripts"
	torchlib := venv + string(os.PathSeparator) + "Lib" + string(os.PathSeparator) + "site-packages" + string(os.PathSeparator) + "torch" + string(os.PathSeparator) + "lib"
	oldPath := envMap["PATH"]
	newPath := fmt.Sprintf("%s;%s;%s;%s",
		oldPath,
		venv,
		python,
		torchlib,
	)
	envMap["PATH"] = newPath

	// Add other environment variables
	envMap["PYTHONIOENCODING"] = "utf-8"
	envMap["AIGCPANEL_SERVER_PLACEHOLDER_CONFIG"] = "" // Will be set per execution
	envMap["AIGCPANEL_SERVER_PLACEHOLDER_ROOT"] = es.ServerInfo.LocalPath

	// Add custom environment variables from config
	if es.ServerConfig.EasyServer.Envs != nil {
		for _, e := range es.ServerConfig.EasyServer.Envs {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	return envMap
}

// executeCommand executes the model command
func (es *EasyServer) executeCommand(
	command []string,
	envMap map[string]string,
	configPath string,
	taskID string,
	launcherResult *LauncherResultType,
) error {

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = es.ServerInfo.LocalPath

	env := []string{}
	for k, v := range envMap {
		v = strings.ReplaceAll(v, "${CONFIG}", configPath)
		v = strings.ReplaceAll(v, "${ROOT}", es.ServerInfo.LocalPath)
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	es.controller = cmd

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)

	// 关键：只监听输出，不等进程退出
	read := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 8*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)

			result, ok := ExtractResultFromLogs(taskID, line)
			if ok && result != nil {

				for k, v := range result {
					launcherResult.Result[k] = v
				}

				// ⭐⭐⭐ 关键：收到最终结果立即返回
				if _, hasUrl := result["url"]; hasUrl ||
					result["records"] != nil ||
					result["error"] != nil {

					now := time.Now().Unix()
					launcherResult.EndTime = &now

					cmd.Process.Kill()
					done <- nil
					return
				}
			}
		}
	}

	go read(stdout)
	go read(stderr)

	// 无输出超时（10分钟）
	timeout := time.After(10 * time.Minute)

	select {
	case <-done:
		return nil
	case <-timeout:
		cmd.Process.Kill()
		return fmt.Errorf("model timeout (no terminal result)")
	}
}

// SoundTts handles sound TTS function
func (es *EasyServer) SoundTts(data ServerFunctionDataType) (*TaskResult, error) {
	configCalculator := func(data ServerFunctionDataType) (map[string]interface{}, error) {
		return map[string]interface{}{
			"id":   data.ID,
			"mode": "local",
			"modelConfig": map[string]interface{}{
				"type":  "soundTts",
				"param": data.Param,
				"text":  data.Text,
			},
		}, nil
	}

	resultDataCalculator := func(data ServerFunctionDataType, launcherResult LauncherResultType) (map[string]interface{}, error) {
		if _, ok := launcherResult.Result["url"]; !ok {
			if errMsg, ok := launcherResult.Result["error"]; ok {
				return nil, fmt.Errorf("%v", errMsg)
			}
			return nil, fmt.Errorf("执行失败，请查看模型日志")
		}

		return map[string]interface{}{
			"url": launcherResult.Result["url"],
		}, nil
	}

	return es.CallFunc(data, configCalculator, resultDataCalculator)
}

// SoundClone handles sound clone function
func (es *EasyServer) SoundClone(data ServerFunctionDataType) (*TaskResult, error) {
	configCalculator := func(data ServerFunctionDataType) (map[string]interface{}, error) {
		return map[string]interface{}{
			"id":   data.ID,
			"mode": "local",
			"modelConfig": map[string]interface{}{
				"type":        "soundClone",
				"param":       data.Param,
				"text":        data.Text,
				"promptAudio": data.PromptAudio,
				"promptText":  data.PromptText,
			},
		}, nil
	}

	resultDataCalculator := func(data ServerFunctionDataType, launcherResult LauncherResultType) (map[string]interface{}, error) {
		if _, ok := launcherResult.Result["url"]; !ok {
			if errMsg, ok := launcherResult.Result["error"]; ok {
				return nil, fmt.Errorf("%v", errMsg)
			}
			return nil, fmt.Errorf("执行失败，请查看模型日志")
		}

		return map[string]interface{}{
			"url": launcherResult.Result["url"],
		}, nil
	}

	return es.CallFunc(data, configCalculator, resultDataCalculator)
}

// VideoGen handles video generation function
func (es *EasyServer) VideoGen(data ServerFunctionDataType) (*TaskResult, error) {
	configCalculator := func(data ServerFunctionDataType) (map[string]interface{}, error) {
		return map[string]interface{}{
			"id":   data.ID,
			"mode": "local",
			"modelConfig": map[string]interface{}{
				"type":  "videoGen",
				"param": data.Param,
				"video": data.Video,
				"audio": data.Audio,
			},
		}, nil
	}

	resultDataCalculator := func(data ServerFunctionDataType, launcherResult LauncherResultType) (map[string]interface{}, error) {
		if _, ok := launcherResult.Result["url"]; !ok {
			if errMsg, ok := launcherResult.Result["error"]; ok {
				return nil, fmt.Errorf("%v", errMsg)
			}
			return nil, fmt.Errorf("执行失败，请查看模型日志")
		}

		return map[string]interface{}{
			"url": launcherResult.Result["url"],
		}, nil
	}

	return es.CallFunc(data, configCalculator, resultDataCalculator)
}

// Asr handles ASR function
func (es *EasyServer) Asr(data ServerFunctionDataType) (*TaskResult, error) {
	configCalculator := func(data ServerFunctionDataType) (map[string]interface{}, error) {
		return map[string]interface{}{
			"id":   data.ID,
			"mode": "local",
			"modelConfig": map[string]interface{}{
				"audio": data.Audio,
				"param": data.Param,
			},
		}, nil
	}

	resultDataCalculator := func(data ServerFunctionDataType, launcherResult LauncherResultType) (map[string]interface{}, error) {
		if _, ok := launcherResult.Result["records"]; !ok {
			if errMsg, ok := launcherResult.Result["error"]; ok {
				return nil, fmt.Errorf("%v", errMsg)
			}
			return nil, fmt.Errorf("执行失败，请查看模型日志")
		}

		return map[string]interface{}{
			"records": launcherResult.Result["records"],
		}, nil
	}

	return es.CallFunc(data, configCalculator, resultDataCalculator)
}
