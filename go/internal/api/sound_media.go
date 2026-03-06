package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"
	"xiacutai-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type SoundMediaCreateRequest struct {
	Title      string `json:"title"`
	FilePath   string `json:"filePath"`
	PromptText string `json:"promptText"`
}
type SoundMediaListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
type SoundMediaUpdateRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func SoundMediaGet(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	record, err := service.DataStorage.GetStorage(id)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, record)
}

type SoundMediaResp struct {
	ID        int64              `json:"ID"`
	CreatedAt int64              `json:"CreatedAt"`
	UpdatedAt int64              `json:"UpdatedAt"`
	Sort      int64              `json:"Sort"`
	Biz       string             `json:"Biz"`
	Title     string             `json:"Title"`
	Content   SoundPromptContent `json:"content"`
}
type SoundPromptContent struct {
	AsrStatus  string `json:"asrStatus"`
	PromptText string `json:"promptText"`
	URL        string `json:"url"`
}

func SoundMediaList(ctx *gin.Context) {
	var req storageListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	list, err := service.DataStorage.ListStorages(service.StorageFilters{
		Biz:  BizSoundPrompt,
		Page: req.Page,
		Size: req.Size,
	})
	if err != nil {
		Err(ctx, err)
		return
	}

	var resp []SoundMediaResp

	for _, item := range list {
		var content SoundPromptContent
		if item.Content != "" {
			_ = json.Unmarshal([]byte(item.Content), &content)
		}

		resp = append(resp, SoundMediaResp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Sort:      item.Sort,
			Biz:       item.Biz,
			Title:     item.Title,
			Content:   content,
		})
	}

	OK(ctx, gin.H{
		"data": resp,
	})
}

func SoundMediaCreate(ctx *gin.Context) {
	var req SoundMediaCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	// 复制文件并且重命名
	// 复制文件到 storage
	newPath, err := utils.CopyToStorage(req.FilePath)
	if err != nil {
		Err(ctx, err)
		return
	}
	promptText := strings.TrimSpace(req.PromptText)
	contentRaw, err := buildSoundPromptStorageContent(newPath, promptText, "")
	if err != nil {
		Err(ctx, err)
		return
	}
	record := domain.DataStorageModel{
		Biz:     BizSoundPrompt,
		Title:   req.Title,
		Content: string(contentRaw),
	}
	created, err := service.DataStorage.CreateStorage(record)
	if err != nil {
		Err(ctx, err)
		return
	}
	if promptText == "" {
		go recognizeAndUpdateSoundPrompt(created.ID, newPath)
	}
	OK(ctx, gin.H{
		"data": created,
	})
}

func SoundMediaUpdate(ctx *gin.Context) {
	var req storageUpdateRequest
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
	task, err := service.DataStorage.UpdateStorage(req.ID, updateMap)

	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{
		"data": task,
	})
}

func SoundMediaDelete(ctx *gin.Context) {
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

	if err := service.DataStorage.DeleteStorage(req.ID); err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": req.ID,
	})
}

func SoundMediaClear(ctx *gin.Context) {
	biz := strings.TrimSpace(ctx.Query("biz"))
	if biz == "" {
		Err(ctx, errs.ParamError)
		return
	}
	if err := service.DataStorage.DeleteStoragesByBiz(biz); err != nil {
		Err(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
