package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoExitMain is an analyzer that reports direct calls to osExit inside the
// main function of package main. Programs should use graceful shutdown where
// possible instead of abruptly calling osExit from main.
var NoExitMain = &analysis.Analyzer{
	Name: "noexitmain",
	Doc:  "reports direct calls to osExit in main.main; prefer graceful shutdown",
	Run:  run,
}

// run executes the analysis pass.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if isNotMainPackage(file) {
			continue
		}
		mainFunction := getMainFuncDecl(file)
		if mainFunction == nil {
			continue
		}
		ast.Inspect(mainFunction, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if ok {
					packageName, ok := selector.X.(*ast.Ident)
					if ok {
						if selector.Sel.Name == "Exit" && packageName.Name == "os" {
							pass.Reportf(node.Pos(), "osExit call within main package main func")
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// isMainPackageFile checks if the given AST file belongs to package main.
func isNotMainPackage(fileNode *ast.File) bool {
	return fileNode.Name.Name != "main"
}

// getMainFuncDecl returns the *ast.FuncDecl for the main function in the given file,
func getMainFuncDecl(fileNode *ast.File) *ast.FuncDecl {
	for _, topDeclaration := range fileNode.Decls {
		function, ok := topDeclaration.(*ast.FuncDecl)
		if ok {
			if function.Name.Name == "main" {
				return function
			}
		}
	}
	return nil
}
