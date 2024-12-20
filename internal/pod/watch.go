package pod

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"
)

func (syncer *syncer) setupWatcher() {
	// Create new watcher.
	watcher, err := fsnotify.NewBufferedWatcher(1024)

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	syncer.watcher = watcher

	if err != nil {
		log.Fatal(err)
	}

	filesToWatch := make([]string, 0)

	fsys := os.DirFS(".")

	err = doublestar.GlobWalk(fsys, syncer.watchCnf.Pattern, func(path string, d fs.DirEntry) error {
		for _, exclude := range syncer.watchCnf.Excludes {
			if strings.Contains(path, exclude) {
				return doublestar.SkipDir
			}
		}

		// Verarbeiten
		filesToWatch = append(filesToWatch, path)

		return nil
	})

	if err != nil {
		for _, errorCh := range syncer.onErrorCallbacks {
			errorCh <- err.Error()
		}
	}

	slices.Sort(filesToWatch)
	slices.Compact(filesToWatch)

	for _, path := range filesToWatch {
		err := watcher.Add(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	for _, path := range filesToWatch {
		syncer.add(path)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				syncer.add(event.Name)
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
			//log.Println("error:", err)
		}
	}

}
