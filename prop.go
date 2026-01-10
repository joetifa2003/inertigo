package inertia

import (
	"context"
	"slices"
)

// PropFunc is the resolver function signature for lazy-evaluated props.
type PropFunc func(ctx context.Context) (any, error)

// Prop is the interface that all prop types must implement.
type Prop interface {
	// shouldInclude determines if the prop should be included in the response
	// based on the request headers and partial reload state.
	shouldInclude(key string, headers *inertiaHeaders) bool

	// resolve executes the prop's resolver or returns its static value.
	resolve(ctx context.Context) (any, error)

	// modifyProcessedProps allows the prop to modify the processed props
	// (e.g., adding to deferredProps, onceProps, scrollProps, etc.)
	modifyProcessedProps(key string, headers *inertiaHeaders, p *processedProps)
}

// Props is a map of prop names to values.
type Props map[string]Prop

// oncePropData is the JSON structure for once props in response.
type oncePropData struct {
	Prop      string `json:"prop"`
	ExpiresAt string `json:"expiresAt"`
}

// defaultShouldInclude is the shared logic for default/lazy/value/scroll props.
// It includes the prop on full load, and on partial reload only if requested
// (or if no specific props are requested) and not excluded.
func defaultShouldInclude(key string, headers *inertiaHeaders) bool {
	if headers.IsPartial {
		includedByData := len(headers.PartialData) == 0
		if len(headers.PartialData) > 0 {
			includedByData = slices.Contains(headers.PartialData, key)
		}

		excluded := false
		if len(headers.PartialExcept) > 0 {
			excluded = slices.Contains(headers.PartialExcept, key)
		}

		return includedByData && !excluded
	}
	return true
}
