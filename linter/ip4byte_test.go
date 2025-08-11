package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	// Ensure you import the analyzer package. In this case, it's in the same package.
)

func TestIPv4Linter(t *testing.T) {
	// The analysistest.Run function executes the analyzer against the test data.
	// The first argument is the testing.T instance.
	// The second argument specifies the directory where the test files are located.
	// The third argument is the analyzer you want to test.
	analysistest.Run(t, analysistest.TestData()+"/ip4byte", AnalyzerIP4Byte)
}

// Below is the content for the test data file: testdata/src/a/a.go
// This file will be automatically found and analyzed by analysistest.Run.
