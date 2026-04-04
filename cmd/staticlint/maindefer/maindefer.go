// Package maindefer provides an analyzer that forbids defer statements in
// main() of package main. Defers there are skipped by os.Exit and log.Fatal*
// calls; use a run() helper to ensure cleanup runs.
package maindefer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "maindefer",
	Doc:      "forbid defer statements in main() of package main; use a run() helper instead",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	for cur := range insp.Root().Preorder(nodeFilter...) {
		fn := cur.Node().(*ast.FuncDecl)
		if fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
			continue
		}
		checkBody(pass, fn.Body)
	}
	return nil, nil
}

// checkBody walks the body of main() reporting any defer statements found,
// pruning descent into function literals (they have their own scope).
func checkBody(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false // don't descend into closures
		}
		if d, ok := n.(*ast.DeferStmt); ok {
			pass.Reportf(d.Pos(), "defer in main(): use a run() helper instead")
		}
		return true
	})
}
