package api

import (
	"aigcpanel/go/internal/service"
	"github.com/gin-gonic/gin"
)

func ServerAdd(ctx *gin.Context) {
	var req struct {
		ConfigPath string `json:"configPath"`
	}
	if err := ctx.ShouldBindJSON(req); err != nil {
		Err(ctx, err)
		return
	}
	out, err := service.Server.ServerAdd(req.ConfigPath)
	if err != nil {
		Err(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"data": out,
	})
}
