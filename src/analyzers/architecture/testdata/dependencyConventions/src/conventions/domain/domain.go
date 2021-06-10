package domain

import "conventions/domain/subDirectory/interfaces" // no error

type DomainStruct struct {
	name string
	interfaces.DomainInterfacesStruct
}
