package inject

import (
	"go/ast"
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
	Requires: []*analysis.Analyzer{inspect.Analyzer, ReceiverAnalyzer},
}

type structObject struct {
	typeSpec   *ast.TypeSpec
	structType *ast.StructType
	fieldList  []*ast.Field
}

// "inject tags" should be used for config injection only, otherwise inject method should be used.
func runTagAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// Using the results of the boundToReference.TagAnalyzer who provides all Inject-Functions
	injectFunctions := pass.ResultOf[ReceiverAnalyzer].([]*ast.FuncDecl)

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
		// If struct don't have fields with inject-tags stop here
		if len(structObject.fieldList) == 0 {
			return
		}
		// Evaluate the Inject-Tags
		for _, field := range structObject.fieldList {
			if expFindEmptyInject.FindStringSubmatch(field.Tag.Value) != nil {
				flanalysis.Report(pass, "Empty Inject-Tags are not allowed! Add more specific naming or use the Inject function for non configuration injections", field.Tag)
			} else {
				if !isTagReferencedInInjectMethod(structObject, field, injectFunctions) {
					flanalysis.Report(pass, "Injections should be referenced in the Inject function! References in the Inject-Function should be found in the same package!", field.Tag)
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return nil, nil
}

// If a inject tag is outside the type must be referenced in the Inject-Function
func isTagReferencedInInjectMethod(structEntity structObject, field *ast.Field, injectFunctions []*ast.FuncDecl) bool {
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
