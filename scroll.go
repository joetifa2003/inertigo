package inertia

// ScrollConfig holds the pagination metadata for infinite scroll.
type ScrollConfig struct {
	PageName     string `json:"pageName"`     // Query parameter name for pagination (e.g., "page")
	PreviousPage any    `json:"previousPage"` // Previous page number/cursor (nil if no previous)
	NextPage     any    `json:"nextPage"`     // Next page number/cursor (nil if no next)
	CurrentPage  any    `json:"currentPage"`  // Current page number/cursor
}

// ScrollProp wraps data with scroll configuration for infinite scroll support.
// This is used by the Inertia client to implement seamless infinite scrolling.
type ScrollProp struct {
	Data      any          // The paginated data (e.g., struct with Data field)
	Config    ScrollConfig // Pagination metadata
	MergePath string       // Path to merge at (e.g., "posts.data")
}

// Scroll creates a ScrollProp for infinite scroll functionality.
// The data is typically a struct with a Data field containing the items.
// The config provides pagination metadata (page numbers/cursors).
// The mergePath specifies where the data should be merged (e.g., "users.data").
//
// Example:
//
//	Scroll(
//	    map[string]any{"data": users},
//	    ScrollConfig{
//	        PageName:     "page",
//	        CurrentPage:  1,
//	        NextPage:     2,
//	        PreviousPage: nil,
//	    },
//	    "users.data",
//	)
func Scroll(data any, config ScrollConfig, mergePath string) ScrollProp {
	return ScrollProp{
		Data:      data,
		Config:    config,
		MergePath: mergePath,
	}
}
