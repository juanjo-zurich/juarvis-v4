package watcher

import (
	"context"
	"fmt"
	"juarvis/pkg/output"
	"juarvis/pkg/snapshot"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	config    WatcherConfig
	fsWatcher *fsnotify.Watcher
	debouncer *Debouncer
}

func NewWatcher(cfg WatcherConfig) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creando fsnotify watcher: %w", err)
	}

	return &Watcher{
		config:    cfg,
		fsWatcher: fsWatcher,
		debouncer: NewDebouncer(cfg.DebounceMs),
	}, nil
}

func (w *Watcher) Start(ctx context.Context) error {
	for _, dir := range w.config.WatchDirs {
		if err := w.addRecursive(dir); err != nil {
			return err
		}
	}

	output.Success("Watching for changes in %s...", w.config.WatchDirs[0])
	output.Info("Press Ctrl+C to stop.")

	changeCount := 0
	lastSnapshotTime := time.Now()

	go func() {
		for batch := range w.debouncer.Events() {
			changeCount += len(batch)
			EvaluateFileChanges(batch)

			if !w.config.NoAutoSnapshot && changeCount >= w.config.AutoSnapshotThreshold {
				if time.Since(lastSnapshotTime) > 2*time.Minute {
					output.Info("Auto-snapshot: %d files changed", changeCount)
					snapshot.CreateSnapshot(fmt.Sprintf("auto-watch-%d", time.Now().Unix()))
					changeCount = 0
					lastSnapshotTime = time.Now()
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return w.Stop()
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}
			if w.config.ShouldIgnore(event.Name) {
				continue
			}

			eventType := "unknown"
			if event.Op&fsnotify.Create == fsnotify.Create {
				eventType = "create"
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.addRecursive(event.Name)
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				eventType = "write"
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				eventType = "remove"
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				eventType = "rename"
			}

			if eventType != "unknown" {
				w.debouncer.Add(event.Name, eventType)
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			output.Warning("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) Stop() error {
	output.Info("Stopping watcher...")
	w.debouncer.Stop()
	return w.fsWatcher.Close()
}

func GetFileScore(path string) int {
	score := 0
	ext := filepath.Ext(path)

	highPriorityExts := map[string]int{
		".go": 50, ".ts": 50, ".tsx": 50, ".js": 50, ".jsx": 50,
		".py": 45, ".rs": 45, ".java": 40,
		".sql": 40, ".graphql": 40, ".proto": 35,
		".yaml": 30, ".yml": 30, ".json": 30,
		".toml": 25,
	}

	if v, ok := highPriorityExts[ext]; ok {
		score += v
	}

	baseName := filepath.Base(path)
	if baseName == "go.mod" || baseName == "go.sum" ||
		baseName == "package.json" || baseName == "tsconfig.json" ||
		baseName == "Dockerfile" || baseName == "Makefile" {
		score += 25
	}

	dirParts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range dirParts {
		if part == "vendor" || part == "node_modules" || part == ".git" || part == "dist" {
			score -= 30
		}
		if part == "internal" || part == "pkg" || part == "cmd" {
			score += 10 * (len(dirParts) - i)
		}
	}

	return score
}

func ShouldSkip(path string, score int) bool {
	skipPatterns := []string{
		"/vendor/", "/node_modules/", "/.git/", "/dist/",
		"/ Coverage/", "/_test.go", "_test.go",
		"/.cache/", "/.tmp/", "/.生成的/",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	if score < 10 {
		return true
	}

	return false
}

func (w *Watcher) addRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if w.config.ShouldIgnore(path + "/") {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(path)
		}

		score := GetFileScore(path)
		if ShouldSkip(path, score) {
			return nil
		}

		if score >= 150 {
			output.Info("High-priority file detected (score %d): %s", score, path)
			snapshot.CreateSnapshot(fmt.Sprintf("immediate-%s", filepath.Base(path)))
		}

		return nil
	})
}
