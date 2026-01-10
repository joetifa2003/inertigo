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

// scrollProp represents a paginated property for infinite scrolling.
type scrollProp struct {
	resolver PropFunc
	wrapper  string
	metadata *ScrollMetadata
}

// ScrollOption configures ScrollProp creation.
type ScrollOption func(*scrollProp)

// WithWrapper sets the data wrapper key path (default: "data").
func WithWrapper(wrapper string) ScrollOption {
	return func(s *scrollProp) {
		s.wrapper = wrapper
	}
}

// WithScrollMetadata sets static scroll metadata.
func WithScrollMetadata(metadata ScrollMetadata) ScrollOption {
	return func(s *scrollProp) {
		s.metadata = &metadata
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
func Scroll[T any](resolver func(ctx context.Context) ([]T, error), opts ...ScrollOption) Prop {
	sp := scrollProp{
		wrapper: "data",
	}
	for _, opt := range opts {
		opt(&sp)
	}

	// Wrap the typed resolver with automatic data wrapping
	sp.resolver = func(ctx context.Context) (any, error) {
		items, err := resolver(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{sp.wrapper: items}, nil
	}

	return sp
}

func (p scrollProp) ShouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p scrollProp) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p scrollProp) ModifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	if !p.ShouldInclude(key, headers) {
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

func (p *scrollProp) getMetadata() *ScrollMetadata {
	if p.metadata != nil {
		return p.metadata
	}
	return &ScrollMetadata{PageName: "page"}
}
