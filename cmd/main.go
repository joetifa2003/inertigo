package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"

	inertia "go-ssr-experiment"
)

func main() {
	isDev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	i, err := inertia.New(
		inertia.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),

		inertia.WithViteFS(os.DirFS("assets")),
		inertia.WithViteDistFS(os.DirFS("assets/dist")),
		inertia.WithRooHtmlPathFS(os.DirFS("assets"), "index.html"),
		inertia.WithEntryPoint("ts/app.tsx"),

		inertia.WithViteURL("http://localhost:5173"),
		inertia.WithReactRefresh(),
		inertia.WithDevMode(*isDev),

		inertia.WithSSR(true, func() (inertia.SSREngine, error) {
			return inertia.NewQJSEngine(
				inertia.WithSrcPath("assets/dist/server/ssr.js"),
				inertia.WithClusterSize(16),
			)
		}),
	)
	must(err)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = i.Render(w, r, "index", inertia.Props{
			"title": "Hello, world!",
		})
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		err = i.Render(w, r, "about", nil)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})

	i.Logger().Log(context.Background(), slog.LevelInfo, "starting server", slog.String("url", "http://localhost:8000"))

	http.ListenAndServe(":8000", mux)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
