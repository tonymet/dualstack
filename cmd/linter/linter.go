package main

import (
	"github.com/tonymet/dualstack/linter" // IMPORTANT: Replace 'your_module_name' with the actual path to your module

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(linter.Analyzers...)
}
