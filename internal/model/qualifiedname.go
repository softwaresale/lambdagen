package model

import (
	"go/types"
)

type QualifiedName struct {
	Package string
	Name    string
}

func (name QualifiedName) Empty() bool {
	return len(name.Name) == 0
}

func NewQualifiedName(varType types.Type) QualifiedName {
	panic("todo")
}
