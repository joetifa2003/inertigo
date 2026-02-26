package props

import (
	"context"
)

// Merge wraps a resolver and marks it for client-side merging.
type Merge[T any] struct {
	resolver     Resolver[T]
	appendPaths  []string
	prependPaths []string
	matchOn      []string
	deepMerge    bool
}

// MergeOption configures Merge prop behavior.
type MergeOption func(*mergeConfig)

type mergeConfig struct {
	appendPaths  []string
	prependPaths []string
	matchOn      []string
	deepMerge    bool
}

// Append specifies nested paths to append to.
func Append(paths ...string) MergeOption {
	return func(c *mergeConfig) { c.appendPaths = paths }
}

// Prepend specifies nested paths to prepend to.
func Prepend(paths ...string) MergeOption {
	return func(c *mergeConfig) { c.prependPaths = paths }
}

// MatchOn specifies fields to match when merging arrays.
func MatchOn(pathMatchPairs ...string) MergeOption {
	return func(c *mergeConfig) { c.matchOn = pathMatchPairs }
}

// DeepMerge enables deep merging of the entire structure.
func DeepMerge() MergeOption {
	return func(c *mergeConfig) { c.deepMerge = true }
}

// NewMerge creates a prop that will be merged with existing client-side data
// during partial reloads, instead of replacing it entirely.
func NewMerge[T any](resolver Resolver[T], opts ...MergeOption) Merge[T] {
	cfg := &mergeConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return Merge[T]{
		resolver:     resolver,
		appendPaths:  cfg.appendPaths,
		prependPaths: cfg.prependPaths,
		matchOn:      cfg.matchOn,
		deepMerge:    cfg.deepMerge,
	}
}

func (p Merge[T]) ShouldInclude(key string, headers *Headers) bool {
	return DefaultShouldInclude(key, headers)
}

func (p Merge[T]) Resolve(ctx context.Context) (any, error) {
	return p.resolver(ctx)
}

func (p Merge[T]) ModifyProcessedProps(key string, headers *Headers, pp *ProcessedProps) {
	if p.deepMerge {
		pp.DeepMergeProps = append(pp.DeepMergeProps, key)
	} else if len(p.prependPaths) > 0 {
		for _, path := range p.prependPaths {
			pp.PrependProps = append(pp.PrependProps, key+"."+path)
		}
	} else if len(p.appendPaths) > 0 {
		for _, path := range p.appendPaths {
			pp.MergeProps = append(pp.MergeProps, key+"."+path)
		}
	} else {
		pp.MergeProps = append(pp.MergeProps, key)
	}

	for _, match := range p.matchOn {
		pp.MatchPropsOn = append(pp.MatchPropsOn, key+"."+match)
	}
}
