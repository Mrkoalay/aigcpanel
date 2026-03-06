package router

import "xiacutai-server/internal/api"

func init() {
	soudMediaGroup := router.Group("/sound/media")
	// 音色管理
	{
		soudMediaGroup.POST("/add", api.SoundMediaCreate)
		soudMediaGroup.POST("/list", api.SoundMediaList)
		soudMediaGroup.POST("/delete", api.SoundMediaDelete)
		soudMediaGroup.POST("/update", api.SoundMediaUpdate)
	}

	// 声音克隆
	soudCloneGroup := router.Group("/sound/clone")
	{

		soudCloneGroup.POST("/add", api.SoundCloneCreate)
		soudCloneGroup.POST("/list", api.SoundCloneList)
		soudCloneGroup.POST("/cancel", api.SoundCloneCancel)
		soudCloneGroup.POST("/continue", api.SoundCloneContinue)
		soudCloneGroup.POST("/delete", api.SoundCloneDelete)
		soudCloneGroup.POST("/update", api.SoundCloneUpdate)
		soudCloneGroup.POST("/sound-replace/confirm", api.SoundCloneSoundReplaceConfirm)

	}
}
