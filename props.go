package inertia

import (
	"context"
	"fmt"
	"slices"
	"time"
)

type PropFunc func(ctx context.Context) (any, error)

type PropType int

const (
	PropTypeDefault PropType = iota
	PropTypeDeferred
	PropTypeOptional
	PropTypeAlways
	PropTypeOnce
)

type Prop struct {
	Type      PropType
	Value     any
	Resolver  PropFunc
	ExpiresAt *string // For Once props, null means forever
	Group     string  // For Deferred props grouping
}

func Deferred(resolver PropFunc) Prop {
	return Prop{Type: PropTypeDeferred, Resolver: resolver}
}

func DeferredGroup(group string, resolver PropFunc) Prop {
	return Prop{Type: PropTypeDeferred, Resolver: resolver, Group: group}
}

func Optional(resolver PropFunc) Prop {
	return Prop{Type: PropTypeOptional, Resolver: resolver}
}

func Always(value any) Prop {
	return Prop{Type: PropTypeAlways, Value: value}
}

func Once(resolver PropFunc) Prop {
	return Prop{Type: PropTypeOnce, Resolver: resolver}
}

func OnceWithExpiration(resolver PropFunc, expiration time.Duration) Prop {
	expiresAt := fmt.Sprintf("%d", time.Now().Add(expiration).UnixMilli())
	return Prop{Type: PropTypeOnce, Resolver: resolver, ExpiresAt: &expiresAt}
}

type Props map[string]any

type onceProp struct {
	Prop      string `json:"prop"`
	ExpiresAt string `json:"expiresAt"`
}

// Resolve executes the prop's resolver or returns its value.
func (p *Prop) Resolve(ctx context.Context) (any, error) {
	if p.Resolver != nil {
		return p.Resolver(ctx)
	}
	return p.Value, nil
}

// ShouldInclude determines if the prop should be included in the response based on the request headers.
// It returns true if the prop should be included in the 'props' key of the PageObject.
func (p *Prop) ShouldInclude(key string, headers *inertiaHeaders) bool {
	switch p.Type {
	case PropTypeDefault:
		if headers.IsPartial {
			// Default props:
			// Include if explicitly requested (Data)
			// OR if Data is empty (meaning "all") AND not excluded (Except)
			includedBySize := len(headers.PartialData) == 0
			if len(headers.PartialData) > 0 {
				includedBySize = slices.Contains(headers.PartialData, key)
			}

			excluded := false
			if len(headers.PartialExcept) > 0 {
				excluded = slices.Contains(headers.PartialExcept, key)
			}

			return includedBySize && !excluded
		}
		return true

	case PropTypeAlways:
		return true

	case PropTypeOptional:
		// Optional props (Lazy):
		// Default: Exclude
		// Partial: Include ONLY if explicitly requested in Data
		if headers.IsPartial && len(headers.PartialData) > 0 {
			if slices.Contains(headers.PartialData, key) {
				return true
			}
		}
		if len(headers.PartialExcept) > 0 && slices.Contains(headers.PartialExcept, key) {
			return false
		}
		return false

	case PropTypeDeferred:
		// Deferred props:
		// Initial visit: Exclude.
		// Partial visit:
		//   If requested (Data), Include.
		//   Else, Exclude.
		requested := false
		if headers.IsPartial && len(headers.PartialData) > 0 {
			requested = slices.Contains(headers.PartialData, key)
		}

		if requested {
			return true
		}
		return false

	case PropTypeOnce:
		if len(headers.ExceptOnceProps) > 0 && slices.Contains(headers.ExceptOnceProps, key) {
			return false
		}
		return true
	}

	return true
}
