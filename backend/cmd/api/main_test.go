package main

import (
	"slices"
	"testing"
)

func TestParseOrigins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		expected []string
	}{
		{
			name:     "single origin",
			raw:      "http://localhost:5173",
			expected: []string{"http://localhost:5173"},
		},
		{
			name:     "several origins",
			raw:      "http://localhost:5173,https://calculator.example",
			expected: []string{"http://localhost:5173", "https://calculator.example"},
		},
		{
			name:     "surrounding whitespace is trimmed",
			raw:      " http://localhost:5173 , https://calculator.example ",
			expected: []string{"http://localhost:5173", "https://calculator.example"},
		},
		{
			name: "trailing slash is trimmed, since a browser never sends one",
			raw:  "http://localhost:5173/",
			// A browser sends "http://localhost:5173" in the Origin header and
			// the comparison is literal, so a configured slash would silently
			// reject every request.
			expected: []string{"http://localhost:5173"},
		},
		{
			name:     "empty entries are skipped",
			raw:      "http://localhost:5173,,",
			expected: []string{"http://localhost:5173"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOrigins(tt.raw)
			if err != nil {
				t.Fatalf("parseOrigins(%q) unexpected err: %v", tt.raw, err)
			}

			if !slices.Equal(got, tt.expected) {
				t.Errorf("parseOrigins(%q) = %v, want %v", tt.raw, got, tt.expected)
			}
		})
	}
}

func TestParseOriginsRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "empty configuration",
			raw:  "",
		},
		{
			name: "only separators",
			raw:  " , ",
		},
		{
			name: "missing scheme",
			raw:  "localhost:5173",
		},
		{
			name: "unsupported scheme",
			raw:  "ftp://localhost:5173",
		},
		{
			name: "origin carrying a path",
			raw:  "http://localhost:5173/api",
		},
		{
			name: "one bad origin rejects the whole list",
			raw:  "http://localhost:5173,notanorigin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOrigins(tt.raw)
			if err == nil {
				t.Fatalf("parseOrigins(%q) = %v, want an error", tt.raw, got)
			}
		})
	}
}
