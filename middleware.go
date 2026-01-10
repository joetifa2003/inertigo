package inertia

import (
	"context"
	"net/http"
)

type contextKey string

const (
	// inertiaErrorsKey is the context key for storing flashed validation errors
	inertiaErrorsKey contextKey = "inertia_errors"
)

// Middleware wraps an http.Handler to handle Inertia-specific concerns:
// - Asset versioning (409 Conflict on version mismatch for GET requests)
// - Retrieving flashed validation errors from session and adding them to context
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve flashed validation errors from session and add to context
		if errors, _ := i.session.Get(w, r, "errors"); errors != nil {
			ctx := context.WithValue(r.Context(), inertiaErrorsKey, errors)
			r = r.WithContext(ctx)
		}

		// Handle version mismatch on Inertia GET requests
		if r.Method == http.MethodGet &&
			r.Header.Get(XInertia) == "true" &&
			i.version != "" {
			clientVersion := r.Header.Get(XInertiaVersion)
			if clientVersion != "" && clientVersion != i.version {
				w.Header().Set(XInertiaLocation, r.URL.String())
				w.WriteHeader(http.StatusConflict)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
