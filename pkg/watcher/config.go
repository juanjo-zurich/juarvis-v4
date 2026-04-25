package watcher

import (
	"juarvis/pkg/config"
	"path/filepath"
	"strings"
	"time"
)

type WatcherConfig struct {
	DebounceMs            int
	AutoSnapshotThreshold int
	NoAutoSnapshot        bool
	QuietMode             bool
	SummaryInterval       time.Duration
	Verbose               bool
	IgnorePatterns        []string
	WatchDirs             []string
	AutoRestart           bool
	MaxRetries            int
	BaseRetryDelay        time.Duration
	MaxRetryDelay         time.Duration
}

func DefaultWatcherConfig(rootPath string) WatcherConfig {
	return WatcherConfig{
		DebounceMs:            500,
		AutoSnapshotThreshold: 5,
		NoAutoSnapshot:        false,
		QuietMode:             true,
		SummaryInterval:       5 * time.Minute,
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
		WatchDirs:      []string{rootPath},
		AutoRestart:    true,
		MaxRetries:     5,
		BaseRetryDelay: 1 * time.Second,
		MaxRetryDelay:  60 * time.Second,
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
