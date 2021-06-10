package infrastructure

import (
	"conventions/application/infrastructure"              // allowed
	"conventions/interfaces/subDirectory/SubSubDirectory" // allowed
	"conventions/interfaces/subDirectory_2"               // allowed
)

type InfrastructureStruct struct {
	name string
	SubSubDirectory.SubSubDirectoryStruct
	subDirectory_2.SubDirectory2Struct
	infrastructure.ApplicationInfrastructureStruct
}
