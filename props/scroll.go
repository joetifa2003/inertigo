package props

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

// Scroll represents a paginated property for infinite scrolling.
type Scroll[T any] struct {
	resolver Resolver[[]T]
	wrapper  string
	metadata *ScrollMetadata
}

// ScrollOption configures Scroll prop creation.
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

// NewScroll creates a new Scroll prop for infinite scrolling with type-safe resolver.
func NewScroll[T any](resolver Resolver[[]T], opts ...ScrollOption) Scroll[T] {
	cfg := &scrollConfig{
		wrapper: "data",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return Scroll[T]{
		resolver: resolver,
		wrapper:  cfg.wrapper,
		metadata: cfg.metadata,
	}
}

func (p Scroll[T]) ShouldInclude(key string, headers *Headers) bool {
	return DefaultShouldInclude(key, headers)
}

func (p Scroll[T]) Resolve(ctx context.Context) (any, error) {
	items, err := p.resolver(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{p.wrapper: items}, nil
}

func (p Scroll[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {
	if !p.ShouldInclude(key, headers) {
		return
	}

	meta := p.getMetadata()
	pp.ScrollProps[key] = ScrollPropMetadata{
		PageName:     meta.PageName,
		PreviousPage: meta.PreviousPage,
		NextPage:     meta.NextPage,
		CurrentPage:  meta.CurrentPage,
		Reset:        slices.Contains(headers.ResetProps, key),
	}

	mergePath := key + "." + p.wrapper
	if headers.InfiniteScrollMerge == "prepend" {
		pp.PrependProps = append(pp.PrependProps, mergePath)
	} else {
		pp.MergeProps = append(pp.MergeProps, mergePath)
	}
}

func (p Scroll[T]) getMetadata() *ScrollMetadata {
	if p.metadata != nil {
		return p.metadata
	}
	return &ScrollMetadata{PageName: "page"}
}
