package main

import (
	"os"

	"github.com/north238/lrv/internal/cli"
)

func main() {
	err := cli.NewRootCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}
