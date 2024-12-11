package main

import (
	"fmt"
	"lambdagen/internal/parsing"
)

func main() {
	services, err := parsing.ParseHandlerFile("api/employee-api.go")
	if err != nil {
		panic(err)
	}

	for _, service := range services {
		fmt.Printf("found service %s with %d handlers", service.Type.TypeSpec.Name.Name, len(service.Handlers))
		fmt.Printf("%v\n", service)
	}
}
