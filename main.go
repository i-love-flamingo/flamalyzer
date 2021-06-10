package main

import (
	"flamingo.me/dingo"
	"flamingo.me/flamalyzer/src/analyzers/architecture"
	dingoAnalyzer "flamingo.me/flamalyzer/src/analyzers/dingo"
	"flamingo.me/flamalyzer/src/flamalyzer"
)

func main() {
	flamalyzer.Run(
		[]dingo.Module{
			new(dingoAnalyzer.Module),
			new(architecture.Module),
		},
	)
}
