package router

import (
	"aigcpanel/go/internal/api"
)

func init() {
	group := router.Group("/model")
	{
		group.POST("/add", api.ModelAdd)
	}
}
