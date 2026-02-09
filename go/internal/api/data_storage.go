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

type storageCreateRequest struct {
	Title      string `json:"title"`
	FilePath   string `json:"filePath"`
	PromptText string `json:"promptText"`
}
type storageListRequest struct {
	Biz  string `form:"biz"`
	Page int    `form:"page"`
	Size int    `form:"size"`
}
type storageUpdateRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func DataStorageGet(ctx *gin.Context) {
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

func DataStorageList(ctx *gin.Context) {
	var req storageListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	list, err := service.DataStorage.ListStorages(service.StorageFilters{
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

func DataStorageSoundCreate(ctx *gin.Context) {
	var req storageCreateRequest
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
	contentMap := map[string]any{
		"url":        newPath,
		"promptText": req.PromptText,
	}

	contentRaw, err := json.Marshal(contentMap)
	record := domain.DataStorageModel{
		Biz:     "SoundPrompt",
		Title:   req.Title,
		Content: string(contentRaw),
	}
	created, err := service.DataStorage.CreateStorage(record)
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": created,
	})
}

func DataStorageUpdate(ctx *gin.Context) {
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

func DataStorageDelete(ctx *gin.Context) {
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

func DataStorageClear(ctx *gin.Context) {
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
