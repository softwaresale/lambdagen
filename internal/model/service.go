package model

import (
	"go/types"
	"golang.org/x/tools/go/packages"
)

// ServiceDefinition describes an API with a collection of endpoint handlers
type ServiceDefinition struct {
	Pkg      *packages.Package   // Pkg is the package that contains this service def. Used for translation stuff
	Type     types.Type          // Type is the type of the struct that is designated the struct handler
	Init     types.Object        // Init is the function responsible for initializing this service
	Handlers []HandlerDefinition // Handlers is the collection of handler methods
	Config   map[string]string   // Config is service-level configuration variables provided in the header line
}

type HandlerDefinition struct {
	Method            string
	Path              string
	Config            HandlerConfig
	HandlerMethodName string
}

type HandlerConfig struct {
	Type  *types.Named
	Query []VariableDefinition
	Path  []VariableDefinition
	Body  VariableDefinition
}

type VariableDefinition struct {
	Name      string
	Type      types.Type
	FieldName string
}
