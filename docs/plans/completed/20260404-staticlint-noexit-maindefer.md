# staticlint: noexit + maindefer analyzers

## Overview
Two `analysis.Analyzer`s in `cmd/staticlint/`:
- **noexit**: exit-like calls (`os.Exit`, `log.Fatal*`, `(*log.Logger).Fatal*`) allowed only in `main()` of `package main`.
- **maindefer**: no `defer` anywhere in `main()` of `package main` (enforces `run()` pattern).

Replaces existing `cmd/staticlint/osexit/` skeleton. Registered via `multichecker.Main`.

## Context
- Existing: `cmd/staticlint/staticlint.go` (skeleton), `cmd/staticlint/osexit/osexit.go` (skeleton — delete).
- `golang.org/x/tools v0.41.0` already in `go.mod` (direct). No new deps.
- Project uses `go tool` for dev tools; tests via `gotestsum`.

## Development Approach
- Testing: **TDD via `analysistest`** — write testdata fixtures first, then analyzer logic.
- Each task: code + tests + `make test` green before next.
- Small commits per task preferred but not required.

## Testing Strategy
- `golang.org/x/tools/go/analysis/analysistest` with `testdata/src/<pkg>/*.go` fixtures.
- Expected diagnostics via `// want "regex"` comments.
- No unit helpers — fixtures test end-to-end behavior.

## Solution Overview
- **noexit**: `inspector.Root().Preorder((*ast.CallExpr)(nil))` → resolve via `pass.TypesInfo.Uses` → check forbidden set → walk `cursor.Parent()` to nearest `*ast.FuncDecl`/`*ast.FuncLit` → report if not `main` in `package main`.
- **maindefer**: early-return if `pass.Pkg.Name() != "main"` → iterate `*ast.FuncDecl` → skip non-`main` / methods → `ast.Inspect` body pruning at `*ast.FuncLit` → report each `*ast.DeferStmt`.
- Both depend on `inspect.Analyzer` via `Requires`.

## Technical Details

### Forbidden set (noexit)
Resolved through `pass.TypesInfo.Uses[ident].(*types.Func)` on the call's `Fun.Sel`:
- `os.Exit`
- `log.Fatal`, `log.Fatalf`, `log.Fatalln`
- `(*log.Logger).Fatal`, `(*log.Logger).Fatalf`, `(*log.Logger).Fatalln`

Match strategy — decomposed check (robust):
- `obj.Pkg().Path() == "os" && obj.Name() == "Exit"` → forbidden
- `obj.Pkg().Path() == "log" && obj.Name() in {Fatal,Fatalf,Fatalln}`:
  - no recv → package-level func
  - recv type is `*log.Logger` → method
  - both forbidden

Avoids fragile string matching on `FullName()` format.

### Cursor parent walk
```go
for p := cur.Parent(); p.Node() != nil; p = p.Parent() {
    switch p.Node().(type) {
    case *ast.FuncDecl, *ast.FuncLit:
        return p.Node()
    }
}
```

### Defer scan (maindefer)
```go
ast.Inspect(mainFn.Body, func(n ast.Node) bool {
    if _, ok := n.(*ast.FuncLit); ok { return false } // prune closures
    if d, ok := n.(*ast.DeferStmt); ok { pass.Reportf(d.Pos(), "...") }
    return true
})
```

## What Goes Where
- **Implementation Steps**: analyzer packages, tests, multichecker wiring, remove old skeleton.
- **Post-Completion**: run `make lint` manually; verify via `go run ./cmd/staticlint ./...` on real project code.

## Implementation Steps

### Task 1: noexit analyzer + tests

**Files:**
- Create: `cmd/staticlint/noexit/noexit.go`
- Create: `cmd/staticlint/noexit/noexit_test.go`
- Create: `cmd/staticlint/noexit/testdata/src/allowed/main.go`
- Create: `cmd/staticlint/noexit/testdata/src/forbidden/lib.go`
- Create: `cmd/staticlint/noexit/testdata/src/forbidden/init_main.go`
- Create: `cmd/staticlint/noexit/testdata/src/forbidden/shadow.go`

- [ ] testdata `allowed/main.go`: `package main` with `os.Exit(0)` in `main()` — no `want`.
- [ ] testdata `forbidden/lib.go`: helper func with `os.Exit`, method with `os.Exit`, closure inside main with `os.Exit`, `log.Fatal`/`Fatalf`/`Fatalln` in helper, `(*log.Logger).Fatal*` via var in helper — each with `// want "..."` on call line.
- [ ] testdata `forbidden/init_main.go`: `package main` with `func init() { os.Exit(1) }` — `// want` (init is not main).
- [ ] testdata `forbidden/shadow.go`: **function-scope** local var `os := struct{ Exit func(int) }{...}` calling `os.Exit(1)` — not reported (same package can still import real `os` in other files).
- [ ] `noexit.go`: `Analyzer` with `Name`, `Doc`, `Requires: inspect.Analyzer`; `run` using `inspector.Root().Preorder((*ast.CallExpr)(nil))`; decomposed match via `obj.Pkg().Path()` + `obj.Name()` + receiver check (see Technical Details); parent walk to enclosing `FuncDecl`/`FuncLit`; report if not `main` of `package main`.
- [ ] `noexit_test.go`: `analysistest.Run(t, analysistest.TestData(), noexit.Analyzer, "allowed", "forbidden")`.
- [ ] `make test` green.

### Task 2: maindefer analyzer + tests

**Files:**
- Create: `cmd/staticlint/maindefer/maindefer.go`
- Create: `cmd/staticlint/maindefer/maindefer_test.go`
- Create: `cmd/staticlint/maindefer/testdata/src/badmain/main.go`
- Create: `cmd/staticlint/maindefer/testdata/src/okmain/main.go`
- Create: `cmd/staticlint/maindefer/testdata/src/lib/lib.go`

- [ ] testdata `badmain/main.go`: `main()` with top-level `defer` + nested `if { defer ... }` — both `// want`.
- [ ] testdata `okmain/main.go`: `main()` containing closure with `defer` (closure pruned, not reported); also nested-block closure `if true { func() { defer fmt.Println("x") }() }`; no top-level defers.
- [ ] testdata `lib/lib.go`: non-main package with `defer` in function (not reported).
- [ ] `maindefer.go`: `Analyzer` with `Name`, `Doc`, `Requires: inspect.Analyzer`; early return if `pass.Pkg.Name() != "main"`; preorder `*ast.FuncDecl`, filter `Name=="main" && Recv==nil`; `ast.Inspect` body pruning `FuncLit`; report `DeferStmt`.
- [ ] `maindefer_test.go`: `analysistest.Run` over `badmain`, `okmain`, `lib`.
- [ ] `make test` green.

### Task 3: Wire multichecker + remove old skeleton

**Files:**
- Modify: `cmd/staticlint/staticlint.go`
- Delete: `cmd/staticlint/osexit/osexit.go` (and parent dir)

- [ ] Rewrite `staticlint.go`: imports limited to `multichecker`, `noexit`, `maindefer` (drop existing unused `analysis` import); body = `multichecker.Main(noexit.Analyzer, maindefer.Analyzer)`.
- [ ] `rm -rf cmd/staticlint/osexit`.
- [ ] `go build ./cmd/staticlint` succeeds.
- [ ] `go run ./cmd/staticlint ./...` on the project runs without crashing (diagnostics on real code are expected/acceptable — do not fix here).
- [ ] `make test` green.

### Task 4: Verify acceptance criteria
- [ ] noexit flags: helper/method/closure `os.Exit`, `log.Fatal*`, `(*log.Logger).Fatal*`.
- [ ] noexit ignores: `main()` in `package main`, shadowed `os` identifier.
- [ ] maindefer flags: any `defer` in `main()` (top-level + nested blocks).
- [ ] maindefer ignores: `defer` in closures inside main, `defer` in helper funcs, non-main packages.
- [ ] `make lint` passes on new code.
- [ ] `make test` full suite green.

### Task 5: Close out
- [ ] Move plan to `docs/plans/completed/`.

## Post-Completion
- Run `go run ./cmd/staticlint ./...` on the full project; fix any legitimate findings in a separate PR (not this one).
