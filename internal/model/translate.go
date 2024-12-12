package model

// TranslateServiceDefinition 1:1 translates a method handler a lambda model, which is in turn translated into a single
// compilation unit
func TranslateServiceDefinition(serviceDef *ServiceDefinition) ([]LambdaModel, error) {

	// get the qualified name for the service handler type
	handlerTypeName := serviceDef.Type.TypeSpec.Name.Name
	serviceHandlerTypeName := &QualifiedName{
		Package: serviceDef.PackageName,
		Name:    handlerTypeName,
	}

	// get the qualified name for the initializer
	initializerFunc := &QualifiedName{
		Name:    serviceDef.Init.Name.Name,
		Package: serviceDef.PackageName,
	}

	var models []LambdaModel
	for _, handler := range serviceDef.Handlers {

		model := LambdaModel{
			HandlerType: serviceHandlerTypeName,
			InitFunc:    initializerFunc,
			HandlerFunc: handler.Func.Name.Name,
			Config:      handler.Config,
		}

		models = append(models, model)
	}

	return models, nil
}
