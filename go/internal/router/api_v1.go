package router

import "aigcpanel/go/internal/api"

func init() {
	group := router.Group("/api/v1")
	{
		app := group.Group("/app")
		{
			app.GET("/tasks", api.TaskList)
			app.POST("/tasks", api.TaskCreate)
			app.GET("/tasks/:id", api.TaskGet)
			app.PATCH("/tasks/:id", api.TaskUpdate)
			app.DELETE("/tasks/:id", api.TaskDelete)
		}
		sound := group.Group("/sound")
		{
			sound.POST("/tts", api.SoundTTSCreate)
			sound.GET("/tts/:id", api.SoundTTSGet)
		}
	}
}
