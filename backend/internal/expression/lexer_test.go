package expression

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// lexAll drains the lexer, returning the tokens in source order.
func lexAll(input string) ([]token, error) {
	lx, err := newLexer(strings.NewReader(input))
	if err != nil {
		return nil, err
	}

	tokens := []token{}

	for {
		tok := lx.next()
		if tok == tokenEOF {
			break
		}

		tokens = append(tokens, tok)
	}

	return tokens, nil
}

func TestNewLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    []token
		expectedErr error
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []token{},
		},
		{
			name:     "whitespace only",
			input:    " \t\n ",
			expected: []token{},
		},
		{
			name:     "single digit",
			input:    "1",
			expected: []token{newToken(tokenKindAtom, "1")},
		},
		{
			name:     "multi digit",
			input:    "120",
			expected: []token{newToken(tokenKindAtom, "120")},
		},
		{
			name:     "decimal",
			input:    "120.75",
			expected: []token{newToken(tokenKindAtom, "120.75")},
		},
		{
			name:  "operator between atoms",
			input: "1 + 2",
			expected: []token{
				newToken(tokenKindAtom, "1"),
				newToken(tokenKindOperator, "+"),
				newToken(tokenKindAtom, "2"),
			},
		},
		{
			name:  "no surrounding whitespace",
			input: "12+34",
			expected: []token{
				newToken(tokenKindAtom, "12"),
				newToken(tokenKindOperator, "+"),
				newToken(tokenKindAtom, "34"),
			},
		},
		{
			name:  "decimal followed by operator",
			input: "1.5×2.25",
			expected: []token{
				newToken(tokenKindAtom, "1.5"),
				newToken(tokenKindOperator, "×"),
				newToken(tokenKindAtom, "2.25"),
			},
		},
		{
			name:  "all operators",
			input: "1 + 2 − 3 × 4 ÷ 5",
			expected: []token{
				newToken(tokenKindAtom, "1"),
				newToken(tokenKindOperator, "+"),
				newToken(tokenKindAtom, "2"),
				newToken(tokenKindOperator, "−"),
				newToken(tokenKindAtom, "3"),
				newToken(tokenKindOperator, "×"),
				newToken(tokenKindAtom, "4"),
				newToken(tokenKindOperator, "÷"),
				newToken(tokenKindAtom, "5"),
			},
		},
		{
			name:  "irregular whitespace",
			input: "  1\t\t+\n2  ",
			expected: []token{
				newToken(tokenKindAtom, "1"),
				newToken(tokenKindOperator, "+"),
				newToken(tokenKindAtom, "2"),
			},
		},
		{
			name:  "adjacent atoms are lexed, not rejected",
			input: "1 2",
			expected: []token{
				newToken(tokenKindAtom, "1"),
				newToken(tokenKindAtom, "2"),
			},
		},
		{
			name:        "ascii asterisk",
			input:       "1 * 2",
			expectedErr: ErrUnrecognizedToken,
		},
		{
			name:        "ascii hyphen",
			input:       "1 - 2",
			expectedErr: ErrUnrecognizedToken,
		},
		{
			name:        "leading decimal point",
			input:       ".5",
			expectedErr: ErrUnrecognizedToken,
		},
		{
			name:        "trailing decimal point",
			input:       "1.",
			expectedErr: ErrInvalidNumber,
		},
		{
			name:        "trailing decimal point before operator",
			input:       "1.+2",
			expectedErr: ErrInvalidNumber,
		},
		{
			name:        "multiple decimal points",
			input:       "1.2.3",
			expectedErr: ErrInvalidNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := lexAll(tt.input)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("lexAll(%q) err: %v, expected: %v", tt.input, err, tt.expectedErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("lexAll(%q) unexpected err: %v", tt.input, err)
			}

			if !slices.Equal(got, tt.expected) {
				t.Fatalf("lexAll(%q) = %v, expected: %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLexerPeek(t *testing.T) {
	t.Parallel()

	lx, err := newLexer(strings.NewReader("1 + 2"))
	if err != nil {
		t.Fatalf("newLexer() unexpected err: %v", err)
	}

	expected := newToken(tokenKindAtom, "1")

	if got := lx.peek(); got != expected {
		t.Fatalf("first peek() = %v, expected: %v", got, expected)
	}

	if got := lx.peek(); got != expected {
		t.Fatalf("second peek() = %v, expected: %v — peek must not consume", got, expected)
	}

	if got := lx.next(); got != expected {
		t.Fatalf("next() = %v, expected: %v", got, expected)
	}

	expected = newToken(tokenKindOperator, "+")

	if got := lx.peek(); got != expected {
		t.Fatalf("peek() after next() = %v, expected: %v", got, expected)
	}
}

func TestLexerExhausted(t *testing.T) {
	t.Parallel()

	lx, err := newLexer(strings.NewReader("1"))
	if err != nil {
		t.Fatalf("newLexer() unexpected err: %v", err)
	}

	if got := lx.next(); got != newToken(tokenKindAtom, "1") {
		t.Fatalf("next() = %v, expected the sole atom", got)
	}

	if got := lx.next(); got != tokenEOF {
		t.Fatalf("next() past the end = %v, expected: %v", got, tokenEOF)
	}

	if got := lx.peek(); got != tokenEOF {
		t.Fatalf("peek() past the end = %v, expected: %v", got, tokenEOF)
	}
}
