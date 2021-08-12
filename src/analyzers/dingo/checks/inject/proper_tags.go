package inject

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/helper"
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
// - Inject tags can be defined in the Inject-Function or must be referenced if defined outside
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

var dingoPkgPath = "flamingo.me/dingo"

// Todo think about this
var passObject *analysis.Pass

// "inject tags" should be used for config injection only, otherwise inject method should be used.
func runTagAnalyzer(pass *analysis.Pass) (interface{}, error) {
	passObject = pass
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// Using the results of the boundToReference.TagAnalyzer who provides all Inject-Functions
	injectFunctions := pass.ResultOf[ReceiverAnalyzer].([]*ast.FuncDecl)
	configureFunctions := pass.ResultOf[helper.ConfigureDeclAnalyzer].([]*ast.FuncDecl)
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}
	checkFunction := func(n ast.Node) {
		// Check regex at https://www.debuggex.com/#cheatsheet
		expFindInject := regexp.MustCompile("(`(inject:)[\"].*\"`)|(\"(inject:)[`].*`\")")
		expFindEmptyInject := regexp.MustCompile("(`(inject:)[\"]\"`)|(\"(inject:)[`]`\")")

		typeObject := n.(*ast.TypeSpec)
		var structObject = structObject{}

		// If found type is a struct-type go on, else stop here
		if assertion, ok := typeObject.Type.(*ast.StructType); ok {
			structObject.structType = assertion
			structObject.typeSpec = typeObject
		} else {
			return
		}

		// Search for tags in the given struct, if there is a tag, safe them to the fieldList
		for _, f := range structObject.structType.Fields.List {
			if f.Tag != nil {
				if expFindInject.FindStringSubmatch(f.Tag.Value) != nil {
					structObject.fieldList = append(structObject.fieldList, f)
				}
			}
		}
		// If struct doesn't have fields with inject-tags stop here
		if len(structObject.fieldList) == 0 {
			return
		}
		// Evaluate the Inject-Tags
		for _, field := range structObject.fieldList {
			if expFindEmptyInject.FindStringSubmatch(field.Tag.Value) != nil {
				flanalysis.Report(pass, "Empty Inject-Tags are not allowed! Add more specific naming or use the Inject function for non configuration injections", field.Tag)
			} else {
				// TODO Bug here? we dont use field
				if !isTagReferencedInInjectMethod(structObject, field, injectFunctions, configureFunctions) {
					flanalysis.Report(pass, "Injections should be referenced in the Inject function! References in the Inject-Function should be found in the same package!", field.Tag)
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return nil, nil
}

func isTypeUsedAsProvider(configureFunctions []*ast.FuncDecl, structEntity structObject) bool {
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
					firstFunc, _ := typeutil.Callee(passObject.TypesInfo, firstCall).(*types.Func)
					secondFunc, _ := typeutil.Callee(passObject.TypesInfo, secondCall).(*types.Func)
					if firstFunc == nil || secondFunc == nil {
						continue
					}

					// Todo maybe make dingoPkgPath more robust since we use it at two places now
					// Make sure we are using "flamingo.me/dingo"
					if firstFunc.Pkg().Path() != dingoPkgPath || secondFunc.Pkg().Path() != dingoPkgPath {
						continue
					}

					// Make sure the called function is one that "binds" something to "ToProvider"
					bindCalls := map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
					toCalls := map[string]bool{"ToProvider": true}

					if ok := bindCalls[firstFunc.Name()] && toCalls[secondFunc.Name()]; ok {
						// get secondfunc argument type?
						for _, provider := range secondCall.Args {
							// Todo change it so work for selector calls as parameter too -> toProvider(d.func) and toProvider(func)
							var providerParameters = provider.(*ast.Ident).Obj.Decl.(*ast.FuncDecl).Type.Params.List
							// TODO schöner pls
							// check if the given type has a parameter which matches the type of the object with the inject tags
							if providerParameters != nil {
								for _, param := range providerParameters {
									// TODO typsicherheit
									//vergleiche provider parameter typ mit typ in dem das inject tag vorkommt -> POC
									if param.Type.(*ast.StarExpr).X.(*ast.Ident).Name == structEntity.typeSpec.Name.Name{
										return true
									}
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

// If an inject tag is outside the type must be referenced in the Inject-Function
func isTagReferencedInInjectMethod(structEntity structObject, field *ast.Field, injectFunctions []*ast.FuncDecl, configureFunctions []*ast.FuncDecl) bool {
	// if type is referenced as a provider
	if isTypeUsedAsProvider(configureFunctions, structEntity) {
		return true
	}
	if injectFunctions == nil {
		return false
	}
	// Check against all Inject-Functions in the Package
	for _, injectFunc := range injectFunctions {
		for _, param := range injectFunc.Type.Params.List {
			// If parameter has a pointer
			if param, ok := param.Type.(*ast.StarExpr); ok {
				// If the parameter has no selectors, meaning he is directly defined in the current package, compare their name and return true if they are the same.
				if param, ok := param.X.(*ast.Ident); ok {
					if param.Name == structEntity.typeSpec.Name.Name {
						return true
					}
				}
			}
		}
	}
	return false
}
