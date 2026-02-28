package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
	"smpark.in/paste"
)

func main() {
	publicDir := "./public"
	spaIndex := filepath.Join(publicDir, "app", "index.html")

	store := paste.NewStore()

	go paste.StartExpiryWorker(context.Background(), store, 30*time.Second)

	rl := paste.NewRateLimiter(rate.Every(30*time.Second), 1)

	handlers := paste.NewHandlers(store)
	http.HandleFunc("POST /api/paste", rl.Middleware(handlers.Create))
	http.HandleFunc("GET /api/paste/{id}/raw", handlers.GetRaw)
	http.HandleFunc("GET /api/paste/{id}", handlers.Get)

	fs := http.FileServer(http.Dir(publicDir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(publicDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, spaIndex)
	})

	addr := ":8080"
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server error", "err", err)
	}
}
