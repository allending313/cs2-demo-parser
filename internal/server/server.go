package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	models "github.com/allending313/cs2-demo-parser/internal/model"
	"github.com/allending313/cs2-demo-parser/internal/parser"
	"github.com/allending313/cs2-demo-parser/internal/util"
)

type Server struct {
	mux        *http.ServeMux
	uploadDir  string
	matchDir   string
	webFS      fs.FS
	mapsFS     fs.FS
	mapConfigs map[string]*models.MapConfig
	jobs       *models.JobStore
	logger     *slog.Logger
}

type Config struct {
	UploadDir string
	MatchDir  string
	WebFS     fs.FS
	MapsFS    fs.FS
}

func New(cfg Config, logger *slog.Logger) (*Server, error) {
	for _, dir := range []string{cfg.UploadDir, cfg.MatchDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	mapConfigs, err := models.LoadMapConfigs(cfg.MapsFS, "configs")
	if err != nil {
		logger.Warn("failed to load map configs, continuing without them", "error", err)
		mapConfigs = make(map[string]*models.MapConfig)
	}

	s := &Server{
		mux:        http.NewServeMux(),
		uploadDir:  cfg.UploadDir,
		matchDir:   cfg.MatchDir,
		webFS:      cfg.WebFS,
		mapsFS:     cfg.MapsFS,
		mapConfigs: mapConfigs,
		jobs:       models.NewJobStore(),
		logger:     logger,
	}

	s.routes()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/parse", s.handleParse)
	s.mux.HandleFunc("GET /api/match/{id}/status", s.handleMatchStatus)
	s.mux.HandleFunc("GET /api/match/{id}", s.handleGetMatch)
	s.mux.HandleFunc("GET /api/maps/{name}/radar.png", s.handleMapRadar)
	s.mux.HandleFunc("GET /api/maps", s.handleListMaps)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	// Serve React SPA from embedded filesystem
	fileServer := http.FileServer(http.FS(s.webFS))
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for paths that don't match a file
		path := r.URL.Path
		if path != "/" {
			name := strings.TrimPrefix(path, "/")
			if _, err := fs.Stat(s.webFS, name); err != nil {
				index, err := fs.ReadFile(s.webFS, "index.html")
				if err != nil {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html")
				w.Write(index)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	// 500MB limit
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	if err := r.ParseMultipartForm(500 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("demo")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'demo' file field"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".dem") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file must be a .dem file"})
		return
	}

	id := util.GenerateID()
	uploadPath := filepath.Join(s.uploadDir, id+".dem")

	dst, err := os.Create(uploadPath)
	if err != nil {
		s.logger.Error("failed to create upload file", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		s.logger.Error("failed to save upload", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	dst.Close()

	job := s.jobs.Create(id)
	go s.parseInBackground(id, uploadPath)

	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) parseInBackground(id, demoPath string) {
	s.logger.Info("starting parse", "id", id)

	match, err := parser.ParseDemo(demoPath, id, func(progress float32) {
		s.jobs.SetProgress(id, progress)
	})

	os.Remove(demoPath)

	if err != nil {
		s.logger.Error("parse failed", "id", id, "error", err)
		s.jobs.Fail(id, err)
		return
	}

	if cfg, ok := s.mapConfigs[match.Map]; ok {
		match.MapConfig = cfg
	}

	matchPath := filepath.Join(s.matchDir, id+".json")
	if err := writeMatchJSON(matchPath, match); err != nil {
		s.logger.Error("failed to write match JSON", "id", id, "error", err)
		s.jobs.Fail(id, fmt.Errorf("writing match data: %w", err))
		return
	}

	s.jobs.Complete(id)
	s.logger.Info("parse complete", "id", id, "map", match.Map, "rounds", len(match.Rounds))
}

func (s *Server) handleMatchStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.jobs.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "match not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if job, ok := s.jobs.Get(id); ok && job.Status != models.JobStatusReady {
		writeJSON(w, http.StatusOK, job)
		return
	}

	matchPath := filepath.Join(s.matchDir, id+".json")
	data, err := os.ReadFile(matchPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "match not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gz.Write(data)
		return
	}

	w.Write(data)
}

func (s *Server) handleMapRadar(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}

	data, err := fs.ReadFile(s.mapsFS, "radars/"+name+".png")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func (s *Server) handleListMaps(w http.ResponseWriter, _ *http.Request) {
	maps := make([]map[string]string, 0, len(s.mapConfigs))
	for name, cfg := range s.mapConfigs {
		maps = append(maps, map[string]string{
			"name":        name,
			"displayName": cfg.DisplayName,
		})
	}
	writeJSON(w, http.StatusOK, maps)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeMatchJSON(path string, match *models.Match) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(match)
}
