package api

import (
	"strings"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"
	"xiacutai-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type dataVideoTemplateCreateRequest struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath"`
}

type dataVideoTemplateListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type dataVideoTemplateUpdateRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func DataVideoTemplateCreate(ctx *gin.Context) {
	var req dataVideoTemplateCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	info, err := resolveVideoAndInfo(req.FilePath)
	if err != nil {
		Err(ctx, err)
		return
	}

	record := domain.DataVideoTemplateModel{
		Name:  strings.TrimSpace(req.Name),
		Video: req.FilePath,
		Info:  info,
	}
	created, err := service.DataVideoTemplate.Create(record)
	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{"data": created})
}

func DataVideoTemplateList(ctx *gin.Context) {
	var req dataVideoTemplateListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}

	list, err := service.DataVideoTemplate.List(service.DataVideoTemplateFilters{
		Page: req.Page,
		Size: req.Size,
	})
	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{"data": list})
}

func DataVideoTemplateUpdate(ctx *gin.Context) {
	var req dataVideoTemplateUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}

	updates := map[string]any{
		"name": strings.TrimSpace(req.Name),
	}

	template, err := service.DataVideoTemplate.Update(req.ID, updates)
	if err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{"data": template})
}

func DataVideoTemplateDelete(ctx *gin.Context) {
	var req taskOperateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	if req.ID <= 0 {
		Err(ctx, errs.ParamError)
		return
	}

	if err := service.DataVideoTemplate.Delete(req.ID); err != nil {
		Err(ctx, err)
		return
	}

	OK(ctx, gin.H{"data": req.ID})
}

func resolveVideoAndInfo(filePath string) (string, error) {
	filePath = strings.TrimSpace(filePath)
	info, err := utils.ProbeVideoInfo(filePath)
	if err != nil {
		return "", err
	}
	return info, nil
}
