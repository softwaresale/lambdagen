package parsing

import (
	"fmt"
	"go/ast"
	"lambdagen/internal/model"
	"regexp"
)

func (builder *LambdaModelBuilder) findServiceHandlers(file *ast.File, serviceTp *model.ServiceType) ([]model.HandlerDefinition, error) {
	// find all top level function declarations with serviceTp as the receiver

	var handlers []model.HandlerDefinition

	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			// verify that this is labelled as a handler
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

			if role.Type != model.ObjectRoleHandlerTp {
				continue
			}

			// this function is marked as a handler. Check it
			err := builder.checkValidHandlerReceiver(decl.Recv, serviceTp)
			if err != nil {
				return nil, err
			}

			// check the parameters and find the arguments type
			configParamIdent, err := builder.checkHandlerParams(decl)
			if err != nil {
				return nil, fmt.Errorf("error while checking function params: %w", err)
			}

			// pull out the request config object, which involves finding the appropriate declaration
			requestConfig, err := builder.extractHandlerConfig(file, configParamIdent)

			// parse the information from the role args to determine the method and path
			method, path, err := parseHandlerArgString(role.Args)
			if err != nil {
				return nil, fmt.Errorf("error while parsing request args: %w", err)
			}

			methodDefinition := model.HandlerDefinition{
				Func:   decl,
				Method: method,
				Path:   path,
				Config: requestConfig,
			}

			// we found a valid handler
			handlers = append(handlers, methodDefinition)
		default:
			continue
		}
	}

	return handlers, nil
}

func (builder *LambdaModelBuilder) checkServiceInitResults(returns *ast.FieldList, serviceTp *model.ServiceType) error {
	// we found a service init role. Now we need to do some type checking
	if returns.NumFields() != 2 {
		return fmt.Errorf("%s: service initializer must return 2 values: (<service>, error)", builder.fset.Position(returns.Pos()))
	}

	// verify that the first arg is a pointer to our serviceTp
	serviceRetTypeExpr := returns.List[0].Type

	var returnTypeExpr ast.Expr
	switch ret := serviceRetTypeExpr.(type) {
	case *ast.StarExpr:
		// assert that the type is the same as our serviceTp
		returnTypeExpr = ret.X

	case *ast.Ident:
		return fmt.Errorf("%s: must return a pointer, not an identifier", builder.fset.Position(ret.Pos()))

	default:
		panic(fmt.Sprintf("unexpected service ret type: %v", ret))
	}

	// is the returnTypeExpr an identifier to our type?
	if !checkReturnTypeEqual(returnTypeExpr, serviceTp.GetName()) {
		return fmt.Errorf("%s: a service initializer must return the same type as service", builder.fset.Position(returnTypeExpr.Pos()))
	}

	// now, check that the second is error
	errorTpExpr := returns.List[1].Type
	if !checkReturnTypeEqual(errorTpExpr, "error") {
		return fmt.Errorf("%s: a service initializer second arg must be error", builder.fset.Position(errorTpExpr.Pos()))
	}

	return nil
}

func checkReturnTypeEqual(identExpr ast.Expr, name string) bool {
	switch returnType := identExpr.(type) {
	case *ast.Ident:
		return returnType.Name == name

	default:
		return false
	}
}

func (builder *LambdaModelBuilder) checkValidHandlerReceiver(recvList *ast.FieldList, serviceTp *model.ServiceType) error {
	if recvList.NumFields() == 0 || recvList.NumFields() > 1 {
		return fmt.Errorf("%s: handler must take exactly the service as a receiver", builder.fset.Position(serviceTp.TypeSpec.Pos()))
	}

	recv := recvList.List[0]

	var identExpr ast.Expr
	switch ret := recv.Type.(type) {
	case *ast.StarExpr:
		identExpr = ret.X

	case *ast.Ident:
		identExpr = ret

	default:
		panic(fmt.Sprintf("unexpected service ret type: %v", ret))
	}

	// is the returnTypeExpr an identifier to our type?
	if !checkReturnTypeEqual(identExpr, serviceTp.GetName()) {
		return fmt.Errorf("%s: the receiver for a handler must be a valid service", builder.fset.Position(identExpr.Pos()))
	}

	return nil
}

func (builder *LambdaModelBuilder) checkHandlerParams(decl *ast.FuncDecl) (*ast.Ident, error) {
	// assert that the handler has two arguments
	if decl.Type.Params.NumFields() > 2 || decl.Type.Results.NumFields() == 0 {
		return nil, fmt.Errorf(
			"at function %s:\nparams %s: handler function should take 1-2 arguments. "+
				"The first should be a context, and the second is an optional struct with request parameters",
			builder.fset.Position(decl.Pos()),
			builder.fset.Position(decl.Type.Params.Pos()),
		)
	}

	// assert that the first argument is of type context.Context
	err := builder.verifyContextParam(decl.Type.Params.List[0])
	if err != nil {
		return nil, err
	}

	// next, we should figure out if there is a second argument. if there is not, then bail
	if decl.Type.Params.NumFields() < 2 {
		return nil, nil
	}

	// pull the next argument, which should be an identifier
	// TODO: basically, this paradigm forces all request configs to be in the same file as the handler
	configParam := decl.Type.Params.List[1]
	switch typeExpr := configParam.Type.(type) {
	case *ast.Ident:
		return typeExpr, nil

	default:
		return nil, fmt.Errorf("%s: config arg must be an identifier. Selector's coming soon!", builder.fset.Position(configParam.Pos()))
	}
}

func (builder *LambdaModelBuilder) determineHandlerType(decl *ast.FuncDecl) error {
	if decl.Recv.NumFields() == 0 {
		return nil
	}

	receiverField := decl.Recv.List[0]

	var _ *ast.Ident

	switch decl := receiverField.Type.(type) {
	case *ast.StarExpr:
		_ = decl.X.(*ast.Ident)

	case *ast.Ident:
		_ = decl

	default:
		return fmt.Errorf("%s: invalid receiver", builder.fset.Position(receiverField.Pos()))
	}

	return nil
}

func (builder *LambdaModelBuilder) verifyContextParam(field *ast.Field) error {

	ctxArgTp, ok := field.Type.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("%s: first argument should be builtin `context.Context` type", builder.fset.Position(field.Pos()))
	}

	pkg, ok := ctxArgTp.X.(*ast.Ident)
	if !ok {
		return fmt.Errorf("%s: first argument type selector is not an identifier", builder.fset.Position(field.Pos()))
	}

	if pkg.Name != "context" || ctxArgTp.Sel.Name != "Context" {
		return fmt.Errorf("%s: first argument is not `context.Context`, instead its `%s.%s`", builder.fset.Position(field.Pos()), pkg.Name, ctxArgTp.Sel.Name)
	}

	return nil
}

func parseHandlerArgString(args string) (string, string, error) {
	parser := regexp.MustCompile(`(GET|POST|PUT|DELETE)\s+(\S*)`)
	match := parser.FindStringSubmatch(args)
	if match == nil {
		return "", "", fmt.Errorf("invalid request definition")
	}

	return match[1], match[2], nil
}
