package linter

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
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

// run is the main analysis function.
func runParseIP(pass *analysis.Pass) (interface{}, error) {
	// A map to store the positions of net.ParseIP calls that need to be checked.
	// The key is the variable that holds the result of ParseIP, and the value is the position.
	parsedIPs := make(map[types.Object]token.Pos)

	// Traverse the AST to find function calls.
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		// Look for assignment statements that might contain a call to net.ParseIP.
		if assign, ok := n.(*ast.AssignStmt); ok {
			// Check if the right-hand side is a function call.
			if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
				// Check if the call is to `net.ParseIP`.
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					// We need to check if the function is `net.ParseIP` specifically.
					// This requires type information, which we get from the analysis.Pass.
					obj := pass.TypesInfo.ObjectOf(sel.Sel)
					if obj != nil && obj.Pkg().Path() == "net" && obj.Name() == "ParseIP" {
						// Store the variable that receives the IP address.
						if len(assign.Lhs) > 0 {
							if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
								if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
									parsedIPs[obj] = assign.Pos()
								}
							}
						}
					}
				}
			}
		}

		// Look for method calls on the IP variable, specifically To4() or Is4().
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Check if the method call is `To4` or `Is4`.
				if sel.Sel.Name == "To4" || sel.Sel.Name == "Is4" {
					// Check if the receiver of the method call is one of the IP variables we tracked.
					if ident, ok := sel.X.(*ast.Ident); ok {
						if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
							// If a check is found, remove the variable from our map.
							// This marks it as "handled."
							if _, ok := parsedIPs[obj]; ok {
								delete(parsedIPs, obj)
							}
						}
					}
				}
			}
		}

		return true
	})

	// After traversing the entire file, any remaining entries in parsedIPs
	// represent calls to net.ParseIP without a subsequent IPv4 check.
	for _, pos := range parsedIPs {
		pass.Reportf(pos, "call to `net.ParseIP` should be followed by a check for IPv4 or handle IPv6 compatibility")
	}

	return nil, nil
}
