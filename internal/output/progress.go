package output

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/north238/lrv/internal/review"
)

type StderrProgressReporter struct {
	w  io.Writer
	mx sync.Mutex
}

var _ review.ProgressReporter = (*StderrProgressReporter)(nil)

// 初期化
func NewStderrProgressReporter(w io.Writer) *StderrProgressReporter {
	return &StderrProgressReporter{
		w: w,
	}
}

// 変更ファイル取得文言
func (r *StderrProgressReporter) DiffFetched(fileCount int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	fmt.Fprintf(r.w, "変更ファイルを取得しました（%d 件）\n", fileCount)
}

// 観点の開始文言
func (r *StderrProgressReporter) AspectStarted(aspect review.Aspect) {
	r.mx.Lock()
	defer r.mx.Unlock()

	fmt.Fprintf(r.w, "%s観点：開始します\n", aspectDisplayNames[aspect])
}

// 観点の終了文言
func (r *StderrProgressReporter) AspectCompleted(aspect review.Aspect, findingCount int, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err != nil {
		fmt.Fprintf(r.w, "%s観点：失敗しました\n", aspectDisplayNames[aspect])
	} else {
		fmt.Fprintf(r.w, "%s観点：完了（%d 件）\n", aspectDisplayNames[aspect], findingCount)
	}
}

// 終了文言
func (r *StderrProgressReporter) Finished(duration time.Duration, totalFindings int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	fmt.Fprintf(r.w, "完了しました（%.1f 秒、計%d 件）\n", duration.Seconds(), totalFindings)
}
