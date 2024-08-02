package main

import (
	"context"
	"embed"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrymomot/httpserver"
)

var (
	staticURLPrefix = "/static"
	staticCacheTTL  = 30 * time.Second
)

//go:embed static
var staticFS embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Handle(staticURLPrefix+"/*", httpserver.EmbeddedStaticHandler(staticFS, staticCacheTTL))

	if err := httpserver.Run(ctx, "localhost:8080", r); err != nil {
		panic(err)
	}
}
