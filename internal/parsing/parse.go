package parsing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"lambdagen/internal/model"
)

func ParseHandlerFile(filename string) ([]model.ServiceDefinition, error) {

	fset := token.NewFileSet()
	parsedAst, err := parser.ParseFile(fset, filename, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	builder := LambdaModelBuilder{
		fset: fset,
	}

	/*
		err = ast.Print(fset, parsedAst)
		if err != nil {
			return fmt.Errorf("failed to print AST: %w", err)
		}
	*/

	return builder.findServices(parsedAst)
}

type LambdaModelBuilder struct {
	fset *token.FileSet
}

func (builder *LambdaModelBuilder) findServices(parsedAst *ast.File) ([]model.ServiceDefinition, error) {
	var err error
	serviceDecls, err := builder.findServiceDecls(parsedAst)
	if err != nil {
		return nil, err
	}

	var services []model.ServiceDefinition
	for _, serviceDecl := range serviceDecls {
		// find an initializer for this service
		initDecl, err := builder.findServiceInitFunc(parsedAst, &serviceDecl)
		if err != nil {
			return nil, err
		}

		handlers, err := builder.findServiceHandlers(parsedAst, &serviceDecl)
		if err != nil {
			return nil, err
		}

		serviceDefinition := model.ServiceDefinition{
			Type:     serviceDecl,
			Init:     initDecl,
			Handlers: handlers,
		}

		services = append(services, serviceDefinition)
	}

	return services, nil
}
