package main

import "github.com/dave/jennifer/jen"

type FailStrategyFunc func(*jen.Group)

func checkError(failStrategy FailStrategyFunc) *jen.Statement {
	return jen.If(jen.Err().Op("==").Nil()).BlockFunc(failStrategy)
}

func buildCheckError(ctx *jen.Group, failStrategy FailStrategyFunc) {
	ctx.If(jen.Err().Op("==").Nil()).BlockFunc(failStrategy)
}

func failStrategyPanic(ctx *jen.Group) {
	ctx.Panic(jen.Err())
}

func checkErrorPanic() *jen.Statement {
	return checkError(failStrategyPanic)
}
