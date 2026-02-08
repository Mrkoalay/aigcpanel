package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"xiacutai-server/internal/service"
)

func ModelAdd(ctx *gin.Context) {
	var req struct {
		ConfigPath string `json:"configPath"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Err(ctx, err)
		return
	}
	out, err := service.Model.ModelAdd(req.ConfigPath)
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": out,
	})
}
func ModelList(ctx *gin.Context) {

	out, err := service.Model.ModelList()
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": out,
	})
}

type UpdateModelSettingReq struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Setting map[string]any `json:"setting"`
}

func ModelSetting(c *gin.Context) {

	var req UpdateModelSettingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	err := service.Model.ModelUpdateSetting(req.Name, req.Version, req.Setting)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	OK(c, gin.H{})
}

type deleteReq struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ModelDelete(c *gin.Context) {

	var req deleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	if err := service.Model.ModelDelete(req.Name, req.Version); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	OK(c, gin.H{})
}
