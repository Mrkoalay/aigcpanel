package api

import (
	"aigcpanel/go/internal/service"
	"github.com/gin-gonic/gin"
)

func ModelAdd(ctx *gin.Context) {
	var req struct {
		ConfigPath string `json:"config_path"`
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
