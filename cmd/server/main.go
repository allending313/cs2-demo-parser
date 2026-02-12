package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/allending313/cs2-demo-parser/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	port := envOrDefault("PORT", "3001")

	srv, err := server.New(server.Config{
		UploadDir: envOrDefault("UPLOAD_DIR", "./data/uploads"),
		MatchDir:  envOrDefault("MATCH_DIR", "./data/matches"),
		MapsDir:   envOrDefault("MAPS_DIR", "./assets/maps"),
		WebDir:    envOrDefault("WEB_DIR", "./web/dist"),
	}, logger)
	if err != nil {
		logger.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}

	handler := server.CORSMiddleware(srv)

	logger.Info("server starting", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
