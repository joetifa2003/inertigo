package inertia_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/joetifa2003/inertigo"
)

func TestRedirect(t *testing.T) {
	inertia := &Inertia{}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request should use 302 Found",
			method:         http.MethodGet,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "POST request should use 302 Found",
			method:         http.MethodPost,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "PUT request should use 303 See Other",
			method:         http.MethodPut,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "PATCH request should use 303 See Other",
			method:         http.MethodPatch,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "DELETE request should use 303 See Other",
			method:         http.MethodDelete,
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/current", nil)

			inertia.Redirect(w, r, "/target")

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if location := w.Header().Get("Location"); location != "/target" {
				t.Errorf("expected Location header %q, got %q", "/target", location)
			}
		})
	}
}

func TestLocation(t *testing.T) {
	inertia := &Inertia{}

	t.Run("Inertia request should use 409 Conflict and X-Inertia-Location header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/current", nil)
		r.Header.Set("X-Inertia", "true")

		inertia.Location(w, r, "https://external.com")

		if w.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
		}

		if location := w.Header().Get("X-Inertia-Location"); location != "https://external.com" {
			t.Errorf("expected X-Inertia-Location header %q, got %q", "https://external.com", location)
		}
	})

	t.Run("Standard request should use 302 Found (or 303 See Other)", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/current", nil)

		inertia.Location(w, r, "https://external.com")

		if w.Code != http.StatusFound {
			t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
		}

		if location := w.Header().Get("Location"); location != "https://external.com" {
			t.Errorf("expected Location header %q, got %q", "https://external.com", location)
		}
	})
}
