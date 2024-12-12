package model

import "go/ast"

type TypeExprMode uint8

const (
	TypeExprModeIdent    TypeExprMode = 1
	TypeExprModeSelector              = 2
)

type TypeExpr struct {
	expr ast.Expr
	mode TypeExprMode
}

func TypeExprOf(expr ast.Expr) TypeExpr {
	switch expr.(type) {
	case *ast.Ident:
		return TypeExpr{expr: expr, mode: TypeExprModeIdent}
	case *ast.SelectorExpr:
		return TypeExpr{expr: expr, mode: TypeExprModeSelector}
	default:
		return TypeExpr{expr: nil, mode: 0}
	}
}

func (e *TypeExpr) Valid() bool {
	switch e.mode {
	case TypeExprModeIdent, TypeExprModeSelector:
		return true
	default:
		return false
	}
}

func (e *TypeExpr) Mode() TypeExprMode {
	return e.mode
}

func (e *TypeExpr) IsIdent() bool {
	return e.mode == TypeExprModeIdent
}

func (e *TypeExpr) IsSelector() bool {
	return e.mode == TypeExprModeSelector
}

func (e *TypeExpr) AsIdent() *ast.Ident {
	if e.mode != TypeExprModeIdent {
		return nil
	}

	return e.expr.(*ast.Ident)
}

func (e *TypeExpr) AsSelector() *ast.SelectorExpr {
	if e.mode != TypeExprModeSelector {
		return nil
	}

	return e.expr.(*ast.SelectorExpr)
}
