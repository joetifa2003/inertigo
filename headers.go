package inertia

import (
	"net/http"
	"strings"
)

const (
	XInertia                          = "X-Inertia"
	XInertiaVersion                   = "X-Inertia-Version"
	XInertiaLocation                  = "X-Inertia-Location"
	XInertiaPartialData               = "X-Inertia-Partial-Data"
	XInertiaPartialComponent          = "X-Inertia-Partial-Component"
	XInertiaPartialExcept             = "X-Inertia-Partial-Except"
	XInertiaErrorBag                  = "X-Inertia-Error-Bag"
	XInertiaInfiniteScrollMergeIntent = "X-Inertia-Infinite-Scroll-Merge-Intent"
	XInertiaExceptOnceProps           = "X-Inertia-Except-Once-Props"
	HeaderPrecognition                = "Precognition"
	HeaderPrecognitionValidateOnly    = "Precognition-Validate-Only"
	HeaderPrecognitionSuccess         = "Precognition-Success"
)

type inertiaHeaders struct {
	Component           string
	PartialData         []string
	PartialExcept       []string
	ExceptOnceProps     []string
	InfiniteScrollMerge string // "append" or "prepend"
	IsPartial           bool
	IsInertia           bool
}

func parseInertiaHeaders(r *http.Request, component string) *inertiaHeaders {
	headers := &inertiaHeaders{
		Component:           r.Header.Get(XInertiaPartialComponent),
		IsInertia:           r.Header.Get(XInertia) == "true",
		InfiniteScrollMerge: r.Header.Get(XInertiaInfiniteScrollMergeIntent),
	}

	headers.IsPartial = headers.IsInertia && headers.Component == component

	if partialData := r.Header.Get(XInertiaPartialData); partialData != "" {
		headers.PartialData = strings.Split(partialData, ",")
	}
	if partialExcept := r.Header.Get(XInertiaPartialExcept); partialExcept != "" {
		headers.PartialExcept = strings.Split(partialExcept, ",")
	}
	if exceptOnce := r.Header.Get(XInertiaExceptOnceProps); exceptOnce != "" {
		headers.ExceptOnceProps = strings.Split(exceptOnce, ",")
	}

	return headers
}
