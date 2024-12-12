package model

import (
	"go/types"
)

// ServiceDefinition describes an API with a collection of endpoint handlers
type ServiceDefinition struct {
	Type     types.Type   // Type is the type of the struct that is designated the struct handler
	Init     types.Object // Init is the function responsible for initializing this service
	Handlers []HandlerDefinition
}

type HandlerDefinition struct {
	Method            string
	Path              string
	Config            HandlerConfig
	HandlerMethodName string
}

type HandlerConfig struct {
	Query []VariableDefinition
	Path  []VariableDefinition
	Body  VariableDefinition
}

type VariableDefinition struct {
	Name      string
	Type      types.Type
	FieldName string
}
