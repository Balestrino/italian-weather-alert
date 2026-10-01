package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Balestrino/italian-weather-alert/internal/frontend"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

func main() {
	if !run() {
		os.Exit(1)
	}
}

func run() bool {
	if len(os.Args) != 1 {
		slog.Error("unknown command")
		return false
	}
	origin := os.Getenv("IWA_PUBLIC_BACKEND_URL")
	if origin == "" {
		origin = "http://127.0.0.1:8080"
	}
	listen := os.Getenv("IWA_FRONTEND_LISTEN")
	if listen == "" {
		listen = "127.0.0.1:8082"
	}
	handler, err := frontend.Handler(origin)
	if err != nil {
		slog.Error("frontend configuration invalid")
		return false
	}
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		slog.Error("frontend listener unavailable")
		return false
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	slog.Info("public frontend started")
	if err := httpserver.Serve(ctx, ln, handler); err != nil {
		slog.Error("public frontend stopped unexpectedly")
		return false
	}
	return true
}
