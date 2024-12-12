package model

import "go/ast"

type ServiceType struct {
	TypeSpec *ast.TypeSpec
	Config   map[string]string
}

func (s *ServiceType) GetName() string {
	return s.TypeSpec.Name.String()
}

type HandlerDefinition struct {
	Func   *ast.FuncDecl
	Method string
	Path   string
	Config RequestConfig
}

type ServiceDefinition struct {
	PackageName string
	Type        ServiceType
	Init        *ast.FuncDecl
	Handlers    []HandlerDefinition
}

type VariableDefinition struct {
	Name      string   // The name of the path variable as it appears in the request
	Type      TypeExpr // the type of this variable. Type is guaranteed to be either a selector expr or an Ident
	FieldName string   // the field in the request config object that should take this path variable
}

type RequestConfig struct {
	Type           *ast.Ident // the type of the struct that defines our request
	PathVariables  []VariableDefinition
	QueryVariables []VariableDefinition
	RequestBody    VariableDefinition
}
