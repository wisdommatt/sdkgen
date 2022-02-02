// Code generated by sdkgen; DO NOT EDIT.

package example

import (
    "time"
    "context"
)

{{ $schema := . }}

{{/* Generating Go types for graphql Unions */}}
{{ range $union := .Unions }}

{{ $unionName := toCamelCase $union.Name }}

type {{ $unionName }} interface {
    Is{{ $unionName }}()
}

{{ end }}

{{/* Generating Go types for graphql Enums */}}
{{ range $enum := .Enums }}
{{ if isExported $enum.Name }}
{{ $enumName := toCamelCase $enum.Name }}

type {{ $enumName }} string

var (
    {{ range $val := $enum.EnumValues }} {{ $enumName }}{{ toCamelCase $val.Name }} {{ $enumName }} = "{{ toCamelCase $val.Name }}"
    {{ end }}
)

func (e {{ $enumName }}) IsValid() bool {
    switch e {
    case {{ range $key, $val := $enum.EnumValues }} {{ $enumName }}{{ toCamelCase $val.Name }} {{ if not (isLastEnumField $enum.EnumValues $key) }}, {{ end }} {{ end }}:
        return true
    }
    return false
}

func (e {{ $enumName }}) String() string {
    return string(e)
}
{{ end }}
{{ end }}

{{/* Generating Go types for graphql Types (inputs, objects) */}}
{{ range $val := .Objects }}
{{ if isExported $val.Name }}
type {{ $val.Name }} struct {
    {{ range $field := $val.Fields }} {{ if and (isExported $field.Name) (isExported $field.Type.Name) }} {{ toCamelCase $field.Name }} {{ extractFieldTypeName $schema $field.Name $field.Type }} `json:"{{ $field.Name }}"` {{ end }}
    {{ end }}
}
{{ end }}
{{ end }}

type GraphqlClient interface {
    Mutations
    Queries
    Subscriptions
}

{{/* Generating mutations interface methods */}}
type Mutations interface {
    {{ range $mutation := $schema.Mutations }} {{ toCamelCase $mutation.Name }}(ctx context.Context, {{ range $arg := $mutation.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }}) ({{ extractFieldTypeName $schema $mutation.Name $mutation.Type }}, error)
    {{ end }}
}

type Queries interface {

}

type Subscriptions interface {

}

// GqlClient is the default implementation for 
// GraphqlClient interface.
type GqlClient struct {

}

// NewGraphqlClient returns a new graphql client.
func NewGraphqlClient() *GqlClient {
    return &GqlClient{}
}

