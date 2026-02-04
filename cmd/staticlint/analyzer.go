package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// NoExitMain is an analyzer that reports direct calls to os.Exit inside the
// main function of package main. Programs should use graceful shutdown where
// possible instead of abruptly calling os.Exit from main.
var NoExitMain = &analysis.Analyzer{
	Name: "noexitmain",
	Doc:  "reports direct calls to os.Exit in main.main; prefer graceful shutdown",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		// Only check package main
		if pass.Pkg.Name() != "main" {
			return nil, nil
		}

		for _, f := range pass.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// We're interested in selector calls like os.Exit
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident := sel.Sel

				obj := pass.TypesInfo.Uses[ident]
				if obj == nil {
					return true
				}

				fn, ok := obj.(*types.Func)
				if !ok {
					return true
				}
				if fn.Pkg() == nil {
					return true
				}
				if fn.Pkg().Path() != "os" || fn.Name() != "Exit" {
					return true
				}

				// Confirm the call is inside a function declaration named main
				for _, decl := range f.Decls {
					fd, ok := decl.(*ast.FuncDecl)
					if !ok || fd.Body == nil {
						continue
					}
					if fd.Name.Name != "main" {
						continue
					}
					// Check positions: call must be inside fd.Body
					if call.Pos() >= fd.Body.Pos() && call.End() <= fd.Body.End() {
						pass.Reportf(call.Lparen, "avoid direct os.Exit call in main; use graceful shutdown instead")
						// one report is enough for this call
						return true
					}
				}

				return true
			})
		}
		return nil, nil
	},
}
