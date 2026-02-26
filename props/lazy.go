package props

import "context"

// Lazy is a prop with lazy evaluation that is always included on full load.
type Lazy[T any] struct {
	resolver Resolver[T]
}

// NewLazy creates a prop with lazy evaluation. The resolver is called
// during rendering to produce the value. Unlike Optional or Deferred,
// Lazy props are always included in the response (on full load).
func NewLazy[T any](resolver Resolver[T]) Lazy[T] {
	return Lazy[T]{resolver: resolver}
}

func (p Lazy[T]) ShouldInclude(key string, headers *Headers) bool {
	return DefaultShouldInclude(key, headers)
}

func (p Lazy[T]) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p Lazy[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {}
