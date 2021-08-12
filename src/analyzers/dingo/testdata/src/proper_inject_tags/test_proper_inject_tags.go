package proper_inject_tags

import "flamingo.me/dingo"
// TODO Empty Inject Tags without "" are not tested e.g X string `inject:`
type A struct {
	X string `inject:""` // want `Empty Inject-Tags are not allowed! Add more specific naming or use the Inject function for non configuration injections`
}

type B struct {
	X bool `inject:"config:flag,optional"` // want `Injections should be referenced in the Inject function! References in the Inject-Function should be found in the same package!`
}

type C struct {
	X bool `inject:"err:flag,optional"` // want `Injections should be referenced in the Inject function! References in the Inject-Function should be found in the same package!`
	Y *A   `inject:"annotated"`         // want `Injections should be referenced in the Inject function! References in the Inject-Function should be found in the same package!`
}

type Mapper interface {
	Map()
}

type D struct {
	service  *A
	isDebug bool
	mapper  Mapper
}



type E struct {
	service  *A `inject:"this:is:fun"` // this is allowed since this type referenced in the configure method as provider
}

type specialInjectCfg struct {
	// this is allowed because it is used as argument type in an inject function
	IsDebug bool `inject:"config:isDebug"`
}

type Z struct {
	X bool `inject:""` // want `Empty Inject-Tags are not allowed! Add more specific naming or use the Inject function for non configuration injections`
}

func providerFunc(e *E) interface{}{
	return new(interface{})
}

func (d *D) Inject(
	service *A,
	z *Z,
	cfg *specialInjectCfg,
	annotated *struct {
		Mapper Mapper `inject:"my-annotations,optional"` // this is allowed as it is an argument of the Inject function
	},
) *D {

	d.service = service
	if cfg != nil {
		d.isDebug = cfg.IsDebug
	}

	return d
}


func (d *D) Configure(injector *dingo.Injector) bool{
	injector.Bind(new(interface{})).ToProvider(providerFunc)

	return true
}
