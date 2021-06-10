package pointer_receiver

type A struct{}
type B struct{}

func (a A) Inject() { // want `Missing pointer in function receiver. Inject method must have a pointer receiver!`
}
func (b *B) Inject() { // no Error
}
