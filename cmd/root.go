package cmd

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"io/fs"
	"log"
	"nfs/internal/config"
	"nfs/internal/pod"
	"os"
	"slices"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "nfs",
	Short: "Sync files on change to pods",
	Run: func(cmd *cobra.Command, args []string) {

		_ = cmd.Help()
		config := config.Parse()

		// Create new watcher.
		watcher, err := fsnotify.NewBufferedWatcher(1024)
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		if err != nil {
			log.Fatal(err)
		}

		syncer := pod.NewSyncer(config)
		syncer.StartSyncing()

		filesToWatch := make([]string, 0)

		fsys := os.DirFS(".")

		for _, watchConfig := range config.WatchConfig {

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

		fmt.Printf("Setup watchers for %d files.\n", len(filesToWatch))

		for _, path := range filesToWatch {
			err := watcher.Add(path)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		fmt.Printf("Syncing initially...\n")

		for _, fileToWatch := range filesToWatch {
			syncer.Add(fileToWatch)
		}

		fmt.Printf("Listening for changes...\n")

		// Todo move this into pod syncer struct?
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					syncer.Add(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
