package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestIPv4Linter(t *testing.T) {
	analysistest.Run(t, analysistest.TestData()+"/ip4byte", AnalyzerIP4Byte)
}
