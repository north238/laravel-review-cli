package git

import (
	"context"
	"os/exec"
)

// 現在のディレクトリがGitリポジトリ配下かどうかを確認する
func IsGitRepository(ctx context.Context, dir string) error {
	// コマンドの作成
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")

	cmd.Dir = dir

	_, err := cmd.Output()
	if err != nil {
		return ErrNotGitRepository
	}

	return nil
}
