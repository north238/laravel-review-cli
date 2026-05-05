package git

import (
	"context"
	"os/exec"
)

func IsGitRepository(ctx context.Context) error {
	// コマンドの作成
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")

	_, err := cmd.Output()
	if err != nil {
		return ErrNotGitRepository
	}

	return nil
}
