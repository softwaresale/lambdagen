package model

import "go/ast"

type QualifiedName struct {
	Package string
	Name    string
}

func (name QualifiedName) Empty() bool {
	return len(name.Name) == 0
}

func NewQualifiedName(typeExpr TypeExpr, containingPackage string) QualifiedName {
	switch typeExpr.Mode() {
	case TypeExprModeIdent:
		return QualifiedName{
			Package: containingPackage,
			Name:    typeExpr.AsIdent().Name,
		}

	case TypeExprModeSelector:
		ident, ok := typeExpr.AsSelector().X.(*ast.Ident)
		if !ok {
			panic("expected ident, but got something else")
		}

		return QualifiedName{
			Package: ident.Name,
			Name:    typeExpr.AsSelector().Sel.Name,
		}

	default:
		panic("unreachable")
	}
}
