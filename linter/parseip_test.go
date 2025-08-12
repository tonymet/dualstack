package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestParseIP(t *testing.T) {
	analysistest.Run(t, analysistest.TestData()+"/parseip", AnalyzerParseIP)
}
