package noexit_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nkiryanov/go-metrics/cmd/staticlint/noexit"
)

func TestNoexit(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), noexit.Analyzer,
		"allowed",
		"forbidden",
		"forbidden_main",
		"shadow",
	)
}
