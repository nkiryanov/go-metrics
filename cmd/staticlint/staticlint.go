package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/nkiryanov/go-metrics/cmd/staticlint/maindefer"
	"github.com/nkiryanov/go-metrics/cmd/staticlint/noexit"
)

func main() {
	multichecker.Main(
		noexit.Analyzer,
		maindefer.Analyzer,
	)
}
