package inertia

import "net/http"

// Middleware wraps an http.Handler to handle Inertia-specific concerns:
// - Asset versioning (409 Conflict on version mismatch for GET requests)
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle version mismatch on Inertia GET requests
		if r.Method == http.MethodGet &&
			r.Header.Get(XInertia) == "true" &&
			i.version != "" {
			clientVersion := r.Header.Get("X-Inertia-Version")
			if clientVersion != "" && clientVersion != i.version {
				w.Header().Set("X-Inertia-Location", r.URL.String())
				w.WriteHeader(http.StatusConflict)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
