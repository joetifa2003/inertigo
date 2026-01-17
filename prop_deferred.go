package inertia

import (
	"context"
	"slices"
)

// deferredProp is excluded on initial load and fetched later via partial reload.
type deferredProp struct {
	resolver PropFunc
	group    string
}

// Deferred creates a prop that is excluded on initial page load.
// The frontend will automatically request it via a partial reload.
func Deferred(resolver PropFunc) Prop {
	return deferredProp{resolver: resolver}
}

// DeferredGroup creates a deferred prop with a specific group.
// Props in the same group are fetched together in a single partial reload.
func DeferredGroup(group string, resolver PropFunc) Prop {
	return deferredProp{resolver: resolver, group: group}
}

func (p deferredProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	// Only include if explicitly requested in partial reload
	if headers.IsPartial && len(headers.PartialData) > 0 {
		return slices.Contains(headers.PartialData, key)
	}
	return false
}

func (p deferredProp) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p deferredProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	if p.shouldInclude(key, headers) {
		return
	}
	group := p.group
	if group == "" {
		group = "default"
	}
	pp.deferredProps[group] = append(pp.deferredProps[group], key)
}
