package inertia

// RenderedPage represents the result of server-side rendering.
type RenderedPage struct {
	// Head contains HTML strings to be injected into the <head> section.
	Head []string `json:"head"`
	// Body contains the rendered HTML for the page body.
	Body string `json:"body"`
}

// SSREngine is the interface for server-side rendering engines.
// Implementations can use QuickJS, Node.js, Deno, or any JavaScript runtime.
type SSREngine interface {
	// Name returns a human-readable name for the SSR engine.
	Name() string
	// Render renders a PageObject and returns the rendered HTML.
	Render(pageObject PageObject) (RenderedPage, error)
}
