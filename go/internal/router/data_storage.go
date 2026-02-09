package router

import "xiacutai-server/internal/api"

func init() {
	group := router.Group("/datastorage")
	{
		group.POST("/sound/add", api.DataStorageSoundCreate)
		group.POST("/list", api.DataStorageList)
		group.POST("/delete", api.DataStorageDelete)
		group.POST("/update", api.DataStorageUpdate)
	}
}
