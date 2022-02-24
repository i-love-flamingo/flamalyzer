package inject

import (
	"go/ast"
	"reflect"

	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ReceiverAnalyzer checks if an inject-function is bound to a pointer-receiver
// returns all inject functions as a resultType
var ReceiverAnalyzer = &analysis.Analyzer{
	Name:       "checkPointerReceiver",
	Doc:        "check if the inject method is bound to a pointer receiver",
	Run:        runReceiverAnalyzer,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeOf(*new([]*ast.FuncDecl)),
}

// If a function is an "Inject" function it must be bound to a pointer receiver
// this function simply returns a List of all Inject-Functions found in the AST used by another analyzer
func runReceiverAnalyzer(pass *analysis.Pass) (interface{}, error) {
	input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	var injectFunctions []*ast.FuncDecl

	// Name must be "Inject" and must have a *function receiver
	checkFunction := func(n ast.Node) {
		funcdecl := n.(*ast.FuncDecl)
		// If function has no Receiver, pass
		if funcdecl.Recv == nil {
			return
		}
		if funcdecl.Name.Name == "Inject" {
			for _, rec := range funcdecl.Recv.List {
				var _, isStarType = rec.Type.(*ast.StarExpr)

				// If it is an inject-function but the pointer(*) is missing, report with a suggested fix
				if !isStarType {
					suggestedFix := &analysis.SuggestedFix{
						Message: "Add missing Pointer",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     rec.Pos(),
								End:     rec.End(),
								NewText: []byte(rec.Names[0].Name + " *" + rec.Type.(*ast.Ident).Name),
							},
						},
					}
					flanalysis.ReportWithSuggestedFixes(pass, "Missing pointer in function receiver. Inject method must have a pointer receiver!\n %s", funcdecl.Recv.List[0].Type, []analysis.SuggestedFix{*suggestedFix})
				} else {

					// If it is an inject-function add it to the list of injectFunctions
					injectFunctions = append(injectFunctions, funcdecl)
				}
			}
		}
	}
	input.Preorder(nodeFilter, checkFunction)
	return injectFunctions, nil
}
