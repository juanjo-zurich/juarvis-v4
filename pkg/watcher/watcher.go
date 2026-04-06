package watcher

import (
	"context"
	"fmt"
	"juarvis/pkg/output"
	"juarvis/pkg/snapshot"
	"os"
	"path/filepath"
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
		return nil
	})
}
