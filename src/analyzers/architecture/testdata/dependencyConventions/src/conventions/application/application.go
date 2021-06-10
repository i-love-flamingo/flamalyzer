package application

import "conventions/domain/subDirectory_2" // allowed

type ApplicationStruct struct {
	name string
	subDirectory_2.SubDirectory2Struct
}
