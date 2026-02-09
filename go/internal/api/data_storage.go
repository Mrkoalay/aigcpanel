package api

import (
	"net/http"
	"strconv"
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"

	"github.com/gin-gonic/gin"
)

type storageCreateRequest struct {
	ID        int64  `json:"id"`
	Biz       string `json:"biz"`
	Sort      int64  `json:"sort"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type storageUpdateRequest struct {
	Biz     *string `json:"biz"`
	Sort    *int64  `json:"sort"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
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
	biz := strings.TrimSpace(ctx.Query("biz"))
	list, err := service.DataStorage.ListStorages(biz)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, list)
}

func DataStorageCreate(ctx *gin.Context) {
	var req storageCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if strings.TrimSpace(req.Biz) == "" {
		Err(ctx, errs.ParamError)
		return
	}
	record := domain.DataStorageModel{
		Biz:       req.Biz,
		Sort:      req.Sort,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
	created, err := service.DataStorage.CreateStorage(record)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func DataStorageUpdate(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req storageUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	updates := map[string]any{}
	if req.Biz != nil {
		updates["biz"] = strings.TrimSpace(*req.Biz)
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	record, err := service.DataStorage.UpdateStorage(id, updates)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, record)
}

func DataStorageDelete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := service.DataStorage.DeleteStorage(id); err != nil {
		Err(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
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
