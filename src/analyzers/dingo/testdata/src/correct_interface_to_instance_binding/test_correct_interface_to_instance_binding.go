package correct_interface_to_instance_binding

import "flamingo.me/dingo"

type I interface {
	funA()
	funB()
}

type A struct{}
type B struct{}

type C struct{}
type D struct{}

func (a *A) funB() {}
func (a *A) funA() {}

func (a *B) funB() {}

type otherType struct {
}

func (ot *otherType) Bind(what interface{}) *otherType {
	return ot
}

func (*otherType) To(what interface{}) {
}

func (*C) Configure(injector *dingo.Injector, otherType *otherType) {
	// check different bind possibilities
	injector.Bind(new(I)).To(new(A))
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"

	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"

	// check if string literals are allowed
	injector.Bind(new(I)).ToInstance(A{})
	injector.Bind(new(I)).ToInstance(B{}) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"

	injector.BindMulti(new(I)).To(new(A))
	injector.BindMulti(new(I)).To(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"
	injector.BindMulti(new(I)).ToInstance(new(A))
	injector.BindMulti(new(I)).ToInstance(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"

	otherType.Bind(new(I)).To(new(A))
	otherType.Bind(new(I)).To(new(B))
}

// check if it works in other functions than "configure"
func (*D) functionName(injector *dingo.Injector, otherType *otherType) {
	injector.Bind(new(I)).To(new(A))
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"
	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"
}

// check if it works in functions that have no Receiver
func noReceiverFunc(injector *dingo.Injector, otherType *otherType) {
	injector.Bind(new(I)).To(new(A))
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"
	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! `*correct_interface_to_instance_binding.B` must implement Interface `correct_interface_to_instance_binding.I`"
}
