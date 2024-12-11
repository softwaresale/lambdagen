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
}

type ServiceDefinition struct {
	Type     ServiceType
	Init     *ast.FuncDecl
	Handlers []HandlerDefinition
}
