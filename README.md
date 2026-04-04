# Shaker

Shaker is a lightweight wrapper around **Gin-Gonic** that lets you define handlers as plain Go functions with optional typed input and output.

### Why use Shaker

- **Function-style handlers**: define handlers as `func(ctx *shaker.Context, input *In) (Out, error)` or `func(ctx *shaker.Context) error`
- **Optional typed input/output**: input and output structs are optional, and request data is bound automatically from URI, query, header, or body tags
- **Standardized error handling**: validation errors return `400 Bad Request`, mapped custom errors return configured status codes, and unknown errors return `500 Internal Server Error`
- **Common HTTP methods**: supports `GET`, `POST`, `PUT`, and `DELETE`

## Getting Started

### Prerequisites

- **Go version**: Shaker requires [Go](https://go.dev/) version `1.25` or later
- **Basic Go knowledge**: Familiarity with Go syntax and modules is helpful

### Installation

With Go modules, import the package in your code and Go will fetch it automatically:

```go
import "github.com/kessaro/shaker"
```

### Your First Shaker Application

```go
package main

import (
  "log"
  "net/http"

  "github.com/kessaro/shaker"
)

type In struct {
    Var string `uri:"var"`
    Opt string `form:"option"`
}

type Out struct {
    Var    string `json:"var"`
    Option string `json:"option"`
}

func copyStringHandler(ctx *shaker.Context, input *In) (Out, error) {
    return Out{
        Var:    input.Var,
        Option: input.Opt,
    }, nil
}

func main() {
    shaker := shaker.NewShaker(nil)

    if err := shaker.Get("/copy/:var", copyStringHandler, http.StatusOK); err != nil {
        log.Fatalf("failed to register handler: %v", err)
    }

    if err := shaker.Shake(); err != nil {
        log.Fatalf("failed to run server: %v", err)
    }
}
```

## Contributing

Contributions are welcome! Open issues or pull requests for:

- bug reports
- feature requests
- documentation improvements
- test coverage

If you plan to contribute, feel free to open a discussion so maintainers can help you get started.
