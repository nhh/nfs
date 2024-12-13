package cmd

import (
	"github.com/spf13/cobra"
	"nfs/internal/config"
	"nfs/internal/pod"
)

var rootCmd = &cobra.Command{
	Use:   "nfs",
	Short: "Sync files on change to pods",
	Run: func(cmd *cobra.Command, args []string) {

		_ = cmd.Help()
		config := config.Parse()

		syncer := pod.NewSyncer(config)
		syncer.StartSyncing()

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
