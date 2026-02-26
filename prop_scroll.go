package inertia

import (
	"context"
	"slices"
)

// ScrollMetadata provides pagination metadata for infinite scrolling.
type ScrollMetadata struct {
	PageName     string // Query param name (default: "page")
	PreviousPage any    // nil, int, or string (for cursor pagination)
	NextPage     any    // nil, int, or string (for cursor pagination)
	CurrentPage  any    // int or string (for cursor pagination)
}

// ScrollProp represents a paginated property for infinite scrolling.
type ScrollProp[T any] struct {
	resolver Resolver[[]T]
	wrapper  string
	metadata *ScrollMetadata
}

// ScrollOption configures ScrollProp creation.
type ScrollOption func(*scrollConfig)

type scrollConfig struct {
	wrapper  string
	metadata *ScrollMetadata
}

// WithWrapper sets the data wrapper key path (default: "data").
func WithWrapper(wrapper string) ScrollOption {
	return func(c *scrollConfig) {
		c.wrapper = wrapper
	}
}

// WithScrollMetadata sets static scroll metadata.
func WithScrollMetadata(metadata ScrollMetadata) ScrollOption {
	return func(c *scrollConfig) {
		c.metadata = &metadata
	}
}

// Scroll creates a new ScrollProp for infinite scrolling with type-safe resolver.
// The resolver returns a typed slice which is automatically wrapped.
//
// Example:
//
//	"posts": inertia.Scroll(func(ctx context.Context) ([]Post, error) {
//	    return db.GetPosts(page, 20)
//	}, inertia.WithScrollMetadata(inertia.ScrollMetadata{
//	    PageName:    "page",
//	    CurrentPage: page,
//	    NextPage:    page + 1,
//	}))
//
// This produces response: {"posts": {"data": [...]}}
// With merge path: "posts.data"
func Scroll[T any](resolver Resolver[[]T], opts ...ScrollOption) ScrollProp[T] {
	cfg := &scrollConfig{
		wrapper: "data",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return ScrollProp[T]{
		resolver: resolver,
		wrapper:  cfg.wrapper,
		metadata: cfg.metadata,
	}
}

func (p ScrollProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p ScrollProp[T]) resolve(ctx context.Context) (any, error) {
	items, err := p.resolver(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{p.wrapper: items}, nil
}

func (p ScrollProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	if !p.shouldInclude(key, headers) {
		return
	}

	meta := p.getMetadata()
	pp.scrollProps[key] = scrollPropMetadata{
		PageName:     meta.PageName,
		PreviousPage: meta.PreviousPage,
		NextPage:     meta.NextPage,
		CurrentPage:  meta.CurrentPage,
		Reset:        slices.Contains(headers.ResetProps, key),
	}

	mergePath := key + "." + p.wrapper
	if headers.InfiniteScrollMerge == "prepend" {
		pp.prependProps = append(pp.prependProps, mergePath)
	} else {
		pp.mergeProps = append(pp.mergeProps, mergePath)
	}
}

func (p ScrollProp[T]) getMetadata() *ScrollMetadata {
	if p.metadata != nil {
		return p.metadata
	}
	return &ScrollMetadata{PageName: "page"}
}
