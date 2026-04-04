package maindefer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nkiryanov/go-metrics/cmd/staticlint/maindefer"
)

func TestMaindefer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), maindefer.Analyzer,
		"badmain",
		"okmain",
		"lib",
	)
}
