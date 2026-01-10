package inertia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsPrecognition(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{"With Header", "true", true},
		{"Without Header", "", false},
		{"With False Header", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set(HeaderPrecognition, tt.header)
			}
			if got := IsPrecognition(req); got != tt.expected {
				t.Errorf("IsPrecognition() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPrecognitionFields(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected []string
	}{
		{"No Header", "", nil},
		{"Single Field", "name", []string{"name"}},
		{"Multiple Fields", "name,email", []string{"name", "email"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set(HeaderPrecognitionValidateOnly, tt.header)
			}
			got := PrecognitionFields(req)
			if len(got) != len(tt.expected) {
				t.Errorf("PrecognitionFields() len = %v, want %v", len(got), len(tt.expected))
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("PrecognitionFields()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestShouldValidateField(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		field    string
		expected bool
	}{
		{"No Filter", "", "any", true},
		{"Filter Match", "name", "name", true},
		{"Filter No Match", "name", "email", false},
		{"Multiple Filter Match", "name,email", "email", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set(HeaderPrecognitionValidateOnly, tt.header)
			}
			if got := ShouldValidateField(req, tt.field); got != tt.expected {
				t.Errorf("ShouldValidateField() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPrecognitionSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	PrecognitionSuccess(rec)

	if rec.Code != http.StatusNoContent {
		t.Errorf("PrecognitionSuccess code = %v, want %v", rec.Code, http.StatusNoContent)
	}

	if rec.Header().Get("Vary") != HeaderPrecognition {
		t.Errorf("Vary header = %v, want %v", rec.Header().Get("Vary"), HeaderPrecognition)
	}
	if rec.Header().Get(HeaderPrecognition) != "true" {
		t.Errorf("Precognition header = %v, want true", rec.Header().Get(HeaderPrecognition))
	}
	if rec.Header().Get(HeaderPrecognitionSuccess) != "true" {
		t.Errorf("Precognition-Success header = %v, want true", rec.Header().Get(HeaderPrecognitionSuccess))
	}
}

func TestPrecognitionError(t *testing.T) {
	rec := httptest.NewRecorder()
	errors := map[string]any{"name": "Required"}
	err := PrecognitionError(rec, errors)
	if err != nil {
		t.Fatalf("PrecognitionError() error = %v", err)
	}

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("PrecognitionError code = %v, want %v", rec.Code, http.StatusUnprocessableEntity)
	}

	if rec.Header().Get("Vary") != HeaderPrecognition {
		t.Errorf("Vary header = %v, want %v", rec.Header().Get("Vary"), HeaderPrecognition)
	}
	if rec.Header().Get(HeaderPrecognition) != "true" {
		t.Errorf("Precognition header = %v, want true", rec.Header().Get(HeaderPrecognition))
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", rec.Header().Get("Content-Type"))
	}

	var body map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to decode body: %v", err)
	}

	if body["errors"]["name"] != "Required" {
		t.Errorf("Body errors.name = %v, want Required", body["errors"]["name"])
	}
}
