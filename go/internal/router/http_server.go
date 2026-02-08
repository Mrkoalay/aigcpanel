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
	addr := getenv("PORT", ":8088")
	//dsn := getenv("AIGCPANEL_DSN", "data/aigcpanel.json")
	router.Use(middleware.Cors())
	/*    st, err := store.NewJSONStore(dsn)
	      if err != nil {
	          log.Fatalf("open store: %v", err)
	      }

	      h := api.NewServer(app.NewService(st)).Routes()*/

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen:", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown:", zap.Error(err))
	}
	// catching ctx.Done(). timeout of 5 seconds.
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
