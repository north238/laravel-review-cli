package git

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
)

type DiffContext struct {
	BaseBranch    string
	CurrentBranch string
	Files         []ChangedFile
}

type ChangedFile struct {
	Path        string
	Diff        string
	FullContent string
}

// 現在のブランチとベースブランチ間の差分情報を取得する
func GetDiff(ctx context.Context, baseBranch string, dir string) (*DiffContext, error) {
	if err := IsGitRepository(ctx, dir); err != nil {
		return nil, err
	}

	// Diffコマンド作成
	diffCmd := exec.CommandContext(ctx, "git", "diff", "--name-only", baseBranch+"...HEAD")

	opts, err := diffCmd.Output()
	if err != nil {
		return nil, ErrBranchNotFound
	}

	// 出力結果の分割
	lines := strings.Split(strings.TrimSpace(string(opts)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}

	// 選択中のブランチを取得
	parseCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := parseCmd.Output()
	if err != nil {
		return nil, ErrBranchNotFound
	}
	currentBranch := strings.TrimSpace(string(branchOut))

	// 出力結果の抽出
	var files []ChangedFile
	for _, path := range lines {
		// diffを取得、失敗したらcontinue
		diffOutCmd := exec.CommandContext(ctx, "git", "diff", baseBranch+"...HEAD", "--", path)
		diffOut, err := diffOutCmd.Output()
		if err != nil {
			slog.Warn("failed to get diff", "path", path, "error", err)
			continue
		}
		diffStr := strings.TrimSpace(string(diffOut))

		// fullContentを取得、失敗したらcontinue
		fullOutCmd := exec.CommandContext(ctx, "git", "show", "HEAD:"+path)
		fullOut, err := fullOutCmd.Output()
		if err != nil {
			slog.Warn("failed to get content", "path", path, "error", err)
			continue
		}
		fullStr := string(fullOut)

		files = append(files, ChangedFile{
			Path:        path,
			Diff:        diffStr,
			FullContent: fullStr,
		})
	}

	return &DiffContext{
		BaseBranch:    baseBranch,
		CurrentBranch: currentBranch,
		Files:         files,
	}, nil
}
