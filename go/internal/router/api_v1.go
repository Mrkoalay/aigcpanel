package router

import "xiacutai-server/internal/api"

func init() {
	group := router.Group("/sound/tts")
	{

		group.POST("/add", api.SoundTTSCreate)

	}
}
