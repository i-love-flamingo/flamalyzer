package bind

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/configure"
	"flamingo.me/flamalyzer/src/analyzers/dingo/globals"
	"fmt"
	"go/ast"
	"go/types"
	"reflect"

	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// Analyzer checks if a dingo binding to an interface really implements the interface
// Returns all dingo-binding declarations
var Analyzer = &analysis.Analyzer{
	Name:       "checkBindingImplementsInterface",
	Doc:        "check if the Binding of an Interface to an Implementation with the Bind() -Function is possible",
	Run:        run,
	Requires:   []*analysis.Analyzer{configure.ReceiverAnalyzer},
	ResultType: reflect.TypeOf(*new(DingoBindings)),
}
var dingoBindings DingoBindings

// can be DingoInstanceBinding or DingoProviderBinding
type DingoBindings []interface{}

// Todo better use getters?

type DingoInstanceBinding struct {
	BindCall *ast.CallExpr
	ToCall   *ast.CallExpr
	BindFunc types.Object
	ToFunc   types.Object
}
type DingoProviderBinding struct {
	BindCall *ast.CallExpr
	ToCall   *ast.Ident
	BindFunc types.Object
	ToFunc   types.Object
}

// This function checks if the given instance can be bound to the interface by the bind functions of Dingo.
// example: injector.Bind(someInterface).To(mustImplementSomeInterface)
func run(pass *analysis.Pass) (interface{}, error) {
	configureFunctions := pass.ResultOf[configure.ReceiverAnalyzer].(*configure.FunctionDeclarations).GetAll()

	for _, f := range configureFunctions {
		checkBlockStatmenetForCorrectBindings(f.Body, pass)
	}
	return dingoBindings, nil
}

// Checks a blockStatement like a function body for correct bindings
func checkBlockStatmenetForCorrectBindings(block *ast.BlockStmt, pass *analysis.Pass) {
	for _, stmt := range block.List {
		if exp, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := exp.X.(*ast.CallExpr); ok {

				var bindCall interface{}
				var toCall interface{}
				var bindFunc types.Object
				var toFunc types.Object

				switch node := call.Fun.(*ast.SelectorExpr).X.(type) {
				// if it is a concatenated binding
				case *ast.CallExpr:
					bindCall = node
					toCall = call
					bindFunc, _ = typeutil.Callee(pass.TypesInfo, bindCall.(*ast.CallExpr)).(*types.Func)
					toFunc, _ = typeutil.Callee(pass.TypesInfo, toCall.(*ast.CallExpr)).(*types.Func)
					// Make sure we are using "flamingo.me/dingo"
					if bindFunc.Pkg().Path() != globals.DingoPkgPath || toFunc.Pkg().Path() != globals.DingoPkgPath {
						continue
					}
				// if it is a split binding
				case *ast.Ident:
					bindCall, ok = node.Obj.Decl.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
					evoNode := node
					for !ok {
						bindCall, ok = evoNode.Obj.Decl.(*ast.AssignStmt).Rhs[0].(*ast.Ident).Obj.Decl.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
						evoNode, _ = evoNode.Obj.Decl.(*ast.AssignStmt).Rhs[0].(*ast.Ident)
					}
					toCall, ok = call.Args[0].(*ast.CallExpr)
					// !ok -> split provider binding
					if !ok {
						// e.g something.ToProvider(Provider)
						if _, isIdent := call.Args[0].(*ast.Ident); isIdent {
							toCall = call.Args[0].(*ast.Ident)

						// if selector expression e.g something.ToProvider(selector.Provider)
						} else if _, isSelectorExpr := call.Args[0].(*ast.SelectorExpr); isSelectorExpr {
							toCall = call.Args[0].(*ast.SelectorExpr).Sel
						} else{
							continue
						}
					}

					// TODO toProvider and bindings: selector expression references outside of package
					bindFunc, _ = typeutil.Callee(pass.TypesInfo, bindCall.(*ast.CallExpr)).(*types.Func)
					toFunc, _ = pass.TypesInfo.ObjectOf(call.Fun.(*ast.SelectorExpr).Sel).(types.Object)

					// Make sure we are using "flamingo.me/dingo"
					if bindFunc.Pkg().Path() != globals.DingoPkgPath {
						continue
					}
				default:
					continue
				}

				if bindFunc == nil || toFunc == nil {
					continue
				}
				// Now we have potential function bindings add them to stack if they have the correct name (BindCalls & toCalls)
				fillReturnData(bindCall.(*ast.CallExpr), toCall, bindFunc, toFunc)

				// Make sure the called function is one that "binds" something "to" something
				bindCalls := map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
				// TODO probably check for "toProvider" too?
				toCalls := map[string]bool{"To": true, "ToInstance": true}
				if ok := bindCalls[bindFunc.Name()] && toCalls[toFunc.Name()]; ok {
					bindType := pass.TypesInfo.Types[bindCall.(*ast.CallExpr).Args[0]].Type
					toType := pass.TypesInfo.Types[toCall.(*ast.CallExpr).Args[0]].Type
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
							flanalysis.Report(pass, message, toCall.(*ast.CallExpr).Args[0])
						}
					case *types.Signature:
						if !types.AssignableTo(toType, what) {
							message := fmt.Sprintf("Incorrect Binding! %q must have Signature of %q", toType.String(), what.String())
							flanalysis.Report(pass, message, toCall.(*ast.CallExpr).Args[0])
						}
					default:
						if !types.AssignableTo(toType, bindType) {
							message := fmt.Sprintf("Incorrect Binding! %q must be assignable to %q", toType.String(), bindType.String())
							flanalysis.Report(pass, message, toCall.(*ast.CallExpr).Args[0])
						}
					}
				}
			}

		}
	}
}

func fillReturnData(bindCall *ast.CallExpr, toCall interface{}, bindFunc types.Object, toFunc types.Object) {
	bindCalls := map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
	toCalls := map[string]bool{"To": true, "ToInstance": true, "ToProvider": true}

	if ok := bindCalls[bindFunc.Name()] && toCalls[toFunc.Name()]; ok {
		switch c := toCall.(type) {
		case *ast.CallExpr:
			var binding = DingoInstanceBinding{
				BindCall: bindCall,
				ToCall:   c,
				BindFunc: bindFunc,
				ToFunc:   toFunc,
			}
			dingoBindings = append(dingoBindings, binding)
		case *ast.Ident:
			var binding = DingoProviderBinding{
				BindCall: bindCall,
				ToCall:   c,
				BindFunc: bindFunc,
				ToFunc:   toFunc,
			}
			dingoBindings = append(dingoBindings, binding)
		}
	}

}
