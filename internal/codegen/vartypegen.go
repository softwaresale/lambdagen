package codegen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/types"
)

func ConversionCode(ctx *jen.Group, tp types.Type, rawVariable, convertedVariable string) {

	switch tp := tp.(type) {
	case *types.Basic:
		switch tp.Kind() {
		case types.Bool:
			// if a number, needs to be converted using strconv
			ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "ParseBool").Call(jen.Id(rawVariable))
			// TODO actually throw an API error, but that doesn't exist yet
			buildCheckError(ctx, failStrategyPanic)

		case types.String:
			ctx.Id(convertedVariable).Op(":=").Id(rawVariable)

		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64, types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			// if a number, needs to be converted using strconv
			ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "Atoi").Call(jen.Id(rawVariable))
			// TODO actually throw an API error, but that doesn't exist yet
			buildCheckError(ctx, failStrategyPanic)

		default:
			panic(fmt.Sprintf("unsupported basic type: %v", tp.Kind()))
		}

	case *types.Named:
		panic("not yet implemented")

	default:
		panic("unsupported variable type")
	}

}
