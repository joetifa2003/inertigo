package inertia

import "context"

// LazyProp is a prop with lazy evaluation that is always included on full load.
// This is equivalent to Laravel's closure props: 'users' => fn () => User::all()
type LazyProp[T any] struct {
	resolver Resolver[T]
}

// Lazy creates a prop with lazy evaluation. The resolver is called
// during rendering to produce the value. Unlike Optional or Deferred,
// Lazy props are always included in the response (on full load).
func Lazy[T any](resolver Resolver[T]) LazyProp[T] {
	return LazyProp[T]{resolver: resolver}
}

func (p LazyProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p LazyProp[T]) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p LazyProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
