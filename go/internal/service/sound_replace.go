package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/modelcall"
	"xiacutai-server/internal/component/modelcall/easyserver"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/utils"
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
	videoPath := cfg.Video
	if videoPath == "" {
		return errs.New("video is required")
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
	if strings.TrimSpace(task.JobResult) != "" && task.JobResult != "{}" {
		_ = json.Unmarshal([]byte(task.JobResult), &job)
	}
	step := asString(job["step"])
	if step == "" {
		step = "ToAudio"
		job["step"] = step
	}

	if step == "ToAudio" || step == "SoundAsr" || step == "Confirm" {
		return runSoundReplaceAsrPhase(task, cfg, job)
	}
	return runSoundReplaceGeneratePhase(task, cfg, job)
}

func runSoundReplaceAsrPhase(task domain.DataTaskModel, cfg *taskConfig, job map[string]any) error {
	videoPath := cfg.Video
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
		return errs.New("soundAsr.serverKey is required")
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
		return errs.New("asr records empty")
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

func runSoundReplaceGeneratePhase(task domain.DataTaskModel, cfg *taskConfig, job map[string]any) error {
	confirm := asMap(job["Confirm"])
	confirmRecords, err := parseSoundReplaceRecords(confirm["records"])
	if err != nil {
		return err
	}
	if len(confirmRecords) == 0 {
		return errs.New("confirm records empty")
	}

	persistDir := filepath.Dir(cfg.Video)
	stamp := time.Now().UnixMilli()
	job["step"] = "SoundGenerate"
	confirm["status"] = "success"
	job["Confirm"] = confirm
	jobGen := asMap(job["SoundGenerate"])
	jobGen["status"] = "running"

	genRecords, err := parseSoundReplaceRecords(jobGen["records"])
	if err != nil || len(genRecords) != len(confirmRecords) {
		genRecords = make([]*soundReplaceRecord, 0, len(confirmRecords))
		for _, rec := range confirmRecords {
			genRecords = append(genRecords, &soundReplaceRecord{Text: rec.Text, Start: rec.Start, End: rec.End, Audio: "", ActualStart: 0, ActualEnd: 0})
		}
	}
	jobGen["records"] = genRecords
	job["SoundGenerate"] = jobGen
	if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusRunning, job, nil, ""); err != nil {
		return err
	}

	serverKey := asString(cfg.SoundGenerate["ttsServerKey"])
	if strings.Contains(strings.ToLower(asString(cfg.SoundGenerate["type"])), "clone") {
		serverKey = asString(cfg.SoundGenerate["cloneServerKey"])
	}
	if serverKey == "" {
		return errs.New("soundGenerate server key is required")
	}
	server, err := startEasyServerByKey(serverKey)
	if err != nil {
		return err
	}
	registerTaskServer(task.ID, server)
	defer func() {
		_ = server.Stop()
		unregisterTaskServer(task.ID)
	}()

	generatedWavs := make([]string, 0, len(genRecords))
	for i, rec := range genRecords {
		if strings.TrimSpace(rec.Audio) != "" {
			continue
		}
		aligned := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_seg_%d.wav", stamp, i))
		targetMs := rec.End - rec.Start
		if targetMs <= 0 {
			targetMs = 1
		}

		text := strings.TrimSpace(rec.Text)
		if text == "" {
			if err := createSilenceAudio(aligned, targetMs); err != nil {
				return err
			}
		} else {
			rec.Text = text
			rawOutput := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_seg_%d_raw.wav", stamp, i))
			if err := generateSpeechForRecord(task.ID, i, rec, cfg.SoundGenerate, server, rawOutput); err != nil {
				if err := createSilenceAudio(aligned, targetMs); err != nil {
					return errs.New(fmt.Sprintf("segment %d generate failed: %v", i, err))
				}
			} else if err := alignAudioDuration(rawOutput, aligned, targetMs); err != nil {
				if err := createSilenceAudio(aligned, targetMs); err != nil {
					return err
				}
			}
		}

		rec.Audio = aligned
		rec.ActualStart = rec.Start
		rec.ActualEnd = rec.End
		generatedWavs = append(generatedWavs, aligned)
		jobGen["records"] = genRecords
		job["SoundGenerate"] = jobGen
		if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusRunning, job, nil, ""); err != nil {
			return err
		}
	}

	jobGen["status"] = "success"
	job["SoundGenerate"] = jobGen
	job["step"] = "Combine"
	jobCombine := asMap(job["Combine"])
	jobCombine["status"] = "running"
	job["Combine"] = jobCombine
	if err := saveSoundReplaceProgress(task.ID, domain.TaskStatusRunning, job, nil, ""); err != nil {
		return err
	}

	concatFiles := make([]string, 0, len(genRecords)*2)
	cursor := int64(0)
	for i, rec := range genRecords {
		if rec.Start > cursor {
			silence := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_silence_%d.wav", stamp, i))
			if err := createSilenceAudio(silence, rec.Start-cursor); err != nil {
				return err
			}
			concatFiles = append(concatFiles, silence)
		}
		concatFiles = append(concatFiles, rec.Audio)
		cursor = rec.End
	}
	if endMs := toInt64(asMap(job["SoundAsr"])["duration"]); endMs > cursor {
		silence := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_silence_end.wav", stamp))
		if err := createSilenceAudio(silence, endMs-cursor); err != nil {
			return err
		}
		concatFiles = append(concatFiles, silence)
	}

	combinedWav := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_combined.wav", stamp))
	if err := ffmpegConcatAudio(concatFiles, combinedWav); err != nil {
		return err
	}
	combinedMp3 := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_combined.mp3", stamp))
	if err := ffmpegEncodeMp3(combinedWav, combinedMp3); err != nil {
		return err
	}
	videoOutput := filepath.Join(persistDir, fmt.Sprintf("sound_replace_%d_output.mp4", stamp))
	if err := ffmpegReplaceVideoAudio(cfg.Video, combinedMp3, videoOutput); err != nil {
		return err
	}

	jobCombine["status"] = "success"
	jobCombine["audio"] = combinedMp3
	jobCombine["file"] = videoOutput
	job["Combine"] = jobCombine
	job["step"] = "End"
	combineConfirm := asMap(job["CombineConfirm"])
	combineConfirm["status"] = "success"
	job["CombineConfirm"] = combineConfirm

	result := map[string]any{
		"url":     videoOutput,
		"audio":   combinedMp3,
		"records": genRecords,
	}
	return saveSoundReplaceProgress(task.ID, domain.TaskStatusSuccess, job, result, "")
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
		return nil, errs.New("asr records not found")
	}
	records := make([]*soundReplaceRecord, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		segments := asArray(m["segments"])
		if len(segments) == 0 {
			continue
		}
		for _, segment := range segments {
			start := toInt64(segment["start"])
			end := toInt64(segment["end"])
			text := asString(segment["text"])
			if end <= start {
				continue
			}
			records = append(records, &soundReplaceRecord{Text: text, Start: start, End: end})
		}

	}
	return records, nil
}

func parseSoundReplaceRecords(v any) ([]*soundReplaceRecord, error) {
	if v == nil {
		return nil, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var records []*soundReplaceRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
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
		return errs.New("sound generate result url empty")
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
		return runCommand(GetFFmpegPath(), "-y", "-i", input, "-t", targetSec, "-acodec", "pcm_s16le", output)
	}
	if actualMs < targetMs {
		return runCommand(GetFFmpegPath(), "-y", "-i", input, "-af", "apad", "-t", targetSec, "-acodec", "pcm_s16le", output)
	}
	return copyFile(input, output)
}

func GetFFmpegPath() string {
	name := "ffmpeg.exe"

	return filepath.Join(utils.GetExeDir(), "binary", name)
}
func GetFFprobePath() string {
	name := "ffprobe.exe"

	return filepath.Join(utils.GetExeDir(), "binary", name)
}
func ffmpegExtractAudio(video, output string) error {
	return runCommand(GetFFmpegPath(), "-y", "-i", video, "-vn", "-acodec", "pcm_s16le", output)
}

func ffmpegEncodeMp3(input, output string) error {
	return runCommand(GetFFmpegPath(), "-y", "-i", input, "-codec:a", "libmp3lame", "-q:a", "2", output)
}

func createSilenceAudio(output string, durationMs int64) error {
	if durationMs <= 0 {
		durationMs = 1
	}
	durSec := fmt.Sprintf("%.3f", float64(durationMs)/1000.0)
	return runCommand(GetFFmpegPath(), "-y", "-f", "lavfi", "-i", "anullsrc=r=16000:cl=mono", "-t", durSec, "-acodec", "pcm_s16le", output)
}

func ffmpegConcatAudio(files []string, output string) error {
	if len(files) == 0 {
		return errs.New("no audio files to concat")
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
	out, err := exec.Command(GetFFprobePath(), "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", file).CombinedOutput()
	if err != nil {
		return 0, errs.New(fmt.Sprintf("ffprobe failed: %w, output: %s", err, string(out)))
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
		return errs.New(fmt.Sprintf("%s failed: %w, output: %s", cmd, err, string(out)))
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
func asArray(v any) []map[string]any {
	if v == nil {
		return []map[string]any{}
	}

	// 1️⃣ 已经是目标类型
	if arr, ok := v.([]map[string]any); ok {
		return arr
	}

	// 2️⃣ []interface{}
	if arr, ok := v.([]any); ok {
		out := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			out = append(out, asMap(item))
		}
		return out
	}

	// 3️⃣ json string
	if s, ok := v.(string); ok {
		var out []map[string]any
		if json.Unmarshal([]byte(s), &out) == nil {
			return out
		}
	}

	// 4️⃣ json bytes
	if b, ok := v.([]byte); ok {
		var out []map[string]any
		if json.Unmarshal(b, &out) == nil {
			return out
		}
	}

	// 5️⃣ struct slice (最关键)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		out := make([]map[string]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, asMap(rv.Index(i).Interface()))
		}
		return out
	}

	return []map[string]any{}
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
