package inertia

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/peterbourgon/mergemap"

	"github.com/joetifa2003/inertigo/internal/pool"
)

type Inertia struct {
	logger  Logger
	version string

	rootTemplate *template.Template
	bundler      Bundler

	ssrEnabled       bool
	ssrEngineFactory func() (SSREngine, error)
	ssrEngine        SSREngine
	ssrInitLock      sync.Mutex

	session Session

	csrfEnabled bool
	csrfConfig  csrfConfig
}

type inertiaConfig struct {
	rootTemplateFS   fs.FS
	rootTemplatePath string

	ssrEnabled       bool
	ssrEngineFactory func() (SSREngine, error)

	logger  Logger
	version string

	session Session

	csrfEnabled bool
	csrfConfig  csrfConfig
}

type InertiaOption func(config *inertiaConfig) error

func WithRootHtmlPath(root string) InertiaOption {
	return func(config *inertiaConfig) error {
		config.rootTemplatePath = root
		return nil
	}
}

func WithRooHtmlPathFS(fsys fs.FS, path string) InertiaOption {
	return func(config *inertiaConfig) error {
		config.rootTemplateFS = fsys
		config.rootTemplatePath = path
		return nil
	}
}

func WithSSR(enabled bool, engineFactory func() (SSREngine, error)) InertiaOption {
	return func(config *inertiaConfig) error {
		config.ssrEnabled = true
		config.ssrEngineFactory = engineFactory
		return nil
	}
}

func WithLogger(logger Logger) InertiaOption {
	return func(config *inertiaConfig) error {
		config.logger = logger
		return nil
	}
}

// WithVersion sets a static version string for asset versioning.
func WithVersion(version string) InertiaOption {
	return func(config *inertiaConfig) error {
		config.version = version
		return nil
	}
}

// WithVersionFromFile computes version from file checksum (MD5 hash).
func WithVersionFromFile(path string) InertiaOption {
	return func(config *inertiaConfig) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read version file: %w", err)
		}
		hash := md5.Sum(data)
		config.version = hex.EncodeToString(hash[:])
		return nil
	}
}

// WithVersionFromFileFS computes version from file checksum using fs.FS.
func WithVersionFromFileFS(fsys fs.FS, path string) InertiaOption {
	return func(config *inertiaConfig) error {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read version file: %w", err)
		}
		hash := md5.Sum(data)
		config.version = hex.EncodeToString(hash[:])
		return nil
	}
}

// WithSession sets a custom session implementation for flash data.
// If not set, a default in-memory session is used.
func WithSession(session Session) InertiaOption {
	return func(config *inertiaConfig) error {
		config.session = session
		return nil
	}
}

// WithCSRF enables CSRF protection.
// enabled: whether to enable CSRF protection
// cookieSecure: set to true for HTTPS-only cookies (recommended for production)
func WithCSRF(enabled bool, cookieSecure bool) InertiaOption {
	return func(config *inertiaConfig) error {
		config.csrfEnabled = enabled
		if enabled {
			config.csrfConfig = csrfConfig{
				cookieSecure: cookieSecure,
			}
		}
		return nil
	}
}

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

// New creates a new Inertia instance.
// bundler is the asset bundler to use for resolving script/link tags.
func New(b Bundler, options ...InertiaOption) (*Inertia, error) {
	var err error

	config := &inertiaConfig{}

	for _, option := range options {
		err = option(config)
		if err != nil {
			return nil, err
		}
	}

	i := Inertia{
		bundler:          b,
		ssrEnabled:       config.ssrEnabled,
		ssrEngineFactory: config.ssrEngineFactory,
		logger:           config.logger,
		version:          config.version,
		session:          config.session,
		csrfEnabled:      config.csrfEnabled,
		csrfConfig:       config.csrfConfig,
	}

	if i.session == nil {
		i.session = NewMemorySession("sid")
	}

	// Parse root template with bundler's template functions
	if config.rootTemplatePath != "" {
		tmpl := template.New("index.html")
		tmpl = tmpl.Funcs(b.TemplateFuncs())

		if config.rootTemplateFS != nil {
			i.rootTemplate, err = tmpl.ParseFS(config.rootTemplateFS, config.rootTemplatePath)
		} else {
			i.rootTemplate, err = tmpl.ParseFiles(config.rootTemplatePath)
		}
		if err != nil {
			return nil, err
		}
	}

	if i.logger == nil {
		i.logger = slog.New(slog.DiscardHandler)
	}

	return &i, nil
}

type RootHtmlView struct {
	InertiaHead template.HTML
	InertiaBody template.HTML
}

var inertiaBodyTemplate = template.Must(template.New("inertiaBody").Parse(`<div id="app" data-page="{{ . }}"></div>`))

func (i *Inertia) getSSREngine(ctx context.Context) (SSREngine, error) {
	if i.ssrEngine == nil {
		i.ssrInitLock.Lock()
		defer i.ssrInitLock.Unlock()

		i.logger.LogAttrs(
			ctx, slog.LevelInfo,
			"starting ssr engine",
		)

		if i.ssrEngine == nil {
			// Check if bundler implements BundlerDevSSR for dev mode
			if devSSR, ok := i.bundler.(BundlerDevSSR); ok && devSSR.IsDev() {
				engine, err := devSSR.DevSSREngine()
				if err != nil {
					return nil, err
				}
				i.ssrEngine = engine
				i.logger.LogAttrs(
					ctx, slog.LevelInfo,
					"dev mode enabled, starting bundler ssr engine",
					slog.String("engine", engine.Name()),
				)
			} else {
				t1 := time.Now()
				engine, err := i.ssrEngineFactory()
				if err != nil {
					return nil, err
				}
				i.ssrEngine = engine
				i.logger.LogAttrs(
					ctx, slog.LevelInfo,
					"ssr engine started",
					slog.String("engine", engine.Name()),
					slog.String("dur", time.Since(t1).String()),
				)
			}
		}
	}

	return i.ssrEngine, nil
}

func (i *Inertia) Logger() Logger {
	return i.logger
}

type processedProps struct {
	finalProps    map[string]any
	deferredProps map[string][]string
	onceProps     map[string]oncePropData
	scrollProps   map[string]scrollPropMetadata
	mergeProps    []string // Prop paths to merge on navigation
	prependProps  []string // Prop paths to prepend on navigation
}

func newProcessedProps() processedProps {
	return processedProps{
		finalProps:    make(map[string]any),
		deferredProps: make(map[string][]string),
		onceProps:     make(map[string]oncePropData),
		scrollProps:   make(map[string]scrollPropMetadata),
	}
}

var processedPropsPool = pool.NewPool(newProcessedProps, pool.WithPoolBeforeGet[processedProps](func(p processedProps) {
	clear(p.finalProps)
	clear(p.deferredProps)
	clear(p.onceProps)
	clear(p.scrollProps)
	p.mergeProps = p.mergeProps[:0]
	p.prependProps = p.prependProps[:0]
}))

var emptyJsonObject = json.RawMessage("{}")

func (i *Inertia) processProps(ctx context.Context, props Props, headers *inertiaHeaders) (processedProps, error) {
	p := processedPropsPool.Get()

	var existingErrorsMap map[string]any

	// if existingErrorProp prop is not an errorsProp, ignore it
	switch existingErrorProp := props["errors"].(type) {
	case errorsProp:
		existingErrorsMap = existingErrorProp.value
	}

	// merge flashed validation errors from context (set by Middleware)
	if errorsFromFlash := ctx.Value(inertiaErrorsKey); errorsFromFlash != nil {
		switch errorsFromFlash := errorsFromFlash.(type) {
		case map[string]any:
			if existingErrorsMap == nil {
				existingErrorsMap = errorsFromFlash
			} else {
				mergemap.Merge(existingErrorsMap, errorsFromFlash)
			}
		}
	}

	// if we have any error, wrap it back in an errorsProp
	if existingErrorsMap != nil {
		props["errors"] = Errors(existingErrorsMap)
	}

	for key, prop := range props {
		if prop.shouldInclude(key, headers) {
			resolved, err := prop.resolve(ctx)
			if err != nil {
				return p, err
			}
			p.finalProps[key] = resolved
		}

		prop.modifyProcessedProps(key, headers, &p)
	}

	// if there were no errors at all, set it to an empty object
	if p.finalProps["errors"] == nil {
		p.finalProps["errors"] = emptyJsonObject
	}

	return p, nil
}

func (i *Inertia) renderJSON(w http.ResponseWriter, r *http.Request, page *PageObject) error {
	i.logger.LogAttrs(r.Context(),
		slog.LevelDebug, "inertia request detected, rendering json",
		slog.String("component", page.Component),
		slog.String("url", page.URL),
	)

	w.Header().Set(XInertia, "true")
	w.Header().Set("Vary", XInertia)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(page)
}

func (i *Inertia) renderHTML(w http.ResponseWriter, r *http.Request, page *PageObject) error {
	i.logger.LogAttrs(
		r.Context(), slog.LevelDebug, "rendering full page",
		slog.String("component", page.Component),
		slog.String("url", page.URL),
	)

	pageObjectJSON, err := json.Marshal(page)
	if err != nil {
		return err
	}

	var head []string
	var body string

	if i.ssrEnabled {
		t1 := time.Now()
		engine, err := i.getSSREngine(r.Context())
		if err != nil {
			return err
		}

		renderedPage, err := engine.Render(*page)
		if err != nil {
			return err
		}
		i.logger.LogAttrs(
			r.Context(), slog.LevelInfo,
			"ssr engine rendered page",
			slog.String("engine", engine.Name()),
			slog.String("dur", time.Since(t1).String()),
		)

		head = renderedPage.Head
		body = renderedPage.Body
	} else {
		inertiaBodyBuf := bytes.NewBuffer(nil)
		err = inertiaBodyTemplate.Execute(inertiaBodyBuf, string(pageObjectJSON))
		if err != nil {
			return err
		}
		body = inertiaBodyBuf.String()
	}

	view := RootHtmlView{
		InertiaHead: template.HTML(strings.Join(head, "\n")),
		InertiaBody: template.HTML(body),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return i.rootTemplate.Execute(w, view)
}

type PageObject struct {
	Component      string                        `json:"component"`
	URL            string                        `json:"url"`
	Props          map[string]any                `json:"props"`
	Version        string                        `json:"version"`
	EncryptHistory bool                          `json:"encryptHistory"`
	ClearHistory   bool                          `json:"clearHistory"`
	MergeProps     []string                      `json:"mergeProps"`
	PrependProps   []string                      `json:"prependProps"`
	DeepMergeProps []string                      `json:"deepMergeProps"`
	MatchPropsOn   []string                      `json:"matchPropsOn"`
	DeferredProps  map[string][]string           `json:"deferredProps"`
	OnceProps      map[string]oncePropData       `json:"onceProps"`
	ScrollProps    map[string]scrollPropMetadata `json:"scrollProps,omitempty"`
}

// scrollPropMetadata contains pagination metadata for infinite scrolling.
type scrollPropMetadata struct {
	PageName     string `json:"pageName"`
	PreviousPage any    `json:"previousPage"`
	NextPage     any    `json:"nextPage"`
	CurrentPage  any    `json:"currentPage"`
	Reset        bool   `json:"reset"`
}

// renderConfig holds per-render configuration options
type renderConfig struct {
	encryptHistory *bool
	clearHistory   *bool
	mergeProps     []string
	prependProps   []string
	deepMergeProps []string
	matchPropsOn   []string
}

// RenderOption configures the behavior of a single Render call
type RenderOption func(config *renderConfig)

// WithEncryptHistory sets whether to encrypt the page's history state
func WithEncryptHistory(encrypt bool) RenderOption {
	return func(config *renderConfig) {
		config.encryptHistory = &encrypt
	}
}

// WithClearHistory sets whether to clear encrypted history state
func WithClearHistory(clear bool) RenderOption {
	return func(config *renderConfig) {
		config.clearHistory = &clear
	}
}

// WithMergeProps specifies prop keys that should be appended during navigation
func WithMergeProps(props ...string) RenderOption {
	return func(config *renderConfig) {
		config.mergeProps = append(config.mergeProps, props...)
	}
}

// WithPrependProps specifies prop keys that should be prepended during navigation
func WithPrependProps(props ...string) RenderOption {
	return func(config *renderConfig) {
		config.prependProps = append(config.prependProps, props...)
	}
}

// WithDeepMergeProps specifies prop keys that should be deep merged during navigation
func WithDeepMergeProps(props ...string) RenderOption {
	return func(config *renderConfig) {
		config.deepMergeProps = append(config.deepMergeProps, props...)
	}
}

// WithMatchPropsOn specifies prop keys used for matching when merging
func WithMatchPropsOn(props ...string) RenderOption {
	return func(config *renderConfig) {
		config.matchPropsOn = append(config.matchPropsOn, props...)
	}
}

func (i *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props Props, options ...RenderOption) error {
	if props == nil {
		props = Props{}
	}

	config := &renderConfig{}
	for _, opt := range options {
		opt(config)
	}

	headers := parseInertiaHeaders(r, component)
	p, err := i.processProps(r.Context(), props, headers)
	defer processedPropsPool.Put(p)
	if err != nil {
		return err
	}

	pageObject := &PageObject{
		Component:      component,
		URL:            r.URL.Path,
		Props:          p.finalProps,
		Version:        i.version,
		EncryptHistory: config.encryptHistory != nil && *config.encryptHistory,
		ClearHistory:   config.clearHistory != nil && *config.clearHistory,
		MergeProps:     slices.Concat(config.mergeProps, p.mergeProps),
		PrependProps:   slices.Concat(config.prependProps, p.prependProps),
		DeepMergeProps: config.deepMergeProps,
		MatchPropsOn:   config.matchPropsOn,
		DeferredProps:  p.deferredProps,
		OnceProps:      p.onceProps,
		ScrollProps:    p.scrollProps,
	}

	if headers.IsInertia {
		return i.renderJSON(w, r, pageObject)
	}

	return i.renderHTML(w, r, pageObject)
}

// Redirect performs a server-side redirect.
// It automatically uses HTTP 303 (See Other) for PUT, PATCH, and DELETE requests
// to prevent double form submissions, and 302 (Found) for other methods.
func (i *Inertia) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	status := http.StatusFound
	if r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
		status = http.StatusSeeOther
	}
	http.Redirect(w, r, url, status)
}

// RedirectBack redirects the user back to the previous page using the Referer header.
// Falls back to "/" if Referer is not available.
// It automatically uses HTTP 303 (See Other) for PUT, PATCH, and DELETE requests
// to prevent double form submissions, and 302 (Found) for other methods.
func (i *Inertia) RedirectBack(w http.ResponseWriter, r *http.Request) {
	url := r.Header.Get("Referer")
	if url == "" {
		url = "/"
	}

	i.Redirect(w, r, url)
}

// RenderErrors handles validation errors for both Precognition and standard Inertia requests.
// For Precognition requests, it returns appropriate precognition responses.
// For standard requests with errors, it flashes the errors to session and redirects back.
func (i *Inertia) RenderErrors(w http.ResponseWriter, r *http.Request, errors map[string]any) error {
	if len(errors) == 0 {
		if IsPrecognition(r) {
			PrecognitionSuccess(w)
		}
		return nil
	}

	if bag := r.Header.Get(XInertiaErrorBag); bag != "" {
		errors = map[string]any{
			bag: errors,
		}
	}

	if IsPrecognition(r) {
		return PrecognitionError(w, errors)
	}

	if err := i.session.Flash(w, r, "errors", errors); err != nil {
		return err
	}
	i.RedirectBack(w, r)

	return nil
}
