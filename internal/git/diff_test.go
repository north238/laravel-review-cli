package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetDiff(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	// Gitとファイルセットアップ
	runGit(t, ctx, dir, "init")

	// ファイル新規作成
	newFilepath := filepath.Join(dir, "test.txt")
	err := os.WriteFile(newFilepath, []byte("hello main world"), 0700)
	if err != nil {
		t.Fatalf("WriteFile() failed to newFilepath: %v", err)
	}

	runGit(t, ctx, dir, "config", "user.name", "test user")
	runGit(t, ctx, dir, "config", "user.email", "test@example.com")

	runGit(t, ctx, dir, "add", ".")
	runGit(t, ctx, dir, "commit", "-m", "first commit main")

	runGit(t, ctx, dir, "branch", "-M", "main")
	runGit(t, ctx, dir, "checkout", "-b", "feature")

	// ファイルを変更と追加する
	err = os.WriteFile(newFilepath, []byte("hello feature world"), 0700)
	if err != nil {
		t.Fatalf("WriteFile() failed to newFilepath: %v", err)
	}

	addFilePath := filepath.Join(dir, "feature.txt")
	err = os.WriteFile(addFilePath, []byte("add feature file"), 0700)
	if err != nil {
		t.Fatalf("WriteFile() failed to addFilePath: %v", err)
	}

	runGit(t, ctx, dir, "add", ".")
	runGit(t, ctx, dir, "commit", "-m", "first commit feature")

	diffCtx, err := GetDiff(ctx, "main", dir)
	if err != nil {
		t.Fatalf("GetDiff() failed: %v", err)
	}

	if diffCtx.BaseBranch != "main" {
		t.Errorf("BaseBranch = %v, want %v", diffCtx.BaseBranch, "main")
	}

	if diffCtx.CurrentBranch != "feature" {
		t.Errorf("CurrentBranch = %v, want %v", diffCtx.CurrentBranch, "feature")
	}

	if len(diffCtx.Files) != 2 {
		t.Errorf("Files length = %v, want %v", len(diffCtx.Files), 2)
	}

	m := make(map[string]string)
	for _, file := range diffCtx.Files {
		m[file.Path] = file.FullContent
	}

	content, ok := m["test.txt"]
	if !ok {
		t.Fatal("test.txt file not found")
	}

	if content != "hello feature world" {
		t.Errorf("content = %v, want %v", content, "hello feature world")
	}
}

// ヘルパー関数 Gitコマンド実行
func runGit(t *testing.T, ctx context.Context, dir string, args ...string) {
	cmd := exec.CommandContext(ctx, "git", args...)

	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		t.Fatalf("runGit() failed: %v", err)
	}
}
