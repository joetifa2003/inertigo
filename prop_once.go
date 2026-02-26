package inertia

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// OnceProp is included once and then excluded on subsequent page visits.
type OnceProp[T any] struct {
	resolver  Resolver[T]
	expiresAt *string
	fresh     bool   // Force refresh even if client has cached
	alias     string // Custom key for client-side caching
}

// OnceOption configures Once prop behavior
type OnceOption func(*onceConfig)

type onceConfig struct {
	expiresAt *string
	fresh     bool
	alias     string
}

// Fresh forces the prop to be resolved even if the client has it cached.
func Fresh() OnceOption {
	return func(c *onceConfig) { c.fresh = true }
}

// FreshWhen conditionally forces the prop to be resolved.
func FreshWhen(condition bool) OnceOption {
	return func(c *onceConfig) { c.fresh = condition }
}

// OnceAs assigns a custom key for client-side caching.
// This allows sharing data across multiple pages while using different prop names.
func OnceAs(alias string) OnceOption {
	return func(c *onceConfig) { c.alias = alias }
}

// OnceUntil sets an expiration time for the once prop.
// After expiration, the prop will be resolved again on subsequent visits.
func OnceUntil(duration time.Duration) OnceOption {
	return func(c *onceConfig) {
		expiresAt := fmt.Sprintf("%d", time.Now().Add(duration).UnixMilli())
		c.expiresAt = &expiresAt
	}
}

// Once creates a prop that is included on the first visit and then
// excluded on subsequent visits to the same page.
func Once[T any](resolver Resolver[T], opts ...OnceOption) OnceProp[T] {
	cfg := &onceConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return OnceProp[T]{
		resolver:  resolver,
		expiresAt: cfg.expiresAt,
		fresh:     cfg.fresh,
		alias:     cfg.alias,
	}
}

func (p OnceProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
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

func (p OnceProp[T]) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p OnceProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	data := oncePropData{Prop: key}
	if p.expiresAt != nil {
		data.ExpiresAt = *p.expiresAt
	}
	if p.alias != "" {
		data.Alias = p.alias
	}
	pp.onceProps[key] = data
}
