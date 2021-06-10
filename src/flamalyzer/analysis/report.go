// Package analysis offers helpers for an easier and nicer reporting
package analysis

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// returns the pretty-print of a given node
func prettyPrint(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}

// Report an Error with nice printing. The Node which is the Point-of-Failure must be passed.
//
// Example:
// analysis.Report(pass,"your error message", pointOfFailure)
func Report(pass *analysis.Pass, message string, corruptNode ast.Node, args ...interface{}) {
	args = append(args, prettyPrint(pass.Fset, corruptNode))
	pass.Reportf(corruptNode.Pos(), message+"\n%s", args...)
}

// ReportWithSuggestedFixes reports an Error with nice printing and suggested fixes. The Node which is the Point-of-Failure must be passed.
//
// Example:
//
// toolbox.ReportWithSuggestedFixes(pass,"your error message %s", pointOfFailure, []analysis.SuggestedFix{
//	{
//		Message: "make code correct",
//		TextEdits: []analysis.TextEdit{
//			{
//				Pos:     n.Pos(),
//				End:     n.End(),
//				NewText: []byte("correct Code"),
//	 		},
//	 	},
//	 },
// })
func ReportWithSuggestedFixes(pass *analysis.Pass, format string, corruptNode ast.Node, suggestedFixes []analysis.SuggestedFix) {
	msg := fmt.Sprintf(format, prettyPrint(pass.Fset, corruptNode))
	pass.Report(analysis.Diagnostic{Pos: corruptNode.Pos(), Message: msg, SuggestedFixes: suggestedFixes})
}
