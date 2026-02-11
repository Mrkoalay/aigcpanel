package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xiacutai-server/internal/component/modelcall"
	"xiacutai-server/internal/component/modelcall/easyserver"
	"xiacutai-server/internal/domain"
)

type soundReplaceRecord struct {
	Text        string `json:"text"`
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Audio       string `json:"audio,omitempty"`
	ActualStart int64  `json:"actualStart,omitempty"`
	ActualEnd   int64  `json:"actualEnd,omitempty"`
}

func runSoundReplaceTask(task domain.DataTaskModel, cfg *taskConfig) error {
	if task.Status == domain.TaskStatusWait {
		return nil
	}

	videoPath := cfg.Video
	if videoPath == "" {
		return fmt.Errorf("video is required")
	}

	job := map[string]any{
		"step":           "ToAudio",
		"ToAudio":        map[string]any{"status": "queue", "file": ""},
		"SoundAsr":       map[string]any{"status": "queue", "start": 0, "end": 0, "duration": 0, "records": []any{}},
		"Confirm":        map[string]any{"status": "queue", "records": []any{}},
		"SoundGenerate":  map[string]any{"status": "queue", "records": nil},
		"Combine":        map[string]any{"status": "queue"},
		"CombineConfirm": map[string]any{"status": "queue"},
	}
	if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusRunning, job, nil, ""); err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("sound-replace-%d-*", task.ID))
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	persistDir := filepath.Dir(videoPath)
	stamp := time.Now().UnixMilli()
	persistSourceAudio := filepath.Join(persistDir, fmt.Sprintf("sound_replace_source_%d.mp3", stamp))

	audioPath := filepath.Join(workDir, "source_audio.wav")
	if err := ffmpegExtractAudio(videoPath, audioPath); err != nil {
		return err
	}
	if err := ffmpegEncodeMp3(audioPath, persistSourceAudio); err != nil {
		return err
	}
	job["step"] = "SoundAsr"
	jobToAudio := asMap(job["ToAudio"])
	jobToAudio["status"] = "success"
	jobToAudio["file"] = persistSourceAudio
	job["ToAudio"] = jobToAudio
	jobAsr := asMap(job["SoundAsr"])
	jobAsr["status"] = "running"
	job["SoundAsr"] = jobAsr
	if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusRunning, job, nil, ""); err != nil {
		return err
	}

	asrStart := time.Now().UnixMilli()
	asrServerKey := asString(cfg.SoundAsr["serverKey"])
	if asrServerKey == "" {
		return fmt.Errorf("soundAsr.serverKey is required")
	}
	asrParam := asMap(cfg.SoundAsr["param"])
	asrServer, err := startEasyServerByKey(asrServerKey)
	if err != nil {
		return err
	}
	registerTaskServer(task.ID, asrServer)
	asrResult, err := asrServer.Asr(easyserver.ServerFunctionDataType{ID: fmt.Sprintf("task-%d-asr", task.ID), Param: asrParam, Result: map[string]interface{}{}, Audio: audioPath})
	_ = asrServer.Stop()
	if err != nil {
		unregisterTaskServer(task.ID)
		return err
	}
	asrData, err := extractResultData(asrResult)
	if err != nil {
		unregisterTaskServer(task.ID)
		return err
	}
	records, err := parseAsrRecords(asrData)
	if err != nil {
		unregisterTaskServer(task.ID)
		return err
	}
	if len(records) == 0 {
		unregisterTaskServer(task.ID)
		return fmt.Errorf("asr records empty")
	}
	asrEnd := time.Now().UnixMilli()
	jobAsr["status"] = "success"
	jobAsr["start"] = asrStart
	jobAsr["end"] = asrEnd
	jobAsr["duration"] = asrEnd - asrStart
	jobAsr["records"] = records
	job["SoundAsr"] = jobAsr
	jobConfirm := asMap(job["Confirm"])
	jobConfirm["status"] = "pending"
	jobConfirm["records"] = records
	job["Confirm"] = jobConfirm
	job["step"] = "Confirm"
	jobGen := asMap(job["SoundGenerate"])
	jobGen["status"] = "queue"
	jobGen["records"] = nil
	job["SoundGenerate"] = jobGen
	if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusWait, job, nil, ""); err != nil {
		unregisterTaskServer(task.ID)
		return err
	}
	unregisterTaskServer(task.ID)
	return nil
}

func saveSoundReplaceProgress(taskID int64, status string, jobResult map[string]any, result map[string]any, statusMsg string) error {
	jobRaw, err := json.Marshal(jobResult)
	if err != nil {
		return err
	}
	updates := map[string]any{"status": status, "jobResult": string(jobRaw)}
	if statusMsg != "" {
		updates["statusMsg"] = statusMsg
	}
	if status == domain.TaskStatusRunning {
		updates["startTime"] = time.Now().UnixMilli()
	}
	if status == domain.TaskStatusSuccess || status == domain.TaskStatusFail {
		updates["endTime"] = time.Now().UnixMilli()
	}
	if result != nil {
		retRaw, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updates["result"] = string(retRaw)
	}
	_, err = DataTask.UpdateTask(taskID, updates)
	return err
}

func startEasyServerByKey(serverKey string) (*easyserver.EasyServer, error) {
	modelInfo, err := Model.Get(serverKey)
	if err != nil {
		return nil, err
	}
	serverConfig, err := modelcall.LoadConfigFromJSON(modelInfo.Path + "/config.json")
	if err != nil {
		return nil, err
	}
	serverInfo := &easyserver.ServerInfo{LocalPath: modelInfo.Path, Name: modelInfo.Name, Version: modelInfo.Version, Setting: modelInfo.Setting, Config: *serverConfig}
	server := easyserver.NewEasyServer(*serverConfig)
	server.ServerInfo = serverInfo
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}

func parseAsrRecords(asrData map[string]any) ([]*soundReplaceRecord, error) {
	items, ok := asrData["records"].([]interface{})
	if !ok || len(items) == 0 {
		if one, ok := asrData["record"].([]interface{}); ok {
			items = one
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("asr records not found")
	}
	records := make([]*soundReplaceRecord, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		text := asString(m["text"])
		if text == "" {
			continue
		}
		start := toInt64(m["start"])
		end := toInt64(m["end"])
		if end <= start {
			continue
		}
		records = append(records, &soundReplaceRecord{Text: text, Start: start, End: end})
	}
	return records, nil
}

func generateSpeechForRecord(taskID int64, idx int, rec *soundReplaceRecord, soundGenerate map[string]any, server *easyserver.EasyServer, outputPath string) error {
	generateType := strings.ToLower(asString(soundGenerate["type"]))
	param := asMap(soundGenerate["ttsParam"])
	if strings.Contains(generateType, "clone") {
		param = asMap(soundGenerate["cloneParam"])
	}
	data := easyserver.ServerFunctionDataType{ID: fmt.Sprintf("task-%d-gen-%d", taskID, idx), Result: map[string]interface{}{}, Param: param, Text: rec.Text}
	var result *easyserver.TaskResult
	var err error
	if strings.Contains(generateType, "clone") {
		data.PromptAudio = asString(soundGenerate["promptUrl"])
		data.PromptText = asString(soundGenerate["promptText"])
		result, err = server.SoundClone(data)
	} else {
		result, err = server.SoundTts(data)
	}
	if err != nil {
		return err
	}
	resultData, err := extractResultData(result)
	if err != nil {
		return err
	}
	src := asString(resultData["url"])
	if src == "" {
		return fmt.Errorf("sound generate result url empty")
	}
	return copyFile(src, outputPath)
}

func alignAudioDuration(input, output string, targetMs int64) error {
	actualMs, err := ffprobeDurationMs(input)
	if err != nil {
		return err
	}
	targetSec := fmt.Sprintf("%.3f", float64(targetMs)/1000.0)
	if actualMs > targetMs {
		return runCommand("ffmpeg", "-y", "-i", input, "-t", targetSec, "-acodec", "pcm_s16le", output)
	}
	if actualMs < targetMs {
		return runCommand("ffmpeg", "-y", "-i", input, "-af", "apad", "-t", targetSec, "-acodec", "pcm_s16le", output)
	}
	return copyFile(input, output)
}

func ffmpegExtractAudio(video, output string) error {
	return runCommand("ffmpeg", "-y", "-i", video, "-vn", "-acodec", "pcm_s16le", output)
}

func ffmpegEncodeMp3(input, output string) error {
	return runCommand("ffmpeg", "-y", "-i", input, "-codec:a", "libmp3lame", "-q:a", "2", output)
}

func ffmpegConcatAudio(files []string, output string) error {
	if len(files) == 0 {
		return fmt.Errorf("no audio files to concat")
	}
	listFile := filepath.Join(filepath.Dir(output), "concat.txt")
	f, err := os.Create(listFile)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, file := range files {
		_, _ = w.WriteString("file '" + strings.ReplaceAll(file, "'", "'\\''") + "'\n")
	}
	_ = w.Flush()
	_ = f.Close()
	return runCommand("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-acodec", "pcm_s16le", output)
}

func ffmpegReplaceVideoAudio(video, audio, output string) error {
	return runCommand("ffmpeg", "-y", "-i", video, "-i", audio, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy", "-shortest", output)
}

func ffprobeDurationMs(file string) (int64, error) {
	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", file).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(out))
	}
	sec, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return int64(sec * 1000), nil
}

func runCommand(cmd string, args ...string) error {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w, output: %s", cmd, err, string(out))
	}
	return nil
}

func asString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func asMap(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]any{}
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		i, _ := t.Int64()
		return i
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return int64(f)
	default:
		return 0
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := out.ReadFrom(in); err != nil {
		return err
	}
	return out.Sync()
}
