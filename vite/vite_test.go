package vite

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDistFS creates a mock filesystem with a manifest and some assets
func mockDistFS() fs.FS {
	return fstest.MapFS{
		".vite/manifest.json": &fstest.MapFile{
			Data: []byte(`{
				"ts/app.tsx": {
					"file": "assets/app-abc123.js",
					"src": "ts/app.tsx",
					"isEntry": true,
					"css": ["assets/app-abc123.css"],
					"imports": ["ts/vendor.tsx"]
				},
				"ts/vendor.tsx": {
					"file": "assets/vendor-def456.js",
					"src": "ts/vendor.tsx",
					"css": ["assets/vendor-def456.css"]
				}
			}`),
		},
		"assets/app-abc123.js": &fstest.MapFile{
			Data: []byte(`console.log("app");`),
		},
		"assets/app-abc123.css": &fstest.MapFile{
			Data: []byte(`body { color: red; }`),
		},
		"assets/vendor-def456.js": &fstest.MapFile{
			Data: []byte(`console.log("vendor");`),
		},
		"assets/vendor-def456.css": &fstest.MapFile{
			Data: []byte(`body { margin: 0; }`),
		},
	}
}

func TestBundler_Handler(t *testing.T) {
	t.Run("returns not found in dev mode", func(t *testing.T) {
		b, err := New(mockDistFS(), WithDevMode(true))
		require.NoError(t, err)

		handler := b.Handler()
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "404 page not found\n", rec.Body.String())
	})

	t.Run("returns not found with nil distFS", func(t *testing.T) {
		b, err := New(nil, WithDevMode(true))
		require.NoError(t, err)

		handler := b.Handler()
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "404 page not found\n", rec.Body.String())
	})

	t.Run("serves assets with default prefix", func(t *testing.T) {
		b, err := New(mockDistFS())
		require.NoError(t, err)

		handler := b.Handler()
		require.NotNil(t, handler)

		req := httptest.NewRequest(http.MethodGet, "/static/assets/app-abc123.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
	})

	t.Run("serves assets with custom prefix", func(t *testing.T) {
		b, err := New(mockDistFS(), WithAssetPrefix("/static/"))
		require.NoError(t, err)

		handler := b.Handler()
		require.NotNil(t, handler)

		req := httptest.NewRequest(http.MethodGet, "/static/assets/app-abc123.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
	})

	t.Run("serves CSS files", func(t *testing.T) {
		b, err := New(mockDistFS())
		require.NoError(t, err)

		handler := b.Handler()
		req := httptest.NewRequest(http.MethodGet, "/static/assets/app-abc123.css", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "color: red")
	})
}

func TestBundler_prodTags(t *testing.T) {
	t.Run("generates URLs with default prefix", func(t *testing.T) {
		b, err := New(mockDistFS())
		require.NoError(t, err)

		html := string(b.viteTagsFunc("ts/app.tsx"))

		// Check main entry script
		assert.Contains(t, html, `src="/static/assets/app-abc123.js"`)

		// Check CSS link
		assert.Contains(t, html, `href="/static/assets/app-abc123.css"`)

		// Check preloaded vendor CSS
		assert.Contains(t, html, `href="/static/assets/vendor-def456.css"`)

		// Check preloaded vendor JS
		assert.Contains(t, html, `href="/static/assets/vendor-def456.js"`)
	})

	t.Run("generates URLs with custom prefix", func(t *testing.T) {
		b, err := New(mockDistFS(), WithAssetPrefix("/static/"))
		require.NoError(t, err)

		html := string(b.viteTagsFunc("ts/app.tsx"))

		// Check main entry script
		assert.Contains(t, html, `src="/static/assets/app-abc123.js"`)

		// Check CSS link
		assert.Contains(t, html, `href="/static/assets/app-abc123.css"`)

		// Check preloaded vendor CSS
		assert.Contains(t, html, `href="/static/assets/vendor-def456.css"`)

		// Check preloaded vendor JS
		assert.Contains(t, html, `href="/static/assets/vendor-def456.js"`)
	})

	t.Run("handles missing entry gracefully", func(t *testing.T) {
		b, err := New(mockDistFS())
		require.NoError(t, err)

		html := string(b.viteTagsFunc("nonexistent.tsx"))

		assert.Contains(t, html, "not found in manifest")
	})
}

func TestBundler_AssetPrefix(t *testing.T) {
	t.Run("returns default prefix", func(t *testing.T) {
		b, err := New(mockDistFS())
		require.NoError(t, err)
		assert.Equal(t, "/static/", b.AssetPrefix())
	})

	t.Run("returns custom prefix", func(t *testing.T) {
		b, err := New(mockDistFS(), WithAssetPrefix("/my-assets/"))
		require.NoError(t, err)
		assert.Equal(t, "/my-assets/", b.AssetPrefix())
	})
}
