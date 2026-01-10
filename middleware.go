package inertia

import (
	"context"
	"net/http"
)

// Middleware wraps an http.Handler to handle Inertia-specific concerns:
// - Asset versioning (409 Conflict on version mismatch for GET requests)
// - Managing shared and flash props via pooled inertiaContext
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get pooled inertiaContext
		ic := inertiaContextPool.Get()
		defer inertiaContextPool.Put(ic)

		// Load flash data from session
		if flashData, _ := i.session.Get(w, r); flashData != nil {
			ic.flash = flashData
		}

		// Inject inertiaContext into request
		ctx := context.WithValue(r.Context(), inertiaContextKey, &ic)
		r = r.WithContext(ctx)

		// CSRF Protection
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
