package bind

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/globals"
	"fmt"
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
					paramType := pass.TypesInfo.ObjectOf(param.Sel)
					// check if param is type dingo.Injector
					if paramType.Pkg().Path() == globals.DingoPkgPath && paramType.Name() == "Injector" {
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
				if firstFunc.Pkg().Path() != globals.DingoPkgPath || secondFunc.Pkg().Path() != globals.DingoPkgPath {
					continue
				}

				// Make sure the called function is one that "binds" something "to" something
				bindCalls := map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
				// TODO probably check for "toProvider" too?
				toCalls := map[string]bool{"To": true, "ToInstance": true}
				if ok := bindCalls[firstFunc.Name()] && toCalls[secondFunc.Name()]; ok {
					bindType := pass.TypesInfo.Types[firstCall.Args[0]].Type
					toType := pass.TypesInfo.Types[secondCall.Args[0]].Type
					// If struct literal is used, get the toType into the correct format
					_, ok := toType.(*types.Named)
					if ok {
						toType = types.NewPointer(toType)
					}
					switch what := bindType.(*types.Pointer).Elem().(*types.Named).Underlying().(type) {
					case *types.Interface:
						// in case of interface to interface binding
						to := toType.(*types.Pointer).Elem().Underlying()
						if !types.Implements(toType, what) && !types.Implements(to, what) {
							message := fmt.Sprintf("Incorrect Binding! %q must implement Interface %q", toType.Underlying().String(), bindType.Underlying().String())
							flanalysis.Report(pass, message, secondCall.Args[0])
						}
					case *types.Signature:
						if !types.AssignableTo(toType, what) {
							message := fmt.Sprintf("Incorrect Binding! %q must have Signature of %q", toType.String(), what.String())
							flanalysis.Report(pass, message, secondCall.Args[0])
						}
					default:
						if !types.AssignableTo(toType, bindType) {
							message := fmt.Sprintf("Incorrect Binding! %q must be assignable to %q", toType.String(), bindType.String())
							flanalysis.Report(pass, message, secondCall.Args[0])
						}
					}
				}
			}
		}
	}
}
