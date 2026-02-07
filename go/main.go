package main

import (
	"aigcpanel/go/internal/component/sqllite"
	"aigcpanel/go/internal/router"
)

func main() {
	sqllite.Init()
	router.Run()

}
