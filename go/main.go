package main

import (
	"xiacutai-server/internal/component/sqllite"
	"xiacutai-server/internal/router"
)

func main() {
	sqllite.Init()
	router.Run()

}
