package inertia

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRender_PartialReload(t *testing.T) {
	i, err := New(
		WithViteURL("http://localhost:5173"),
		WithEntryPoint("test.js"),
		WithDevMode(true),
		WithViteFS(os.DirFS(".")),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		headers         map[string]string
		props           Props
		expectedProps   []string
		unexpectedProps []string
	}{
		{
			name: "Full Load",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: Props{
				"foo": "bar",
				"baz": "qux",
			},
			expectedProps: []string{"foo", "baz", "errors"},
		},
		{
			name: "Partial Reload - Select One",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Data":      "foo",
			},
			props: Props{
				"foo": "bar",
				"baz": "qux",
			},
			expectedProps:   []string{"foo", "errors"},
			unexpectedProps: []string{"baz"},
		},
		{
			name: "Partial Reload - Select multiple",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Data":      "foo,baz",
			},
			props: Props{
				"foo": "bar",
				"baz": "qux",
			},
			expectedProps: []string{"foo", "baz", "errors"},
		},
		{
			name: "Partial Reload - Except One",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Except":    "foo",
			},
			props: Props{
				"foo": "bar",
				"baz": "qux",
			},
			expectedProps:   []string{"baz", "errors"},
			unexpectedProps: []string{"foo"},
		},
		{
			name: "Deferred Prop - Initial Load",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: Props{
				"foo": "bar",
				"def": Deferred(func(ctx context.Context) (any, error) { return "deferred", nil }),
			},
			expectedProps:   []string{"foo", "errors"},
			unexpectedProps: []string{"def"},
		},
		{
			name: "Deferred Prop - Partial Load Requested",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Data":      "def",
			},
			props: Props{
				"foo": "bar",
				"def": Deferred(func(ctx context.Context) (any, error) { return "deferred", nil }),
			},
			expectedProps:   []string{"def", "errors"},
			unexpectedProps: []string{"foo"},
		},
		{
			name: "Optional Prop - Initial Load (Excluded)",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: Props{
				"opt": Optional(func(ctx context.Context) (any, error) { return "optional", nil }),
			},
			expectedProps:   []string{"errors"},
			unexpectedProps: []string{"opt"},
		},
		{
			name: "Optional Prop - Partial Load Requested",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Data":      "opt",
			},
			props: Props{
				"opt": Optional(func(ctx context.Context) (any, error) { return "optional", nil }),
			},
			expectedProps: []string{"opt", "errors"},
		},
		{
			name: "Always Prop - Partial Load (Not Requested)",
			headers: map[string]string{
				"X-Inertia":                   "true",
				"X-Inertia-Partial-Component": "TestComponent",
				"X-Inertia-Partial-Data":      "other",
			},
			props: Props{
				"alw":   Always("always"),
				"other": "other",
			},
			expectedProps: []string{"alw", "other", "errors"},
		},
		{
			name: "Once Prop - Initial Load",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: Props{
				"onc": Once(func(ctx context.Context) (any, error) { return "once", nil }),
			},
			expectedProps: []string{"onc", "errors"},
		},
		{
			name: "Once Prop With Expiration",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: Props{
				"onc_exp": OnceWithExpiration(func(ctx context.Context) (any, error) { return "once_exp", nil }, 1*time.Hour),
			},
			expectedProps: []string{"onc_exp", "errors"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", tt.props)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			var resp PageObject
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}

			// Verify expected props present
			for _, k := range tt.expectedProps {
				if _, ok := resp.Props[k]; !ok {
					t.Errorf("expected prop %q missing", k)
				}
			}

			// Verify unexpected props missing
			for _, k := range tt.unexpectedProps {
				if _, ok := resp.Props[k]; ok {
					t.Errorf("unexpected prop %q present", k)
				}
			}
		})
	}
}
