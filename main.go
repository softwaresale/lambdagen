package main

import (
	"flag"
	"fmt"
	"github.com/softwaresale/lambdagen/internal/output"
	"github.com/softwaresale/lambdagen/internal/parsing"
	"log"
	"os"
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
	outputManager := output.NewManager(args.RootModuleDir, args.OutputModName)

	err = outputManager.CreateOutputDir()
	if err != nil {
		return fmt.Errorf("while creating base output directory: %w", err)
	}

	for _, service := range services {
		for _, handler := range service.Handlers {

			err = outputManager.Register(&service, &handler)
			if err != nil {
				return fmt.Errorf("while registering handler for module %s:\n%w", mod, err)
			}

		}
	}

	err = outputManager.Render()
	if err != nil {
		return fmt.Errorf("while rendering lambdas for module %s:\n%w", mod, err)
	}

	return nil
}
