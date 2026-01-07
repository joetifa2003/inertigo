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
	"strings"
	"sync"
	"time"

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
}

type inertiaConfig struct {
	rootTemplateFS   fs.FS
	rootTemplatePath string

	ssrEnabled       bool
	ssrEngineFactory func() (SSREngine, error)

	logger  Logger
	version string
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

const (
	XInertia = "X-Inertia"
)

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
	finalProps    Props
	deferredProps map[string][]string
	onceProps     map[string]onceProp
}

func newProcessedProps() processedProps {
	return processedProps{
		finalProps:    make(Props),
		deferredProps: make(map[string][]string),
		onceProps:     make(map[string]onceProp),
	}
}

var processedPropsPool = pool.NewPool(newProcessedProps, pool.WithPoolBeforeGet[processedProps](func(p processedProps) {
	clear(p.finalProps)
	clear(p.deferredProps)
	clear(p.onceProps)
}))

func (i *Inertia) processProps(ctx context.Context, props Props, headers *inertiaHeaders) (processedProps, error) {
	p := processedPropsPool.Get()

	if _, ok := props["errors"]; !ok {
		props["errors"] = Always(json.RawMessage(`{}`))
	}

	for key, value := range props {
		var prop Prop
		if p, ok := value.(Prop); ok {
			prop = p
		} else {
			prop = Prop{Type: PropTypeDefault, Value: value}
		}

		shouldInclude := prop.ShouldInclude(key, headers)

		if shouldInclude {
			val, err := prop.Resolve(ctx)
			if err != nil {
				return p, err
			}
			p.finalProps[key] = val
		} else {
			switch prop.Type {
			case PropTypeDeferred:
				group := "default"
				if prop.Group != "" {
					group = prop.Group
				}
				p.deferredProps[group] = append(p.deferredProps[group], key)
			}
		}

		if prop.Type == PropTypeOnce {
			op := onceProp{Prop: key}
			if prop.ExpiresAt != nil {
				op.ExpiresAt = *prop.ExpiresAt
			}
			p.onceProps[key] = op
		}
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

	return i.rootTemplate.Execute(w, view)
}

type PageObject struct {
	Component      string              `json:"component"`
	URL            string              `json:"url"`
	Props          Props               `json:"props"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
	MergeProps     []string            `json:"mergeProps"`
	PrependProps   []string            `json:"prependProps"`
	DeepMergeProps []string            `json:"deepMergeProps"`
	MatchPropsOn   []string            `json:"matchPropsOn"`
	ScrollProps    Props               `json:"scrollProps"`
	DeferredProps  map[string][]string `json:"deferredProps"`
	OnceProps      map[string]onceProp `json:"onceProps"`
}

func (i *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props Props) error {
	if props == nil {
		props = Props{}
	}

	headers := parseInertiaHeaders(r, component)
	p, err := i.processProps(r.Context(), props, headers)
	defer processedPropsPool.Put(p)
	if err != nil {
		return err
	}

	pageObject := &PageObject{
		Component:     component,
		URL:           r.URL.Path,
		Props:         p.finalProps,
		Version:       i.version,
		DeferredProps: p.deferredProps,
		OnceProps:     p.onceProps,
	}

	if headers.IsInertia {
		return i.renderJSON(w, r, pageObject)
	}

	return i.renderHTML(w, r, pageObject)
}
