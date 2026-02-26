package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/modelcall/easyserver"
	"xiacutai-server/internal/domain"
)

func runVideoGenFlowTask(task domain.DataTaskModel, cfg *taskConfig) error {
	videoPath := strings.TrimSpace(cfg.VideoTemplateURL)
	if videoPath == "" {
		videoPath = strings.TrimSpace(cfg.Video)
	}
	if videoPath == "" {
		return errs.New("videoTemplateUrl is required")
	}

	audioPath, soundResult, err := runVideoGenFlowSoundGenerate(task, cfg)
	if err != nil {
		return err
	}

	videoServerKey := cfg.ServerKey
	if videoServerKey == "" {
		return errs.New("serverKey is required")
	}
	videoServer, err := startEasyServerByKey(videoServerKey)
	if err != nil {
		return err
	}
	registerTaskServer(task.ID, videoServer)
	defer func() {
		_ = videoServer.Stop()
		unregisterTaskServer(task.ID)
	}()

	videoRes, err := videoServer.VideoGen(easyserver.ServerFunctionDataType{
		ID:     fmt.Sprintf("task-%d-video-gen", task.ID),
		Result: map[string]interface{}{},
		Param:  cfg.VideoParam,
		Video:  videoPath,
		Audio:  audioPath,
	})
	if err != nil {
		return err
	}
	videoData, err := extractResultData(videoRes)
	if err != nil {
		return err
	}

	jobResult := map[string]any{soundResult.name: soundResult.raw, "videoGen": videoRes}
	jobResultRaw, _ := json.Marshal(jobResult)

	resultData := map[string]any{"urlSound": audioPath}
	for k, v := range videoData {
		resultData[k] = v
	}
	resultRaw, _ := json.Marshal(resultData)

	_, err = DataTask.UpdateTask(task.ID, map[string]any{
		"status":    domain.TaskStatusSuccess,
		"jobResult": string(jobResultRaw),
		"result":    string(resultRaw),
		"endTime":   time.Now().UnixMilli(),
	})
	return err
}

type soundResultPayload struct {
	name string
	raw  *easyserver.TaskResult
}

func runVideoGenFlowSoundGenerate(task domain.DataTaskModel, cfg *taskConfig) (string, *soundResultPayload, error) {
	soundGenerate := cfg.SoundGenerate
	if len(soundGenerate) == 0 {
		return "", nil, errs.New("soundGenerate is required")
	}

	generateType := strings.ToLower(asString(soundGenerate["type"]))
	serverKey := asString(soundGenerate["ttsServerKey"])
	method := "soundTts"
	param := asMap(soundGenerate["ttsParam"])
	callData := easyserver.ServerFunctionDataType{
		ID:     fmt.Sprintf("task-%d-sound-gen", task.ID),
		Result: map[string]interface{}{},
		Param:  param,
		Text:   cfg.Text,
	}
	jobResultName := "soundTts"

	if strings.Contains(generateType, "clone") {
		serverKey = asString(soundGenerate["cloneServerKey"])
		method = "soundClone"
		param = asMap(soundGenerate["cloneParam"])
		callData.Param = param
		callData.PromptAudio = asString(soundGenerate["promptUrl"])
		callData.PromptText = asString(soundGenerate["promptText"])
		jobResultName = "soundClone"
	}

	if serverKey == "" {
		return "", nil, errs.New("soundGenerate server key is required")
	}

	soundServer, err := startEasyServerByKey(serverKey)
	if err != nil {
		return "", nil, err
	}
	registerTaskServer(task.ID, soundServer)
	defer func() {
		_ = soundServer.Stop()
		unregisterTaskServer(task.ID)
	}()

	var soundRes *easyserver.TaskResult
	if method == "soundClone" {
		soundRes, err = soundServer.SoundClone(callData)
	} else {
		soundRes, err = soundServer.SoundTts(callData)
	}
	if err != nil {
		return "", nil, err
	}
	soundData, err := extractResultData(soundRes)
	if err != nil {
		return "", nil, err
	}
	audioPath := asString(soundData["url"])
	if audioPath == "" {
		return "", nil, errs.New("soundGenerate result missing url")
	}

	return audioPath, &soundResultPayload{name: jobResultName, raw: soundRes}, nil
}
