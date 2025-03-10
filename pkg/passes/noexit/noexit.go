package noexit

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// Analyzer статический анализатор проверяющий прямой вызов os exit в main.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "check os exit call in main function",
	Run:  run,
}

func getOSExitCallPos(list []ast.Stmt) token.Pos {
	for _, stmt := range list {
		expr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}

		call, ok := expr.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		if selector.Sel.Name == "Exit" {
			ident, ok := selector.X.(*ast.Ident)

			if ok && ident.Name == "os" {
				return ident.NamePos
			}
		}
	}

	return -1
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if x, ok := node.(*ast.FuncDecl); ok {
				if x.Name.Name == "main" {
					osExitPos := getOSExitCallPos(x.Body.List)

					if osExitPos != -1 {
						pass.Reportf(osExitPos, "os exit call in main function")
					}
				}
			}

			return true
		})
	}

	//nolint: nilnil
	return nil, nil
}
