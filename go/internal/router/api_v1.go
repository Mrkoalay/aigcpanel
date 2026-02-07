package router

import "aigcpanel/go/internal/api"

func init() {
	group := router.Group("/sound/tts")
	{

		group.POST("/add", api.SoundTTSCreate)

	}
}
