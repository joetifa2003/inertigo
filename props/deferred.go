package props

import (
	"context"
	"slices"
)

// Deferred is a prop that is excluded on initial page load.
// The frontend will automatically request it via a partial reload.
type Deferred[T any] struct {
	resolver Resolver[T]
	group    string
}

// NewDeferred creates a prop that is excluded on initial page load.
// The frontend will automatically request it via a partial reload.
func NewDeferred[T any](resolver Resolver[T]) Deferred[T] {
	return Deferred[T]{resolver: resolver}
}

// NewDeferredGroup creates a deferred prop with a specific group.
// Props in the same group are fetched together in a single partial reload.
func NewDeferredGroup[T any](group string, resolver Resolver[T]) Deferred[T] {
	return Deferred[T]{resolver: resolver, group: group}
}

func (p Deferred[T]) ShouldInclude(key string, headers *Headers) bool {
	if headers.IsPartial && len(headers.PartialData) > 0 {
		return slices.Contains(headers.PartialData, key)
	}
	return false
}

func (p Deferred[T]) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p Deferred[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {
	if headers.IsPartial {
		return
	}

	group := p.group
	if group == "" {
		group = "default"
	}
	pp.DeferredProps[group] = append(pp.DeferredProps[group], key)
}
