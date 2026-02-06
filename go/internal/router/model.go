package router

import (
	"aigcpanel/go/internal/api"
)

func init() {
	group := router.Group("/model")
	{
		group.POST("/add", api.ModelAdd)
		group.POST("/list", api.ModelList)
		group.POST("/setting", api.ModelSetting)
		group.POST("/delete", api.ModelDelete)
	}
}
