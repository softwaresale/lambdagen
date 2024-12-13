package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/softwaresale/lambdagen/internal/codegen"
	"github.com/softwaresale/lambdagen/internal/model"
	"github.com/softwaresale/lambdagen/internal/parsing"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Args struct {
	RootModuleDir string
	Modules       []string
	OutputModName string
}

var args Args

func init() {
	flag.StringVar(&args.RootModuleDir, "project", "", "root directory of project to generate lambdas for")
	flag.StringVar(&args.OutputModName, "output", "lambda", "directory to store lambdas in")
}

func main() {

	flag.Parse()
	args.Modules = flag.Args()

	var err error

	if len(args.RootModuleDir) == 0 {
		args.RootModuleDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("while falling back to get current directory: %s", err)
		}
	}

	if len(args.Modules) == 0 {
		log.Fatal("no handler modules provided")
	}

	for _, module := range args.Modules {
		err = createHandlersForModule(module)
		if err != nil {
			log.Println(err)
		}
	}
}

func createHandlersForModule(mod string) error {
	services, err := parsing.ParseServices(args.RootModuleDir, mod)
	if err != nil {
		return fmt.Errorf("while parsing module %s:\n%w", mod, err)
	}

	// lambda output directory
	outputDirectory := filepath.Join(args.RootModuleDir, args.OutputModName)
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
				return fmt.Errorf("error while making output dir: %w", err)
			}

			outputFile, err := os.Create(filepath.Join(lambdaOutputDir, "main.go"))
			if err != nil {
				return fmt.Errorf("error while making main file: %w", err)
			}

			err = codegen.TranslateHandler(outputFile, service, handler)
			if err != nil {
				return fmt.Errorf("error while translating handler: %w", err)
			}

			err = outputMetadata(handler, lambdaOutputDir)
			if err != nil {
				return fmt.Errorf("error while writing metadata: %w", err)
			}
		}
	}

	return nil
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
