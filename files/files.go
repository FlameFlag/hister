// SPDX-FileContributor: slowerloris <taylor@teukka.tech>
//
// SPDX-License-Identifier: AGPL-3.0-or-later
package files

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/asciimoo/hister/config"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// MatchesFilters reports whether a filename passes the given filetype, pattern, and exclude filters.
func MatchesFilters(name string, filetypes, patterns, excludes []string) bool {
	if len(excludes) > 0 {
		for _, pattern := range excludes {
			if matched, _ := filepath.Match(pattern, name); matched {
				return false
			}
		}
	}
	if len(filetypes) > 0 {
		ext := strings.TrimPrefix(filepath.Ext(name), ".")
		if !slices.Contains(filetypes, ext) {
			return false
		}
	}
	if len(patterns) > 0 {
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, name); matched {
				return true
			}
		}
		return false
	}
	return true
}

// Debounce so we don't spam the index as write events can file multiple times before closing a file after editing
const debounceTime = 200 * time.Millisecond

// findMatchingDir returns the Directory config whose expanded path contains filePath, or nil.
func findMatchingDir(dirs []config.Directory, filePath string) *config.Directory {
	for i := range dirs {
		dirPath := filepath.Clean(ExpandHome(dirs[i].Path))
		if strings.HasPrefix(filePath, dirPath+"/") || filePath == dirPath {
			return &dirs[i]
		}
	}
	return nil
}

func WatchDirectories(dirs []config.Directory, callback func(string)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start file watcher")
	}
	defer watcher.Close()

	go func() {
		log.Debug().Msg("Starting file watcher")
		debounced := make(map[string]time.Timer)
		go func() {
			for path, timer := range maps.All(debounced) {
				<-timer.C
				callback(path)
			}
		}()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					if dir := findMatchingDir(dirs, event.Name); dir != nil {
						if MatchesFilters(filepath.Base(event.Name), dir.Filetypes, dir.Patterns, dir.Excludes) {
							if debounceTimer, ok := debounced[event.Name]; ok {
								debounceTimer.Reset(debounceTime)
							} else {
								debounced[event.Name] = *time.NewTimer(debounceTime)
							}
						}
					}
				}
				if event.Has(fsnotify.Create) {
					st, err := os.Stat(event.Name)
					if err == nil {
						if st.IsDir() && !slices.Contains(watcher.WatchList(), event.Name) {
							watcher.Add(event.Name)
						} else if dir := findMatchingDir(dirs, event.Name); dir != nil {
							if MatchesFilters(filepath.Base(event.Name), dir.Filetypes, dir.Patterns, dir.Excludes) {
								callback(event.Name)
							}
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("Watcher failed to process event")
			}
		}
	}()
	for _, dir := range dirs {
		expanded := ExpandHome(dir.Path)
		err = watcher.Add(expanded)
		if err != nil {
			log.Error().Err(err).Str("path", expanded).Msg("Failed to add path to file watcher")
		}
		err = filepath.WalkDir(expanded, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				watcher.Add(path)
			}
			return nil
		})
	}
	<-make(chan struct{})
}
