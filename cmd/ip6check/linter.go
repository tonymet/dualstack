package main

import (
	"github.com/tonymet/dualstack/linter"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(linter.Analyzers...)
}
