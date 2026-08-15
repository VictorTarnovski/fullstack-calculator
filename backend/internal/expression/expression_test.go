package expression

import (
	"errors"
	"strings"
	"testing"
)

func TestExpressionEval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    float32
		expectedErr error
	}{
		{
			name:     "atom",
			input:    "12",
			expected: 12,
		},
		{
			name:     "addition",
			input:    "1 + 1",
			expected: 2,
		},
		{
			name:     "precedence",
			input:    "120 + 34 × 5",
			expected: 290,
		},
		{
			name:     "subtraction and division",
			input:    "8 − 6 ÷ 2",
			expected: 5,
		},
		{
			name:     "decimals",
			input:    "1.5 + 2.25",
			expected: 3.75,
		},
		{
			name:     "decimal precedence",
			input:    "0.5 + 1.5 × 2",
			expected: 3.5,
		},
		{
			name:        "division by zero",
			input:       "6 ÷ 0",
			expectedErr: ErrDivisionByZero,
		},
		{
			name:        "nested division by zero",
			input:       "1 + 6 ÷ 0",
			expectedErr: ErrDivisionByZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr, err := New(strings.NewReader(tt.input))
			if err != nil {
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

func TestNewOperationExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		operands    []Expression
		expectedErr error
	}{
		{
			name:        "no operands",
			operands:    []Expression{},
			expectedErr: ErrOperandCount,
		},
		{
			name:        "one operand",
			operands:    []Expression{newAtomExpression("1")},
			expectedErr: ErrOperandCount,
		},
		{
			name:     "two operands",
			operands: []Expression{newAtomExpression("1"), newAtomExpression("2")},
		},
		{
			name: "three operands",
			operands: []Expression{
				newAtomExpression("1"),
				newAtomExpression("2"),
				newAtomExpression("3"),
			},
			expectedErr: ErrOperandCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newBinaryExpression(OperatorPlus, tt.operands...)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("newBinaryExpression(%d operands) err: %v, expected: %v",
					len(tt.operands), err, tt.expectedErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:     "atom",
			input:    "1",
			expected: "1",
		},
		{
			name:     "operator",
			input:    "1 + 2",
			expected: "(+ 1 2)",
		},
		{
			name:     "binding power",
			input:    "1 + 2 × 3",
			expected: "(+ 1 (× 2 3))",
		},
		{
			name:     "nesting",
			input:    "1 + 3 × 2 × 4",
			expected: "(+ 1 (× (× 3 2) 4))",
		},
		{
			name:     "minus and division",
			input:    "8 − 6 ÷ 2",
			expected: "(− 8 (÷ 6 2))",
		},
		{
			name:     "multi digit atom",
			input:    "12",
			expected: "12",
		},
		{
			name:     "multi digit operands",
			input:    "120 + 34 × 5",
			expected: "(+ 120 (× 34 5))",
		},
		{
			name:     "decimal atom",
			input:    "120.75",
			expected: "120.75",
		},
		{
			name:     "decimal operands",
			input:    "1.5 + 2.25 × 3.0",
			expected: "(+ 1.5 (× 2.25 3.0))",
		},
		{
			name:        "lexer error propagates",
			input:       "1 * 2",
			expectedErr: ErrUnrecognizedToken,
		},
		{
			name:        "adjacent atoms",
			input:       "1 2",
			expectedErr: ErrUnexpectedToken,
		},
		{
			name:        "leading operator",
			input:       "+ 1",
			expectedErr: ErrUnexpectedToken,
		},
		{
			name:        "empty input",
			input:       "",
			expectedErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr, err := New(strings.NewReader(tt.input))

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("New(%q) err: %v, expected: %v", tt.input, err, tt.expectedErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("New(%q) unexpected err: %v", tt.input, err)
			}

			if expr.String() != tt.expected {
				t.Fatalf("New(%q) expr: %s, expected: %s", tt.input, expr, tt.expected)
			}
		})
	}
}
