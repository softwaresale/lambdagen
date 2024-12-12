package model

type LambdaModel struct {
	HandlerType *QualifiedName
	InitFunc    *QualifiedName
	HandlerFunc string
	Config      RequestConfig
}
