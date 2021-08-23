package configure

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/helper"
	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"reflect"
)

// Analyzer which checks a function for a correct 'Configure' definition and returns all correct ones
var FunctionHasReceiver = &analysis.Analyzer{
	Name:       "ConfigureHasReceiver",
	Doc:        "Search for missing receivers in 'Configure function' declarations",
	Run:        runConfigureHasReceiver,
	Requires:   []*analysis.Analyzer{helper.ConfigureDeclAnalyzer},
	ResultType: reflect.TypeOf(*new([]*ast.FuncDecl)),
}

// If a function is a "Configure" function it must be bound to a pointer receiver
// this function simply returns a List of all Configure-Functions found in the AST, including ones with different names than "Configure" (needed if delegated)
func runConfigureHasReceiver(pass *analysis.Pass) (interface{}, error) {
	configureFunctions := pass.ResultOf[helper.ConfigureDeclAnalyzer].([]*ast.FuncDecl)
	// TODO make it so this test returns a result of correct configuration functions so it can be used by other analyzers
	var correctFunctions [] *ast.FuncDecl
	for _, f := range configureFunctions {
		if f.Recv == nil {
			flanalysis.Report(pass, "Configure function has no Receiver! A type must implement the Module interface!", f)
		}else{
			correctFunctions = append(correctFunctions, f)
		}
	}
	return correctFunctions, nil
}

