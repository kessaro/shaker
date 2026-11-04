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

    api := shaker.Group("/api")
    if err := api.Get("/copy/:var", copyStringHandler, http.StatusOK); err != nil {
        log.Fatalf("failed to register handler: %v", err)
    }

    if err := shaker.Shake(); err != nil {
        log.Fatalf("failed to run server: %v", err)
    }
}
```

### Grouping routes

Use `Group` to register related endpoints under a shared prefix. Groups can be nested, and support the same HTTP methods and handler signatures as `Shaker`:

```go
api := shaker.Group("/api")
v1 := api.Group("/v1")

if err := v1.Get("/users", listUsersHandler, http.StatusOK); err != nil {
    log.Fatalf("failed to register handler: %v", err)
}
if err := v1.Post("/users", createUserHandler, http.StatusCreated); err != nil {
    log.Fatalf("failed to register handler: %v", err)
}
```

These routes are registered at `/api/v1/users`.

### Middleware

Add middleware to a group with `Use` to apply it to all routes registered in that group. You can also pass middleware to an individual route method after its default status code. Group middleware runs before route-specific middleware, and both run before the endpoint handler.

```go
api := shaker.Group("/api")
api.Use(func(ctx *shaker.Context) {
    ctx.Header("X-API", "v1")
    ctx.Next()
})

if err := api.Get("/private", privateHandler, http.StatusOK, authMiddleware); err != nil {
    log.Fatalf("failed to register handler: %v", err)
}
```

Individual-route middleware can also be passed to `Shaker` methods such as `shaker.Get`, not only to group methods.

## Contributing

Contributions are welcome! Open issues or pull requests for:

- bug reports
- feature requests
- documentation improvements
- test coverage

If you plan to contribute, feel free to open a discussion so maintainers can help you get started.
