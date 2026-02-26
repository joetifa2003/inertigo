package inertia

import "context"

// valueProp wraps raw values (strings, ints, structs, etc.) as props.
// This is used internally to wrap plain struct fields.
type valueProp struct {
	value any
}

func (p valueProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p valueProp) resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p valueProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
