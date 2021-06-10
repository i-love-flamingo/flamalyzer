package dependency

import (
	"go/ast"
	"regexp"
	"strings"

	flanalysis "flamingo.me/flamalyzer/src/flamalyzer/analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type analyzer struct {
	Analyzer   *analysis.Analyzer
	Groups     map[string][]string
	EntryPaths []string
}

// NewAnalyzer creates a new dependency-conventions Analyzer with the passed configuration
// This analysis checks that the provided architecture-conventions are respected
// configuration example:
// groups:
//	infrastructure: ["infrastructure", "interfaces", "application", "domain"]
//	interfaces: ["interfaces", "application", "domain"]
//	application: ["application", "domain"]
//	domain: ["domain"]
func NewAnalyzer(groups map[string][]string, entryPaths []string) *analyzer {
	analyzer := new(analyzer)
	analyzer.Groups = groups
	analyzer.EntryPaths = entryPaths
	analyzer.Analyzer = &analysis.Analyzer{
		Name:     "checkDependencyConventions",
		Doc:      "check if the architecture conventions are respected",
		Run:      analyzer.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	return analyzer
}

// Checks if one file has valid imports regarding the specified conventions
func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	// if there are no imports there is no need to check anything
	if pass.Pkg.Imports() == nil {
		return nil, nil
	}
	packagePath := pass.Pkg.Path()
	// if current package is not part of an allowed entryPath, skip
	if !a.allowedEntryPath(a.EntryPaths, packagePath) {
		return nil, nil
	}
	fileGroup := a.getAssociatedGroupFromPath(packagePath)

	// If this file is located inside a group of the defined architecture, check the imports
	if fileGroup != "" {
		input := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		// Run a check for every importSpecification
		nodeFilter := []ast.Node{
			(*ast.ImportSpec)(nil),
		}
		checkFunction := func(n ast.Node) {
			importSpec := n.(*ast.ImportSpec)
			// If importSpec is not part of an allowed entryPath, skip
			if !a.allowedEntryPath(a.EntryPaths, importSpec.Path.Value) {
				return
			}
			importGroup := a.getAssociatedGroupFromPath(importSpec.Path.Value)

			// If the import is located inside of a group, check if the import is allowed
			if importGroup != "" {
				approved := false
				for _, allowedDependency := range a.Groups[fileGroup] {
					if allowedDependency == importGroup {
						approved = true
						break
					}
				}
				if !approved {
					message := "Import Dependency Violation: The `" + fileGroup + "` group is not allowed to have a dependency on `" + importGroup + "`!\n Faulty Import:"
					flanalysis.Report(pass, message, importSpec)
				}
			}
		}
		input.Preorder(nodeFilter, checkFunction)
	}
	return nil, nil
}

// Compares entryPaths with the current filePath
func (a *analyzer) allowedEntryPath(entryPaths []string, path string) bool {
	if len(entryPaths) == 0 {
		return true
	}
	for _, e := range entryPaths {
		if strings.Contains(strings.ToLower(path), strings.ToLower(e)) {
			return true
		}
	}
	return false
}

// Checks if a path belongs to a group
func (a *analyzer) getAssociatedGroupFromPath(path string) string {
	var lowestIndex *int
	var associatedGroup string
	for groupName := range a.Groups {
		exp := groupName + `["]?$|/` + groupName + `/`
		expRes := regexp.MustCompile(exp).FindStringSubmatchIndex(path)
		if expRes != nil {
			if lowestIndex == nil {
				lowestIndex = &expRes[0]
				associatedGroup = groupName
			} else if expRes[0] < *lowestIndex {
				lowestIndex = &expRes[0]
				associatedGroup = groupName
			}
		}
	}
	if associatedGroup != "" {
		return associatedGroup
	}
	return ""
}
