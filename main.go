package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"
	"smpark.in/paste"
	"smpark.in/site"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := paste.NewStore()
	go paste.StartExpiryWorker(ctx, store, 30*time.Second)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           newMux(store),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	slog.Info("listening", "addr", *addr)

	select {
	case err := <-errc:
		slog.Error("server error", "err", err)
		os.Exit(1)
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}

func newMux(store *paste.Store) *http.ServeMux {
	mux := http.NewServeMux()

	rl := paste.NewRateLimiter(rate.Every(30*time.Second), 1)
	h := paste.NewHandlers(store)
	mux.HandleFunc("POST /api/paste", rl.Middleware(h.Create))
	mux.HandleFunc("GET /api/paste/{id}", h.Get)
	mux.HandleFunc("GET /api/paste/{id}/raw", h.GetRaw)

	site.Register(mux)
	return mux
}
