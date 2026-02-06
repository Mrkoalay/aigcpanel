package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/errs"
	"aigcpanel/go/internal/service"
	"github.com/gin-gonic/gin"
)

type soundTTSRequest struct {
	Text          string                 `json:"text"`
	ServerName    string                 `json:"serverName"`
	ServerTitle   string                 `json:"serverTitle"`
	ServerVersion string                 `json:"serverVersion"`
	TtsServerKey  string                 `json:"ttsServerKey"`
	TtsParam      map[string]any         `json:"ttsParam"`
	Extra         map[string]interface{} `json:"-"`
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

func SoundTTSCreate(ctx *gin.Context) {
	var req soundTTSRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if strings.TrimSpace(req.Text) == "" || strings.TrimSpace(req.ServerName) == "" {
		Err(ctx, errs.ParamError)
		return
	}
	modelConfig := map[string]any{
		"type":         "SoundTts",
		"ttsServerKey": req.TtsServerKey,
		"ttsParam":     req.TtsParam,
		"text":         req.Text,
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
	task := domain.AppTask{
		Biz:           "SoundGenerate",
		Title:         req.Text,
		Status:        "queue",
		ServerName:    req.ServerName,
		ServerTitle:   req.ServerTitle,
		ServerVersion: req.ServerVersion,
		Param:         string(paramRaw),
		ModelConfig:   string(modelConfigRaw),
		JobResult:     "{}",
		Result:        "{}",
		Type:          1,
	}
	created, err := service.Task.CreateTask(task)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func SoundGenerateCreate(ctx *gin.Context) {
	var req soundGenerateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	req.ServerName = strings.TrimSpace(req.ServerName)
	req.Type = strings.TrimSpace(req.Type)
	if req.Text == "" || req.ServerName == "" {
		Err(ctx, errs.ParamError)
		return
	}
	if req.Type == "" {
		req.Type = "SoundTts"
	}
	switch req.Type {
	case "SoundTts":
		if strings.TrimSpace(req.TtsServerKey) == "" {
			Err(ctx, errs.ParamError)
			return
		}
	case "SoundClone":
		if strings.TrimSpace(req.CloneServerKey) == "" || req.PromptID == 0 {
			Err(ctx, errs.ParamError)
			return
		}
	default:
		Err(ctx, errs.ParamError)
		return
	}
	modelConfig := map[string]any{
		"type":           req.Type,
		"ttsServerKey":   req.TtsServerKey,
		"ttsParam":       req.TtsParam,
		"cloneServerKey": req.CloneServerKey,
		"cloneParam":     req.CloneParam,
		"promptId":       req.PromptID,
		"promptTitle":    req.PromptTitle,
		"promptUrl":      req.PromptURL,
		"promptText":     req.PromptText,
		"text":           req.Text,
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
	now := time.Now().UnixMilli()
	task := domain.AppTask{
		Biz:           "SoundGenerate",
		Title:         req.Text,
		Status:        "queue",
		StartTime:     now,
		ServerName:    req.ServerName,
		ServerTitle:   req.ServerTitle,
		ServerVersion: req.ServerVersion,
		Param:         string(paramRaw),
		ModelConfig:   string(modelConfigRaw),
		JobResult:     "{}",
		Result:        "{}",
		Type:          1,
	}
	created, err := service.Task.CreateTask(task)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func SoundTTSGet(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Err(ctx, errs.ParamError)
		return
	}
	task, err := service.Task.GetTask(id)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}

func SoundGenerateGet(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Err(ctx, errs.ParamError)
		return
	}
	task, err := service.Task.GetTask(id)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}
