package props

import (
	"context"
	"slices"
)

// Optional is excluded by default and only included when explicitly requested.
type Optional[T any] struct {
	resolver Resolver[T]
}

// NewOptional creates a prop that is excluded by default.
// It is only included when explicitly requested in a partial reload.
func NewOptional[T any](resolver Resolver[T]) Optional[T] {
	return Optional[T]{resolver: resolver}
}

func (p Optional[T]) ShouldInclude(key string, headers *Headers) bool {
	if headers.IsPartial && len(headers.PartialData) > 0 {
		return slices.Contains(headers.PartialData, key)
	}
	return false
}

func (p Optional[T]) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p Optional[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {}
