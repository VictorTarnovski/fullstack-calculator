package expression

import "testing"

func TestTokenKindString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     tokenKind
		expected string
	}{
		{name: "unknown", kind: tokenKindUnknown, expected: "UNKNOWN"},
		{name: "eof", kind: tokenKindEOF, expected: "EOF"},
		{name: "atom", kind: tokenKindAtom, expected: "ATOM"},
		{name: "operator", kind: tokenKindOperator, expected: "OPERATOR"},
		{name: "zero value", kind: tokenKind(0), expected: "UNKNOWN"},
		{name: "out of range", kind: tokenKind(99), expected: "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.kind.String(); got != tt.expected {
				t.Fatalf("tokenKind(%d).String() = %q, expected: %q", tt.kind, got, tt.expected)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tok      token
		expected string
	}{
		{name: "atom", tok: newToken(tokenKindAtom, "12"), expected: "ATOM(12)"},
		{name: "operator", tok: newToken(tokenKindOperator, "+"), expected: "OPERATOR(+)"},
		{name: "eof", tok: tokenEOF, expected: "EOF(EOF)"},
		{name: "zero value", tok: token{}, expected: "UNKNOWN()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.tok.String(); got != tt.expected {
				t.Fatalf("token.String() = %q, expected: %q", got, tt.expected)
			}
		})
	}
}
