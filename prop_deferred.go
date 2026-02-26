package inertia

import (
	"context"
	"slices"
)

// DeferredProp is a prop that is excluded on initial page load.
// The frontend will automatically request it via a partial reload.
type DeferredProp[T any] struct {
	resolver Resolver[T]
	group    string
}

// Deferred creates a prop that is excluded on initial page load.
// The frontend will automatically request it via a partial reload.
func Deferred[T any](resolver Resolver[T]) DeferredProp[T] {
	return DeferredProp[T]{resolver: resolver}
}

// DeferredGroup creates a deferred prop with a specific group.
// Props in the same group are fetched together in a single partial reload.
func DeferredGroup[T any](group string, resolver Resolver[T]) DeferredProp[T] {
	return DeferredProp[T]{resolver: resolver, group: group}
}

func (p DeferredProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	// Only include if explicitly requested in partial reload
	if headers.IsPartial && len(headers.PartialData) > 0 {
		return slices.Contains(headers.PartialData, key)
	}
	return false
}

func (p DeferredProp[T]) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p DeferredProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	if headers.IsPartial {
		return
	}

	group := p.group
	if group == "" {
		group = "default"
	}
	pp.deferredProps[group] = append(pp.deferredProps[group], key)
}
