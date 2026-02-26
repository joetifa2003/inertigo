package props

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// Once is included once and then excluded on subsequent page visits.
type Once[T any] struct {
	resolver  Resolver[T]
	expiresAt *string
	fresh     bool
	alias     string
}

// OnceOption configures Once prop behavior.
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

// As assigns a custom key for client-side caching.
func As(alias string) OnceOption {
	return func(c *onceConfig) { c.alias = alias }
}

// Until sets an expiration time for the once prop.
func Until(duration time.Duration) OnceOption {
	return func(c *onceConfig) {
		expiresAt := fmt.Sprintf("%d", time.Now().Add(duration).UnixMilli())
		c.expiresAt = &expiresAt
	}
}

// NewOnce creates a prop that is included on the first visit and then
// excluded on subsequent visits to the same page.
func NewOnce[T any](resolver Resolver[T], opts ...OnceOption) Once[T] {
	cfg := &onceConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return Once[T]{
		resolver:  resolver,
		expiresAt: cfg.expiresAt,
		fresh:     cfg.fresh,
		alias:     cfg.alias,
	}
}

func (p Once[T]) ShouldInclude(key string, headers *Headers) bool {
	if p.fresh {
		return true
	}
	cacheKey := key
	if p.alias != "" {
		cacheKey = p.alias
	}
	if len(headers.ExceptOnceProps) > 0 && slices.Contains(headers.ExceptOnceProps, cacheKey) {
		return false
	}
	return true
}

func (p Once[T]) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p Once[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {
	data := OncePropData{Prop: key}
	if p.expiresAt != nil {
		data.ExpiresAt = *p.expiresAt
	}
	if p.alias != "" {
		data.Alias = p.alias
	}
	pp.OnceProps[key] = data
}
