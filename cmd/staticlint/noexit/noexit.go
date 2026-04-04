// Package noexit provides an analyzer that forbids exit-like calls outside
// of the main function of package main.
//
// Forbidden: os.Exit, log.Fatal/Fatalf/Fatalln, (*log.Logger).Fatal/Fatalf/Fatalln.
package noexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "forbid os.Exit and log.Fatal* calls outside of main() in package main",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	forbidden := buildForbidden(pass.Pkg)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	for cur := range insp.Root().Preorder(nodeFilter...) {
		call := cur.Node().(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		obj, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
		if !ok {
			continue
		}
		label, ok := forbidden[obj]
		if !ok {
			continue
		}

		if inMainFunc(pass, cur) {
			continue
		}

		pass.Reportf(call.Pos(), "%s is only allowed in main() of package main", label)
	}
	return nil, nil
}

// buildForbidden builds a map from canonical *types.Func to a display name
// for each function that must only be called from main() of package main.
// Only packages actually imported by pkg are included.
func buildForbidden(pkg *types.Package) map[*types.Func]string {
	set := make(map[*types.Func]string)
	for _, imp := range pkg.Imports() {
		switch imp.Path() {
		case "os":
			if f, ok := imp.Scope().Lookup("Exit").(*types.Func); ok {
				set[f] = "os.Exit"
			}
		case "log":
			for _, name := range []string{"Fatal", "Fatalf", "Fatalln"} {
				if f, ok := imp.Scope().Lookup(name).(*types.Func); ok {
					set[f] = "log." + name
				}
			}
			loggerTypeName, ok := imp.Scope().Lookup("Logger").(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := loggerTypeName.Type().(*types.Named)
			if !ok {
				continue
			}
			for m := range named.Methods() {
				if m.Name() == "Fatal" || m.Name() == "Fatalf" || m.Name() == "Fatalln" {
					set[m] = "(*log.Logger)." + m.Name()
				}
			}
		}
	}
	return set
}

// inMainFunc reports whether cur is directly inside the main() function of
// package main (not inside a closure within main).
func inMainFunc(pass *analysis.Pass, cur inspector.Cursor) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}
	for cur = cur.Parent(); cur.Node() != nil; cur = cur.Parent() {
		switch n := cur.Node().(type) {
		case *ast.FuncLit:
			// nearest enclosing is a closure — not main
			return false
		case *ast.FuncDecl:
			return n.Recv == nil && n.Name.Name == "main"
		}
	}
	return false
}
