package git

import "errors"

var (
	ErrNotGitRepository = errors.New("not a git repository")
	ErrBranchNotFound   = errors.New("branch not found")
	ErrNoBaseDetected   = errors.New("base branch could not be detected")
)
