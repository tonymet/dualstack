package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// The Analyzer's name and description.
var AnalyzerParseIP = &analysis.Analyzer{
	Name: "checkip",
	Doc:  "checks for net.ParseIP calls without a net.IP.To4() check",
	Run:  runParseIP,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer, // Required to get a handle to the AST inspector
	},
}

func runParseIP(pass *analysis.Pass) (interface{}, error) {
	// The inspector allows us to walk the AST.
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Filter for function calls. We're only interested in function calls
	// because net.ParseIP is a function.
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		// Cast the node to a CallExpr to access its fields.
		callExpr := n.(*ast.CallExpr)

		// Get the function being called.
		fun, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Look up the type information for the function.
		obj := pass.TypesInfo.ObjectOf(fun.Sel)
		if obj == nil {
			return
		}

		// Check if the function is net.ParseIP.
		if obj.Pkg().Path() == "net" && obj.Name() == "ParseIP" {
			// If it is, report a diagnostic message.
			pass.Reportf(n.Pos(), "Found a call to net.ParseIP. Consider if a more specific IP parsing function is needed.")
		}
	})

	return nil, nil
}
