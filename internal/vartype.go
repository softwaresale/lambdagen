package main

import "github.com/dave/jennifer/jen"

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

func (tp VariableTp) ConversionCode(ctx *jen.Group, rawVariable, convertedVariable string) {
	if tp == VariableTpString {
		// incoming variable tp is a string, then nothing needs to be done. Just remap the name
		ctx.Id(convertedVariable).Op(":=").Id(rawVariable)
		return
	}

	if tp == VariableTpNumber {
		// if a number, needs to be converted using strconv
		ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "Atoi").Call(jen.Id(rawVariable))
		// TODO actually throw an API error, but that doesn't exist yet
		buildCheckError(ctx, failStrategyPanic)
		return
	}

	if tp == VariableTpBool {
		// if a number, needs to be converted using strconv
		ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "ParseBool").Call(jen.Id(rawVariable))
		// TODO actually throw an API error, but that doesn't exist yet
		buildCheckError(ctx, failStrategyPanic)
		return
	}

	panic("uncovered conversion type")
}
