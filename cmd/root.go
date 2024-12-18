package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"nfs/internal/config"
	"nfs/internal/pod"
	"sync"
	"time"
)

var wg sync.WaitGroup

var rootCmd = &cobra.Command{
	Use:   "nfs",
	Short: "Sync files on change to pods",
	Run: func(cmd *cobra.Command, args []string) {

		_ = cmd.Help()
		config := config.Parse()

		fmt.Println()

		for _, cnf := range config.WatchConfig {
			syncer := pod.NewSyncer(cnf, config.PodConfig, time.Duration(config.Interval)*time.Second)
			syncer.StartSyncing()
			wg.Add(1)
		}

		wg.Wait()

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
