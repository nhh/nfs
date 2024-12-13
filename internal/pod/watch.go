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

func (syncer *syncerImpl) setupWatcher() {
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

	for _, watchConfig := range syncer.cnf.WatchConfig {

		err = doublestar.GlobWalk(fsys, watchConfig.Pattern, func(path string, d fs.DirEntry) error {
			for _, exclude := range watchConfig.Excludes {
				if strings.Contains(path, exclude) {
					return doublestar.SkipDir
				}
			}

			// Verarbeiten
			filesToWatch = append(filesToWatch, path)

			return nil
		})

		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	slices.Sort(filesToWatch)
	slices.Compact(filesToWatch)

	fmt.Println()

	fmt.Printf("Watching %d files\n", len(filesToWatch))

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
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}
