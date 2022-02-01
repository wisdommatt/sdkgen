
package example

{{ range $val := .Types }}
type {{ $val.Name }} struct {
    {{ range $field := $val.Fields }} {{ $field.Name }} {{ extractFieldTypeName $field }}
    {{ end }}
}
{{ end }}