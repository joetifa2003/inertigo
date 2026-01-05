package vite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	inertia "github.com/joetifa2003/inertigo"
)

// Bundler implements the inertia.Bundler interface for Vite projects.
type Bundler struct {
	isDev            bool
	viteURL          string
	manifest         map[string]manifestChunk
	withReactRefresh bool
	distFS           fs.FS
	assetPrefix      string
}

type manifestChunk struct {
	File           string   `json:"file"`
	Name           string   `json:"name"`
	Src            string   `json:"src"`
	CSS            []string `json:"css"`
	IsEntry        bool     `json:"isEntry"`
	IsDynamicEntry bool     `json:"isDynamicEntry"`
	Imports        []string `json:"imports"`
	DynamicImports []string `json:"dynamicImports"`
}

type config struct {
	isDev            bool
	viteURL          string
	withReactRefresh bool
	assetPrefix      string
}

// Option is a functional option for configuring the Vite bundler.
type Option func(*config)

// WithDevMode enables development mode.
// In dev mode, assets are loaded from the Vite dev server.
func WithDevMode(isDev bool) Option {
	return func(c *config) {
		c.isDev = isDev
	}
}

// WithURL sets the Vite dev server URL.
// Default: "http://localhost:5173"
func WithURL(url string) Option {
	return func(c *config) {
		c.viteURL = url
	}
}

// WithReactRefresh enables React Refresh HMR support in dev mode.
func WithReactRefresh() Option {
	return func(c *config) {
		c.withReactRefresh = true
	}
}

// WithAssetPrefix sets the URL prefix for production assets.
// Default: "static" will result in URLs like "/static/app.js"
func WithAssetPrefix(prefix string) Option {
	return func(c *config) {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}

		c.assetPrefix = prefix
	}
}

// New creates a new Vite bundler.
// distFS is the filesystem containing the Vite build output (dist directory).
// It is required for production mode to load the manifest.
func New(distFS fs.FS, options ...Option) (*Bundler, error) {
	cfg := &config{
		viteURL:     "http://localhost:5173",
		assetPrefix: "/static/",
	}

	for _, opt := range options {
		opt(cfg)
	}

	v := &Bundler{
		isDev:            cfg.isDev,
		viteURL:          cfg.viteURL,
		withReactRefresh: cfg.withReactRefresh,
		distFS:           distFS,
		assetPrefix:      cfg.assetPrefix,
	}

	// In production mode, load the manifest
	if !cfg.isDev && distFS != nil {
		manifestData, err := fs.ReadFile(distFS, ".vite/manifest.json")
		if err != nil {
			return nil, fmt.Errorf("failed to read vite manifest: %w", err)
		}

		if err := json.Unmarshal(manifestData, &v.manifest); err != nil {
			return nil, fmt.Errorf("failed to parse vite manifest: %w", err)
		}
	}

	return v, nil
}

func (v *Bundler) IsDev() bool { return v.isDev }

// TemplateFuncs returns template functions for use in HTML templates.
// It provides a "vite" function that generates script/link tags for assets.
func (v *Bundler) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"vite": v.viteTagsFunc,
	}
}

func (v *Bundler) viteTagsFunc(entry string) template.HTML {
	if v.isDev {
		return v.devTags(entry)
	}
	return v.prodTags(entry)
}

func (v *Bundler) devTags(entry string) template.HTML {
	var buf bytes.Buffer

	// Vite client for HMR
	fmt.Fprintf(&buf, `<script type="module" src="%s/@vite/client"></script>`+"\n", v.viteURL)

	// React Refresh preamble (must come before React)
	if v.withReactRefresh {
		fmt.Fprintf(&buf, `<script type="module">
import RefreshRuntime from '%s/@react-refresh'
RefreshRuntime.injectIntoGlobalHook(window)
window.$RefreshReg$ = () => {}
window.$RefreshSig$ = () => (type) => type
window.__vite_plugin_react_preamble_installed__ = true
</script>`+"\n", v.viteURL)
	}

	// Entry point
	fmt.Fprintf(&buf, `<script type="module" src="%s/%s"></script>`+"\n", v.viteURL, entry)

	return template.HTML(buf.String())
}

func (v *Bundler) prodTags(entry string) template.HTML {
	chunk, ok := v.manifest[entry]
	if !ok {
		return template.HTML(fmt.Sprintf("<!-- vite: entry %q not found in manifest -->", entry))
	}

	var buf bytes.Buffer

	// CSS files
	for _, cssFile := range chunk.CSS {
		fmt.Fprintf(&buf, `<link rel="stylesheet" href="%s%s">`+"\n", v.assetPrefix, cssFile)
	}

	// Preload imported chunks
	v.writePreloads(&buf, chunk.Imports, make(map[string]bool))

	// Main entry script
	fmt.Fprintf(&buf, `<script type="module" src="%s%s"></script>`+"\n", v.assetPrefix, chunk.File)

	return template.HTML(buf.String())
}

func (v *Bundler) writePreloads(buf *bytes.Buffer, imports []string, visited map[string]bool) {
	for _, importPath := range imports {
		if visited[importPath] {
			continue
		}
		visited[importPath] = true

		importedChunk, ok := v.manifest[importPath]
		if !ok {
			continue
		}

		// Preload the chunk's CSS
		for _, cssFile := range importedChunk.CSS {
			fmt.Fprintf(buf, `<link rel="stylesheet" href="%s%s">`+"\n", v.assetPrefix, cssFile)
		}

		// Preload the JS file
		fmt.Fprintf(buf, `<link rel="modulepreload" href="%s%s">`+"\n", v.assetPrefix, importedChunk.File)

		// Recursively preload dependencies
		v.writePreloads(buf, importedChunk.Imports, visited)
	}
}

// Handler returns an http.Handler that serves production assets from distFS.
func (v *Bundler) Handler() http.Handler {
	if v.isDev || v.distFS == nil {
		return http.NotFoundHandler()
	}
	return http.StripPrefix(v.assetPrefix, http.FileServerFS(v.distFS))
}

// AssetPrefix returns the configured asset prefix.
func (v *Bundler) AssetPrefix() string {
	return v.assetPrefix
}

// DevSSREngine implements inertia.BundlerDevSSR interface.
// It returns an SSR engine that calls the Vite dev server's /render endpoint.
func (v *Bundler) DevSSREngine() (inertia.SSREngine, error) {
	return &devSSREngine{viteURL: v.viteURL}, nil
}

type devSSREngine struct {
	viteURL string
}

func (e *devSSREngine) Render(page inertia.PageObject) (inertia.RenderedPage, error) {
	pageJSON, err := json.Marshal(page)
	if err != nil {
		return inertia.RenderedPage{}, err
	}

	resp, err := http.Post(e.viteURL+"/render", "application/json", bytes.NewBuffer(pageJSON))
	if err != nil {
		return inertia.RenderedPage{}, err
	}
	defer resp.Body.Close()

	var result inertia.RenderedPage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return inertia.RenderedPage{}, err
	}

	return result, nil
}

func (e *devSSREngine) Name() string { return "vite-dev-ssr" }
