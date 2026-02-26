package props

import (
	"context"
	"reflect"
	"slices"
	"strings"
)

// Resolver is the generic resolver function signature for lazy-evaluated props.
type Resolver[T any] func(ctx context.Context) (T, error)

// Prop is the interface that all prop types implement.
type Prop interface {
	// ShouldInclude determines if the prop should be included in the response
	// based on the request headers and partial reload state.
	ShouldInclude(key string, headers *Headers) bool

	// Resolve executes the prop's resolver or returns its static value.
	Resolve(ctx context.Context) (any, error)

	// ModifyProcessedProps allows the prop to modify the processed props
	// (e.g., adding to deferredProps, onceProps, scrollProps, etc.)
	ModifyProcessedProps(key string, headers *Headers, p *ProcessedProps)
}

// Headers contains the request header data needed for prop evaluation.
type Headers struct {
	Component           string
	PartialData         []string
	PartialExcept       []string
	ExceptOnceProps     []string
	ResetProps          []string
	InfiniteScrollMerge string
	IsPartial           bool
	IsInertia           bool
}

// OncePropData is the JSON structure for once props in response.
type OncePropData struct {
	Prop      string `json:"prop"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	Alias     string `json:"alias,omitempty"`
}

// ScrollPropMetadata contains pagination metadata for infinite scrolling.
type ScrollPropMetadata struct {
	PageName     string `json:"pageName"`
	PreviousPage any    `json:"previousPage"`
	NextPage     any    `json:"nextPage"`
	CurrentPage  any    `json:"currentPage"`
	Reset        bool   `json:"reset"`
}

// ProcessedProps holds the accumulated prop processing results.
type ProcessedProps struct {
	FinalProps     map[string]any
	DeferredProps  map[string][]string
	OnceProps      map[string]OncePropData
	ScrollProps    map[string]ScrollPropMetadata
	MergeProps     []string
	PrependProps   []string
	DeepMergeProps []string
	MatchPropsOn   []string
}

// NewProcessedProps creates a new ProcessedProps with initialized maps.
func NewProcessedProps() *ProcessedProps {
	return &ProcessedProps{
		FinalProps:    make(map[string]any),
		DeferredProps: make(map[string][]string),
		OnceProps:     make(map[string]OncePropData),
		ScrollProps:   make(map[string]ScrollPropMetadata),
	}
}

// ResetProcessedProps clears a ProcessedProps for reuse.
func ResetProcessedProps(p *ProcessedProps) {
	clear(p.FinalProps)
	clear(p.DeferredProps)
	clear(p.OnceProps)
	clear(p.ScrollProps)
	p.MergeProps = p.MergeProps[:0]
	p.PrependProps = p.PrependProps[:0]
	p.DeepMergeProps = p.DeepMergeProps[:0]
	p.MatchPropsOn = p.MatchPropsOn[:0]
}

// DefaultShouldInclude is the shared logic for default/lazy/value/scroll props.
func DefaultShouldInclude(key string, headers *Headers) bool {
	if headers.IsPartial {
		includedByData := len(headers.PartialData) == 0
		if len(headers.PartialData) > 0 {
			includedByData = slices.Contains(headers.PartialData, key)
		}

		excluded := false
		if len(headers.PartialExcept) > 0 {
			excluded = slices.Contains(headers.PartialExcept, key)
		}

		return includedByData && !excluded
	}
	return true
}

// StructToProps converts a struct (or nil) into a map[string]Prop by reflecting
// over its exported fields. Fields are keyed by their json tag name (or lowercased
// field name if no tag). Fields implementing the Prop interface are used directly;
// all other values are wrapped as Value props.
func StructToProps(propsStruct any) map[string]Prop {
	if propsStruct == nil {
		return nil
	}

	v := reflect.ValueOf(propsStruct)
	t := v.Type()

	// Handle pointer to struct
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]Prop, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		key := fieldKey(field)
		if key == "-" {
			continue
		}

		fieldVal := v.Field(i)

		// Check if the field value implements the Prop interface
		if p, ok := fieldVal.Interface().(Prop); ok {
			result[key] = p
		} else {
			// Also check pointer to field
			if fieldVal.CanAddr() {
				if p, ok := fieldVal.Addr().Interface().(Prop); ok {
					result[key] = p
					continue
				}
			}
			// Wrap plain values as Value
			result[key] = Value[any]{Val: fieldVal.Interface()}
		}
	}

	return result
}

// fieldKey returns the JSON key for a struct field.
func fieldKey(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(f.Name[:1]) + f.Name[1:]
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return strings.ToLower(f.Name[:1]) + f.Name[1:]
	}
	return name
}
