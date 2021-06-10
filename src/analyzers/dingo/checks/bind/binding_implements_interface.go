package bind

import (
	"go/ast"
	"go/types"

	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

// Analyzer checks if a dingo binding to an interface really implements the interface
var Analyzer = &analysis.Analyzer{
	Name:     "checkCorrectInterfaceToInstanceBinding",
	Doc:      "check if the Binding of an Interface to an Implementation with the Bind() -Function is possible",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var dingoTypeDecl = "dingo.Injector"
var dingoPkgPath = "flamingo.me/dingo"

// This function checks if the given instance can be bound to the interface by the bind functions of Dingo.
// example: injector.Bind(someInterface).To(mustImplementSomeInterface)
func run(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	checkFunction := func(n ast.Node) {
		funcdecl := n.(*ast.FuncDecl)

		// Make sure we check functions with "*dingo.Injector" in the parameters
		for _, param := range funcdecl.Type.Params.List {
			if param, ok := param.Type.(*ast.StarExpr); ok {
				if param, ok := param.X.(*ast.SelectorExpr); ok {
					if isSelectorExprTypeOf(param, dingoTypeDecl) {
						checkBlockStatmenetForCorrectBindings(funcdecl.Body, pass)
					}
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return nil, nil
}

// Checks a blockStatement like a function body for correct bindings
func checkBlockStatmenetForCorrectBindings(block *ast.BlockStmt, pass *analysis.Pass) {
	for _, stmt := range block.List {
		if exp, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := exp.X.(*ast.CallExpr); ok {

				// make sure we have a concatenated function
				firstCall, _ := call.Fun.(*ast.SelectorExpr).X.(*ast.CallExpr)
				secondCall := call
				if firstCall == nil {
					continue
				}
				firstFunc, _ := typeutil.Callee(pass.TypesInfo, firstCall).(*types.Func)
				secondFunc, _ := typeutil.Callee(pass.TypesInfo, secondCall).(*types.Func)
				if firstFunc == nil || secondFunc == nil {
					continue
				}

				// Make sure we are using "flamingo.me/dingo"
				if firstFunc.Pkg().Path() != dingoPkgPath || secondFunc.Pkg().Path() != dingoPkgPath {
					continue
				}

				// Make sure the called function is one that "binds" something "to" something
				bindCalls := map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
				toCalls := map[string]bool{"To": true, "ToInstance": true}
				if ok := bindCalls[firstFunc.Name()] && toCalls[secondFunc.Name()]; ok {
					bindType := pass.TypesInfo.Types[firstCall.Args[0]].Type
					toType := pass.TypesInfo.Types[secondCall.Args[0]].Type
					iface := bindType.(*types.Pointer).Elem().(*types.Named).Underlying().(*types.Interface)

					// If struct literal is used, get the toType into the correct format
					_, ok := toType.(*types.Named)
					if ok {
						toType = types.NewPointer(toType)
					}
					if !types.Implements(toType, iface) {
						message := "Incorrect Binding! `" + toType.Underlying().String()[1:] + "` must implement Interface `" + bindType.Underlying().String()[1:] + "`"
						flanalysis.Report(pass, message, secondCall.Args[0])
					}

				}
			}
		}
	}
}

// checks if a given selectorExpression matches the given type (as string)
// example: "dingo.Injector"
func isSelectorExprTypeOf(expr *ast.SelectorExpr, typ string) bool {
	if n, ok := expr.X.(*ast.Ident); ok {
		fullName := n.Name + "." + expr.Sel.Name
		if fullName == typ {
			return true
		}
	}
	return false
}
