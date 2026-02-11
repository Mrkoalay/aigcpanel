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

type taskCreateRequest struct {
	Text          string                 `json:"text"`
	Type          string                 `json:"type"` // 类型
	ServerKey     string                 `json:"serverKey"`
	Param         map[string]any         `json:"param"`
	SoundAsr      map[string]any         `json:"soundAsr"`
	SoundGenerate map[string]any         `json:"soundGenerate"`
	Extra         map[string]interface{} `json:"-"`
	PromptId      int64                  `json:"promptId"` // 声音克隆-声音ID
	Audio         string                 `json:"audio"`    // 语音转文字-声音文件
	Video         string                 `json:"video"`    // 声音替换-视频文件
}
type taskOperateRequest struct {
	ID int64 `json:"id"`
}
type taskUpdateRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

var TypeBizMap = map[string]string{
	"soundTts":     "SoundGenerate",
	"soundClone":   "SoundGenerate",
	"soundAsr":     "SoundAsr",
	"videoGen":     "VideoGen",
	"soundReplace": "SoundReplace",
}

type PromptContent struct {
	URL        string `json:"url"`
	PromptText string `json:"promptText"`
}

func DataTaskCreate(ctx *gin.Context) {
	var req taskCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	typeStr := req.Type
	modelConfig := map[string]any{}

	serverKey := req.ServerKey
	model := &domain.LocalModelConfigInfo{}
	if serverKey != "" {

		dbModel, err := service.Model.Get(serverKey)
		if err != nil {
			Err(ctx, err)
			return
		}
		model = dbModel
	}

	switch typeStr {
	case domain.FunctionSoundTts:

		modelConfig = map[string]any{
			"type":         typeStr,
			"ttsServerKey": serverKey,
			"ttsParam":     req.Param,
			"text":         req.Text,
		}
	case domain.FunctionSoundClone:

		promptId := req.PromptId
		storageModel, _ := service.DataStorage.GetStorage(promptId)
		var promptContent PromptContent
		json.Unmarshal([]byte(storageModel.Content), &promptContent)
		modelConfig = map[string]any{
			"type":           typeStr,
			"cloneServerKey": serverKey,
			"cloneParam":     req.Param,
			"text":           req.Text,
			"promptId":       promptId,
			"promptTitle":    storageModel.Title,
			"promptUrl":      promptContent.URL,
			"promptText":     promptContent.PromptText,
		}
	case domain.FunctionVideoGen:
		modelConfig = map[string]any{
			"type":      typeStr,
			"serverKey": serverKey,
			"video":     req.Video,
			"audio":     req.Audio,
			"param":     req.Param,
		}
	case domain.FunctionSoundReplace:

		// 补充soundAsr

		req.Param = map[string]any{}

		// 补充 soundGenerate
		cloneServerKey, _ := req.SoundGenerate["cloneServerKey"].(string)
		cloneModel, err := service.Model.Get(cloneServerKey)
		if err != nil {
			Err(ctx, err)
			return
		}
		req.SoundGenerate["serverName"] = cloneModel.Name
		req.SoundGenerate["serverTitle"] = cloneModel.Title
		req.SoundGenerate["serverVersion"] = cloneModel.Version

		promptId := req.SoundGenerate["promptId"].(float64)
		storageModel, err := service.DataStorage.GetStorage(int64(promptId))
		if err != nil {
			Err(ctx, err)
			return
		}
		var promptContent PromptContent
		json.Unmarshal([]byte(storageModel.Content), &promptContent)
		req.SoundGenerate["promptTitle"] = storageModel.Title
		req.SoundGenerate["promptUrl"] = promptContent.URL
		req.SoundGenerate["promptText"] = promptContent.PromptText

		modelConfig = map[string]any{
			"type":          typeStr,
			"video":         req.Video,
			"soundAsr":      req.SoundAsr,
			"soundGenerate": req.SoundGenerate,
		}
	case domain.FunctionSoundAsr:
		serverKey := req.ServerKey
		paramRaw, err := json.Marshal(map[string]any{})
		if err != nil {
			Err(ctx, err)
			return
		}
		modelConfig = map[string]any{
			"type":      typeStr,
			"serverKey": serverKey,
			"audio":     req.Audio,
			"param":     string(paramRaw),
		}
	default:
		Err(ctx, errs.New("不支持的功能"))
		return
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
		Biz:           TypeBizMap[typeStr],
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

type dataTaskListRequest struct {
	Biz  string `form:"biz"`
	Page int    `form:"page"`
	Size int    `form:"size"`
}

func DataTaskList(ctx *gin.Context) {
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

func DataTaskCancel(ctx *gin.Context) {
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

func DataTaskContinue(ctx *gin.Context) {
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

func DataTaskDelete(ctx *gin.Context) {
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
func DataTaskUpdate(ctx *gin.Context) {
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
