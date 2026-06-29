package review

import (
	"time"
)

type ProgressReporter interface {
	DiffFetched(fileCount int)
	AspectStarted(aspect Aspect)
	AspectCompleted(aspect Aspect, findingCount int, err error)
	Finished(duration time.Duration, totalFindings int)
}
