package inertia

import "context"

// AlwaysProp is always included in the response, even on partial reloads.
type AlwaysProp[T any] struct {
	value T
}

// Always creates a prop that is always included in the response,
// even when not explicitly requested in a partial reload.
func Always[T any](value T) AlwaysProp[T] {
	return AlwaysProp[T]{value: value}
}

func (p AlwaysProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	return true
}

func (p AlwaysProp[T]) resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p AlwaysProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
}
