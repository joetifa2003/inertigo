package inertia

import (
	"context"
	"reflect"
	"slices"
	"strings"
)

// Resolver is the generic resolver function signature for lazy-evaluated props.
type Resolver[T any] func(ctx context.Context) (T, error)

// propFunc is the internal untyped resolver function signature.
type propFunc func(ctx context.Context) (any, error)

// prop is the internal interface that all prop types must implement.
type prop interface {
	// shouldInclude determines if the prop should be included in the response
	// based on the request headers and partial reload state.
	shouldInclude(key string, headers *inertiaHeaders) bool

	// resolve executes the prop's resolver or returns its static value.
	resolve(ctx context.Context) (any, error)

	// modifyProcessedProps allows the prop to modify the processed props
	// (e.g., adding to deferredProps, onceProps, scrollProps, etc.)
	modifyProcessedProps(key string, headers *inertiaHeaders, p *processedProps)
}

// oncePropData is the JSON structure for once props in response.
type oncePropData struct {
	Prop      string `json:"prop"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	Alias     string `json:"alias,omitempty"` // Custom key for client-side caching
}

// defaultShouldInclude is the shared logic for default/lazy/value/scroll props.
// It includes the prop on full load, and on partial reload only if requested
// (or if no specific props are requested) and not excluded.
func defaultShouldInclude(key string, headers *inertiaHeaders) bool {
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

// structToProps converts a struct (or nil) into a map[string]prop by reflecting
// over its exported fields. Fields are keyed by their json tag name (or lowercased
// field name if no tag). Fields implementing the prop interface are used directly;
// all other values are wrapped as valueProp.
func structToProps(propsStruct any) map[string]prop {
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

	result := make(map[string]prop, t.NumField())

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

		// Check if the field value implements the prop interface
		if p, ok := fieldVal.Interface().(prop); ok {
			result[key] = p
		} else {
			// Also check pointer to field
			if fieldVal.CanAddr() {
				if p, ok := fieldVal.Addr().Interface().(prop); ok {
					result[key] = p
					continue
				}
			}
			// Wrap plain values as valueProp
			result[key] = valueProp{value: fieldVal.Interface()}
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
