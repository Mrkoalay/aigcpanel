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
			envMap[parts[0]] = parts[1]
		}
	}

	// Add custom paths
	path := envMap["PATH"]
	envMap["PATH"] = fmt.Sprintf("%s:%s:%s/binary", path, es.ServerInfo.LocalPath, es.ServerInfo.LocalPath)

	// 添加 _aienv 方便调试
	venv := es.ServerInfo.LocalPath + string(os.PathSeparator) + "_aienv"
	python := venv + string(os.PathSeparator) + "Scripts"
	torchlib := venv + string(os.PathSeparator) + "Lib" + string(os.PathSeparator) + "site-packages" + string(os.PathSeparator) + "torch" + string(os.PathSeparator) + "lib"
	oldPath := envMap["PATH"]
	envMap["PATH"] = fmt.Sprintf("%s;%s;%s;%s",
		oldPath,
		venv,
		python,
		torchlib,
	)

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

	fmt.Printf("Executing command: %v\n", command)
	fmt.Printf("Working directory: %s\n", es.ServerInfo.LocalPath)

	// Create command
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = es.ServerInfo.LocalPath

	// Set environment variables
	env := []string{}
	for k, v := range envMap {
		// Replace placeholders in environment variables
		v = strings.ReplaceAll(v, "${CONFIG}", configPath)
		v = strings.ReplaceAll(v, "${ROOT}", es.ServerInfo.LocalPath)
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Env = env

	// Print environment variables for debugging
	fmt.Printf("Environment variables:\n")
	for _, e := range env {
		fmt.Printf("  %s\n", e)
	}

	// Set the controller
	es.controller = cmd

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	fmt.Printf("Command started with PID: %d\n", cmd.Process.Pid)

	// Create channels to signal when we're done reading
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)
	cmdDone := make(chan error, 1)

	// Buffer to store logs
	var logBuffer strings.Builder

	// Function to read from a pipe
	readPipe := func(pipe io.Reader, pipeName string, doneChan chan error) {
		defer func() {
			doneChan <- nil // Signal that this goroutine is done
		}()

		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			line := scanner.Text()
			logBuffer.WriteString(line + "\n")
			fmt.Printf("[%s] %s\n", pipeName, line)

			// Try to extract result from logs
			result, err := ExtractResultFromLogs(taskID, line)
			if err == nil && result != nil {
				// Merge the result into launcherResult
				for k, v := range result {
					launcherResult.Result[k] = v
				}
				fmt.Printf("Extracted result for task %s: %+v\n", taskID, result)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading from %s: %v\n", pipeName, err)
			doneChan <- fmt.Errorf("error reading from %s: %v", pipeName, err)
		}
	}

	// Read from both stdout and stderr concurrently
	go readPipe(stdout, "stdout", stdoutDone)
	go readPipe(stderr, "stderr", stderrDone)

	// Wait for the command to finish
	go func() {
		err := cmd.Wait()
		cmdDone <- err // Send the command result
	}()

	// Wait for all goroutines to complete or timeout
	var cmdErr error
	completed := 0
	expectedCompletions := 3 // stdout, stderr, and command

	for completed < expectedCompletions {
		select {
		case err := <-stdoutDone:
			completed++
			if err != nil && cmdErr == nil {
				cmdErr = err
			}
		case err := <-stderrDone:
			completed++
			if err != nil && cmdErr == nil {
				cmdErr = err
			}
		case err := <-cmdDone:
			completed++
			if err != nil && cmdErr == nil {
				cmdErr = err
			}
		case <-time.After(120 * time.Second): // Increased timeout to 120 seconds for voice cloning
			// Terminate the process
			if cmd.Process != nil {
				fmt.Printf("Command timed out, terminating process\n")
				cmd.Process.Kill()
			}
			return fmt.Errorf("command timed out after 120 seconds")
		}
	}

	if cmdErr != nil {
		fmt.Printf("Command finished with error: %v\n", cmdErr)
		// Even if there's an error, we might have extracted some results
		// Only return error if we didn't get a result
		if len(launcherResult.Result) == 0 {
			return fmt.Errorf("command failed: %v", cmdErr)
		}
		// If we have results, we consider it a success even with some errors
		fmt.Printf("Command completed with partial errors, but results were extracted\n")
	} else {
		fmt.Printf("Command finished successfully\n")
	}

	// Set end time
	endTime := time.Now().Unix()
	launcherResult.EndTime = &endTime

	// Print final log buffer for debugging
	fmt.Printf("Full log output:\n%s\n", logBuffer.String())

	return nil
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
