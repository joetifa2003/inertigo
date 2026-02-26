package pages

import "github.com/joetifa2003/inertigo/props"

type Index struct {
	Message string                   `json:"message"`
	Reviews props.Deferred[[]string] `json:"reviews"`
}
