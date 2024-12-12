package parsing

import (
	"fmt"
	"go/ast"
	"lambdagen/internal/model"
	"regexp"
	"strings"
)

func (builder *LambdaModelBuilder) extractHandlerConfig(parsedAst *ast.File, typeName *ast.Ident) (model.RequestConfig, error) {
	var typeSpec *ast.TypeSpec

searchLoop:
	for _, decl := range parsedAst.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			tp := genDeclIsTypeSpec(decl)
			if tp == nil {
				continue
			}

			// verify that the type names match
			if tp.Name.String() != typeName.String() {
				continue
			}

			// we have what we want, so break our loop
			typeSpec = tp
			break searchLoop

		default:
			continue
		}
	}

	if typeSpec == nil {
		return model.RequestConfig{}, fmt.Errorf("no handler config found for type %s", typeName)
	}

	configStruct, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return model.RequestConfig{}, fmt.Errorf("%s: expected request config arg to be a struct, got %s instead", builder.fset.Position(configStruct.Pos()), typeSpec.Type)
	}

	config := model.RequestConfig{
		Type:           typeName,
		PathVariables:  []model.VariableDefinition{},
		QueryVariables: []model.VariableDefinition{},
		RequestBody:    model.VariableDefinition{},
	}

	// flip through the fields and extract stuff based on the tags
	for _, field := range configStruct.Fields.List {
		// Needs a tag and a name
		if field.Tag == nil || len(field.Names) == 0 {
			continue
		}

		// find a role in the tag. If there is no role, we skip this field
		role, hasRole := findRoleInFieldTag(field.Tag.Value)
		if !hasRole {
			continue
		}

		// ensure that there is only one name -- don't share types
		if len(field.Names) > 1 {
			return model.RequestConfig{}, fmt.Errorf("%s: only one name is allowed", builder.fset.Position(field.Pos()))
		}

		var typeExpr model.TypeExpr
		var err error
		switch role.Type {
		case model.ObjectRoleBody, model.ObjectRolePathVar, model.ObjectRoleQueryParam:
			typeExpr, err = builder.checkVariableType(field.Type)
		default:
			return model.RequestConfig{}, fmt.Errorf("%s: invalid role %s on field", builder.fset.Position(field.Pos()), role.Type)
		}

		if err != nil {
			return model.RequestConfig{}, fmt.Errorf("%s: %w", builder.fset.Position(field.Pos()), err)
		}

		// TODO dirty way to yank the name, good for now
		fieldName := field.Names[0].Name

		def := model.VariableDefinition{
			Name:      role.Args,
			Type:      typeExpr,
			FieldName: fieldName,
		}

		switch role.Type {
		case model.ObjectRoleBody:
			config.RequestBody = def
		case model.ObjectRolePathVar:
			config.QueryVariables = append(config.QueryVariables, def)
		case model.ObjectRoleQueryParam:
			config.QueryVariables = append(config.QueryVariables, def)
		default:
			panic("shouldn't be any of these")
		}
	}

	return config, nil
}

func (builder *LambdaModelBuilder) checkVariableType(fieldTypeExpr ast.Expr) (model.TypeExpr, error) {
	switch fieldTypeExpr := fieldTypeExpr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return model.TypeExprOf(fieldTypeExpr), nil

	default:
		return model.TypeExpr{}, fmt.Errorf("only ")
	}
}

func findRoleInFieldTag(tag string) (model.ObjectRole, bool) {
	parser := regexp.MustCompile(`lambdagen:"([^)]+)"`)
	match := parser.FindStringSubmatch(tag)
	if match == nil {
		return model.ObjectRole{}, false
	}

	tagValues := strings.Split(match[1], ",")
	if len(tagValues) < 1 {
		return model.ObjectRole{}, false
	}

	if !model.IsValidRoleStr(tagValues[0]) {
		return model.ObjectRole{}, false
	}

	// check if there is an override of the name
	overrideName := ""
	if len(tagValues) > 1 {
		overrideName = tagValues[1]
	}

	return model.ObjectRole{
		Type: tagValues[0],
		Args: overrideName,
	}, true
}
