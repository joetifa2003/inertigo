package inertia

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const (
	csrfCookieName  = "XSRF-TOKEN"
	csrfHeaderName  = "X-XSRF-TOKEN"
	csrfTokenLength = 32
)

type csrfConfig struct {
	cookieSecure bool
}

func generateCSRFToken() string {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		panic("impossible, read never returns an error")
	}
	return hex.EncodeToString(bytes)
}

func needsNewCSRFToken(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	return err != nil || cookie.Value == ""
}

func setCSRFCookie(w http.ResponseWriter, token string, config csrfConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Must be false so axios can read it
		Secure:   config.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func validateCSRF(r *http.Request) bool {
	// Get token from cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	// Get token from header
	headerToken := r.Header.Get(csrfHeaderName)
	if headerToken == "" {
		return false
	}

	// Constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) == 1
}

func isStateChangingMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}
