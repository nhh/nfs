package cmd

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"log"
	"nfs/internal/config/v1"
	"nfs/internal/pod"
)

var rootCmd = &cobra.Command{
	Use:   "nfs",
	Short: "Sync files on change to pods",
	Run: func(cmd *cobra.Command, args []string) {

		_ = cmd.Help()
		config := v1.Parse()

		// Create new watcher.
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		// Add a path.
		err = watcher.Add("/home/niklas-hanft/Projects/nfs/")

		if err != nil {
			log.Fatal(err)
		}

		syncer := pod.NewSyncer(config)
		syncer.StartWatching()

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
