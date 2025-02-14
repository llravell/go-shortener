package noexit_test

import (
	"testing"

	"github.com/llravell/go-shortener/pkg/passes/noexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), noexit.Analyzer, "./...")
}
