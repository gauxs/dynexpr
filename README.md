[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gauxs/dynexpr/go.yml?branch=master)](https://github.com/gauxs/dynexpr/actions/workflows/go.yml)
![Code Coverage](https://raw.githubusercontent.com/gauxs/dynexpr/badges/.badges/master/coverage.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gauxs/dynexpr)](https://goreportcard.com/report/github.com/gauxs/dynexpr)
![Go Version](https://img.shields.io/badge/go%20version=1.22-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/gauxs/dynexpr)](https://pkg.go.dev/mod/github.com/gauxs/dynexpr)

# Dynexpr

Expression builder for dynamo DB.

## Install

```shell
go get github.com/gauxs/dynexpr && go install github.com/gauxs/dynexpr/...@latest
```

**Note:** Dynexpr uses [Go Modules](https://go.dev/wiki/Modules) to manage dependencies.

## What is Dynexpr?

Dynexpr simplifies the creation of DynamoDB expressions by performing code generation on Go structs representing DynamoDB items. It offers convenient methods to generate expressions for DynamoDB, streamlining the process of building complex queries.

## Usage

```
// dynexpr:generate
type Person struct {
	PK            *string       `json:"pk,omitempty" dynexpr:"partitionKey"`
	SK            *string       `json:"sk,omitempty"  dynexpr:"sortKey"`
	Name          *string       `json:"name,omitempty"`
}

```

## Configurations

1. `dynexpr:generate`: should be declared over the struct which represents a single item of dynamoDB.

```
// dynexpr:generate
type DDBItem struct {
    ...
}
```

2. `dynexpr:"partitionKey"`: to declare that the attribute is partion key of the dynamoDB item.

```
type DDBItem struct {
    PK          *string       `json:"pk,omitempty" dynexpr:"partitionKey"`
    ...
}
```

3. `dynexpr:"sortKey"`: to declare that the attribute is sort key of the dynamoDB item.

```
type DDBItem struct {
    SK          *string       `json:"sk,omitempty" dynexpr:"sortKey"`
    ...
}
```

## Q & A

### Is this fit for your usecase?

## License

The project is licensed under the [MIT License](LICENSE).

## Test

`go run cmd/main.go -output_filename test/expression/data/person_dynexpr.go  test/expression/data`
