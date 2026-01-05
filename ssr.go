package inertia

type RenderedPage struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

type SSREngine interface {
	Name() string
	Render(pageObject PageObject) (RenderedPage, error)
}
