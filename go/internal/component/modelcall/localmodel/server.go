package localmodel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"xiacutai-server/internal/component/errs"
)

// ServerManager manages local AI models
type ServerManager struct {
	servers []ServerRecord
	mutex   sync.RWMutex
}

// NewServerManager creates a new server manager
func NewServerManager() *ServerManager {
	return &ServerManager{
		servers: make([]ServerRecord, 0),
	}
}

// LoadServerConfig loads server configuration from a config.json file
func (sm *ServerManager) LoadServerConfig(configPath string) (*ServerConfig, error) {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errs.New(fmt.Sprintf("failed to read config file: %v", err))
	}

	var config ServerConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, errs.New(fmt.Sprintf("failed to parse config file: %v", err))
	}

	return &config, nil
}

// AddServer adds a new server
func (sm *ServerManager) AddServer(config *ServerConfig, localPath string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Generate server key
	key := fmt.Sprintf("%s-%s", config.Name, config.Version)

	// Check if server already exists
	for _, server := range sm.servers {
		if server.Key == key {
			return errs.New(fmt.Sprintf("server with key %s already exists", key))
		}
	}

	// Create server record
	server := ServerRecord{
		Key:       key,
		Name:      config.Name,
		Title:     config.Title,
		Version:   config.Version,
		Type:      ServerLocalDir,
		Functions: config.Functions,
		LocalPath: localPath,
		AutoStart: config.Entry == "__EasyServer__",
		Settings:  config.Settings,
		Setting:   make(map[string]interface{}),
		Status:    ServerStopped,
		Runtime: &ServerRuntime{
			Status:          ServerStopped,
			AutoStartStatus: ServerStopped,
		},
	}

	// Set default settings
	for _, setting := range config.Settings {
		server.Setting[setting.Name] = setting.Default
	}

	sm.servers = append(sm.servers, server)
	return nil
}

// GetServers returns all servers
func (sm *ServerManager) GetServers() []ServerRecord {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to prevent external modification
	servers := make([]ServerRecord, len(sm.servers))
	copy(servers, sm.servers)
	return servers
}

// GetServerByKey returns a server by its key
func (sm *ServerManager) GetServerByKey(key string) *ServerRecord {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for i := range sm.servers {
		if sm.servers[i].Key == key {
			// Return a copy to prevent external modification
			server := sm.servers[i]
			return &server
		}
	}
	return nil
}

// GetServerByNameVersion returns a server by its name and version
func (sm *ServerManager) GetServerByNameVersion(name, version string) *ServerRecord {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for i := range sm.servers {
		if sm.servers[i].Name == name && sm.servers[i].Version == version {
			// Return a copy to prevent external modification
			server := sm.servers[i]
			return &server
		}
	}
	return nil
}

// StartServer starts a server
func (sm *ServerManager) StartServer(key string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i := range sm.servers {
		if sm.servers[i].Key == key {
			if sm.servers[i].Status == ServerRunning {
				return errs.New("server is already running")
			}

			// Update status
			sm.servers[i].Status = ServerStarting
			if sm.servers[i].Runtime == nil {
				sm.servers[i].Runtime = &ServerRuntime{}
			}
			sm.servers[i].Runtime.Status = ServerStarting
			sm.servers[i].Runtime.StartTime = time.Now().Unix()

			// If this is an EasyServer, we don't actually start a process
			// In a real implementation, this would start the actual model process
			if sm.servers[i].AutoStart {
				sm.servers[i].Status = ServerRunning
				sm.servers[i].Runtime.Status = ServerRunning
			}

			return nil
		}
	}

	return errs.New(fmt.Sprintf("server with key %s not found", key))
}

// StopServer stops a server
func (sm *ServerManager) StopServer(key string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i := range sm.servers {
		if sm.servers[i].Key == key {
			if sm.servers[i].Status != ServerRunning {
				return errs.New("server is not running")
			}

			// Update status
			sm.servers[i].Status = ServerStopping
			if sm.servers[i].Runtime != nil {
				sm.servers[i].Runtime.Status = ServerStopping
			}

			// In a real implementation, this would stop the actual model process
			sm.servers[i].Status = ServerStopped
			if sm.servers[i].Runtime != nil {
				sm.servers[i].Runtime.Status = ServerStopped
				sm.servers[i].Runtime.StartTime = 0
			}

			return nil
		}
	}

	return errs.New(fmt.Sprintf("server with key %s not found", key))
}

// ServerInfo returns information about a server
func (sm *ServerManager) ServerInfo(key string) (*ServerInfo, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for i := range sm.servers {
		if sm.servers[i].Key == key {
			serverInfo := &ServerInfo{
				LocalPath:        sm.servers[i].LocalPath,
				Name:             sm.servers[i].Name,
				Version:          sm.servers[i].Version,
				Setting:          sm.servers[i].Setting,
				LogFile:          "",
				EventChannelName: "",
				Config:           sm.servers[i],
			}

			if sm.servers[i].Runtime != nil {
				serverInfo.LogFile = sm.servers[i].Runtime.LogFile
				serverInfo.EventChannelName = sm.servers[i].Runtime.EventChannelName
			}

			return serverInfo, nil
		}
	}

	return nil, errs.New(fmt.Sprintf("server with key %s not found", key))
}

// PingServer checks if a server is running
func (sm *ServerManager) PingServer(key string) (bool, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for i := range sm.servers {
		if sm.servers[i].Key == key {
			// For EasyServer, we just check if start time is set
			if sm.servers[i].Runtime != nil && sm.servers[i].Runtime.StartTime > 0 {
				return true, nil
			}
			return false, nil
		}
	}

	return false, errs.New(fmt.Sprintf("server with key %s not found", key))
}

// CallFunction calls a function on a server
func (sm *ServerManager) CallFunction(key string, function string, data ServerFunctionDataType) (*TaskResult, error) {
	sm.mutex.RLock()
	server := sm.getServerCopy(key)
	sm.mutex.RUnlock()

	if server == nil {
		return nil, errs.New(fmt.Sprintf("server with key %s not found", key))
	}

	if server.Status != ServerRunning {
		return nil, errs.New("server is not running")
	}

	// Check if function is supported
	supported := false
	for _, f := range server.Functions {
		if string(f) == function {
			supported = true
			break
		}
	}

	if !supported {
		return nil, errs.New(fmt.Sprintf("function %s is not supported by this server", function))
	}

	// Handle different functions
	switch function {
	case "soundTts":
		return sm.callSoundTts(server, data)
	case "soundClone":
		return sm.callSoundClone(server, data)
	case "videoGen":
		return sm.callVideoGen(server, data)
	case "asr":
		return sm.callAsr(server, data)
	default:
		return nil, errs.New(fmt.Sprintf("function %s is not implemented", function))
	}
}

// getServerCopy returns a copy of a server record
func (sm *ServerManager) getServerCopy(key string) *ServerRecord {
	for i := range sm.servers {
		if sm.servers[i].Key == key {
			server := sm.servers[i]
			return &server
		}
	}
	return nil
}

// callSoundTts calls the sound TTS function
func (sm *ServerManager) callSoundTts(server *ServerRecord, data ServerFunctionDataType) (*TaskResult, error) {
	// Prepare configuration data
	configData := SoundTtsModelConfig{
		Type:  "soundTts",
		Param: data.Param,
		Text:  data.Text,
	}

	// Convert to map for JSON serialization
	configMap := map[string]interface{}{
		"type":  configData.Type,
		"param": configData.Param,
		"text":  configData.Text,
	}

	// Prepare config JSON file
	configPath, err := sm.prepareConfigJson(configMap)
	if err != nil {
		return nil, errs.New(fmt.Sprintf("failed to prepare config: %v", err))
	}
	defer os.Remove(configPath)

	// Execute model process
	cmd, err := sm.executeModelProcess(server, configPath)
	if err != nil {
		return nil, errs.New(fmt.Sprintf("failed to execute model process: %v", err))
	}

	// Wait for completion and get result
	err = cmd.Wait()
	if err != nil {
		return nil, errs.New(fmt.Sprintf("model process failed: %v", err))
	}

	return &TaskResult{
		Code: 0,
		Msg:  "Success",
		Data: data.Result,
	}, nil
}

// callSoundClone handles sound clone function call
func (sm *ServerManager) callSoundClone(server *ServerRecord, data ServerFunctionDataType) (*TaskResult, error) {
	// Prepare configuration data
	_ = map[string]interface{}{
		"id":   data.ID,
		"mode": "local",
		"modelConfig": SoundCloneModelConfig{
			Type:        "soundClone",
			Param:       data.Param,
			Text:        data.Text,
			PromptAudio: data.PromptAudio,
			PromptText:  data.PromptText,
		},
		"setting": server.Setting,
	}

	// In a real implementation, this would:
	// 1. Prepare config.json file with configData
	// 2. Execute the model process with the config file
	// 3. Capture stdout/stderr and parse results
	// 4. Return the result

	// For this example, we'll simulate a successful result
	result := &TaskResult{
		Code: 0,
		Msg:  "ok",
		Data: map[string]interface{}{
			"type":  "success",
			"start": time.Now().Unix(),
			"end":   time.Now().Unix() + 10, // Simulate 10 seconds processing
			"data": map[string]interface{}{
				"url": "/path/to/generated/clone.wav",
			},
		},
	}

	return result, nil
}

// callVideoGen handles video generation function call
func (sm *ServerManager) callVideoGen(server *ServerRecord, data ServerFunctionDataType) (*TaskResult, error) {
	// Prepare configuration data
	_ = map[string]interface{}{
		"id":   data.ID,
		"mode": "local",
		"modelConfig": VideoGenModelConfig{
			Type:  "videoGen",
			Param: data.Param,
			Video: data.Video,
			Audio: data.Audio,
		},
		"setting": server.Setting,
	}

	// In a real implementation, this would:
	// 1. Prepare config.json file with configData
	// 2. Execute the model process with the config file
	// 3. Capture stdout/stderr and parse results
	// 4. Return the result

	// For this example, we'll simulate a successful result
	result := &TaskResult{
		Code: 0,
		Msg:  "ok",
		Data: map[string]interface{}{
			"type":  "success",
			"start": time.Now().Unix(),
			"end":   time.Now().Unix() + 30, // Simulate 30 seconds processing
			"data": map[string]interface{}{
				"url": "/path/to/generated/video.mp4",
			},
		},
	}

	return result, nil
}

// callAsr handles ASR (Automatic Speech Recognition) function call
func (sm *ServerManager) callAsr(server *ServerRecord, data ServerFunctionDataType) (*TaskResult, error) {
	// Prepare configuration data
	_ = map[string]interface{}{
		"id":   data.ID,
		"mode": "local",
		"modelConfig": AsrModelConfig{
			Audio: data.Audio,
			Param: data.Param,
		},
		"setting": server.Setting,
	}

	// In a real implementation, this would:
	// 1. Prepare config.json file with configData
	// 2. Execute the model process with the config file
	// 3. Capture stdout/stderr and parse results
	// 4. Return the result

	// For this example, we'll simulate a successful result
	result := &TaskResult{
		Code: 0,
		Msg:  "ok",
		Data: map[string]interface{}{
			"type":  "success",
			"start": time.Now().Unix(),
			"end":   time.Now().Unix() + 5, // Simulate 5 seconds processing
			"data": map[string]interface{}{
				"text": "Recognized speech text would be here",
			},
		},
	}

	return result, nil
}

// prepareConfigJson prepares a config.json file for the model
func (sm *ServerManager) prepareConfigJson(configData map[string]interface{}) (string, error) {
	// Create a temporary file
	tmpFile, err := ioutil.TempFile("", "model-config-*.json")
	if err != nil {
		return "", errs.New(fmt.Sprintf("failed to create temp file: %v", err))
	}
	defer tmpFile.Close()

	// Marshal config data to JSON
	configBytes, err := json.Marshal(configData)
	if err != nil {
		return "", errs.New(fmt.Sprintf("failed to marshal config data: %v", err))
	}

	// Write to file
	_, err = tmpFile.Write(configBytes)
	if err != nil {
		return "", errs.New(fmt.Sprintf("failed to write config data: %v", err))
	}

	return tmpFile.Name(), nil
}

// executeModelProcess executes the model process
func (sm *ServerManager) executeModelProcess(server *ServerRecord, configPath string) (*exec.Cmd, error) {
	if server.Config.EasyServer == nil {
		return nil, errs.New("server is not an EasyServer")
	}

	// Prepare command
	command := []string{server.Config.EasyServer.Entry}
	if server.Config.EasyServer.EntryArgs != nil {
		command = append(command, server.Config.EasyServer.EntryArgs...)
	}

	// Replace placeholders
	for i := range command {
		command[i] = strings.ReplaceAll(command[i], "${CONFIG}", configPath)
		command[i] = strings.ReplaceAll(command[i], "${ROOT}", server.LocalPath)
	}

	// Prepare environment variables
	env := os.Environ()
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Add custom environment variables
	if server.Config.EasyServer.Envs != nil {
		for _, e := range server.Config.EasyServer.Envs {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	// Convert env map back to slice
	env = []string{}
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Create command
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = server.LocalPath
	cmd.Env = env

	return cmd, nil
}
