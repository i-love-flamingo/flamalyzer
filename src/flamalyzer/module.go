package flamalyzer

import (
	"log"

	"flamingo.me/dingo"
	"flamingo.me/flamalyzer/src/flamalyzer/configuration"
)

// module to set up the core functionality of Flamalyzer
type module struct{}

// Configure DI
func (m *module) Configure(injector *dingo.Injector) {
	injector.Bind(new(configuration.Config)).In(dingo.Singleton)
	injector.Bind(new(configuration.AnalyzerConfig)).To(new(configuration.Config))
	injector.Bind(new(configuration.CoreConfig)).To(new(configuration.Config))
}

// Run with the given modules.
func Run(modules []dingo.Module) {
	modules = append([]dingo.Module{new(module)}, modules...)

	injector, err := dingo.NewInjector(
		modules...,
	)
	if err != nil {
		log.Fatal(err)
	}
	service, err := injector.GetInstance(new(Controller))

	if err != nil {
		log.Fatal(err)
	}
	service.(*Controller).Run()
}
