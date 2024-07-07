# httpserver

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/dmitrymomot/httpserver)](https://github.com/dmitrymomot/httpserver)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmitrymomot/httpserver.svg)](https://pkg.go.dev/github.com/dmitrymomot/httpserver)
[![License](https://img.shields.io/github/license/dmitrymomot/httpserver)](https://github.com/dmitrymomot/httpserver/blob/main/LICENSE)

[![Tests](https://github.com/dmitrymomot/httpserver/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/tests.yml)
[![CodeQL Analysis](https://github.com/dmitrymomot/httpserver/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/codeql-analysis.yml)
[![GolangCI Lint](https://github.com/dmitrymomot/httpserver/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/golangci-lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmitrymomot/httpserver)](https://goreportcard.com/report/github.com/dmitrymomot/httpserver)

The `httpserver` package in Go offers a simple and efficient solution for creating, running, and gracefully shutting down HTTP servers. It supports context cancellation and concurrent execution, making it suitable for a wide range of web applications.

## Features

- Easy to set up and start HTTP servers
- Graceful shutdown handling
- Context cancellation support
- Concurrency management with `errgroup`
- Lightweight and flexible design

## Installation

To install the `httpserver` package, use the following command:

```bash
go get github.com/dmitrymomot/httpserver
```

## Usage

Here's a basic example of how to use the `httpserver` package:

```go
package main

import (
    "context"
    "net/http"
    "github.com/dmitrymomot/httpserver"
    "time"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    r := http.NewServeMux()
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    if err := httpserver.Run(ctx, ":8080", r); err != nil {
        panic(err)
    }
}
```

This will start an HTTP server on port 8080 and respond with "Hello, World!" to `GET /` request. The server will be gracefully shut down after 10 minutes.

The `httpserver.Run` function is a shortcut for creating a new HTTP server and starting it. It's equivalent to the following code:

```go
srv := httpserver.NewServer(addr, handler)
if err := srv.Start(ctx); err != nil {
    panic(err)
}
```

### Using with `errgroup`:

```go
package main

import (
    "context"
    "net/http"
    "github.com/dmitrymomot/httpserver"
    "golang.org/x/sync/errgroup"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    r := http.NewServeMux()
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error {
        return httpserver.Run(ctx, ":8080", r)
    })
    g.Go(func() error {
        return httpserver.Run(ctx, ":8081", r)
    })

    if err := g.Wait(); err != nil {
        panic(err)
    }
}
```

The code above will start two HTTP servers on ports 8080 and 8081. Both servers will be gracefully shut down when the context is canceled.

### With options:

```go
package main

import (
    "context"
    "net/http"
    "github.com/dmitrymomot/httpserver"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    r := http.NewServeMux()
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    srv := httpserver.NewServer(
        ":8080", r, 
        httpserver.WithGracefulShutdown(10*time.Second),
        httpserver.WithReadTimeout(5*time.Second),
    )
    if err := srv.Start(ctx); err != nil {
        panic(err)
    }
}
```

The code above will start an HTTP server on port 8080 with a graceful shutdown timeout of 10 seconds and a read timeout of 5 seconds.

### Serve static files:

The code below will serve static files from the `./web/static` directory under the `/static` URL prefix with a cache TTL of 10 minutes.

```go
package main

import (
    "context"
    "net/http"
    "github.com/dmitrymomot/httpserver"
    "time"
)

var (
    staticURLPrefix = "/static"
    staticDir       = "./web/static"
    staticCacheTTL  = 10 * time.Minute
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    r := http.NewServeMux()
	r.HandleFunc(staticURLPrefix+"/*", httpserver.StaticHandler(staticURLPrefix, http.Dir(staticDir), staticCacheTTL))

    if err := httpserver.Run(ctx, ":8080", r); err != nil {
        panic(err)
    }
}
```

## Contributing

Contributions to the `httpserver` package are welcome! Here are some ways you can contribute:

- Reporting bugs
- Additional tests cases
- Suggesting enhancements
- Submitting pull requests
- Sharing the love by telling others about this project

## License

This project is licensed under the [Apache 2.0](LICENSE) - see the `LICENSE` file for details.
