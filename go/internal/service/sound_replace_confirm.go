package service

import (
	"encoding/json"
	"sort"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/domain"
)

type SoundReplaceConfirmRecord struct {
	Text  string `json:"text"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
}

func SubmitSoundReplaceConfirm(taskID int64, records []SoundReplaceConfirmRecord) (domain.DataTaskModel, error) {
	task, err := DataTask.GetTask(taskID)
	if err != nil {
		return domain.DataTaskModel{}, err
	}
	if task.Biz != "SoundReplace" {
		return domain.DataTaskModel{}, errs.New("task is not sound replace")
	}
	if task.Status != domain.TaskStatusWait {
		return domain.DataTaskModel{}, errs.New("task status must be wait")
	}

	job := map[string]any{}
	if strings.TrimSpace(task.JobResult) != "" {
		if err := json.Unmarshal([]byte(task.JobResult), &job); err != nil {
			return domain.DataTaskModel{}, err
		}
	}
	if asString(job["step"]) != "Confirm" {
		return domain.DataTaskModel{}, errs.New("task is not waiting for confirm")
	}

	cleaned := make([]SoundReplaceConfirmRecord, 0, len(records))
	for _, rec := range records {
		text := strings.TrimSpace(rec.Text)
		if text == "" || rec.End <= rec.Start {
			continue
		}
		cleaned = append(cleaned, SoundReplaceConfirmRecord{Text: text, Start: rec.Start, End: rec.End})
	}
	if len(cleaned) == 0 {
		return domain.DataTaskModel{}, errs.New("confirm records empty")
	}
	sort.SliceStable(cleaned, func(i, j int) bool {
		if cleaned[i].Start == cleaned[j].Start {
			return cleaned[i].End < cleaned[j].End
		}
		return cleaned[i].Start < cleaned[j].Start
	})

	confirm := asMap(job["Confirm"])
	confirm["status"] = "success"
	confirm["records"] = cleaned
	job["Confirm"] = confirm
	gen := asMap(job["SoundGenerate"])
	gen["status"] = "queue"
	gen["records"] = nil
	job["SoundGenerate"] = gen
	combine := asMap(job["Combine"])
	combine["status"] = "queue"
	combine["audio"] = ""
	combine["file"] = ""
	job["Combine"] = combine
	combineConfirm := asMap(job["CombineConfirm"])
	combineConfirm["status"] = "queue"
	job["CombineConfirm"] = combineConfirm
	job["step"] = "SoundGenerate"

	jobRaw, err := json.Marshal(job)
	if err != nil {
		return domain.DataTaskModel{}, err
	}
	return DataTask.UpdateTask(taskID, map[string]any{
		"status":    domain.TaskStatusQueue,
		"statusMsg": "",
		"jobResult": string(jobRaw),
	})
}
