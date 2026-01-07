package inertia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inertia "github.com/joetifa2003/inertigo"
	"github.com/joetifa2003/inertigo/vite"
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

func TestRender_VersionInPageObject(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler, inertia.WithVersion("test-v1"))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Inertia", "true")
	w := httptest.NewRecorder()

	err = i.Render(w, req, "TestComponent", inertia.Props{"foo": "bar"})
	require.NoError(t, err)

	var resp inertia.PageObject
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "test-v1", resp.Version)
}

func TestMiddleware_VersionMismatchReturns409(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler, inertia.WithVersion("server-v2"))
	require.NoError(t, err)

	// Create a handler that should not be called on version mismatch
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := i.Middleware(handler)

	// Request with old version
	req := httptest.NewRequest("GET", "/test-page", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "client-v1") // Mismatched version
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "/test-page", w.Header().Get("X-Inertia-Location"))
	assert.False(t, handlerCalled, "handler should not be called on version mismatch")
}

func TestMiddleware_VersionMatchContinues(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler, inertia.WithVersion("v1"))
	require.NoError(t, err)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := i.Middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "v1") // Matching version
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "handler should be called when versions match")
}

func TestMiddleware_POSTRequestNoConflict(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler, inertia.WithVersion("v2"))
	require.NoError(t, err)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := i.Middleware(handler)

	// POST request with mismatched version should NOT trigger 409
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Version", "v1") // Mismatched, but POST
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "handler should be called for POST requests even with version mismatch")
}

func TestMiddleware_NonInertiaRequest(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler, inertia.WithVersion("v2"))
	require.NoError(t, err)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := i.Middleware(handler)

	// Non-Inertia request (no X-Inertia header)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "handler should be called for non-Inertia requests")
}

func TestRender_EncryptHistoryOption(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	tests := []struct {
		name          string
		options       []inertia.RenderOption
		expectedValue bool
	}{
		{
			name:          "no options - defaults to false",
			options:       nil,
			expectedValue: false,
		},
		{
			name:          "WithEncryptHistory(true)",
			options:       []inertia.RenderOption{inertia.WithEncryptHistory(true)},
			expectedValue: true,
		},
		{
			name:          "WithEncryptHistory(false)",
			options:       []inertia.RenderOption{inertia.WithEncryptHistory(false)},
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Inertia", "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", inertia.Props{"foo": "bar"}, tt.options...)
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedValue, resp.EncryptHistory)
		})
	}
}

func TestRender_ClearHistoryOption(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	tests := []struct {
		name          string
		options       []inertia.RenderOption
		expectedValue bool
	}{
		{
			name:          "no options - defaults to false",
			options:       nil,
			expectedValue: false,
		},
		{
			name:          "WithClearHistory(true)",
			options:       []inertia.RenderOption{inertia.WithClearHistory(true)},
			expectedValue: true,
		},
		{
			name:          "WithClearHistory(false)",
			options:       []inertia.RenderOption{inertia.WithClearHistory(false)},
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Inertia", "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", inertia.Props{"foo": "bar"}, tt.options...)
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedValue, resp.ClearHistory)
		})
	}
}

func TestRender_MergePropsOptions(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	tests := []struct {
		name                   string
		options                []inertia.RenderOption
		expectedMergeProps     []string
		expectedPrependProps   []string
		expectedDeepMergeProps []string
		expectedMatchPropsOn   []string
	}{
		{
			name:                   "no options - all nil",
			options:                nil,
			expectedMergeProps:     nil,
			expectedPrependProps:   nil,
			expectedDeepMergeProps: nil,
			expectedMatchPropsOn:   nil,
		},
		{
			name:               "WithMergeProps",
			options:            []inertia.RenderOption{inertia.WithMergeProps("posts", "comments")},
			expectedMergeProps: []string{"posts", "comments"},
		},
		{
			name:                 "WithPrependProps",
			options:              []inertia.RenderOption{inertia.WithPrependProps("notifications")},
			expectedPrependProps: []string{"notifications"},
		},
		{
			name:                   "WithDeepMergeProps",
			options:                []inertia.RenderOption{inertia.WithDeepMergeProps("conversations")},
			expectedDeepMergeProps: []string{"conversations"},
		},
		{
			name:                 "WithMatchPropsOn",
			options:              []inertia.RenderOption{inertia.WithMatchPropsOn("posts.id", "comments.id")},
			expectedMatchPropsOn: []string{"posts.id", "comments.id"},
		},
		{
			name: "multiple options combined",
			options: []inertia.RenderOption{
				inertia.WithMergeProps("posts"),
				inertia.WithPrependProps("notifications"),
				inertia.WithDeepMergeProps("conversations"),
				inertia.WithMatchPropsOn("posts.id"),
			},
			expectedMergeProps:     []string{"posts"},
			expectedPrependProps:   []string{"notifications"},
			expectedDeepMergeProps: []string{"conversations"},
			expectedMatchPropsOn:   []string{"posts.id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Inertia", "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", inertia.Props{"foo": "bar"}, tt.options...)
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMergeProps, resp.MergeProps)
			assert.Equal(t, tt.expectedPrependProps, resp.PrependProps)
			assert.Equal(t, tt.expectedDeepMergeProps, resp.DeepMergeProps)
			assert.Equal(t, tt.expectedMatchPropsOn, resp.MatchPropsOn)
		})
	}
}

func TestRender_MultipleOptionsComposability(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Inertia", "true")
	w := httptest.NewRecorder()

	// Test that all options can be used together
	err = i.Render(w, req, "TestComponent", inertia.Props{"data": "value"},
		inertia.WithEncryptHistory(true),
		inertia.WithClearHistory(true),
		inertia.WithMergeProps("items"),
		inertia.WithPrependProps("newItems"),
		inertia.WithDeepMergeProps("settings"),
		inertia.WithMatchPropsOn("items.id"),
	)
	require.NoError(t, err)

	var resp inertia.PageObject
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.True(t, resp.EncryptHistory)
	assert.True(t, resp.ClearHistory)
	assert.Equal(t, []string{"items"}, resp.MergeProps)
	assert.Equal(t, []string{"newItems"}, resp.PrependProps)
	assert.Equal(t, []string{"settings"}, resp.DeepMergeProps)
	assert.Equal(t, []string{"items.id"}, resp.MatchPropsOn)
}
