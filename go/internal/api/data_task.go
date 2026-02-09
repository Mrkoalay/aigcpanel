package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"

	"github.com/gin-gonic/gin"
)

type taskCreateRequest struct {
	Text      string                 `json:"text"`
	Type      string                 `json:"type"`
	ServerKey string                 `json:"serverKey"`
	Param     map[string]any         `json:"param"`
	Extra     map[string]interface{} `json:"-"`
}
type taskOperateRequest struct {
	ID int64 `json:"id"`
}

type soundGenerateRequest struct {
	Text           string         `json:"text"`
	Type           string         `json:"type"`
	ServerName     string         `json:"serverName"`
	ServerTitle    string         `json:"serverTitle"`
	ServerVersion  string         `json:"serverVersion"`
	TtsServerKey   string         `json:"ttsServerKey"`
	TtsParam       map[string]any `json:"ttsParam"`
	CloneServerKey string         `json:"cloneServerKey"`
	CloneParam     map[string]any `json:"cloneParam"`
	PromptID       int64          `json:"promptId"`
	PromptTitle    string         `json:"promptTitle"`
	PromptURL      string         `json:"promptUrl"`
	PromptText     string         `json:"promptText"`
}

var TypeBizMap = map[string]string{
	"soundTts": "SoundGenerate",
}

func DataTaskCreate(ctx *gin.Context) {
	var req taskCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		Err(ctx, errs.ParamError)
		return
	}
	serverKey := req.ServerKey
	model, err := service.Model.Get(serverKey)

	typeStr := req.Type
	modelConfig := map[string]any{}

	switch typeStr {
	case "soundTts":
		modelConfig = map[string]any{
			"type":         typeStr,
			"ttsServerKey": serverKey,
			"ttsParam":     req.Param,
			"text":         req.Text,
		}
	case "soundClone":
		//result, err = server.SoundClone(*functionData)
	case "videoGen":
		//result, err = server.VideoGen(*functionData)
	case "asr":
		//result, err = server.Asr(*functionData)
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
	ctx.JSON(http.StatusOK, created)
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

	ctx.JSON(http.StatusOK, gin.H{
		"list": list,
	})
}
