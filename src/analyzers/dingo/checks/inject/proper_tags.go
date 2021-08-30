package inject

import (
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/bind"
	ast "go/ast"
	"go/types"
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
	Name:     "checkStrictTagsAndFunctions",
	Doc:      "check if convention of using inject tags is respected",
	Run:      runTagAnalyzer,
	Requires: []*analysis.Analyzer{inspect.Analyzer, ReceiverAnalyzer, bind.Analyzer},
}

var bindCalls = map[string]bool{"Bind": true, "BindMulti": true, "BindMap": true}
var toCalls = map[string]bool{"ToProvider": true}

// "inject tags" should be used for config injection only, otherwise inject method should be used.
func runTagAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// Using the results of the boundToReference.TagAnalyzer who provides all Inject-Functions
	injectFunctions := pass.ResultOf[ReceiverAnalyzer].([]*ast.FuncDecl)
	// Get all dingo-bindings
	dingoBindings := pass.ResultOf[bind.Analyzer].(bind.DingoBindings)

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
						if !(isTagReferencedInInjectMethod(n.(*ast.TypeSpec), injectFunctions) || isTypeUsedAsProvider(pass, structType, dingoBindings)) {
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
func isTypeUsedAsProvider(pass *analysis.Pass, structType *ast.StructType, bindings bind.DingoBindings) bool {
	for _, b := range bindings {
		switch bt := b.(type) {
		case bind.DingoInstanceBinding:
			binding := bt
			toProviderFunc := pass.TypesInfo.Types[binding.ToCall.Args[0]].Type
			// Make sure the called function is one that "binds" something to "ToProvider"
			if ok := bindCalls[binding.BindFunc.Name()] && toCalls[binding.ToFunc.Name()]; ok {
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

		case bind.DingoProviderBinding:
			binding := bt
			toProviderFunc := pass.TypesInfo.ObjectOf(binding.ToCall).Type()

			// TODO more or less same code than above....
			//Make sure the called function is one that "binds" something to "ToProvider"
			if ok := bindCalls[binding.BindFunc.Name()] && toCalls[binding.ToFunc.Name()]; ok {
				for i := 0; i < toProviderFunc.(*types.Signature).Params().Len(); i++ {
					// TODO What if param is not a pointer?
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
		default:
			continue
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
