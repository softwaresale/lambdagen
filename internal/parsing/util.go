package parsing

import (
	"go/ast"
	"go/token"
)

func genDeclIsTypeSpec(genDecl *ast.GenDecl) *ast.TypeSpec {
	// only interested in type decls
	if genDecl.Tok != token.TYPE {
		return nil
	}

	// there should be a single handler
	if len(genDecl.Specs) != 1 {
		return nil
	}

	typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil
	}

	return typeSpec
}
