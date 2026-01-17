package inertia

import "context"

// valueProp wraps raw values (strings, ints, structs, etc.) as Props.
type valueProp struct {
	value any
}

// Value wraps a static value as a Prop.
// Use this for props that don't require lazy evaluation or special behavior.
func Value(value any) Prop {
	return valueProp{value: value}
}

func (p valueProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p valueProp) resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p valueProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
