package props

import "context"

// Value wraps raw values (strings, ints, structs, etc.) as Props.
// This is typically used internally to wrap plain struct fields,
// but can be used directly for static values.
type Value[T any] struct {
	Val T
}

func (p Value[T]) ShouldInclude(key string, headers *Headers) bool {
	return DefaultShouldInclude(key, headers)
}

func (p Value[T]) Resolve(ctx context.Context) (any, error) {
	return p.Val, nil
}

func (p Value[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {}
