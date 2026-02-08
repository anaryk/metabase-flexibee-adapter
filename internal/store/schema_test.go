package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/flexibee"
)

func TestFlexibeeTypeToPG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		flexibeeType string
		expected     string
	}{
		{"string", "TEXT"},
		{"integer", "BIGINT"},
		{"numeric", "NUMERIC"},
		{"date", "DATE"},
		{"datetime", "TIMESTAMPTZ"},
		{"logic", "BOOLEAN"},
		{"relation", "TEXT"},
		{"unknown", "TEXT"},
		{"", "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.flexibeeType, func(t *testing.T) {
			t.Parallel()
			result := FlexibeeTypeToPG(flexibee.Property{Type: tt.flexibeeType})
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", `"simple"`},
		{"with-hyphen", `"with_hyphen"`},
		{"already_underscore", `"already_underscore"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := sanitizeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
