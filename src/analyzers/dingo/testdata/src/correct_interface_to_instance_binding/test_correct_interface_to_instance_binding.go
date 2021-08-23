package correct_interface_to_instance_binding

import (
	"flamingo.me/dingo"
)

type I interface {
	funA()
	funB()
}

type J interface {
	funA()
}

type A struct{}
type B struct{}

type C struct{}
type D struct{}

func (a *A) funB() {}
func (a *A) funA() {}

func (a *B) funB() {}

type F func(a string) bool

func IsF(b string) bool {
	return false
}

func IsNotF() bool {
	return false
}

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
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""

	injector.Bind(new(J)).To(new(I))
	injector.Bind(new(I)).To(new(J)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.J\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""

	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""

	// check if string literals are allowed
	injector.Bind(new(I)).ToInstance(A{})
	injector.Bind(new(I)).ToInstance(B{}) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""

	injector.BindMulti(new(I)).To(new(A))
	injector.BindMulti(new(I)).To(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
	injector.BindMulti(new(I)).ToInstance(new(A))
	injector.BindMulti(new(I)).ToInstance(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""

	// function binding
	injector.Bind(new(F)).ToInstance(IsF)
	injector.Bind(new(F)).ToInstance(IsNotF) // want "Incorrect Binding! \"func\\(\\) bool\" must have Signature of \"func\\(a string\\) bool\""
	injector.Bind(new(F)).ToInstance(func(string) bool { return false })
	injector.Bind(new(F)).ToInstance(func() bool { return false }) // want "Incorrect Binding! \"func\\(\\) bool\" must have Signature of \"func\\(a string\\) bool\""

	otherType.Bind(new(I)).To(new(A))
	otherType.Bind(new(I)).To(new(B))

	// check struct binding
	injector.Bind(new(A)).In(dingo.Singleton)
	injector.Bind(new(A)).ToInstance(new(A))
	injector.Bind(new(A)).ToInstance(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must be assignable to \"\\*correct_interface_to_instance_binding.A\""

	// check uncommon bind arguments
	injector.Bind((*I)(nil)).To(new(A))
	injector.Bind((*I)(nil)).To(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
}

// check if it works in other functions than "configure"
func (*D) functionName(injector *dingo.Injector, otherType *otherType) {
	injector.Bind(new(I)).To(new(A))
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
}
// TODO Add tests for configure_has_receiver and probably remove this one
// check if it works in functions that have no Receiver
func noReceiverFunc(injector *dingo.Injector, otherType *otherType) {
	injector.Bind(new(I)).To(new(A))
	injector.Bind(new(I)).To(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
	injector.Bind(new(I)).ToInstance(new(A))
	injector.Bind(new(I)).ToInstance(new(B)) // want "Incorrect Binding! \"\\*correct_interface_to_instance_binding.B\" must implement Interface \"\\*correct_interface_to_instance_binding.I\""
}
