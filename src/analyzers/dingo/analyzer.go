// Package dingo provides checks targeting dingo-specific problems
package dingo

import (
	"flamingo.me/dingo"
	"flamingo.me/flamalyzer/src/analyzers"
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/bind"
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/configure"
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/inject"
	"flamingo.me/flamalyzer/src/flamalyzer/configuration"
	"golang.org/x/tools/go/analysis"
)

// Module to register the dingo checks
type Module struct{}

// The default properties which are used if there is no config-file
var defaultProps = Props{
	CheckPointerReceiver:            true,
	CheckStrictTagsAndFunctions:     true,
	CheckBindingImplementsInterface: true,
	CheckConfigureHasReceiver:       true,
}

// Props of an analyzer which will be used by the config-module to match the entries
// of a file to these variables. Is used to activate and deactivate checks, for example.
type Props struct {
	Name                            string
	CheckPointerReceiver            bool
	CheckStrictTagsAndFunctions     bool
	CheckBindingImplementsInterface bool
	CheckConfigureHasReceiver       bool
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
	d.props.Name = "dingoAnalyzer"
	d.config = config
}

// ChecksToExecute decides which checks to run
func (d *Analyzer) ChecksToExecute() []*analysis.Analyzer {
	analyzers.DecodeAnalyzerConfigurationsToAnalyzerProps(d.props.Name, d.config, &d.props)

	if d.props.CheckPointerReceiver {
		d.checks = append(d.checks, inject.ReceiverAnalyzer)
	}
	if d.props.CheckStrictTagsAndFunctions {
		d.checks = append(d.checks, inject.TagAnalyzer)
	}
	if d.props.CheckBindingImplementsInterface {
		d.checks = append(d.checks, bind.Analyzer)
	}
	if d.props.CheckConfigureHasReceiver {
		d.checks = append(d.checks, configure.ReceiverAnalyzer)
	}
	return d.checks
}
