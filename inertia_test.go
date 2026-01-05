package inertia_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	inertia "github.com/joetifa2003/inertigo"
	"github.com/joetifa2003/inertigo/vite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender_PartialReload(t *testing.T) {
	bundler, err := vite.New(
		nil,
		vite.WithDevMode(true),
	)
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	tests := []struct {
		name            string
		headers         map[string]string
		props           inertia.Props
		expectedProps   []string
		unexpectedProps []string
	}{
		{
			name: "Full Load",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: inertia.Props{
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
			props: inertia.Props{
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
			props: inertia.Props{
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
			props: inertia.Props{
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
			props: inertia.Props{
				"foo": "bar",
				"def": inertia.Deferred(func(ctx context.Context) (any, error) { return "deferred", nil }),
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
			props: inertia.Props{
				"foo": "bar",
				"def": inertia.Deferred(func(ctx context.Context) (any, error) { return "deferred", nil }),
			},
			expectedProps:   []string{"def", "errors"},
			unexpectedProps: []string{"foo"},
		},
		{
			name: "Optional Prop - Initial Load (Excluded)",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: inertia.Props{
				"opt": inertia.Optional(func(ctx context.Context) (any, error) { return "optional", nil }),
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
			props: inertia.Props{
				"opt": inertia.Optional(func(ctx context.Context) (any, error) { return "optional", nil }),
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
			props: inertia.Props{
				"alw":   inertia.Always("always"),
				"other": "other",
			},
			expectedProps: []string{"alw", "other", "errors"},
		},
		{
			name: "Once Prop - Initial Load",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: inertia.Props{
				"onc": inertia.Once(func(ctx context.Context) (any, error) { return "once", nil }),
			},
			expectedProps: []string{"onc", "errors"},
		},
		{
			name: "Once Prop With Expiration",
			headers: map[string]string{
				"X-Inertia": "true",
			},
			props: inertia.Props{
				"onc_exp": inertia.OnceWithExpiration(func(ctx context.Context) (any, error) { return "once_exp", nil }, 1*time.Hour),
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
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			// Verify expected props present
			for _, k := range tt.expectedProps {
				assert.Contains(t, resp.Props, k, "expected prop %q missing", k)
			}

			// Verify unexpected props missing
			for _, k := range tt.unexpectedProps {
				assert.NotContains(t, resp.Props, k, "unexpected prop %q present", k)
			}
		})
	}
}
