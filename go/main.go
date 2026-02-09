package main

import (
	"context"
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/router"
	"xiacutai-server/internal/service"
)

func main() {
	sqllite.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.StartTaskScheduler(ctx)

	router.Run()

}
