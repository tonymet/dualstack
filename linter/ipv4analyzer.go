package linter

import (
	//"go/analysis"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzers []*analysis.Analyzer = make([]*analysis.Analyzer, 0)

func init() {
	Analyzers = append(Analyzers, AnalyzerIP4)
	Analyzers = append(Analyzers, AnalyzerParseIP)
	Analyzers = append(Analyzers, AnalyzerIP4Byte)
}

// Analyzer is the core component of our static analysis checker.
// It defines the name, documentation, and the function that performs the analysis.
var AnalyzerIP4 = &analysis.Analyzer{
	Name: "ipv4checker",
	Doc:  "Reports calls to net.Listen using a hardcoded IPv4 loopback address.",
	Run:  runIP4,
}

// run is the main function that inspects the AST of the Go source files.
func runIP4(pass *analysis.Pass) (interface{}, error) {
	// Traverse the Abstract Syntax Tree (AST) of each file in the pass.
	for _, file := range pass.Files {
		// ast.Inspect traverses the nodes of the AST.
		ast.Inspect(file, func(node ast.Node) bool {
			// Check if the current node is a function call expression.
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true // Not a function call, continue to the next node.
			}

			// Check if the function being called is 'net.Listen'.
			fun, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true // Not a selector expression (e.g., package.Function), continue.
			}

			// Ensure the selector is 'net.Listen'.
			if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "net" && fun.Sel.Name == "Listen" {
				// We've found a call to net.Listen. Now check its arguments.
				if len(callExpr.Args) != 2 {
					return true // The function call doesn't have the expected two arguments.
				}

				// Check if the first argument is the string literal "tcp".
				networkArg, ok := callExpr.Args[0].(*ast.BasicLit)
				if !ok || networkArg.Kind != token.STRING || strings.Trim(networkArg.Value, `"`) != "tcp" {
					return true // Not a call to net.Listen("tcp", ...), continue.
				}

				// Check if the second argument is the string literal "127.0.0.1".
				addressArg, ok := callExpr.Args[1].(*ast.BasicLit)
				if !ok || addressArg.Kind != token.STRING || !strings.Contains(addressArg.Value, ".") {
					return true // Not a call to net.Listen("...", "127.0.0.1"), continue.
				}

				// All conditions are met. Report the issue.
				pass.Reportf(callExpr.Pos(), "found hardcoded IPv4 loopback address '127.0.0.1'; consider using a dual-stack address like \":PORT\" for better compatibility.")
			}
			return true // Continue inspecting the next node.
		})
	}
	return nil, nil
}
