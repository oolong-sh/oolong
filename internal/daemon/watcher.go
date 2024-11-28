package daemon

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oolong-sh/oolong/internal/config"
	"github.com/oolong-sh/oolong/internal/documents"
)

// Initialize and run file update watcher for notes directories
func runNotesDirsWatcher(dirs ...string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dirIgnores := config.IgnoredDirectories()

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			log.Printf("Error creating watcher on directory '%s': %v\n", dir, err)
			continue
		}

		// TODO: add oolong ignore system to blacklist certain subdirs/files
		if err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}

			// NOTE: this may not be the exact desired behavior for ignores
			// - this logic also needs to be replicated in the document reader
			if slices.Contains(dirIgnores, filepath.Base(path)) {
				return filepath.SkipDir
			}

			// TEST: this may need to add path as absolute to get correct results
			err = watcher.Add(path)
			if err != nil {
				return err
			}
			log.Println("Added watcher on", path)

			return nil
		}); err != nil {
			return err
		}
	}

	// watcher handler
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Watcher event channel returned bad result.")
				return errors.New("Invalid watcher errors channel value.")
			}

			switch {
			case event.Has(fsnotify.Write):
				log.Println("Modified file:", event.Name)

				// write event is sent on write start, wait 500ms for write to finish
				time.Sleep(500)

				// re-read document
				documents.ReadDocuments(event.Name)

				// TODO: add dedup timer to prevent multi-write calls

			case event.Has(fsnotify.Remove):
				log.Println("Removed file/directory", event.Name)
				// TODO: remove from state
				// - need to be careful with remove event as editors use it when writing files
				// - state removal needs to also remove ngrams
				// - should only trigger update on file deletions

			case event.Has(fsnotify.Create):
				log.Println("Created file/directory", event.Name)

				if info, err := os.Stat(event.Name); err == nil {
					if info.IsDir() {
						watcher.Add(event.Name)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("Invalid watcher errors channel value.")
			}
			log.Println("error:", err)
		}
	}
}
