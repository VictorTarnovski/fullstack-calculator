package expression

import "testing"

func TestOperatorIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		operator Operator
		expected bool
	}{
		{name: "plus", operator: OperatorPlus, expected: true},
		{name: "minus", operator: OperatorMinus, expected: true},
		{name: "multiplication", operator: OperatorMultiplication, expected: true},
		{name: "division", operator: OperatorDivision, expected: true},
		{name: "ascii asterisk", operator: Operator('*'), expected: false},
		{name: "ascii hyphen", operator: Operator('-'), expected: false},
		{name: "ascii slash", operator: Operator('/'), expected: false},
		{name: "digit", operator: Operator('1'), expected: false},
		{name: "zero value", operator: Operator(0), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.operator.IsValid(); got != tt.expected {
				t.Fatalf("Operator(%#U).IsValid() = %t, expected: %t", rune(tt.operator), got, tt.expected)
			}
		})
	}
}

func TestOperatorString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		operator Operator
		expected string
	}{
		{name: "plus", operator: OperatorPlus, expected: "+"},
		{name: "minus", operator: OperatorMinus, expected: "−"},
		{name: "multiplication", operator: OperatorMultiplication, expected: "×"},
		{name: "division", operator: OperatorDivision, expected: "÷"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.operator.String(); got != tt.expected {
				t.Fatalf("Operator(%#U).String() = %q, expected: %q", rune(tt.operator), got, tt.expected)
			}
		})
	}
}
