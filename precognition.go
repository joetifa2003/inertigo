package inertia

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

// IsPrecognition checks if the request is a Precognition validation request
func IsPrecognition(r *http.Request) bool {
	return r.Header.Get(HeaderPrecognition) == "true"
}

// PrecognitionFields returns the list of fields to validate from the header
func PrecognitionFields(r *http.Request) []string {
	header := r.Header.Get(HeaderPrecognitionValidateOnly)
	if header == "" {
		return nil
	}
	return strings.Split(header, ",")
}

// ShouldValidateField checks if a specific field should be validated.
// Returns true if no filter is specified (validate all) or if field is in the list.
func ShouldValidateField(r *http.Request, field string) bool {
	fields := PrecognitionFields(r)
	if len(fields) == 0 {
		return true // No filter, validate all
	}
	return slices.Contains(fields, field)
}

// PrecognitionSuccess sends a successful Precognition response (204 No Content)
func PrecognitionSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Vary", HeaderPrecognition)
	w.Header().Set(HeaderPrecognition, "true")
	w.Header().Set(HeaderPrecognitionSuccess, "true")
	w.WriteHeader(http.StatusNoContent)
}

// PrecognitionError sends a Precognition error response (422 Unprocessable Entity)
func PrecognitionError(w http.ResponseWriter, r *http.Request, errors map[string]any) error {
	w.Header().Set("Vary", HeaderPrecognition)
	w.Header().Set(HeaderPrecognition, "true")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	return json.NewEncoder(w).Encode(map[string]any{"errors": errors})
}
