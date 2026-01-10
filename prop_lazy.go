package inertia

import "context"

// lazyProp is a prop with lazy evaluation that is always included on full load.
// This is equivalent to Laravel's closure props: 'users' => fn () => User::all()
type lazyProp struct {
	resolver PropFunc
}

// Lazy creates a prop with lazy evaluation. The resolver is called
// during rendering to produce the value. Unlike Optional or Deferred,
// Lazy props are always included in the response (on full load).
func Lazy(resolver PropFunc) Prop {
	return &lazyProp{resolver: resolver}
}

func (p lazyProp) ShouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p lazyProp) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p lazyProp) ModifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
