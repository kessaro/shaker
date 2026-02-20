# Shaker

Shaker is a wrapper that allows you to define handlers that take an input structure as parameter (optionnal) and returns a pair of output struct (optionnal too) and an error

### **Why choose Shaker**

Shaker uses the powerful **Gin-Gonic** HTTP server with some cool features :

- **Function like handler** : declare your handlers like a function by defining your input and output structures
- **Standardize your handlers** : all handlers have the same behavior about nominal and error management
- **Invisible binding** : Bind several input sources to your input structure by using differents binding tags (json, query, path, ...)
- **All methods supported** : All methods such as GET, PUT, POST, DELETE are natively supported by this wrapper

## Getting Started

### Prerequisites

- **Go version**: Shaker requires [Go](https://go.dev/) version [1.24](https://go.dev/doc/devel/release#go1.24.0) or above
- **Basic Go knowledge**: Familiarity with Go syntax and package management is helpful

### Installation

With [Go's module support](https://go.dev/wiki/Modules#how-to-use-modules), simply import Gin in your code and Go will automatically fetch it during build:

```go
import "github.com/kessaro/shaker"
```

### Your First Gin Application

Here's a complete example that demonstrates Gin's simplicity:

```go
package main

import (
  "log"
  "net/http"

  "github.com/kessaro/shaker"
)

// Define your input and output structures

type in struct {
    Var string `uri:"var"`
    Opt string `form:"option"`
}

type out struct {
    Var    string `json:"var"`
    Option string `json:"option"`
}

func copyStringHandler(ctx *shaker.Context, input *in) (out, error) {
    return out{
        Var: input.Var,
        Option: input.Opt
    }, nil
}

func main() {
  // Create a Shaker wrapper
  shaker := shaker.NewShaker()

  // Register your handlers
  shaker.Get("/copy", copyStringHandler, http.StatusOK)


  // Start server on port 8080 (default)
  // Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
  if err := shaker.Shake(); err != nil {
    log.Fatalf("failed to run server: %v", err)
  }
}
```

## 🤝 Contributing

Gin is the work of hundreds of contributors from around the world. We welcome and appreciate your contributions!

### How to Contribute

- 🐛 **Report bugs** - Help us identify and fix issues
- 💡 **Suggest features** - Share your ideas for improvements
- 📝 **Improve documentation** - Help make our docs clearer
- 🔧 **Submit code** - Fix bugs or implement new features
- 🧪 **Write tests** - Improve our test coverage

### Getting Started with Contributing

1. Check out our [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines
2. Join our community discussions and ask questions

**All contributions are valued and help make Gin better for everyone!**
