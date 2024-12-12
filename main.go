package main

import (
	"fmt"
	"lambdagen/internal/parsing"
	"log"
)

func main() {
	packageRoot := "/home/charlie/Programming/mac-schedule-api"

	services, err := parsing.ParseServices(packageRoot, "mac-schedule-api/api")
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range services {
		fmt.Println(service)
	}
}
