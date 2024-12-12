package model

type VariableTp string

const (
	VariableTpString = "string"
	VariableTpNumber = "number"
	VariableTpBool   = "bool"
)

func (tp VariableTp) Valid() bool {
	switch tp {
	case VariableTpString, VariableTpNumber, VariableTpBool:
		return true
	default:
		return false
	}
}
