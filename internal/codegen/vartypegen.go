package codegen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/ast"
	"lambdagen/internal/model"
)

func ConvertTypeExpr(typeExpr model.TypeExpr) jen.Code {

	switch typeExpr.Mode() {
	case model.TypeExprModeIdent:
		return jen.Id(typeExpr.AsIdent().Name)

	case model.TypeExprModeSelector:
		// TODO for now, i'm assuming that the X part will always be an identifier. if it is not, panic
		selector := typeExpr.AsSelector()
		qualifier, ok := selector.X.(*ast.Ident)
		if !ok {
			panic(fmt.Errorf("expected selector ident, got %T", selector.X))
		}

		return jen.Qual(qualifier.Name, selector.Sel.Name)

	default:
		panic("unreachable")
	}
}

func ConversionCode(ctx *jen.Group, tp model.VariableTp, rawVariable, convertedVariable string) {
	if tp == model.VariableTpString {
		// incoming variable tp is a string, then nothing needs to be done. Just remap the name
		ctx.Id(convertedVariable).Op(":=").Id(rawVariable)
		return
	}

	if tp == model.VariableTpNumber {
		// if a number, needs to be converted using strconv
		ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "Atoi").Call(jen.Id(rawVariable))
		// TODO actually throw an API error, but that doesn't exist yet
		buildCheckError(ctx, failStrategyPanic)
		return
	}

	if tp == model.VariableTpBool {
		// if a number, needs to be converted using strconv
		ctx.List(jen.Id(convertedVariable), jen.Err()).Op(":=").Qual("strconv", "ParseBool").Call(jen.Id(rawVariable))
		// TODO actually throw an API error, but that doesn't exist yet
		buildCheckError(ctx, failStrategyPanic)
		return
	}

	panic("uncovered conversion type")
}
