package inertia

import "context"

type errorsProp struct {
	value map[string]any
}

func Errors(value map[string]any) Prop {
	return errorsProp{value: value}
}

func (p errorsProp) ShouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p errorsProp) Resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p errorsProp) ModifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
