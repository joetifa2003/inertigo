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
	fresh     bool   // Force refresh even if client has cached
	alias     string // Custom key for client-side caching
}

// OnceOption configures Once prop behavior
type OnceOption func(*onceProp)

// Fresh forces the prop to be resolved even if the client has it cached.
func Fresh() OnceOption {
	return func(p *onceProp) { p.fresh = true }
}

// FreshWhen conditionally forces the prop to be resolved.
func FreshWhen(condition bool) OnceOption {
	return func(p *onceProp) { p.fresh = condition }
}

// OnceAs assigns a custom key for client-side caching.
// This allows sharing data across multiple pages while using different prop names.
func OnceAs(alias string) OnceOption {
	return func(p *onceProp) { p.alias = alias }
}

// OnceUntil sets an expiration time for the once prop.
// After expiration, the prop will be resolved again on subsequent visits.
func OnceUntil(duration time.Duration) OnceOption {
	return func(p *onceProp) {
		expiresAt := fmt.Sprintf("%d", time.Now().Add(duration).UnixMilli())
		p.expiresAt = &expiresAt
	}
}

// Once creates a prop that is included on the first visit and then
// excluded on subsequent visits to the same page.
func Once(resolver PropFunc, opts ...OnceOption) Prop {
	p := &onceProp{resolver: resolver}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *onceProp) shouldInclude(key string, headers *inertiaHeaders) bool {
	// Always include if fresh is set
	if p.fresh {
		return true
	}
	// Exclude if client already has this prop cached
	cacheKey := key
	if p.alias != "" {
		cacheKey = p.alias
	}
	if len(headers.ExceptOnceProps) > 0 && slices.Contains(headers.ExceptOnceProps, cacheKey) {
		return false
	}
	return true
}

func (p *onceProp) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p *onceProp) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	data := oncePropData{Prop: key}
	if p.expiresAt != nil {
		data.ExpiresAt = *p.expiresAt
	}
	if p.alias != "" {
		data.Alias = p.alias
	}
	pp.onceProps[key] = data
}
