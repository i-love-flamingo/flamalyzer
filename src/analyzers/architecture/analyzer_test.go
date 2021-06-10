// This module configures the Testing with the "analysistest" package
package architecture

import (
	"path/filepath"
	"testing"

	"flamingo.me/flamalyzer/src/analyzers/architecture/checks/dependency"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestArchitectureConventions(t *testing.T) {
	analysis := dependency.NewAnalyzer(map[string][]string{
		"infrastructure": {"infrastructure", "interfaces", "application", "domain"},
		"interfaces":     {"interfaces", "application", "domain"},
		"application":    {"application", "domain"},
		"domain":         {"domain"},
	}, []string{})
	analysistest.Run(t, filepath.Join(analysistest.TestData(), "/dependencyConventions"), analysis.Analyzer, "./...")
}
