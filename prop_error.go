package inertia

import "context"

type errorsProp struct {
	value map[string]any
}

func Errors(value map[string]any) Prop {
	return errorsProp{value: value}
}

func (p errorsProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p errorsProp) resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p errorsProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
