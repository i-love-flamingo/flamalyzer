package configure_has_receiver

import "flamingo.me/dingo"

type otherType struct {
}

func noReceiverFunc (injector *dingo.Injector, otherType *otherType) { // want "Configure function has no Receiver! A type must implement the dingo.Module interface!"

}
func(*otherType) thisIsFine (injector *dingo.Injector, otherType *otherType) {

}
