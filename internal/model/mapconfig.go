package models

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

func LoadMapConfigs(fsys fs.FS, dir string) (map[string]*MapConfig, error) {
	configs := make(map[string]*MapConfig)

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("reading map configs dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var cfg MapConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		if cfg.RadarWidth == 0 {
			cfg.RadarWidth = 1024
		}
		if cfg.RadarHeight == 0 {
			cfg.RadarHeight = 1024
		}

		configs[cfg.Name] = &cfg
	}

	return configs, nil
}
