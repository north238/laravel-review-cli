package output

import (
	"io"

	"github.com/north238/lrv/internal/review"
)

type Formatter interface {
	Format(w io.Writer, result *review.AggregatedResult) error
}
