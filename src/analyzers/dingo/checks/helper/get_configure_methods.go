package helper

import (
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// helper analyzer which returns all function declarations
var ConfigureDeclAnalyzer = &analysis.Analyzer{
	Name:       "getConfigureMethods",
	Doc:        "helper to get all Configure Methods",
	Run:        runConfigureDeclAnalyzer,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeOf(*new([]*ast.FuncDecl)),
}

// If a function is an "Inject" function it must be bound to a pointer receiver
// this function simply returns a List of all Inject-Functions found in the AST used by another analyzer
func runConfigureDeclAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	var configureMethods []*ast.FuncDecl

	// Name must be "Inject" and must have a *function receiver
	checkFunction := func(n ast.Node) {
		funcdecl := n.(*ast.FuncDecl)
		// If function has no Receiver, pass
		if funcdecl.Recv == nil {
			return
		}
		if funcdecl.Name.Name == "Configure" {
			// If it is an Configure-function add it to the list of configureMethods
			configureMethods = append(configureMethods, funcdecl)
		}

	}
	input.Preorder(nodeFilter, checkFunction)
	return configureMethods, nil
}
