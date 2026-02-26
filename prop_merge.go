package inertia

import (
	"context"
)

// MergeProp wraps a resolver and marks it for client-side merging.
type MergeProp[T any] struct {
	resolver     Resolver[T]
	appendPaths  []string // Nested paths to append (default: root level)
	prependPaths []string // Nested paths to prepend
	matchOn      []string // e.g., ["data.id", "users.uuid"]
	deepMerge    bool     // Deep merge entire structure
}

// MergeOption configures Merge prop behavior
type MergeOption func(*mergeConfig)

type mergeConfig struct {
	appendPaths  []string
	prependPaths []string
	matchOn      []string
	deepMerge    bool
}

// Append specifies nested paths to append to.
// If no paths are provided, appends at root level (default behavior).
func Append(paths ...string) MergeOption {
	return func(c *mergeConfig) { c.appendPaths = paths }
}

// Prepend specifies nested paths to prepend to.
func Prepend(paths ...string) MergeOption {
	return func(c *mergeConfig) { c.prependPaths = paths }
}

// MergeMatchOn specifies fields to match when merging arrays.
// Format: "path.field" e.g., "data.id" or "users.uuid"
func MergeMatchOn(pathMatchPairs ...string) MergeOption {
	return func(c *mergeConfig) { c.matchOn = pathMatchPairs }
}

// MergeDeepMerge enables deep merging of the entire structure.
func MergeDeepMerge() MergeOption {
	return func(c *mergeConfig) { c.deepMerge = true }
}

// Merge creates a prop that will be merged with existing client-side data
// during partial reloads, instead of replacing it entirely.
func Merge[T any](resolver Resolver[T], opts ...MergeOption) MergeProp[T] {
	cfg := &mergeConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return MergeProp[T]{
		resolver:     resolver,
		appendPaths:  cfg.appendPaths,
		prependPaths: cfg.prependPaths,
		matchOn:      cfg.matchOn,
		deepMerge:    cfg.deepMerge,
	}
}

func (p MergeProp[T]) shouldInclude(key string, headers *inertiaHeaders) bool {
	return defaultShouldInclude(key, headers)
}

func (p MergeProp[T]) resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p MergeProp[T]) modifyProcessedProps(key string, headers *inertiaHeaders, pp *processedProps) {
	if p.deepMerge {
		pp.deepMergeProps = append(pp.deepMergeProps, key)
	} else if len(p.prependPaths) > 0 {
		// Prepend at specified paths
		for _, path := range p.prependPaths {
			pp.prependProps = append(pp.prependProps, key+"."+path)
		}
	} else if len(p.appendPaths) > 0 {
		// Append at specified paths
		for _, path := range p.appendPaths {
			pp.mergeProps = append(pp.mergeProps, key+"."+path)
		}
	} else {
		// Default: append at root level
		pp.mergeProps = append(pp.mergeProps, key)
	}

	// Handle matchOn
	for _, match := range p.matchOn {
		pp.matchPropsOn = append(pp.matchPropsOn, key+"."+match)
	}
}
