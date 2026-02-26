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
	"github.com/joetifa2003/inertigo/props"
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

	type FooBazProps struct {
		Foo string `json:"foo"`
		Baz string `json:"baz"`
	}

	type FooDefProps struct {
		Foo string                 `json:"foo"`
		Def props.Deferred[string] `json:"def"`
	}

	type OptProps struct {
		Opt props.Optional[string] `json:"opt"`
	}

	type AlwaysOtherProps struct {
		Alw   props.Always[string] `json:"alw"`
		Other string               `json:"other"`
	}

	type OnceProps struct {
		Onc props.Once[string] `json:"onc"`
	}

	type OnceExpProps struct {
		OncExp props.Once[string] `json:"onc_exp"`
	}

	tests := []struct {
		name            string
		headers         map[string]string
		props           any
		expectedProps   []string
		unexpectedProps []string
	}{
		{
			name: "Full Load",
			headers: map[string]string{
				inertia.XInertia: "true",
			},
			props: FooBazProps{
				Foo: "bar",
				Baz: "qux",
			},
			expectedProps: []string{"foo", "baz", "errors"},
		},
		{
			name: "Partial Reload - Select One",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialData:      "foo",
			},
			props: FooBazProps{
				Foo: "bar",
				Baz: "qux",
			},
			expectedProps:   []string{"foo"},
			unexpectedProps: []string{"baz"},
		},
		{
			name: "Partial Reload - Select multiple",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialData:      "foo,baz",
			},
			props: FooBazProps{
				Foo: "bar",
				Baz: "qux",
			},
			expectedProps: []string{"foo", "baz"},
		},
		{
			name: "Partial Reload - Except One",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialExcept:    "foo",
			},
			props: FooBazProps{
				Foo: "bar",
				Baz: "qux",
			},
			expectedProps:   []string{"baz", "errors"},
			unexpectedProps: []string{"foo"},
		},
		{
			name: "Deferred Prop - Initial Load",
			headers: map[string]string{
				inertia.XInertia: "true",
			},
			props: FooDefProps{
				Foo: "bar",
				Def: props.NewDeferred(func(ctx context.Context) (string, error) { return "deferred", nil }),
			},
			expectedProps:   []string{"foo", "errors"},
			unexpectedProps: []string{"def"},
		},
		{
			name: "Deferred Prop - Partial Load Requested",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialData:      "def",
			},
			props: FooDefProps{
				Foo: "bar",
				Def: props.NewDeferred(func(ctx context.Context) (string, error) { return "deferred", nil }),
			},
			expectedProps:   []string{"def"},
			unexpectedProps: []string{"foo"},
		},
		{
			name: "Optional Prop - Initial Load (Excluded)",
			headers: map[string]string{
				inertia.XInertia: "true",
			},
			props: OptProps{
				Opt: props.NewOptional(func(ctx context.Context) (string, error) { return "optional", nil }),
			},
			expectedProps:   []string{"errors"},
			unexpectedProps: []string{"opt"},
		},
		{
			name: "Optional Prop - Partial Load Requested",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialData:      "opt",
			},
			props: OptProps{
				Opt: props.NewOptional(func(ctx context.Context) (string, error) { return "optional", nil }),
			},
			expectedProps: []string{"opt"},
		},
		{
			name: "Always Prop - Partial Load (Not Requested)",
			headers: map[string]string{
				inertia.XInertia:                 "true",
				inertia.XInertiaPartialComponent: "TestComponent",
				inertia.XInertiaPartialData:      "other",
			},
			props: AlwaysOtherProps{
				Alw:   props.NewAlways("always"),
				Other: "other",
			},
			expectedProps: []string{"alw", "other"},
		},
		{
			name: "Once Prop - Initial Load",
			headers: map[string]string{
				inertia.XInertia: "true",
			},
			props: OnceProps{
				Onc: props.NewOnce(func(ctx context.Context) (string, error) { return "once", nil }),
			},
			expectedProps: []string{"onc", "errors"},
		},
		{
			name: "Once Prop With Expiration",
			headers: map[string]string{
				inertia.XInertia: "true",
			},
			props: OnceExpProps{
				OncExp: props.NewOnce(func(ctx context.Context) (string, error) { return "once_exp", nil }, props.Until(1*time.Hour)),
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

			for _, k := range tt.expectedProps {
				assert.Contains(t, resp.Props, k, "expected prop %q missing", k)
			}

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
	req.Header.Set(inertia.XInertia, "true")
	w := httptest.NewRecorder()

	type FooProps struct {
		Foo string `json:"foo"`
	}

	err = i.Render(w, req, "TestComponent", FooProps{Foo: "bar"})
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

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := i.Middleware(handler)

	req := httptest.NewRequest("GET", "/test-page", nil)
	req.Header.Set(inertia.XInertia, "true")
	req.Header.Set(inertia.XInertiaVersion, "client-v1")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "/test-page", w.Header().Get(inertia.XInertiaLocation))
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
	req.Header.Set(inertia.XInertia, "true")
	req.Header.Set(inertia.XInertiaVersion, "v1")
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

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set(inertia.XInertia, "true")
	req.Header.Set(inertia.XInertiaVersion, "v1")
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

	type FooProps struct {
		Foo string `json:"foo"`
	}

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
			req.Header.Set(inertia.XInertia, "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", FooProps{Foo: "bar"}, tt.options...)
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

	type FooProps struct {
		Foo string `json:"foo"`
	}

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
			req.Header.Set(inertia.XInertia, "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", FooProps{Foo: "bar"}, tt.options...)
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedValue, resp.ClearHistory)
		})
	}
}

func TestRender_MergeProp(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	type BasicMergeProps struct {
		Posts props.Merge[[]string] `json:"posts"`
	}
	type AppendMergeProps struct {
		Results props.Merge[map[string]any] `json:"results"`
	}
	type PrependMergeProps struct {
		Messages props.Merge[[]string] `json:"messages"`
	}
	type DeepMergeTestProps struct {
		Settings props.Merge[map[string]any] `json:"settings"`
	}
	type MatchOnMergeProps struct {
		Users props.Merge[[]map[string]any] `json:"users"`
	}

	tests := []struct {
		name                   string
		props                  any
		expectedMergeProps     []string
		expectedPrependProps   []string
		expectedDeepMergeProps []string
		expectedMatchPropsOn   []string
	}{
		{
			name: "basic Merge prop",
			props: BasicMergeProps{
				Posts: props.NewMerge(func(ctx context.Context) ([]string, error) {
					return []string{"post1", "post2"}, nil
				}),
			},
			expectedMergeProps: []string{"posts"},
		},
		{
			name: "Merge with Append paths",
			props: AppendMergeProps{
				Results: props.NewMerge(func(ctx context.Context) (map[string]any, error) {
					return map[string]any{"data": []string{"item1"}}, nil
				}, props.Append("data")),
			},
			expectedMergeProps: []string{"results.data"},
		},
		{
			name: "Merge with Prepend paths",
			props: PrependMergeProps{
				Messages: props.NewMerge(func(ctx context.Context) ([]string, error) {
					return []string{"msg1"}, nil
				}, props.Prepend("items")),
			},
			expectedPrependProps: []string{"messages.items"},
		},
		{
			name: "Merge with DeepMerge",
			props: DeepMergeTestProps{
				Settings: props.NewMerge(func(ctx context.Context) (map[string]any, error) {
					return map[string]any{"theme": "dark"}, nil
				}, props.DeepMerge()),
			},
			expectedDeepMergeProps: []string{"settings"},
		},
		{
			name: "Merge with MatchOn",
			props: MatchOnMergeProps{
				Users: props.NewMerge(func(ctx context.Context) ([]map[string]any, error) {
					return []map[string]any{{"id": 1, "name": "Alice"}}, nil
				}, props.MatchOn("id")),
			},
			expectedMergeProps:   []string{"users"},
			expectedMatchPropsOn: []string{"users.id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(inertia.XInertia, "true")
			w := httptest.NewRecorder()

			err := i.Render(w, req, "TestComponent", tt.props)
			require.NoError(t, err)

			var resp inertia.PageObject
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.ElementsMatch(t, tt.expectedMergeProps, resp.MergeProps)
			assert.ElementsMatch(t, tt.expectedPrependProps, resp.PrependProps)
			assert.ElementsMatch(t, tt.expectedDeepMergeProps, resp.DeepMergeProps)
			assert.ElementsMatch(t, tt.expectedMatchPropsOn, resp.MatchPropsOn)
		})
	}
}

func TestRender_MultipleOptionsComposability(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(inertia.XInertia, "true")
	w := httptest.NewRecorder()

	type ComposableProps struct {
		Data  string                `json:"data"`
		Items props.Merge[[]string] `json:"items"`
	}

	err = i.Render(w, req, "TestComponent", ComposableProps{
		Data: "value",
		Items: props.NewMerge(func(ctx context.Context) ([]string, error) {
			return []string{"item1"}, nil
		}),
	},
		inertia.WithEncryptHistory(true),
		inertia.WithClearHistory(true),
	)
	require.NoError(t, err)

	var resp inertia.PageObject
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.True(t, resp.EncryptHistory)
	assert.True(t, resp.ClearHistory)
	assert.Equal(t, []string{"items"}, resp.MergeProps)
}

func TestRenderErrors(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	t.Run("Precognition Success (No Errors)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set(inertia.HeaderPrecognition, "true")
		w := httptest.NewRecorder()

		err := i.RenderErrors(w, req, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "true", w.Header().Get(inertia.HeaderPrecognitionSuccess))
	})

	t.Run("Precognition Error (With Errors)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set(inertia.HeaderPrecognition, "true")
		w := httptest.NewRecorder()

		errors := map[string]any{"field": "error"}
		err := i.RenderErrors(w, req, errors)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var body map[string]any
		json.NewDecoder(w.Body).Decode(&body)
		assert.Equal(t, "error", body["errors"].(map[string]any)["field"])
	})

	t.Run("Standard Request (With Errors) - Redirects Back", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users", nil)
		req.Header.Set("Referer", "/register")
		w := httptest.NewRecorder()

		errors := map[string]any{"field": "error"}
		err := i.RenderErrors(w, req, errors)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/register", w.Header().Get("Location"))
	})

	t.Run("Standard Request (With Errors) - Fallback to root", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users", nil)
		w := httptest.NewRecorder()

		errors := map[string]any{"field": "error"}
		err := i.RenderErrors(w, req, errors)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/", w.Header().Get("Location"))
	})

	t.Run("Standard Request (No Errors)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		err := i.RenderErrors(w, req, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 0, w.Body.Len())
	})
}

func TestValidationErrors_FullFlow(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	t.Run("Flashed errors are shared via middleware and rendered", func(t *testing.T) {
		postReq := httptest.NewRequest("POST", "/users", nil)
		postReq.Header.Set("Referer", "/register")
		postW := httptest.NewRecorder()

		errors := map[string]any{"email": "Email is required"}
		err := i.RenderErrors(postW, postReq, errors)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, postW.Code)

		cookies := postW.Result().Cookies()
		require.NotEmpty(t, cookies)

		getReq := httptest.NewRequest("GET", "/register", nil)
		getReq.Header.Set(inertia.XInertia, "true")
		for _, c := range cookies {
			getReq.AddCookie(c)
		}
		getW := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := i.Render(w, r, "register", nil)
			require.NoError(t, err)
		})

		i.Middleware(handler).ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusOK, getW.Code)
		var resp inertia.PageObject
		json.NewDecoder(getW.Body).Decode(&resp)
		assert.Equal(t, "register", resp.Component)
		assert.NotNil(t, resp.Props["errors"])
		errorsMap := resp.Props["errors"].(map[string]any)
		assert.Equal(t, "Email is required", errorsMap["email"])

		getReq2 := httptest.NewRequest("GET", "/register", nil)
		getReq2.Header.Set(inertia.XInertia, "true")
		for _, c := range cookies {
			getReq2.AddCookie(c)
		}
		getW2 := httptest.NewRecorder()

		i.Middleware(handler).ServeHTTP(getW2, getReq2)

		var resp2 inertia.PageObject
		json.NewDecoder(getW2.Body).Decode(&resp2)
		errorsMap2, ok := resp2.Props["errors"].(map[string]any)
		assert.True(t, ok, "errors should be a map")
		assert.Empty(t, errorsMap2, "errors should be empty on second request")
	})

	t.Run("Error Bags are respected", func(t *testing.T) {
		postReq := httptest.NewRequest("POST", "/login", nil)
		postReq.Header.Set("Referer", "/login")
		postReq.Header.Set(inertia.XInertiaErrorBag, "loginBag")
		postW := httptest.NewRecorder()

		errors := map[string]any{"email": "Invalid credentials"}
		err := i.RenderErrors(postW, postReq, errors)
		require.NoError(t, err)

		cookies := postW.Result().Cookies()

		getReq := httptest.NewRequest("GET", "/login", nil)
		getReq.Header.Set(inertia.XInertia, "true")
		for _, c := range cookies {
			getReq.AddCookie(c)
		}
		getW := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := i.Render(w, r, "login", nil)
			require.NoError(t, err)
		})

		i.Middleware(handler).ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusOK, getW.Code)
		var resp inertia.PageObject
		json.NewDecoder(getW.Body).Decode(&resp)

		assert.NotNil(t, resp.Props["errors"])
		errorsMap := resp.Props["errors"].(map[string]any)

		assert.NotNil(t, errorsMap["loginBag"])
		bagMap, ok := errorsMap["loginBag"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Invalid credentials", bagMap["email"])
	})
}

func TestRender_LazyProp(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	t.Run("Lazy prop is always included", func(t *testing.T) {
		callCount := 0
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		w := httptest.NewRecorder()

		type UsersProps struct {
			Users props.Lazy[[]string] `json:"users"`
		}

		err := i.Render(w, req, "TestComponent", UsersProps{
			Users: props.NewLazy(func(ctx context.Context) ([]string, error) {
				callCount++
				return []string{"user1", "user2"}, nil
			}),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Contains(t, resp.Props, "users")
		assert.Equal(t, 1, callCount, "resolver should be called exactly once")
	})

	t.Run("Lazy prop included on partial reload when requested", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		req.Header.Set(inertia.XInertiaPartialComponent, "TestComponent")
		req.Header.Set(inertia.XInertiaPartialData, "users")
		w := httptest.NewRecorder()

		type UsersOtherProps struct {
			Users props.Lazy[[]string] `json:"users"`
			Other string               `json:"other"`
		}

		err := i.Render(w, req, "TestComponent", UsersOtherProps{
			Users: props.NewLazy(func(ctx context.Context) ([]string, error) {
				return []string{"user1", "user2"}, nil
			}),
			Other: "data",
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Contains(t, resp.Props, "users")
		assert.NotContains(t, resp.Props, "other")
	})
}

func TestRender_ScrollProp(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	type PostsProps struct {
		Posts props.Scroll[string] `json:"posts"`
	}

	type FooPostsProps struct {
		Foo   string               `json:"foo"`
		Posts props.Scroll[string] `json:"posts"`
	}

	t.Run("ScrollProp with metadata - initial load", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1", "post2"}, nil
			}, props.WithScrollMetadata(props.ScrollMetadata{
				PageName:     "page",
				CurrentPage:  1,
				PreviousPage: nil,
				NextPage:     2,
			})),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Contains(t, resp.Props, "posts")
		postsData := resp.Props["posts"].(map[string]any)
		assert.Contains(t, postsData, "data")

		assert.Contains(t, resp.MergeProps, "posts.data")

		assert.NotNil(t, resp.ScrollProps)
		scrollMeta := resp.ScrollProps["posts"]
		assert.Equal(t, "page", scrollMeta.PageName)
		assert.Equal(t, float64(1), scrollMeta.CurrentPage)
		assert.Nil(t, scrollMeta.PreviousPage)
		assert.Equal(t, float64(2), scrollMeta.NextPage)
		assert.False(t, scrollMeta.Reset)
	})

	t.Run("ScrollProp with custom wrapper", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1"}, nil
			}, props.WithWrapper("items")),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		postsData := resp.Props["posts"].(map[string]any)
		assert.Contains(t, postsData, "items")
		assert.NotContains(t, postsData, "data")

		assert.Contains(t, resp.MergeProps, "posts.items")
	})

	t.Run("ScrollProp merge intent - prepend", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		req.Header.Set(inertia.XInertiaInfiniteScrollMergeIntent, "prepend")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1"}, nil
			}),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Contains(t, resp.PrependProps, "posts.data")
	})

	t.Run("ScrollProp merge intent - append (default)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1"}, nil
			}),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Contains(t, resp.MergeProps, "posts.data")
	})

	t.Run("ScrollProp with reset flag", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		req.Header.Set(inertia.XInertiaReset, "posts")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1"}, nil
			}),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.True(t, resp.ScrollProps["posts"].Reset)
	})

	t.Run("ScrollProp with cursor pagination (string values)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", PostsProps{
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				return []string{"post1"}, nil
			}, props.WithScrollMetadata(props.ScrollMetadata{
				PageName:     "cursor",
				CurrentPage:  "eyJpZCI6MTB9",
				PreviousPage: "eyJpZCI6NX0=",
				NextPage:     "eyJpZCI6MTV9",
			})),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		scrollMeta := resp.ScrollProps["posts"]
		assert.Equal(t, "cursor", scrollMeta.PageName)
		assert.Equal(t, "eyJpZCI6MTB9", scrollMeta.CurrentPage)
		assert.Equal(t, "eyJpZCI6NX0=", scrollMeta.PreviousPage)
		assert.Equal(t, "eyJpZCI6MTV9", scrollMeta.NextPage)
	})

	t.Run("ScrollProp not resolved on partial reload when not requested", func(t *testing.T) {
		resolverCalled := false
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		req.Header.Set(inertia.XInertiaPartialComponent, "Posts/Index")
		req.Header.Set(inertia.XInertiaPartialData, "foo")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", FooPostsProps{
			Foo: "bar",
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				resolverCalled = true
				return []string{"post1"}, nil
			}, props.WithScrollMetadata(props.ScrollMetadata{
				PageName:    "page",
				CurrentPage: 1,
			})),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.False(t, resolverCalled, "ScrollProp resolver should not be called when not requested")
		assert.NotContains(t, resp.Props, "posts", "posts should not be in props")
		assert.Empty(t, resp.ScrollProps, "scrollProps should be empty")

		assert.Contains(t, resp.Props, "foo")
	})

	t.Run("ScrollProp resolved on partial reload when requested", func(t *testing.T) {
		resolverCalled := false
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(inertia.XInertia, "true")
		req.Header.Set(inertia.XInertiaPartialComponent, "Posts/Index")
		req.Header.Set(inertia.XInertiaPartialData, "posts")
		w := httptest.NewRecorder()

		err := i.Render(w, req, "Posts/Index", FooPostsProps{
			Foo: "bar",
			Posts: props.NewScroll(func(ctx context.Context) ([]string, error) {
				resolverCalled = true
				return []string{"post1"}, nil
			}, props.WithScrollMetadata(props.ScrollMetadata{
				PageName:    "page",
				CurrentPage: 1,
			})),
		})
		require.NoError(t, err)

		var resp inertia.PageObject
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.True(t, resolverCalled, "ScrollProp resolver should be called when requested")
		assert.Contains(t, resp.Props, "posts", "posts should be in props")
		assert.NotEmpty(t, resp.ScrollProps, "scrollProps should not be empty")

		assert.NotContains(t, resp.Props, "foo")
	})
}
