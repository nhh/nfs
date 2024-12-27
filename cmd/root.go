package cmd

import (
	"github.com/spf13/cobra"
	"nfs/internal/config"
	"nfs/internal/pod"
	"nfs/internal/tui"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "nfs",
	Short: "Sync files on change to pods",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.Parse()
		for _, cnf := range config.WatchConfig {
			syncer := pod.NewSyncer(cnf, config.PodConfig, time.Duration(config.Interval)*time.Millisecond)
			syncer.AddOnUpdateListener(tui.GetUpdateChannel())
			syncer.AddOnErrorListener(tui.GetErrorChannel())
			syncer.StartSyncing()
		}
		tui.DisplayApp(config)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
