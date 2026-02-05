package router

import (
	"aigcpanel/go/internal/api"
)

func init() {
	group := router.Group("/server")
	{
		group.POST("/add", api.ServerAdd)
	}
}
