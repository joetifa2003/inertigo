package inertia

import "context"

// alwaysProp is always included in the response, even on partial reloads.
type alwaysProp struct {
	value any
}

// Always creates a prop that is always included in the response,
// even when not explicitly requested in a partial reload.
func Always(value any) Prop {
	return alwaysProp{value: value}
}

func (p alwaysProp) ShouldInclude(key string, headers *inertiaHeaders) bool {
	return true
}

func (p alwaysProp) Resolve(ctx context.Context) (any, error) {
	return p.value, nil
}

func (p alwaysProp) ModifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
