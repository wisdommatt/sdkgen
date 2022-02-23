# Sdkgen

A CLI tool for generating SDK clients for Rest &amp; Graphql APIs using swagger and graphql schema files.


## Installation

```bash
go install github.com/wisdommatt/sdkgen
```


## Usage

**To generate SDK client from GraphQl schema file:**

```bash
sdkgen graphql --schema sample.graphql --output pkg/sample
```

`--schema` and `--output` parameters are required.


**To generate SDK client from OpenAPI | Swagger schema file**:

```bash
sdkgen openapi --schema sample-api.json --output pkg/sample
```

You can generate the client from a **json** or **yaml** schema file, `--schema`and`--output` parameters are required.


## Documentation

[https://pkg.go.dev/github.com/wisdommatt/sdkgen](https://pkg.go.dev/github.com/wisdommatt/sdkgen)
