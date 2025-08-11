package linter

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer is the entry point for our linter.
var AnalyzerIP4Byte = &analysis.Analyzer{
	Name:     "ipv4linter",
	Doc:      "Checks for incorrect IPv4 size assumptions on net.IP variables.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runIP4Byte,
}

func runIP4Byte(pass *analysis.Pass) (interface{}, error) {
	// We'll store all variables we've identified as being of type net.IP.
	// This map lets us quickly check if a variable is an IP address.
	ipv4AssumedVars := make(map[types.Object]bool)

	// --- Pass 1: Identify all variables of type net.IP ---
	// This pass is essential to correctly type-check all variables
	// before we analyze how they are used.
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					// Check if the variable is defined and of type net.IP.
					if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
						if t, ok := obj.Type().(*types.Named); ok {
							if t.Obj().Pkg().Path() == "net" && t.Obj().Name() == "IP" {
								ipv4AssumedVars[obj] = true
							}
						}
					}
				}
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Type != nil && fn.Type.Params != nil {
				for _, param := range fn.Type.Params.List {
					for _, name := range param.Names {
						if obj := pass.TypesInfo.ObjectOf(name); obj != nil {
							if t, ok := obj.Type().(*types.Named); ok {
								if t.Obj().Pkg().Path() == "net" && t.Obj().Name() == "IP" {
									ipv4AssumedVars[obj] = true
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	// --- Pass 2: Check how these net.IP variables are used ---
	// This is the core logic that checks for bad patterns.
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		// Detect fixed-length slicing (e.g., ip[0:4]).
		if slice, ok := n.(*ast.SliceExpr); ok {
			if ident, ok := slice.X.(*ast.Ident); ok {
				if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
					// Check if the variable is one of the net.IP variables we tracked.
					if _, isIP := ipv4AssumedVars[obj]; isIP {
						// Check if the slice assumes a fixed size of 4 bytes.
						if basicLit, ok := slice.High.(*ast.BasicLit); ok && basicLit.Value == "4" {
							pass.Reportf(slice.Pos(), "fixed-length slice of 4 on a net.IP variable may fail with IPv6")
						}
					}
				}
			}
		}

		// Detect indexing on fixed positions (e.g., ip[3]).
		if index, ok := n.(*ast.IndexExpr); ok {
			if ident, ok := index.X.(*ast.Ident); ok {
				if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
					if _, isIP := ipv4AssumedVars[obj]; isIP {
						if basicLit, ok := index.Index.(*ast.BasicLit); ok {
							// Flag any index that is an IPv4-specific position.
							// Note: This is a heuristic and might have false positives,
							// but it catches common IPv4 assumptions.
							if basicLit.Value == "3" || basicLit.Value == "4" {
								pass.Reportf(index.Pos(), "fixed index on a net.IP variable may be an IPv4 assumption")
							}
						}
					}
				}
			}
		}

		return true
	})

	return nil, nil
}
