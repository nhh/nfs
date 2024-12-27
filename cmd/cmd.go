package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"time"
)

func init() {
	rootCmd.AddCommand(cmdCmd)
}

var cmdCmd = &cobra.Command{
	Use:   "cmd",
	Short: "Sync files on change to pods",
	Run: func(cobraCmd *cobra.Command, args []string) {
		fmt.Println("Running command")

		pr, pw := io.Pipe()
		defer pw.Close() // Pipe Writer am Ende schließen, um Deadlocks zu vermeiden

		cmd := exec.Command("/bin/bash")

		cmd.Stdin = pr
		cmd.Stdout = os.Stdout

		go func() {
			for {
				time.Sleep(1 * time.Second)
				pw.Write([]byte("echo 1\n"))
			}
		}()

		err := cmd.Run()

		if err != nil {
			panic(err)
		}
	},
}
