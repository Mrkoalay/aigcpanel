package main

import (
	"log"
	"net/http"
	"os"

	"aigcpanel/go/internal/api"
	"aigcpanel/go/internal/app"
	"aigcpanel/go/internal/store"
)

func main() {
	addr := getenv("AIGCPANEL_ADDR", ":8080")
	dsn := getenv("AIGCPANEL_DSN", "data/aigcpanel.json")

	st, err := store.NewJSONStore(dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	h := api.NewServer(app.NewService(st)).Routes()
	log.Printf("aigcpanel-go listening on %s (dsn=%s)", addr, dsn)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
