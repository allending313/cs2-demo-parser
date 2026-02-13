package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	cs2demoparser "github.com/allending313/cs2-demo-parser"
	"github.com/allending313/cs2-demo-parser/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	port := envOrDefault("PORT", "3001")

	webFS, err := fs.Sub(cs2demoparser.WebFS, "web/dist")
	if err != nil {
		logger.Error("failed to create web sub-filesystem", "error", err)
		os.Exit(1)
	}

	mapsFS, err := fs.Sub(cs2demoparser.MapsFS, "assets/maps")
	if err != nil {
		logger.Error("failed to create maps sub-filesystem", "error", err)
		os.Exit(1)
	}

	srv, err := server.New(server.Config{
		UploadDir: envOrDefault("UPLOAD_DIR", "./data/uploads"),
		MatchDir:  envOrDefault("MATCH_DIR", "./data/matches"),
		WebFS:     webFS,
		MapsFS:    mapsFS,
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
