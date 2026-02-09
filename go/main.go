package main

import (
	"context"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/router"
	"xiacutai-server/internal/service"
	"xiacutai-server/internal/utils"
)

func main() {
	sqllite.Init()
	utils.InitDirs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.StartTaskScheduler(ctx)

	router.Run()

}
