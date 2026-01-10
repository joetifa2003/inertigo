package inertia

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// onceProp is included once and then excluded on subsequent page visits.
type onceProp struct {
	resolver  PropFunc
	expiresAt *string
}

// Once creates a prop that is included on the first visit and then
// excluded on subsequent visits to the same page.
func Once(resolver PropFunc) Prop {
	return onceProp{resolver: resolver}
}

// OnceWithExpiration creates a once prop that expires after the given duration.
// After expiration, the prop will be included again.
func OnceWithExpiration(resolver PropFunc, expiration time.Duration) Prop {
	expiresAt := fmt.Sprintf("%d", time.Now().Add(expiration).UnixMilli())
	return &onceProp{resolver: resolver, expiresAt: &expiresAt}
}

func (p onceProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	if len(headers.ExceptOnceProps) > 0 && slices.Contains(headers.ExceptOnceProps, key) {
		return false
	}
	return true
}

func (p onceProp) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p onceProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	data := oncePropData{Prop: key}
	if p.expiresAt != nil {
		data.ExpiresAt = *p.expiresAt
	}
	pp.onceProps[key] = data
}
