package configure

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/helper"
	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"reflect"
)

// Analyzer which checks a function for a correct 'Configure' definition and returns all correct ones
var ReceiverAnalyzer = &analysis.Analyzer{
	Name:       "checkConfigureHasReceiver",
	Doc:        "Search for missing receivers in 'Configure function' declarations",
	Run:        runConfigureHasReceiver,
	Requires:   []*analysis.Analyzer{helper.ConfigureDeclAnalyzer},
	ResultType: reflect.TypeOf(*new([]*ast.FuncDecl)),
}

// If a function is a "Configure" function there must be a receiver.
// This function simply returns a list of all Configure-Functions found in the AST, including ones with different names than "Configure" (needed if delegated)
func runConfigureHasReceiver(pass *analysis.Pass) (interface{}, error) {
	configureFunctions := pass.ResultOf[helper.ConfigureDeclAnalyzer].([]*ast.FuncDecl)
	var correctFunctions [] *ast.FuncDecl
	for _, f := range configureFunctions {
		if f.Recv == nil {
			flanalysis.Report(pass, "Configure function has no Receiver! A type must implement the dingo.Module interface!", f)
		}else{
			correctFunctions = append(correctFunctions, f)
		}
	}
	return correctFunctions, nil
}

