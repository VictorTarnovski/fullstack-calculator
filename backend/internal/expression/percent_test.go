package expression

import (
	"errors"
	"strings"
	"testing"
)

// Percent is postfix and its meaning depends on the operator it hangs off:
// against + and − it is a share of the left operand, which is what a pocket
// calculator does, and against × and ÷ it is a plain hundredth, because a share
// of the left operand there would mean "200 × 20" for "200 × 10%".
func TestPercentEval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    float32
		expectedErr error
	}{
		{
			name:     "bare percent is a hundredth",
			input:    "10%",
			expected: 0.1,
		},
		{
			name:     "subtracting a percent takes a share of the left operand",
			input:    "200 − 10%",
			expected: 180,
		},
		{
			name:     "adding a percent adds a share of the left operand",
			input:    "50 + 10%",
			expected: 55,
		},
		{
			name:     "the share is of the whole left operand, not its last term",
			input:    "100 + 100 − 10%",
			expected: 180,
		},
		{
			name:     "multiplying by a percent is a plain hundredth",
			input:    "200 × 10%",
			expected: 20,
		},
		{
			name:     "dividing by a percent is a plain hundredth",
			input:    "20 ÷ 10%",
			expected: 200,
		},
		{
			name:     "percent binds tighter than multiplication",
			input:    "2 + 3 × 4%",
			expected: 2.12,
		},
		{
			name:     "decimal percent",
			input:    "200 − 12.5%",
			expected: 175,
		},
		{
			name:     "percent of zero",
			input:    "0 + 10%",
			expected: 0,
		},
		{
			name:        "dividing by a zero percent is still division by zero",
			input:       "20 ÷ 0%",
			expectedErr: ErrDivisionByZero,
		},
		{
			name:        "percent with no operand",
			input:       "% 5",
			expectedErr: ErrUnexpectedToken,
		},
		{
			name:        "percent applied to nothing",
			input:       "%",
			expectedErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr, err := New(strings.NewReader(tt.input))
			if err != nil {
				if tt.expectedErr != nil && errors.Is(err, tt.expectedErr) {
					return
				}

				t.Fatalf("New(%q) unexpected err: %v", tt.input, err)
			}

			got, err := expr.Eval()

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("Eval(%q) err: %v, expected: %v", tt.input, err, tt.expectedErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Eval(%q) unexpected err: %v", tt.input, err)
			}

			if got != tt.expected {
				t.Fatalf("Eval(%q) = %v, expected: %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPercentTree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "postfix on an atom",
			input:    "10%",
			expected: "(% 10)",
		},
		{
			name:     "binds to the atom, not the expression",
			input:    "200 − 10%",
			expected: "(− 200 (% 10))",
		},
		{
			name:     "binds tighter than multiplication",
			input:    "2 + 3 × 4%",
			expected: "(+ 2 (× 3 (% 4)))",
		},
		{
			name:     "repeated postfix",
			input:    "5%%",
			expected: "(% (% 5))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr, err := New(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("New(%q) unexpected err: %v", tt.input, err)
			}

			if got := expr.String(); got != tt.expected {
				t.Fatalf("New(%q).String() = %q, expected: %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPercentOperatorIsValid(t *testing.T) {
	t.Parallel()

	if !OperatorPercent.IsValid() {
		t.Error("OperatorPercent.IsValid() = false, expected true")
	}

	if got := OperatorPercent.String(); got != "%" {
		t.Errorf("OperatorPercent.String() = %q, expected %q", got, "%")
	}
}
