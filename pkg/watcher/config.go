package watcher

import (
	"juarvis/pkg/config"
	"path/filepath"
	"strings"
)

type WatcherConfig struct {
	DebounceMs            int
	AutoSnapshotThreshold int
	NoAutoSnapshot        bool
	IgnorePatterns        []string
	WatchDirs             []string
}

func DefaultWatcherConfig(rootPath string) WatcherConfig {
	return WatcherConfig{
		DebounceMs:            500,
		AutoSnapshotThreshold: 5,
		NoAutoSnapshot:        false,
		IgnorePatterns: []string{
			".git/",
			config.JuarDir + "/",
			config.JuarvisDir + "/",
			config.JuarvisPluginDir + "/",
			"node_modules/",
			"vendor/",
			"__pycache__/",
			".DS_Store",
			".tmp",
			".swp",
		},
		WatchDirs: []string{rootPath},
	}
}

func (c *WatcherConfig) ShouldIgnore(path string) bool {
	for _, pattern := range c.IgnorePatterns {
		if strings.HasSuffix(pattern, "/") {
			if strings.Contains(path, pattern) || strings.HasPrefix(path, pattern) {
				return true
			}
		} else {
			if filepath.Base(path) == pattern || strings.HasSuffix(path, pattern) {
				return true
			}
		}
	}
	return false
}
