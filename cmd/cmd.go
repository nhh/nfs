package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
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

		cmd := exec.Command("/bin/bash")

		stdin, _ := cmd.StdinPipe()
		stdout, _ := cmd.StdoutPipe()

		go func() {
			buf := bufio.NewScanner(stdout)
			for buf.Scan() {
				fmt.Printf("Return value: %s\n", buf.Text()) // Ausgabe lesen
			}
		}()

		go func() {
			for {
				time.Sleep(1 * time.Second)
				stdin.Write([]byte("echo 1\n"))
			}
		}()

		err := cmd.Run()

		if err != nil {
			panic(err)
		}
	},
}
