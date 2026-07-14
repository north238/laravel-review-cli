package git

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestIsGitRepository(t *testing.T) {
	tests := []struct {
		name     string
		needInit bool
		wantErr  error
	}{
		{
			name:     "正常系",
			needInit: true,
			wantErr:  nil,
		},
		{
			name:     "異常系",
			needInit: false,
			wantErr:  ErrNotGitRepository,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// ループごとに新たなディレクトリを作成
			dir := t.TempDir()

			// git intが必要な場合
			if tt.needInit {
				cmd := exec.CommandContext(ctx, "git", "init")

				cmd.Dir = dir

				err := cmd.Run()
				if err != nil {
					t.Fatalf("git init failed: %v", err)
				}
			}
			gotErr := IsGitRepository(ctx, dir)
			// エラー種類不一致
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("IsGitRepository() failed: %v", gotErr)
				return
			}
		})
	}
}
