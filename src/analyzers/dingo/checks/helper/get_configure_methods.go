package helper

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/globals"
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

// If a function is a "Configure" function it must be bound to a pointer receiver
// this function simply returns a List of all Configure-Functions found in the AST
func runConfigureDeclAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	var configureMethods []*ast.FuncDecl

	// Must have a *function receiver
	checkFunction := func(n ast.Node) {
		funcdecl := n.(*ast.FuncDecl)
		// If function has no Receiver, pass
		// TODO disabled until it is known if this is a requirement for a configure function
		// TODO if not think if we want to check the bindings (bind.to) and tags (providercheck) anyways -> probably a new analyzer to check configure funcs declaration?
		//if funcdecl.Recv == nil {
		//	return
		//}

		// Make sure we get functions with "*dingo.Injector" in the parameters
		for _, param := range funcdecl.Type.Params.List {
			if param, ok := param.Type.(*ast.StarExpr); ok {
				if param, ok := param.X.(*ast.SelectorExpr); ok {
					paramType := pass.TypesInfo.ObjectOf(param.Sel)
					if paramType.Pkg().Path() == globals.DingoPkgPath && paramType.Name() == "Injector" {
						configureMethods = append(configureMethods, funcdecl)
					}
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return configureMethods, nil
}
