package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"lambdagen/internal/codegen"
	"lambdagen/internal/model"
	"lambdagen/internal/parsing"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	packageRoot := "/home/charlie/Programming/mac-schedule-api"

	services, err := parsing.ParseServices(packageRoot, "mac-schedule-api/api")
	if err != nil {
		log.Fatal(err)
	}

	// lambda output directory
	outputDirectory := filepath.Join(packageRoot, "lambda")
	for _, service := range services {

		var handlerName string
		switch handlerType := service.Type.(type) {
		case *types.Named:
			handlerName = handlerType.Obj().Name()

		default:
			panic("handler was not a named type")
		}

		for _, handler := range service.Handlers {

			lambdaDirectoryName := strings.Join([]string{handlerName, handler.HandlerMethodName}, "_")

			lambdaOutputDir := filepath.Join(outputDirectory, lambdaDirectoryName)
			err = os.MkdirAll(lambdaOutputDir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}

			outputFile, err := os.Create(filepath.Join(lambdaOutputDir, "main.go"))
			if err != nil {
				log.Fatal(err)
			}

			err = codegen.TranslateHandler(outputFile, service, handler)
			if err != nil {
				log.Fatal(err)
			}

			err = outputMetadata(handler, lambdaOutputDir)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func outputMetadata(handler model.HandlerDefinition, lambdaDir string) error {
	outputPath := filepath.Join(lambdaDir, "spec.json")
	metadata := model.LambdaMetadata{
		Method: handler.Method,
		Path:   handler.Path,
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error while marshalling data: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error while creating output file: %w", err)
	}

	defer func() {
		err := outputFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	_, err = outputFile.Write(bytes)
	if err != nil {
		return fmt.Errorf("error while writing output file: %w", err)
	}

	_, err = outputFile.Write([]byte{'\n'})
	if err != nil {
		return fmt.Errorf("error while writing output file: %w", err)
	}

	return nil
}
