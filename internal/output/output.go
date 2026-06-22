package output

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/north238/lrv/internal/review"
)

// ファイルへの書き込み関数
func Write(outputPath string, result *review.AggregatedResult, formatter Formatter) (err error) {
	var w io.Writer
	var cleanup func() error

	// ファイルパスがある場合
	if outputPath != "" {
		// ファイル名を取り除いてディレクトリ部分を取得
		path := filepath.Dir(outputPath)
		// 指定されたディレクトリの作成
		cerr := os.MkdirAll(path, 0o755)
		if cerr != nil {
			return cerr
		}

		// ファイルを作成
		f, cerr := os.Create(outputPath)
		if cerr != nil {
			return cerr
		}

		w = f
		cleanup = f.Close
	} else {
		w = os.Stdout
		cleanup = func() error {
			return nil
		}
	}

	// f.Closeのエラーをキャッチするため
	defer func() {
		cerr := cleanup()
		if cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	// フォーマット処理呼び出し
	err = formatter.Format(w, result)

	return err
}
