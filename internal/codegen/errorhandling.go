package codegen

import "github.com/dave/jennifer/jen"

func CheckError(group *jen.Group, GenError func(group *jen.Group)) {
	group.If(jen.Err().Op("!=").Nil()).BlockFunc(GenError)
}

func GenerateAPIError(group *jen.Group, status int, message string) {

	// create an API
	apiErrVar := "apiErr"
	group.Id(apiErrVar).Op(":=").Qual("github.com/softwaresale/lambdagen/pkg", "APIError").Values(jen.Dict{
		jen.Id("Message"): jen.Lit(message),
		jen.Id("Error"):   jen.Err(),
	})

	responseBodyVar := "errorResponseBody"

	group.List(group.Id(responseBodyVar), jen.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(jen.Id(apiErrVar))
	group.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.List(
			jen.Qual("github.com/aws/aws-lambda-go/events", "APIGatewayProxyResponse").Values(jen.Dict{}),
			jen.Err(),
		)),
	)

	group.Return(jen.List(
		jen.Qual("github.com/aws/aws-lambda-go/events", "APIGatewayProxyResponse").Values(jen.Dict{
			jen.Id("StatusCode"): jen.Lit(status),
			jen.Id("Body"):       jen.String().Parens(jen.Id(responseBodyVar)),
		}),
		jen.Nil(),
	))
}
