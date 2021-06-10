package interfaces

import (
	"conventions/application"                // want "Import Dependency Violation: The `domain` group is not allowed to have a dependency on `application`!"
	"conventions/application/infrastructure" // want "Import Dependency Violation: The `domain` group is not allowed to have a dependency on `application`!"
)

type DomainInterfaces2Struct struct {
	name string
	application.ApplicationStruct
	infrastructure.ApplicationInfrastructureStruct
}
