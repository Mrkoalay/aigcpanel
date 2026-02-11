package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"xiacutai-server/internal/component/log"
	"xiacutai-server/internal/component/modelcall"
	"xiacutai-server/internal/component/modelcall/easyserver"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/utils"

	"go.uber.org/zap"
)

var errTaskRetry = errors.New("task requires retry")

type taskConfig struct {
	Type           string                 `json:"type"`
	TtsServerKey   string                 `json:"ttsServerKey"`
	TtsParam       map[string]any         `json:"ttsParam"`
	CloneServerKey string                 `json:"cloneServerKey"`
	ServerKey      string                 `json:"serverKey"`
	CloneParam     map[string]any         `json:"cloneParam"`
	PromptURL      string                 `json:"promptUrl"`
	PromptText     string                 `json:"promptText"`
	Video          string                 `json:"video"`
	VideoParam     map[string]any         `json:"param"`
	SoundAsr       map[string]any         `json:"soundAsr"`
	SoundGenerate  map[string]any         `json:"soundGenerate"`
	Audio          string                 `json:"audio"`
	Text           string                 `json:"text"`
	Extra          map[string]interface{} `json:"-"`
}

func StartTaskScheduler(ctx context.Context) {
	interval := getTaskPollInterval()
	ticker := time.NewTicker(interval)

	log.Info("DataTask scheduler started", zap.Duration("interval", interval))

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("DataTask scheduler stopped")
				return
			case <-ticker.C:
				if err := runTaskSchedulerOnce(); err != nil {
					log.Error("DataTask scheduler failed", zap.Error(err))
				}
			}
		}
	}()
}

func getTaskPollInterval() time.Duration {
	const defaultInterval = 2 * time.Second
	raw := utils.GetEnv("AIGCPANEL_TASK_POLL_INTERVAL_MS", "")
	if raw == "" {
		return defaultInterval
	}
	ms, err := time.ParseDuration(raw + "ms")
	if err != nil || ms <= 0 {
		return defaultInterval
	}
	return ms
}

func runTaskSchedulerOnce() error {
	filters := sqllite.TaskFilters{
		Status: []string{domain.TaskStatusQueue},
	}
	tasks, err := DataTask.ListTasks(filters)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		//if task.Biz != "SoundGenerate" {
		//	continue
		//}
		if err := handleSoundTask(task); err != nil {
			log.Error("Handle task failed", zap.Int64("taskId", task.ID), zap.Error(err))
		}
	}
	return nil
}

func handleSoundTask(task domain.DataTaskModel) error {
	if task.Status != domain.TaskStatusQueue {
		return nil
	}

	if err := setTaskRunning(task.ID); err != nil {
		return err
	}

	cfg, err := parseSoundTaskConfig(task.ModelConfig)
	if err != nil {
		return setTaskFailed(task.ID, err)
	}
	if cfg.Type == domain.FunctionSoundReplace {
		if runErr := runSoundReplaceTask(task, cfg); runErr != nil {
			return setTaskFailed(task.ID, runErr)
		}
		return nil
	}

	serverKey := cfg.TtsServerKey
	if cfg.Type == domain.FunctionSoundClone {
		serverKey = cfg.CloneServerKey
	}
	if cfg.Type == domain.FunctionVideoGen {
		serverKey = cfg.ServerKey
	}
	if cfg.Type == domain.FunctionSoundAsr {
		serverKey = cfg.ServerKey
	}
	modelInfo, err := Model.Get(serverKey)
	if err != nil {
		return setTaskFailed(task.ID, err)
	}

	serverConfig, err := modelcall.LoadConfigFromJSON(modelInfo.Path + "/config.json")
	if err != nil {
		return setTaskFailed(task.ID, err)
	}

	serverInfo := &easyserver.ServerInfo{
		LocalPath:        modelInfo.Path,
		Name:             modelInfo.Name,
		Version:          modelInfo.Version,
		Setting:          modelInfo.Setting,
		LogFile:          "",
		EventChannelName: "",
		Config:           *serverConfig,
	}

	server := easyserver.NewEasyServer(*serverConfig)
	registerTaskServer(task.ID, server)
	defer unregisterTaskServer(task.ID)
	server.ServerInfo = serverInfo
	if err := server.Start(); err != nil {
		return setTaskFailed(task.ID, err)
	}

	result, err := callEasyServerTask(task, cfg, server)
	if err != nil {
		return setTaskFailed(task.ID, err)
	}

	if result != nil {
		return updateTaskResult(task.ID, result)
	}

	return setTaskFailed(task.ID, fmt.Errorf("empty task result"))
}

func parseSoundTaskConfig(raw string) (*taskConfig, error) {
	cfg := &taskConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	if cfg.Type == "" {
		cfg.Type = domain.FunctionSoundTts
	}
	return cfg, nil
}

func setTaskRunning(taskID int64) error {
	_, err := DataTask.UpdateTask(taskID, map[string]any{
		"status":    domain.TaskStatusRunning,
		"startTime": time.Now().UnixMilli(),
	})
	return err
}

func setTaskFailed(taskID int64, err error) error {
	statusMsg := ""
	if err != nil {
		statusMsg = err.Error()
	}
	_, updateErr := DataTask.UpdateTask(taskID, map[string]any{
		"status":    domain.TaskStatusFail,
		"statusMsg": statusMsg,
		"endTime":   time.Now().UnixMilli(),
	})
	return updateErr
}

func callEasyServerTask(task domain.DataTaskModel, cfg *taskConfig, server *easyserver.EasyServer) (*easyserver.TaskResult, error) {
	taskID := fmt.Sprintf("task-%d", task.ID)
	data := easyserver.ServerFunctionDataType{
		ID:     taskID,
		Result: map[string]interface{}{},
		Param:  cfg.TtsParam,
		Text:   cfg.Text,
	}

	switch cfg.Type {
	case domain.FunctionSoundClone:
		data.Param = cfg.CloneParam
		data.PromptAudio = cfg.PromptURL
		data.PromptText = cfg.PromptText
		data.Text = cfg.Text
		return server.SoundClone(data)
	case domain.FunctionSoundAsr:
		data.Param = map[string]interface{}{}
		data.Audio = cfg.Audio
		return server.Asr(data)
	case domain.FunctionVideoGen:
		data.Param = cfg.VideoParam
		data.Video = cfg.Video
		data.Audio = cfg.Audio
		return server.VideoGen(data)
	default:
		data.Param = cfg.TtsParam
		data.Text = cfg.Text
		return server.SoundTts(data)
	}
}

func updateTaskResult(taskID int64, result *easyserver.TaskResult) error {
	jobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resultData, err := extractResultData(result)
	if err != nil {
		if errors.Is(err, errTaskRetry) {
			_, updateErr := DataTask.UpdateTask(taskID, map[string]any{
				"status":    domain.TaskStatusQueue,
				"statusMsg": err.Error(),
			})
			return updateErr
		}
		return setTaskFailed(taskID, err)
	}

	resultRaw, err := json.Marshal(resultData)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"jobResult": string(jobResult),
		"result":    string(resultRaw),
		"endTime":   time.Now().UnixMilli(),
		"status":    domain.TaskStatusSuccess,
	}

	_, err = DataTask.UpdateTask(taskID, updates)
	return err
}

func extractResultData(result *easyserver.TaskResult) (map[string]any, error) {
	if result == nil {
		return nil, fmt.Errorf("task result is nil")
	}
	if result.Code != 0 {
		if result.Msg == "" {
			return nil, fmt.Errorf("task failed")
		}
		return nil, fmt.Errorf(result.Msg)
	}

	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("task result data format invalid")
	}

	if dataMap["type"] == "retry" {
		return nil, errTaskRetry
	}

	resultData, ok := dataMap["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("task result payload missing")
	}

	return resultData, nil
}
