package parsing

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"lambdagen/internal/model"
	"log"
	"strings"
)

type ServiceParser struct {
	pkg        *packages.Package
	syntax     *ast.File
	commentMap ast.CommentMap
}

func ParseServices(basePath, modulePath string) ([]model.ServiceDefinition, error) {
	packageConfig := packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  basePath,
	}

	srcPackages, err := packages.Load(&packageConfig, modulePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading packages:\n%w", err)
	}

	var serviceDefinitions []model.ServiceDefinition
	for _, pkg := range srcPackages {
		for _, syntax := range pkg.Syntax {

			commentMap := ast.NewCommentMap(pkg.Fset, syntax, syntax.Comments)

			parser := ServiceParser{
				pkg:        pkg,
				syntax:     syntax,
				commentMap: commentMap,
			}

			handlers, err := parser.parseServiceDefinitions(syntax.Decls)
			if err != nil {
				return nil, err
			}

			serviceDefinitions = append(serviceDefinitions, handlers...)
		}
	}

	return serviceDefinitions, nil
}

func (parser *ServiceParser) parseServiceDefinitions(decls []ast.Decl) ([]model.ServiceDefinition, error) {

	var serviceHandlerObjects []types.Object
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok != token.TYPE {
				continue
			}

			if decl.Doc == nil {
				continue
			}

			// Determine the role
			role, valid := model.ParseObjectRoleDocstring(decl.Doc.Text())
			if !valid {
				continue
			}

			// if the role is not a service handler, bail
			if role.Type != model.ObjectRoleServiceTp {
				continue
			}

			// this type is a service type, so we need to find any structs
			for _, spec := range decl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				handlerObj := parser.pkg.TypesInfo.ObjectOf(typeSpec.Name)
				serviceHandlerObjects = append(serviceHandlerObjects, handlerObj)
			}

		default:
			continue
		}
	}

	var services []model.ServiceDefinition
	for _, handler := range serviceHandlerObjects {
		fmt.Printf("got handler: %s\n", handler.String())

		service, err := parser.parseServiceDefinition(handler)
		if err != nil {
			fmt.Printf("while parsing service def:\n%e\n", err)
			continue
		}

		services = append(services, service)
	}

	return services, nil
}

func (parser *ServiceParser) parseServiceDefinition(handlerObj types.Object) (model.ServiceDefinition, error) {

	// find the service initializer
	serviceInit := parser.findServiceInitializer(handlerObj)
	if serviceInit == nil {
		return model.ServiceDefinition{}, fmt.Errorf("service has now initializer for %s", handlerObj.String())
	}

	initializerFunctionObj := parser.pkg.TypesInfo.ObjectOf(serviceInit.Name)

	// we know the type, let's find the handlers
	handlerDecls := parser.extractHandlerMethods(handlerObj)

	var handlerDefs []model.HandlerDefinition
	for _, decl := range handlerDecls {
		def, err := parser.mapHandlerFunction(decl)
		if err != nil {
			fmt.Printf("while parsing handler def:\n%e\n", err)
			continue
		}

		handlerDefs = append(handlerDefs, def)
	}

	return model.ServiceDefinition{
		Pkg:      parser.pkg,
		Type:     handlerObj.Type(),
		Init:     initializerFunctionObj,
		Handlers: handlerDefs,
	}, nil
}

func (parser *ServiceParser) findServiceInitializer(handlerObj types.Object) *ast.FuncDecl {
	for _, decl := range parser.syntax.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Doc == nil {
				continue
			}

			role, found := model.ParseObjectRoleDocstring(decl.Doc.Text())
			if !found || role.Type != model.ObjectRoleServiceInit {
				continue
			}

			// make sure that the return type is the service handler tp
			err := parser.validateInitializerReturns(handlerObj, decl.Type.Results)
			if err != nil {
				fmt.Printf("warning: %e\n", err)
				continue
			}

			// if we got this far, we found our initializer
			return decl

		default:
			continue
		}
	}

	return nil
}

func (parser *ServiceParser) validateInitializerReturns(handlerObj types.Object, results *ast.FieldList) error {
	if results.NumFields() != 2 {
		return fmt.Errorf("expected 2 return values, but got %d", results.NumFields())
	}

	returnType := parser.pkg.TypesInfo.TypeOf(results.List[0].Type)
	switch returnType := returnType.(type) {
	case *types.Pointer:
		if returnType.Elem() != handlerObj.Type() {
			return fmt.Errorf("expected return value of type %s, but got %s", handlerObj.Type(), returnType.Elem())
		}

	case *types.Named:
		// if this is the same type as our initializer,
		return fmt.Errorf("service initializers require a pointer to be returned")

	default:
		return fmt.Errorf("expected pointer, but got %s", returnType.String())
	}

	// second type should be error
	errorResult := parser.pkg.TypesInfo.TypeOf(results.List[1].Type)
	switch errorResult := errorResult.(type) {
	case *types.Interface:
		if errorResult.String() != "error" {
			return fmt.Errorf("expected second return value of type error, but got %s", errorResult.String())
		}
	}

	return nil
}

func (parser *ServiceParser) extractHandlerMethods(handlerObj types.Object) []*ast.FuncDecl {
	handlerMethodSet := types.NewMethodSet(handlerObj.Type())

	var handlerDecls []*ast.FuncDecl

	handlerDecls = append(handlerDecls, parser.findHandlerMethods(handlerMethodSet)...)

	handlerPtrType := types.NewPointer(handlerObj.Type())
	ptrHandlerSet := types.NewMethodSet(handlerPtrType)

	handlerDecls = append(handlerDecls, parser.findHandlerMethods(ptrHandlerSet)...)

	// get handler specs for everything
	return filterMethods(handlerDecls)
}

func (parser *ServiceParser) findHandlerMethods(methodSet *types.MethodSet) []*ast.FuncDecl {

	var decls []*ast.FuncDecl
	for i := range methodSet.Len() {
		method := methodSet.At(i)
		methodPos := method.Obj().Pos()

		var methodDecl *ast.FuncDecl

		for _, decl := range parser.syntax.Decls {

			switch decl := decl.(type) {
			case *ast.FuncDecl:
				declPos := decl.Name.Pos()
				if methodPos == declPos {
					methodDecl = decl
					break
				}

			default:
				continue
			}
		}

		if methodDecl == nil {
			log.Fatalf("failed to find declaration info for %s\n", method.Obj().Name())
		}

		decls = append(decls, methodDecl)
	}

	return decls
}

func (parser *ServiceParser) mapHandlerFunction(handlerFunc *ast.FuncDecl) (model.HandlerDefinition, error) {
	// verify that the first arg is the context type
	role, valid := model.ParseObjectRoleDocstring(handlerFunc.Doc.Text())
	if !valid {
		return model.HandlerDefinition{}, fmt.Errorf("invalid role for %s", handlerFunc.Name.String())
	}

	// parse the arg for handler stuff
	httpMethod, endpoint, err := ParseHttpInfo(role.Args)
	if err != nil {
		return model.HandlerDefinition{}, fmt.Errorf("error while parsing endpoitn info: %w", err)
	}

	handlerConfig, err := parser.extractHandlerConfig(handlerFunc)
	if err != nil {
		return model.HandlerDefinition{}, fmt.Errorf("error while finding handler config: %w", err)
	}

	return model.HandlerDefinition{
		Method:            httpMethod,
		Path:              endpoint,
		Config:            handlerConfig,
		HandlerMethodName: handlerFunc.Name.String(),
	}, nil
}

func (parser *ServiceParser) extractHandlerConfig(handlerFunc *ast.FuncDecl) (model.HandlerConfig, error) {
	// second arg must be the config
	// TODO this needs to be validated first...
	configStructArg := handlerFunc.Type.Params.List[1]

	argType := parser.pkg.TypesInfo.TypeOf(configStructArg.Type)

	var handlerConfig model.HandlerConfig
	switch argType := argType.(type) {
	case *types.Named:
		structTp, ok := argType.Underlying().(*types.Struct)
		if !ok {
			return model.HandlerConfig{}, fmt.Errorf("expected struct, got %s", argType.String())
		}

		handlerConfig.Type = argType

		for i := range structTp.NumFields() {
			field := structTp.Field(i)
			tag := structTp.Tag(i)

			role, valid := model.ParseObjectRoleTag(tag)
			if !valid {
				continue
			}

			// role
			if !(role.Type == model.ObjectRolePathVar || role.Type == model.ObjectRoleQueryParam || role.Type == model.ObjectRoleBody) {
				return model.HandlerConfig{}, fmt.Errorf("invalid role '%s' for %s", role.Type, field.String())
			}

			// name
			tagName := getVariableName(role.Args)
			if len(tagName) == 0 || role.Type != model.ObjectRoleBody {
				tagName = strcase.ToLowerCamel(field.Name())
			}

			def := model.VariableDefinition{
				Name:      tagName,
				Type:      field.Type(),
				FieldName: field.Name(),
			}

			switch role.Type {
			case model.ObjectRolePathVar:
				handlerConfig.Path = append(handlerConfig.Path, def)
			case model.ObjectRoleQueryParam:
				handlerConfig.Query = append(handlerConfig.Query, def)
			case model.ObjectRoleBody:
				handlerConfig.Body = def
			}
		}

	default:
		return model.HandlerConfig{}, fmt.Errorf("expected named type to be named, but got %s", argType.String())
	}

	return handlerConfig, nil
}

func getVariableName(tagArgs string) string {
	args := strings.Split(tagArgs, ",")
	if len(args) < 1 {
		return ""
	}

	return args[0]
}

func filterMethods(methods []*ast.FuncDecl) []*ast.FuncDecl {
	var filtered []*ast.FuncDecl
	for _, method := range methods {
		if method.Doc == nil {
			continue
		}

		role, found := model.ParseObjectRoleDocstring(method.Doc.Text())
		if !found || role.Type != model.ObjectRoleHandlerTp {
			continue
		}

		filtered = append(filtered, method)
	}

	return filtered
}
