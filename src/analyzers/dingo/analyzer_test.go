package dingo_test

import (
	"testing"

	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/bind"
	"flamingo.me/flamalyzer/src/analyzers/dingo/checks/inject"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestProperInjectTags(t *testing.T) {
	analysis := inject.TagAnalyzer
	analysistest.Run(t, analysistest.TestData(), analysis, "proper_inject_tags")
}

func TestBoundToReference(t *testing.T) {
	analysis := inject.ReceiverAnalyzer
	analysistest.Run(t, analysistest.TestData(), analysis, "pointer_receiver")
}

func TestCorrectInterfaceToInstanceBinding(t *testing.T) {
	analysis := bind.Analyzer
	analysistest.Run(t, analysistest.TestData(), analysis, "correct_interface_to_instance_binding")
}
