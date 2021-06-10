package analyzers

import (
	"reflect"

	"flamingo.me/flamalyzer/src/flamalyzer/configuration"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/tools/go/analysis"
)

// AnalyzerProvider returns all bound instances
type AnalyzerProvider func() []Analyzer

// Analyzer is a collection of checks performed by Flamalyzer.
type Analyzer interface {
	ChecksToExecute() []*analysis.Analyzer
}

// DecodeAnalyzerConfigurationsToAnalyzerProps decodes the props loaded from the config-files to the specific props of an analyzer
// The props musst be passed a Pointer e.g &props
func DecodeAnalyzerConfigurationsToAnalyzerProps(entryName string, config configuration.AnalyzerConfig, propsPtr interface{}) {
	if reflect.ValueOf(propsPtr).Type().Kind() != reflect.Ptr {
		panic("The passed propsPtr must be a pointer, otherwise the result of this function won't be available in the analyzer!")
	}
	inputProps := config.GetProps(entryName)
	err := mapstructure.Decode(inputProps, &propsPtr)
	if err != nil {
		panic(err)
	}

}
