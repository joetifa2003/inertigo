package inertia

import (
	"context"
	"slices"
)

// OptionalProp is excluded by default and only included when explicitly requested.
type OptionalProp[T any] struct {
	resolver Resolver[T]
}

// Optional creates a prop that is excluded by default.
// It is only included when explicitly requested in a partial reload.
func Optional[T any](resolver Resolver[T]) OptionalProp[T] {
	return OptionalProp[T]{resolver: resolver}
}

func (p OptionalProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	// Only include if explicitly requested in a partial reload
	if headers.IsPartial && len(headers.PartialData) > 0 {
		return slices.Contains(headers.PartialData, key)
	}
	return false
}

func (p OptionalProp[T]) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p OptionalProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
}
