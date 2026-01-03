package inertia

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/olivere/vite"
)

type Inertia struct {
	logger Logger

	rootTemplate *template.Template
	vite         *vite.Fragment

	viteURL string
	isDev   bool

	ssrEnabled       bool
	ssrEngineFactory func() (SSREngine, error)
	ssrEngine        SSREngine
	ssrInitLock      sync.Mutex
}

type inertiaConfig struct {
	viteURL          string
	isDev            bool
	entryPoint       string
	viteSource       fs.FS
	viteDist         fs.FS
	withReactRefresh bool
	rootTemplate     *template.Template

	ssrEnabled       bool
	ssrEngineFactory func() (SSREngine, error)

	logger Logger
}

type InertiaOption func(config *inertiaConfig) error

func WithViteURL(url string) InertiaOption {
	return func(config *inertiaConfig) error {
		config.viteURL = url

		return nil
	}
}

func WithDevMode(isDev bool) InertiaOption {
	return func(config *inertiaConfig) error {
		config.isDev = isDev

		return nil
	}
}

func WithEntryPoint(entryPoint string) InertiaOption {
	return func(config *inertiaConfig) error {
		config.entryPoint = entryPoint

		return nil
	}
}

func WithViteFS(source fs.FS) InertiaOption {
	return func(config *inertiaConfig) error {
		config.viteSource = source

		return nil
	}
}

func WithViteDistFS(dist fs.FS) InertiaOption {
	return func(config *inertiaConfig) error {
		config.viteDist = dist

		return nil
	}
}

func WithReactRefresh() InertiaOption {
	return func(config *inertiaConfig) error {
		config.withReactRefresh = true

		return nil
	}
}

func WithRootHtmlPath(root string) InertiaOption {
	return func(config *inertiaConfig) error {
		var err error

		config.rootTemplate, err = template.New("index.html").ParseFiles(root)
		if err != nil {
			return err
		}

		return nil
	}
}

func WithRooHtmlPathFS(fs fs.FS, path string) InertiaOption {
	return func(config *inertiaConfig) error {
		var err error

		config.rootTemplate, err = template.New("index.html").ParseFS(fs, path)
		if err != nil {
			return err
		}

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

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

func New(options ...InertiaOption) (*Inertia, error) {
	var err error

	config := &inertiaConfig{
		viteURL: "http://localhost:5173",
	}

	for _, option := range options {
		err = option(config)
		if err != nil {
			return nil, err
		}
	}

	viteConfig := vite.Config{
		FS:              config.viteDist,
		IsDev:           config.isDev,
		ViteEntry:       config.entryPoint,
		ViteURL:         config.viteURL,
		ViteTemplate:    vite.None,
		AssetsURLPrefix: "foo",
	}

	if config.isDev {
		viteConfig.FS = config.viteSource
	}
	if config.withReactRefresh {
		viteConfig.ViteTemplate = vite.React
	}

	i := Inertia{
		rootTemplate:     config.rootTemplate,
		ssrEnabled:       config.ssrEnabled,
		ssrEngineFactory: config.ssrEngineFactory,
		isDev:            config.isDev,
		viteURL:          config.viteURL,
		logger:           config.logger,
	}
	i.vite, err = vite.HTMLFragment(viteConfig)
	if err != nil {
		return nil, err
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

type Props map[string]any

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

type onceProp struct {
	Prop      string `json:"prop"`
	ExpiresAt string `json:"expiresAt"`
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
			if i.isDev {
				engine, err := newViteSSREngine(i.viteURL)
				if err != nil {
					return nil, err
				}
				i.ssrEngine = engine
				i.logger.LogAttrs(
					ctx, slog.LevelInfo,
					"dev mode enabled, starting vite ssr engine",
					slog.String("engine", "vite"),
					slog.String("url", i.viteURL),
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
					slog.String("engine", "custom"),
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

func (i *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props Props) error {
	if props == nil {
		props = Props{}
	}

	if _, ok := props["error"]; !ok {
		props["errors"] = json.RawMessage(`{}`)
	}

	pageObject := PageObject{
		Component: component,
		URL:       r.URL.Path,
		Props:     props,
	}

	if r.Header.Get(XInertia) == "true" {
		i.logger.LogAttrs(r.Context(),
			slog.LevelDebug, "inertia request detected, rendering json",
			slog.String("component", component),
			slog.String("url", r.URL.Path),
		)

		w.Header().Set(XInertia, "true")
		w.Header().Set("Vary", XInertia)
		return json.NewEncoder(w).Encode(pageObject)
	}

	i.logger.LogAttrs(
		r.Context(), slog.LevelDebug, "rendering full page",
		slog.String("component", component),
		slog.String("url", r.URL.Path),
	)

	pageObjectJSON, err := json.Marshal(pageObject)
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

		renderedPage, err := engine.Render(pageObject)
		if err != nil {
			return err
		}
		i.logger.LogAttrs(
			r.Context(), slog.LevelInfo,
			"ssr engine rendered page",
			slog.String("dur", time.Since(t1).String()),
		)

		head = append([]string{string(i.vite.Tags)}, renderedPage.Head...)
		body = renderedPage.Body
	} else {
		inertiaBodyBuf := bytes.NewBuffer(nil)
		err = inertiaBodyTemplate.Execute(inertiaBodyBuf, string(pageObjectJSON))
		if err != nil {
			return err
		}
		head = []string{string(i.vite.Tags)}
		body = inertiaBodyBuf.String()
	}

	view := RootHtmlView{
		InertiaHead: template.HTML(strings.Join(head, "\n")),
		InertiaBody: template.HTML(body),
	}

	return i.rootTemplate.Execute(w, view)
}
