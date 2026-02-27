package router

import (
	"xiacutai-server/internal/api"
)

func init() {
	group := router.Group("/")
	{
		group.GET("/sys_config", api.SysConfig)

	}
}
