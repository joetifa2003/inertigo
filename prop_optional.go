package inertia

import (
	"context"
	"slices"
)

// optionalProp is excluded by default and only included when explicitly requested.
type optionalProp struct {
	resolver PropFunc
}

// Optional creates a prop that is excluded by default.
// It is only included when explicitly requested in a partial reload.
func Optional(resolver PropFunc) Prop {
	return optionalProp{resolver: resolver}
}

func (p optionalProp) ShouldInclude(key string, headers *inertiaHeaders) bool {
	if headers.IsPartial && len(headers.PartialData) > 0 {
		if slices.Contains(headers.PartialData, key) {
			return true
		}
	}
	if len(headers.PartialExcept) > 0 && slices.Contains(headers.PartialExcept, key) {
		return false
	}
	return false
}

func (p optionalProp) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p optionalProp) ModifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {}
