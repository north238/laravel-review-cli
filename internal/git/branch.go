package git

import (
	"context"
	"os/exec"
	"strings"
)

// ベースブランチを決定する
func DetermineBaseBranch(ctx context.Context, specified string, dir string) (string, error) {
	if specified != "" {
		// ブランチの存在確認
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", specified)

		cmd.Dir = dir
		if _, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(specified)), nil
		}

		return "", ErrBranchNotFound
	}

	if branch, err := detectUpstream(ctx, dir); err == nil {
		return branch, nil
	}

	candidates := []string{"main", "master", "develop", "development"}
	return detectFromCandidates(ctx, candidates, dir)
}

// 現在のブランチのupstream設定からベースブランチを取得する
func detectUpstream(ctx context.Context, dir string) (string, error) {
	// 現在のブランチのupstreamを取得
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "@{upstream}")

	cmd.Dir = dir
	opts, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(opts)), nil
}

// 候補ブランチを順に試し、分岐元のブランチを返す
func detectFromCandidates(ctx context.Context, candidates []string, dir string) (string, error) {
	// 現在のブランチ名を取得
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")

	cmd.Dir = dir
	branch, err := cmd.Output()
	if err != nil {
		return "", ErrBranchNotFound
	}
	currentBranch := strings.TrimSpace(string(branch))

	// 候補から一致するブランチを探す
	for _, candidate := range candidates {
		// 現在のブランチと一致する場合はスキップ
		if candidate == currentBranch {
			continue
		}

		// 指定したブランチから派生しているか確認する
		cmd := exec.CommandContext(ctx, "git", "merge-base", "HEAD", candidate)

		cmd.Dir = dir
		_, err := cmd.Output()
		if err != nil {
			continue
		}
		return candidate, nil
	}

	return "", ErrNoBaseDetected
}
