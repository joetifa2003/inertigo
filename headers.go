package inertia

import (
	"net/http"
	"strings"
)

// Inertia protocol headers.
const (
	// XInertia indicates an Inertia request when set to "true".
	XInertia = "X-Inertia"
	// XInertiaVersion contains the client's asset version for cache busting.
	XInertiaVersion = "X-Inertia-Version"
	// XInertiaLocation is used in 409 responses to trigger a full page reload.
	XInertiaLocation = "X-Inertia-Location"
	// XInertiaPartialData specifies which props to include in a partial reload.
	XInertiaPartialData = "X-Inertia-Partial-Data"
	// XInertiaPartialComponent specifies the component for partial reloads.
	XInertiaPartialComponent = "X-Inertia-Partial-Component"
	// XInertiaPartialExcept specifies which props to exclude in a partial reload.
	XInertiaPartialExcept = "X-Inertia-Partial-Except"
	// XInertiaErrorBag specifies the error bag name for scoped validation errors.
	XInertiaErrorBag = "X-Inertia-Error-Bag"
	// XInertiaInfiniteScrollMergeIntent specifies merge direction ("append" or "prepend").
	XInertiaInfiniteScrollMergeIntent = "X-Inertia-Infinite-Scroll-Merge-Intent"
	// XInertiaExceptOnceProps lists once props the client already has cached.
	XInertiaExceptOnceProps = "X-Inertia-Except-Once-Props"
	// XInertiaReset specifies scroll props to reset.
	XInertiaReset = "X-Inertia-Reset"
)

// Precognition headers for real-time validation.
const (
	// HeaderPrecognition indicates a precognition validation request.
	HeaderPrecognition = "Precognition"
	// HeaderPrecognitionValidateOnly specifies which fields to validate.
	HeaderPrecognitionValidateOnly = "Precognition-Validate-Only"
	// HeaderPrecognitionSuccess indicates successful validation.
	HeaderPrecognitionSuccess = "Precognition-Success"
)

type inertiaHeaders struct {
	Component           string
	PartialData         []string
	PartialExcept       []string
	ExceptOnceProps     []string
	ResetProps          []string // Props to reset from X-Inertia-Reset header
	InfiniteScrollMerge string   // "append" or "prepend"
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
	if resetProps := r.Header.Get(XInertiaReset); resetProps != "" {
		headers.ResetProps = strings.Split(resetProps, ",")
	}

	return headers
}
