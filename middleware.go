package inertia

import (
	"context"
	"net/http"
)

// Middleware wraps an http.Handler to handle Inertia-specific concerns:
// - Asset versioning (409 Conflict on version mismatch for GET requests)
// - Managing shared and flash props via pooled inertiaContext
// - CSRF protection (if enabled)
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ic := inertiaContextPool.Get()
		defer inertiaContextPool.Put(ic)

		if flashData, _ := i.session.Get(w, r); flashData != nil {
			ic.flash = flashData
		}

		ctx := context.WithValue(r.Context(), inertiaContextKey, &ic)
		r = r.WithContext(ctx)

		if i.csrfEnabled {
			if needsNewCSRFToken(r) {
				token := generateCSRFToken()
				setCSRFCookie(w, token, i.csrfConfig)
			}

			if isStateChangingMethod(r.Method) {
				if !validateCSRF(r) {
					http.Error(w, "CSRF token mismatch", http.StatusForbidden)
					return
				}
			}
		}

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
