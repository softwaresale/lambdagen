package model

// LambdaMetadata describes the metadata used by CDK to determine how to specify this lambda
type LambdaMetadata struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
