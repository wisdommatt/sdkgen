// Code generated by sdkgen; DO NOT EDIT.

package example

import (
    "time"
    "context"
    "fmt"
    "github.com/machinebox/graphql"
)

{{ $schema := . }}

{{/* Generating Go types for graphql Unions */}}
{{ range $union := .Unions }}

{{ $unionName := toCamelCase $union.Name }}

type {{ $unionName }} interface {
    Is{{ $unionName }}()
}

{{ range $type := $union.Types }}func (u {{ toCamelCase $type }}) Is{{ $unionName }}() {}
{{ end }}

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
    {{ range $mutation := $schema.Mutations }} {{ if isExported $mutation.Name }} {{ toCamelCase $mutation.Name }}(ctx context.Context, {{ range $arg := $mutation.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }}) ({{ extractFieldTypeName $schema $mutation.Name $mutation.Type }}, error)
    {{ end }}{{ end }}
}

{{/* Generating queries interface methods */}}
type Queries interface {
    {{ range $query := $schema.Queries }} {{ if isExported $query.Name }} {{ toCamelCase $query.Name }}(ctx context.Context, {{ range $arg := $query.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }}) ({{ extractFieldTypeName $schema $query.Name $query.Type }}, error)
    {{ end }}{{ end }}
}

{{/* Generating subscriptions interface methods */}}
type Subscriptions interface {
    {{ range $sub := $schema.Subscriptions }} {{ if isExported $sub.Name }} {{ toCamelCase $sub.Name }}(ctx context.Context, {{ range $arg := $sub.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }}) ({{ extractFieldTypeName $schema $sub.Name $sub.Type }}, error)
    {{ end }}{{ end }}
}

// ClientConfig is the config used for creating a new
// graphql client.
type ClientConfig struct {
	MutationURL        string
	QueryURL           string
	SubscriptionURL    string
	DefaultHTTPHeaders map[string]string
}

// GqlClient is the default implementation for 
// GraphqlClient interface.
type GqlClient struct {
    Mutation *Mutation
    Query *Query
    config ClientConfig
}

// NewClient returns a new graphql client.
func NewClient(config ClientConfig) *GqlClient {
    return &GqlClient{
        Mutation: &Mutation{
            graphClient: graphql.NewClient(config.MutationURL),
            defaultHTTPHeaders: config.DefaultHTTPHeaders,
        },
        Query: &Query{
            graphClient: graphql.NewClient(config.QueryURL),
            defaultHTTPHeaders: config.DefaultHTTPHeaders,
        },
    }
}

type Mutation struct {
	graphClient *graphql.Client
    defaultHTTPHeaders map[string]string
}

{{ range $mutation := $schema.Mutations }} 
{{ if isExported $mutation.Name }} 
    {{ $responseName := extractFieldTypeName $schema $mutation.Name $mutation.Type }}
    {{ $pointerResponse := toPointerTypeName $schema $responseName $mutation.Type }}
    func (m *Mutation) {{ toCamelCase $mutation.Name }}(ctx context.Context, {{ range $arg := $mutation.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }} gqlFields string) ({{ $pointerResponse }}, error) {
        req := graphql.NewRequest(fmt.Sprintf(`
            mutation({{ range $arg := $mutation.Arguments }}${{ $arg.Name }}: {{ $arg.Type }}, {{ end }}) {
                {{ $mutation.Name }}({{ range $arg := $mutation.Arguments }}{{ $arg.Name }}: ${{ $arg.Name }}, {{ end }}) %s
            }
        `, gqlFields))
        {{ range $arg := $mutation.Arguments }} req.Var("{{ $arg.Name }}", {{ $arg.Name }})
        {{ end }}

        for key, value := range m.defaultHTTPHeaders {
            req.Header.Set(key, value)
        }

        var {{ toLowerCamel $mutation.Name }}Response map[string]{{ $pointerResponse }}
        err := m.graphClient.Run(ctx, req, &{{ toLowerCamel $mutation.Name }}Response)
        if err != nil {
            return {{ nilValue $responseName $mutation.Type }}, err
        }
        return {{ toLowerCamel $mutation.Name }}Response["{{ $mutation.Name }}"], nil
    }
{{ end }}{{ end }}

type Query struct {
	graphClient *graphql.Client
    defaultHTTPHeaders map[string]string
}

{{ range $query := $schema.Queries }} 
{{ if isExported $query.Name }} 
    {{ $responseName := extractFieldTypeName $schema $query.Name $query.Type }}
    {{ $pointerResponse := toPointerTypeName $schema $responseName $query.Type }}
    func (q *Query) {{ toCamelCase $query.Name }}(ctx context.Context, {{ range $arg := $query.Arguments }} {{ $arg.Name }} {{ extractFieldTypeName $schema $arg.Name $arg.Type }}, {{ end }} gqlFields string) ({{ $pointerResponse }}, error) {
        req := graphql.NewRequest(fmt.Sprintf(`
            query({{ range $arg := $query.Arguments }}${{ $arg.Name }}: {{ $arg.Type }}, {{ end }}) {
                {{ $query.Name }}({{ range $arg := $query.Arguments }}{{ $arg.Name }}: ${{ $arg.Name }}, {{ end }}) %s
            }
        `, gqlFields))
        {{ range $arg := $query.Arguments }} req.Var("{{ $arg.Name }}", {{ $arg.Name }})
        {{ end }}

        for key, value := range q.defaultHTTPHeaders {
            req.Header.Set(key, value)
        }

        var {{ toLowerCamel $query.Name }}Response map[string]{{ $pointerResponse }}
        err := q.graphClient.Run(ctx, req, &{{ toLowerCamel $query.Name }}Response)
        if err != nil {
            return {{ nilValue $responseName $query.Type }}, err
        }
        return {{ toLowerCamel $query.Name }}Response["{{ $query.Name }}"], nil
    }
{{ end }}{{ end }}