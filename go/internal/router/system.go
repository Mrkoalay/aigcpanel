package router

import (
	"xiacutai-server/internal/api"
)

func init() {
	group := router.Group("/")
	{
		group.POST("/init", api.SysInit)

	}
}
