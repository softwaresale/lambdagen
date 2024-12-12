package parsing

import (
	"errors"
	"fmt"
	"go/ast"
	"lambdagen/internal/model"
	"regexp"
)

func (builder *LambdaModelBuilder) findServiceDecls(parsedAst *ast.File) ([]model.ServiceType, error) {

	var services []model.ServiceType
	for _, decl := range parsedAst.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:

			typeSpec := genDeclIsTypeSpec(decl)
			if typeSpec == nil {
				continue
			}

			// parse the document string for the token "lambdagen:handler"
			if decl.Doc == nil {
				// if not doc string, can't be annotated
				continue
			}

			var docStrings []string
			for _, item := range decl.Doc.List {
				docStrings = append(docStrings, item.Text)
			}

			typeRole, found := model.ParseObjectRoleDocstring(docStrings...)
			if !found {
				fmt.Printf("decl %s has no valid role, so skipping", builder.fset.Position(decl.Specs[0].Pos()))
				continue
			}

			// ensure that the decl is of a struct
			_, isAStruct := typeSpec.Type.(*ast.StructType)
			if isAStruct && typeRole.Type != model.ObjectRoleServiceTp {
				return nil, fmt.Errorf("object role '%s' on a struct is invalid. Only use '%s'", typeRole, model.ObjectRoleServiceTp)
			}

			// parse configuration
			config, err := parseServiceArgs(typeRole.Args)
			if err != nil {
				return nil, fmt.Errorf("error while parsing service config: %w", err)
			}

			serviceType := model.ServiceType{
				TypeSpec: typeSpec,
				Config:   config,
			}

			// if everything passed, then we correctly found a handler
			services = append(services, serviceType)

		default:
			continue
		}
	}

	return services, nil
}

func parseServiceArgs(args string) (map[string]string, error) {
	configValues := make(map[string]string)
	parser := regexp.MustCompile(`([a-zA-Z_]\w*)=(\S+)`)
	for _, matches := range parser.FindAllStringSubmatch(args, -1) {
		key := matches[1]
		value := matches[2]

		configValues[key] = value
	}

	return configValues, nil
}

func (builder *LambdaModelBuilder) findServiceInitFunc(parsedAst *ast.File, serviceTp *model.ServiceType) (*ast.FuncDecl, error) {
	var err error
	// find a top level function designated as initializer
	for _, decl := range parsedAst.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:

			// if no docstring, can't be initializer
			if decl.Doc == nil {
				continue
			}

			var docstrings []string
			for _, doc := range decl.Doc.List {
				docstrings = append(docstrings, doc.Text)
			}

			role, found := model.ParseObjectRoleDocstring(docstrings...)
			if !found {
				continue
			}

			// if role was service_init, then we found it
			if role.Type != model.ObjectRoleServiceInit {
				continue
			}

			// check that the service initializer is valid
			err = builder.checkServiceInitResults(decl.Type.Results, serviceTp)
			if err != nil {
				return nil, err
			}

			// if valid, then return
			return decl, nil

		default:
			continue
		}
	}

	return nil, errors.New("no service initializer declaration found")
}
