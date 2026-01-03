package inertia

type RenderedPage struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

type SSREngine interface {
	Render(pageObject PageObject) (RenderedPage, error)
}
