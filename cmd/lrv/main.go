package main

import (
	"log/slog"
	"os"

	"github.com/north238/lrv/internal/cli"
)

func main() {
	// slogの初期化
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	err := cli.NewRootCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}
