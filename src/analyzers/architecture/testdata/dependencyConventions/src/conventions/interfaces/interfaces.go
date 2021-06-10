package interfaces

import (
	"conventions/application"                    // allowed
	"conventions/application/infrastructure"     // allowed
	infrastructure2 "conventions/infrastructure" // want "Import Dependency Violation: The `interfaces` group is not allowed to have a dependency on `infrastructure`!"
	"conventions/infrastructure/subDirectory"    // want "Import Dependency Violation: The `interfaces` group is not allowed to have a dependency on `infrastructure`!"
)

type InterfacesStruct struct {
	name string
	application.ApplicationStruct
	infrastructure.ApplicationInfrastructureStruct
	infrastructure2.InfrastructureStruct
	subDirectory.SubDirectoryStruct
}
