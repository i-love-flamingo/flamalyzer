package flamalyzer

import (
	"flamingo.me/flamalyzer/src/analyzers"
	"flamingo.me/flamalyzer/src/flamalyzer/configuration"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

// The Controller delegates the application.
// It triggers the loading of the Config and holds all analyzers.
// It also passes the individual checks of the analyzers to the driver (multichecker)
type Controller struct {
	config    configuration.CoreConfig
	analyzers []analyzers.Analyzer
}

// Inject dependencies
func (c *Controller) Inject(config configuration.CoreConfig, analyzerProvider analyzers.AnalyzerProvider) {
	c.config = config
	c.analyzers = analyzerProvider()
}

// Get checks to run from the analyzers
func (c *Controller) runAnalyzers() {
	var analysisChecks []*analysis.Analyzer

	for _, a := range c.analyzers {
		analysisChecks = append(analysisChecks, a.ChecksToExecute()...)
	}
	multichecker.Main(
		analysisChecks...,
	)
}

// Run the analysis
func (c *Controller) Run() {
	c.config.LoadConfigFromFiles()
	c.runAnalyzers()
}
