package router

import "xiacutai-server/internal/api"

func init() {
	group := router.Group("/app/storages")
	{
		group.GET("", api.DataStorageList)
		group.POST("", api.DataStorageCreate)
		group.DELETE("", api.DataStorageClear)
		group.GET("/:id", api.DataStorageGet)
		group.PATCH("/:id", api.DataStorageUpdate)
		group.DELETE("/:id", api.DataStorageDelete)
	}
}
