package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/north238/lrv/internal/cli"
	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/llm"
	"github.com/north238/lrv/internal/review"
)

func main() {
	// slogの初期化
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	err := cli.NewRootCommand().Execute()
	if err != nil {
		exitCode := determineExitCode(err)

		// コード４の場合は固定文言を返却
		if exitCode == 4 {
			fmt.Fprintln(os.Stderr, "an unexpected error occurred; please try again")
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(exitCode)
	}
}

// エラーの振り分け
func determineExitCode(err error) int {
	switch {
	case errors.Is(err, git.ErrBranchNotFound),
		errors.Is(err, review.ErrInvalidAspect):
		return 1
	case errors.Is(err, git.ErrNotGitRepository),
		errors.Is(err, git.ErrNoBaseDetected),
		errors.Is(err, llm.ErrAPIKeyMissing):
		return 2
	case errors.Is(err, llm.ErrAPITimeout),
		errors.Is(err, llm.ErrAPIAuthFailed),
		errors.Is(err, llm.ErrAPIUnexpectedResponse):
		return 3
	default:
		return 4
	}
}
