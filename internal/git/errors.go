package git

import "errors"

var (
	ErrNotGitRepository = errors.New("not a git repository; run lrv inside a git repository")
	ErrBranchNotFound   = errors.New("specified branch does not exist; check the branch name and try again")
	ErrNoBaseDetected   = errors.New("could not detect the base branch automatically; specify one explicitly with --base")
)
