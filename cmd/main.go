package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	inertia "github.com/joetifa2003/inertigo"
	"github.com/joetifa2003/inertigo/qjs"
	"github.com/joetifa2003/inertigo/vite"
)

func main() {
	isDev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	bundler, err := vite.New(
		os.DirFS("assets/dist"),
		vite.WithDevMode(*isDev),
		vite.WithReactRefresh(),
	)
	must(err)

	i, err := inertia.New(
		bundler,
		inertia.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
		inertia.WithRooHtmlPathFS(os.DirFS("assets"), "index.html"),
		inertia.WithSSR(true, func() (inertia.SSREngine, error) {
			return qjs.New(
				qjs.WithSrcPath("assets/dist/server/ssr.js"),
				qjs.WithClusterSize(16),
			)
		}),
	)
	must(err)

	mux := http.NewServeMux()

	mux.Handle(bundler.AssetPrefix(), bundler.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = i.Render(w, r, "index", inertia.Props{
			"title": inertia.Value("Hello, world!"),
			"lazyMessage": inertia.Deferred(func(ctx context.Context) (any, error) {
				time.Sleep(1 * time.Second)
				return "This message is lazy loaded", nil
			}),
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

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		err = i.Render(w, r, "register", nil)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body struct {
				Name     string `json:"name"`
				Email    string `json:"email"`
				Password string `json:"password"`
			}

			if r.Header.Get("Content-Type") == "application/json" {
				json.NewDecoder(r.Body).Decode(&body)
			} else {
				r.ParseForm()
				body.Name = r.FormValue("name")
				body.Email = r.FormValue("email")
				body.Password = r.FormValue("password")
			}

			errors := map[string]any{}

			if inertia.ShouldValidateField(r, "name") {
				if body.Name == "" {
					errors["name"] = "Name is required"
				}
			}

			if inertia.ShouldValidateField(r, "email") {
				if body.Email == "" {
					errors["email"] = "Email is required"
				} else if !strings.Contains(body.Email, "@") {
					errors["email"] = "Email must contain @"
				}
			}

			if inertia.ShouldValidateField(r, "password") {
				if body.Password == "" {
					errors["password"] = "Password is required"
				}
			}

			err := i.RenderErrors(w, r, errors)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			if len(errors) != 0 {
				return
			}

			i.Redirect(w, r, "/")
		}
	})

	i.Logger().Log(context.Background(), slog.LevelInfo, "starting server", slog.String("url", "http://localhost:8001"))

	http.ListenAndServe(":8001", i.Middleware(mux))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
