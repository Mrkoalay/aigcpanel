package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"xiacutai-server/internal/component/log"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xiacutai-server/internal/component/middleware"
)

var router = gin.Default()

func Run() {
	addr := getenv("PORT", "2021") // 默认端口 2021
	addr = ":" + addr
	router.Use(middleware.Cors())

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen:", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown:", zap.Error(err))
	}

	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}

	log.Info("Server exiting")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
