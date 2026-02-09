package router

import "xiacutai-server/internal/api"

func init() {
	group := router.Group("/datatask")
	{

		group.POST("/add", api.DataTaskCreate)
		group.POST("/list", api.DataTaskList)

	}
}
