package inertia

import "html/template"

// Bundler is the interface that asset bundlers must implement.
// It provides template functions for generating script/link tags and
// indicates whether the bundler is running in development mode.
type Bundler interface {
	// TemplateFuncs returns template functions for use in HTML templates.
	TemplateFuncs() template.FuncMap
	// IsDev returns true if the bundler is in development mode.
	IsDev() bool
}

// BundlerDevSSR extends Bundler with development-mode SSR capabilities.
// Bundlers that support SSR in development mode should implement this interface.
type BundlerDevSSR interface {
	Bundler
	// DevSSREngine returns an SSR engine for development mode.
	DevSSREngine() (SSREngine, error)
}
