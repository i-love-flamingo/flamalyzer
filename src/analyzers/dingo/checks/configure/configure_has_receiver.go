package configure

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/globals"
	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"reflect"
)

// Analyzer which checks a function for a correct 'Configure' definition and returns them
var ReceiverAnalyzer = &analysis.Analyzer{
	Name:       "checkConfigureHasReceiver",
	Doc:        "Search for missing receivers in 'Configure function' declarations",
	Run:        runConfigureHasReceiver,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeOf(new(FunctionDeclarations)),
}
type FunctionDeclarations struct {
	all []*ast.FuncDecl
	valid []*ast.FuncDecl
}
// All functions having *dingo.Injector" in the parameters (useful if analyzers should run although configure-function is not perfectly correct)
func (c *FunctionDeclarations) GetAll() []*ast.FuncDecl{
	return c.all
}
// All functions having "*dingo.Injector" in the parameters + existing function receiver (considered correct configure-function)
func (c *FunctionDeclarations) GetValid() []*ast.FuncDecl{
	return c.valid
}

// this function returns a struct containing all configure-functions found in the AST, including ones with different names than "Configure" (needed if delegated)
func runConfigureHasReceiver(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	configurefuncs := new(FunctionDeclarations)

	checkFunction := func(n ast.Node) {
		funcdecl := n.(*ast.FuncDecl)
		// Make sure we get functions with "*dingo.Injector" in the parameters
		for _, param := range funcdecl.Type.Params.List {
			if param, ok := param.Type.(*ast.StarExpr); ok {
				if param, ok := param.X.(*ast.SelectorExpr); ok {
					paramType := pass.TypesInfo.ObjectOf(param.Sel)
					if paramType.Pkg().Path() == globals.DingoPkgPath && paramType.Name() == "Injector" {
						configurefuncs.all = append(configurefuncs.all, funcdecl)
						// If function has no Receiver, throw an error
						if funcdecl.Recv == nil {
							// TODO add tests
							flanalysis.Report(pass, "Configure function has no Receiver! A type must implement the dingo.Module interface!", funcdecl)
							return
						} else {
							configurefuncs.valid = append(configurefuncs.valid, funcdecl)
						}
					}
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return configurefuncs, nil
}


