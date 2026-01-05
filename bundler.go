package inertia

import "html/template"

type Bundler interface {
	TemplateFuncs() template.FuncMap
	IsDev() bool
}

type BundlerDevSSR interface {
	Bundler
	// DevSSREngine returns an SSR engine for development mode.
	// For Vite, this calls the Vite dev server's /render endpoint.
	DevSSREngine() (SSREngine, error)
}
