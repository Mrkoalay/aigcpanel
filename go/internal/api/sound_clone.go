package api

import (
	"encoding/json"
	"time"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"

	"github.com/gin-gonic/gin"
)

type SoundCloneCreateRequest struct {
	Text     string `json:"text"`
	PromptId int64  `json:"promptId"` // 声音克隆-声音ID
}

func SoundCloneCreate(ctx *gin.Context) {
	var req SoundCloneCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	modelList, _ := service.Model.ModelList(TypeSoundClone)
	if len(modelList) == 0 {
		Err(ctx, errs.New("没有找到模型"))
		return
	}

	param := map[string]any{}
	param["crossLingual"] = false
	param["_crossLingual"] = "跨语种复刻"
	param["speed"] = 1
	param["_speed"] = "语速"
	param["seed"] = 403048
	param["_seed"] = "随机种子"

	modelConfig := map[string]any{}

	model := modelList[0]
	serverKey := model.Key

	promptId := req.PromptId
	storageModel, _ := service.DataStorage.GetStorage(promptId)
	var promptContent PromptContent
	json.Unmarshal([]byte(storageModel.Content), &promptContent)
	modelConfig = map[string]any{
		"type":           TypeSoundClone,
		"cloneServerKey": serverKey,
		"cloneParam":     param,
		"text":           req.Text,
		"promptId":       promptId,
		"promptTitle":    storageModel.Title,
		"promptUrl":      promptContent.URL,
		"promptText":     promptContent.PromptText,
	}

	modelConfigRaw, err := json.Marshal(modelConfig)
	if err != nil {
		Err(ctx, err)
		return
	}

	paramRaw, err := json.Marshal(map[string]any{})
	if err != nil {
		Err(ctx, err)
		return
	}
	task := domain.DataTaskModel{
		Biz:           TypeBizMap[TypeSoundClone],
		Title:         req.Text,
		Status:        "queue",
		ServerName:    model.Name,
		ServerTitle:   model.Title,
		ServerVersion: model.Version,
		Param:         string(paramRaw),
		ModelConfig:   string(modelConfigRaw),
		JobResult:     "{}",
		Result:        "{}",
		Type:          1,
	}
	created, err := service.DataTask.CreateTask(task)
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": created,
	})
}

func SoundCloneList(ctx *gin.Context) {
	var req dataTaskListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	list, err := service.DataTask.ListTasks(sqllite.TaskFilters{
		Biz:  req.Biz,
		Page: req.Page,
		Size: req.Size,
	})
	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{
		"data": list,
	})
}

func SoundCloneCancel(ctx *gin.Context) {
	var req taskOperateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}
	current, err := service.DataTask.GetTask(req.ID)
	if err != nil {
		Err(ctx, err)
		return
	}
	switch current.Status {
	case domain.TaskStatusQueue, domain.TaskStatusWait:
	case domain.TaskStatusRunning:
		if err := service.CancelEasyServerTask(req.ID); err != nil {
			Err(ctx, err)
			return
		}
	default:
		Err(ctx, errs.New("任务状态不允许取消"))
		return
	}
	task, err := service.DataTask.UpdateTask(req.ID, map[string]any{
		"status":    domain.TaskStatusFail,
		"statusMsg": "cancelled",
		"endTime":   time.Now().UnixMilli(),
	})
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": task,
	})
}

func SoundCloneContinue(ctx *gin.Context) {
	var req taskOperateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}

	current, err := service.DataTask.GetTask(req.ID)
	if err != nil {
		Err(ctx, err)
		return
	}
	if current.Status == domain.TaskStatusFail {
		task, err := service.DataTask.UpdateTask(req.ID, map[string]any{
			"status":    domain.TaskStatusQueue,
			"statusMsg": "",
		})
		if err != nil {
			Err(ctx, err)
			return
		}
		OK(ctx, gin.H{
			"data": task,
		})
		return
	}

	OK(ctx, gin.H{})
}

func SoundCloneSoundReplaceConfirm(ctx *gin.Context) {
	var req soundReplaceConfirmRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 || len(req.Records) == 0 {
		Err(ctx, errs.ParamError)
		return
	}

	task, err := service.SubmitSoundReplaceConfirm(req.ID, req.Records)
	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{"data": task})
}

func SoundCloneDelete(ctx *gin.Context) {
	var req taskOperateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}
	current, err := service.DataTask.GetTask(req.ID)
	if current.Status == domain.TaskStatusRunning {
		Err(ctx, errs.New("不可删除运行中任务"))
		return
	}
	if err != nil {
		Err(ctx, err)
		return
	}

	if err := service.DataTask.DeleteTask(req.ID); err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": req.ID,
	})
}
func SoundCloneUpdate(ctx *gin.Context) {
	var req taskUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}

	title := req.Title
	updateMap := map[string]any{
		"title": title,
	}
	task, err := service.DataTask.UpdateTask(req.ID, updateMap)

	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{
		"data": task,
	})
}
