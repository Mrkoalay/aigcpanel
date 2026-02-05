package main

import (
	"log"
	"net/http"
	"os"

	"aigcpanel/internal/core"
	"aigcpanel/internal/platform/db"
)

func main() {
	addr := env("AIGCPANEL_ADDR", ":8080")
	dsn := env("AIGCPANEL_DSN", "data/aigcpanel.json")

	database, err := db.OpenFileDB(dsn)
	if err != nil {
		log.Fatalf("open file db failed: %v", err)
	}

	mux := http.NewServeMux()
	h := core.NewHTTPHandler(core.NewRepository(database))
	h.Register(mux)

	log.Printf("aigcpaneld listening at %s, db=%s", addr, dsn)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
