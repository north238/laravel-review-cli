package output_test

import (
	"testing"

	"github.com/north238/lrv/internal/output"
	"github.com/north238/lrv/internal/review"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		outputPath string
		result     *review.AggregatedResult
		formatter  output.Formatter
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := output.Write(tt.outputPath, tt.result, tt.formatter)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Write() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Write() succeeded unexpectedly")
			}
		})
	}
}
