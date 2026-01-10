package inertia

import "context"

// ScrollMetadata provides pagination metadata for infinite scrolling.
type ScrollMetadata struct {
	PageName     string // Query param name (default: "page")
	PreviousPage any    // nil, int, or string (for cursor pagination)
	NextPage     any    // nil, int, or string (for cursor pagination)
	CurrentPage  any    // int or string (for cursor pagination)
}

// ScrollProp represents a paginated property for infinite scrolling.
// Use the generic Scroll[T] constructor to create instances.
type ScrollProp struct {
	resolver     PropFunc
	wrapper      string
	metadata     *ScrollMetadata
	metadataFunc func(value any) *ScrollMetadata
}

// ScrollOption configures ScrollProp creation.
type ScrollOption func(*ScrollProp)

// WithWrapper sets the data wrapper key path (default: "data").
func WithWrapper(wrapper string) ScrollOption {
	return func(s *ScrollProp) {
		s.wrapper = wrapper
	}
}

// WithScrollMetadata sets static scroll metadata.
func WithScrollMetadata(metadata ScrollMetadata) ScrollOption {
	return func(s *ScrollProp) {
		s.metadata = &metadata
	}
}

// WithScrollMetadataFunc sets a function to derive metadata from the resolved value.
func WithScrollMetadataFunc(fn func(value any) *ScrollMetadata) ScrollOption {
	return func(s *ScrollProp) {
		s.metadataFunc = fn
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
func Scroll[T any](resolver func(ctx context.Context) ([]T, error), opts ...ScrollOption) ScrollProp {
	sp := ScrollProp{
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

// Resolve executes the resolver (lazy evaluation).
func (s *ScrollProp) Resolve(ctx context.Context) (any, error) {
	return s.resolver(ctx)
}

// GetMetadata returns the scroll metadata for this prop.
func (s *ScrollProp) GetMetadata(resolvedValue any) *ScrollMetadata {
	if s.metadataFunc != nil {
		return s.metadataFunc(resolvedValue)
	}
	if s.metadata != nil {
		return s.metadata
	}
	// Return default metadata if none provided
	return &ScrollMetadata{PageName: "page"}
}

// GetWrapper returns the wrapper key path.
func (s *ScrollProp) GetWrapper() string {
	return s.wrapper
}
