package inject

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/helper"
	"flamingo.me/flamalyzer/src/analyzers/dingo/globals"
	ast "go/ast"
	"go/types"
	"golang.org/x/tools/go/types/typeutil"
	"regexp"

	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// TagAnalyzer checks if the inject tags are used properly.
// inject tags should be used for config injection only, otherwise an inject method should be used.
// This means:
// - No empty inject tags
// - Inject tags can be defined in the Inject-Function or must be referenced if defined outside. Exception: Those types are used in a Provider-Function
// - They must be declared in the same package as the Inject-Function

var TagAnalyzer = &analysis.Analyzer{
	Name:     "checkProperInjectTags",
	Doc:      "check if convention of using inject tags is respected",
	Run:      runTagAnalyzer,
	Requires: []*analysis.Analyzer{inspect.Analyzer, ReceiverAnalyzer, helper.ConfigureDeclAnalyzer},
}

type structObject struct {
	typeSpec   *ast.TypeSpec
	structType *ast.StructType
	fieldList  []*ast.Field
}

var bindCalls = map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
var toCalls = map[string]bool{"ToProvider": true}

// "inject tags" should be used for config injection only, otherwise inject method should be used.
func runTagAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// Using the results of the boundToReference.TagAnalyzer who provides all Inject-Functions
	injectFunctions := pass.ResultOf[ReceiverAnalyzer].([]*ast.FuncDecl)
	// Get all configureFunctions
	configureFunctions := pass.ResultOf[helper.ConfigureDeclAnalyzer].([]*ast.FuncDecl)

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}
	checkFunction := func(n ast.Node) {
		// Check regex at https://www.debuggex.com/#cheatsheet
		expFindEmptyTag := regexp.MustCompile("(`(inject:)[\"]\"`)|(\"(inject:)[`]`\")")
		expFindTagWithInfo := regexp.MustCompile("(`(inject:)[\"].*\"`)|(\"(inject:)[`].*`\")")

		if structType, ok := n.(*ast.TypeSpec).Type.(*ast.StructType); ok {
			// check structTypes for Tags
			for _, f := range structType.Fields.List {
				if f.Tag != nil {
					if expFindEmptyTag.FindStringSubmatch(f.Tag.Value) != nil {
						flanalysis.Report(pass, "Empty Inject-Tags are not allowed! Add more specific naming or use the Inject function for non configuration injections", f.Tag)
					}
					if expFindTagWithInfo.FindStringSubmatch(f.Tag.Value) != nil {
						// I) Tag referenced as Parameter of an Inject function?
						// II) Tag referenced as Parameter of a Provider? (if Provider used in Configure function)
						if !(isTagReferencedInInjectMethod(n.(*ast.TypeSpec), injectFunctions) || isTypeUsedAsProvider(pass, structType, configureFunctions)) {
							flanalysis.Report(pass, "Injections should be referenced in the Inject/Provider-Function! References in the Inject/Provider-Function should be found in the same package!", f.Tag)

						}
					}
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return nil, nil
}

// Tag referenced as Parameter of a Provider? (if Provider used in Configure function)
func isTypeUsedAsProvider(pass *analysis.Pass, structType *ast.StructType, configureFunctions []*ast.FuncDecl) bool {
	for _, configureFunc := range configureFunctions {
		for _, stmt := range configureFunc.Body.List {
			// TODO this code "checking if we have an injector-bind-to-something" is something we already do in "binding_implements_interface"
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
					toProviderFunc := pass.TypesInfo.Types[secondCall.Args[0]].Type
					// Make sure the called function is one that "binds" something to "ToProvider"
					if ok := bindCalls[firstFunc.Name()] && toCalls[secondFunc.Name()]; ok {
						for i := 0; i < toProviderFunc.(*types.Signature).Params().Len(); i++ {
							// if param is a pointer
							if _, ok := toProviderFunc.(*types.Signature).Params().At(i).Type().(*types.Pointer); ok {
								providerParameter := toProviderFunc.(*types.Signature).Params().At(i).Type().(*types.Pointer).Elem().Underlying()
								typedef := pass.TypesInfo.Types[structType].Type.Underlying().(*types.Struct)
								if types.Identical(providerParameter, typedef) {
									return true
								}

							}
						}
					}
				}
			}
		}
	}
	return false
}

// If an inject tag is defined outside the Inject-Method, the type must be referenced in the Inject-Method
func isTagReferencedInInjectMethod(typeSpec *ast.TypeSpec, injectFunctions []*ast.FuncDecl) bool {
	// Check against all Inject-Functions in the Package
	for _, injectFunc := range injectFunctions {
		for _, param := range injectFunc.Type.Params.List {
			// If parameter has a pointer
			if param, ok := param.Type.(*ast.StarExpr); ok {
				// If the parameter has no selectors, meaning he is directly defined in the current package, compare their name and return true if they are the same.
				if param, ok := param.X.(*ast.Ident); ok {
					if param.Name == typeSpec.Name.Name {
						return true
					}
				}
			}
		}
	}
	return false
}
