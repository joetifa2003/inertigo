package props

import "context"

// Always is always included in the response, even on partial reloads.
type Always[T any] struct {
	value T
}

// NewAlways creates a prop that is always included in the response,
// even when not explicitly requested in a partial reload.
func NewAlways[T any](value T) Always[T] {
	return Always[T]{value: value}
}

func (p Always[T]) ShouldInclude(key string, headers *Headers) bool {
	return true
}

func (p Always[T]) Resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p Always[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {}
