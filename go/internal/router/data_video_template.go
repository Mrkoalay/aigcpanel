package router

import "xiacutai-server/internal/api"

func init() {
	group := router.Group("/datavideotemplate")
	{
		group.POST("/add", api.DataVideoTemplateCreate)
		group.POST("/list", api.DataVideoTemplateList)
		group.POST("/update", api.DataVideoTemplateUpdate)
		group.POST("/delete", api.DataVideoTemplateDelete)
	}
}
