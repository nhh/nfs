package main

import (
	"nfs/cmd"
	"nfs/internal/tui"
)

func main() {
	tui.DisplayApp()
	cmd.Execute()
}
