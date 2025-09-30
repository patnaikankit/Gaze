// internal/watcher/watcher.go
package watcher

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/patnaikankit/Gaze/internal/config"
	"github.com/patnaikankit/Gaze/pkg/logger"
)

type Watcher struct {
	cfg          *config.Config
	fsWatcher    *fsnotify.Watcher
	eventCh      chan struct{}
	stopCh       chan struct{}
	debouncing   bool
	lastActivity time.Time
}

// NewWatcher initializes a new Watcher instance.
func NewWatcher(eventCh chan struct{}, cfg *config.Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		cfg:       cfg,
		fsWatcher: fsWatcher,
		eventCh:   eventCh,
		stopCh:    make(chan struct{}),
	}, nil
}

// Start begins watching the configured directory recursively.
func (w *Watcher) Start() error {
	if err := w.addRecursive(w.cfg.WatchDir); err != nil {
		return err
	}

	go w.processEvents()
	logger.Info("File watcher started on: %s", w.cfg.WatchDir)
	return nil
}

// addRecursive walks through the directory and adds subdirs to the watcher.
func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Apply ignore patterns
		for _, pattern := range w.cfg.IgnorePattern {
			matched, _ := filepath.Match(pattern, path)
			if matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			if err := w.fsWatcher.Add(path); err != nil {
				return err
			}
			logger.Debug("Watching directory: %s", path)
		}
		return nil
	})
}

// processEvents handles fsnotify events and applies debouncing.
func (w *Watcher) processEvents() {
	for {
		select {
		case ev, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			w.handleFsEvent(ev)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			logger.Error("Watcher error: %v", err)

		case <-w.stopCh:
			logger.Warn("Stopping file watcher")
			return
		}
	}
}

func (w *Watcher) handleFsEvent(ev fsnotify.Event) {
	if !shouldTrack(ev.Name) {
		return
	}

	// Add new subdir dynamically
	if ev.Op&fsnotify.Create == fsnotify.Create {
		if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
			_ = w.fsWatcher.Add(ev.Name)
			logger.Info("Added new directory to watcher: %s", ev.Name)
		}
	}

	if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
		w.triggerDebounced()
	}
}

func shouldTrack(file string) bool {
	extensions := []string{
		".go", ".mod", ".sum",
		".json", ".yaml", ".yml", ".toml", ".xml",
		".csv", ".txt", ".env", ".ini", ".conf",
	}
	for _, ext := range extensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

// triggerDebounced ensures changes are debounced before signaling a rebuild.
func (w *Watcher) triggerDebounced() {
	now := time.Now()
	if w.debouncing {
		w.lastActivity = now
		return
	}

	w.debouncing = true
	w.lastActivity = now

	go func(start time.Time) {
		debounce := 300 * time.Millisecond
		for {
			timer := time.NewTimer(debounce)
			select {
			case <-timer.C:
				if time.Since(w.lastActivity) >= debounce || time.Since(start) > 2*time.Second {
					w.eventCh <- struct{}{}
					w.debouncing = false
					return
				}
			case <-w.stopCh:
				return
			}
		}
	}(now)
}

// Stop shuts down the watcher gracefully.
func (w *Watcher) Stop() {
	close(w.stopCh)
	_ = w.fsWatcher.Close()
}
