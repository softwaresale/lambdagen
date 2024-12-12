package main

import (
	"fmt"
	"lambdagen/internal/codegen"
	"lambdagen/internal/model"
	"lambdagen/internal/parsing"
	"os"
)

func main() {
	services, err := parsing.ParseHandlerFile("/home/charlie/Programming/mac-schedule-api/api/employee-api.go")
	if err != nil {
		panic(err)
	}

	var lambdaModels []model.LambdaModel
	for _, service := range services {
		fmt.Printf("found service %s with %d handlers", service.Type.TypeSpec.Name.Name, len(service.Handlers))

		// make lambda models for our service
		lambdas, err := model.TranslateServiceDefinition(&service)
		if err != nil {
			panic(err)
		}

		lambdaModels = append(lambdaModels, lambdas...)
	}

	// actually generate code for each lambda
	for _, lambdaModel := range lambdaModels {
		err = codegen.Transform(os.Stdout, lambdaModel)
		if err != nil {
			panic(err)
		}
	}
}
