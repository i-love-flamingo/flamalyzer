// Package architecture provides checks targeting architecture conventions
package architecture

import (
	"flamingo.me/dingo"
	"flamingo.me/flamalyzer/src/analyzers"
	"flamingo.me/flamalyzer/src/analyzers/architecture/checks/dependency"
	"flamingo.me/flamalyzer/src/flamalyzer/configuration"
	"golang.org/x/tools/go/analysis"
)

// Module to register the architecture checks
type Module struct{}

// The default properties which are used if there is no config-file
var defaultProps = Props{
	EntryPaths:                 []string{},
	CheckDependencyConventions: true,
	Groups: map[string][]string{
		"infrastructure": {"infrastructure", "interfaces", "application", "domain"},
		"interfaces":     {"interfaces", "application", "domain"},
		"application":    {"application", "domain"},
		"domain":         {"domain"},
	},
}

// Props of an analyzer which will be used by the config-module to match the entries
// of a file to this variables. Is used to activate and deactivate checks, for example.
type Props struct {
	Name                       string
	EntryPaths                 []string
	CheckDependencyConventions bool
	Groups                     map[string][]string
}

// The Analyzer holds a set of checks, uses the config and has props that can be defined to get read by the config
type Analyzer struct {
	checks []*analysis.Analyzer
	config configuration.AnalyzerConfig
	props  Props
}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	injector.BindMulti(new(analyzers.Analyzer)).To(new(Analyzer))
}

// Inject dependencies
func (d *Analyzer) Inject(config configuration.AnalyzerConfig) {
	d.props = defaultProps
	d.props.Name = "architectureAnalyzer"
	d.config = config
}

// ChecksToExecute decides which checks to run
func (d *Analyzer) ChecksToExecute() []*analysis.Analyzer {
	analyzers.DecodeAnalyzerConfigurationsToAnalyzerProps(d.props.Name, d.config, &d.props)

	if d.props.CheckDependencyConventions {
		d.checks = append(d.checks, dependency.NewAnalyzer(d.props.Groups, d.props.EntryPaths).Analyzer)
	}
	return d.checks
}
