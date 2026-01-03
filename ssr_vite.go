package inertia

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type viteSSREngine struct {
	viteUrl string
}

func newViteSSREngine(viteUrl string) (SSREngine, error) {
	return &viteSSREngine{
		viteUrl: viteUrl,
	}, nil
}

func (v *viteSSREngine) Render(page PageObject) (RenderedPage, error) {
	pageJson, err := json.Marshal(page)
	if err != nil {
		return RenderedPage{}, err
	}

	resp, err := http.Post(v.viteUrl+"/render", "application/json", bytes.NewBuffer(pageJson))
	if err != nil {
		return RenderedPage{}, err
	}
	defer resp.Body.Close()

	var result RenderedPage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return RenderedPage{}, err
	}

	return result, nil
}
