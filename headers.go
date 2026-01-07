package inertia

import (
	"net/http"
	"strings"
)

type inertiaHeaders struct {
	Component           string
	Version             string
	PartialData         []string
	PartialExcept       []string
	ExceptOnceProps     []string
	InfiniteScrollMerge string // "append" or "prepend"
	IsPartial           bool
	IsInertia           bool
}

func parseInertiaHeaders(r *http.Request, component string) *inertiaHeaders {
	headers := &inertiaHeaders{
		Component:           r.Header.Get("X-Inertia-Partial-Component"),
		Version:             r.Header.Get("X-Inertia-Version"),
		IsInertia:           r.Header.Get(XInertia) == "true",
		InfiniteScrollMerge: r.Header.Get("X-Inertia-Infinite-Scroll-Merge-Intent"),
	}

	headers.IsPartial = headers.IsInertia && headers.Component == component

	if partialData := r.Header.Get("X-Inertia-Partial-Data"); partialData != "" {
		headers.PartialData = strings.Split(partialData, ",")
	}
	if partialExcept := r.Header.Get("X-Inertia-Partial-Except"); partialExcept != "" {
		headers.PartialExcept = strings.Split(partialExcept, ",")
	}
	if exceptOnce := r.Header.Get("X-Inertia-Except-Once-Props"); exceptOnce != "" {
		headers.ExceptOnceProps = strings.Split(exceptOnce, ",")
	}

	return headers
}
