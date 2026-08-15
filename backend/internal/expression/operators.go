package expression

import "slices"

// Operator is a single arithmetic operator, stored as the rune it was lexed
// from. The zero value is not a valid operator; check with IsValid.
type Operator rune

// The supported operators. The four infix ones are the Unicode math
// characters, not their ASCII lookalikes: "1 * 2" and "1 - 2" are rejected as
// unrecognized tokens, while "1 × 2" and "1 − 2" parse. Percent is the
// exception, since ASCII '%' is the percent sign rather than a lookalike for
// one.
const (
	OperatorPlus           Operator = 0x002B // + U+002B PLUS SIGN
	OperatorMinus          Operator = 0x2212 // − U+2212 MINUS SIGN
	OperatorMultiplication Operator = 0x00D7 // × U+00D7 MULTIPLICATION SIGN
	OperatorDivision       Operator = 0x00F7 // ÷ U+00F7 DIVISION SIGN
	OperatorPercent        Operator = 0x0025 // % U+0025 PERCENT SIGN
)

// IsValid reports whether o is one of the supported operators. The lexer uses
// it to decide whether a rune starts an operator token, so an operator that
// fails this check never reaches the parser.
func (o Operator) IsValid() bool {
	return slices.Contains(
		[]Operator{
			OperatorPlus,
			OperatorMinus,
			OperatorMultiplication,
			OperatorDivision,
			OperatorPercent,
		},
		o,
	)
}

// IsPostfix reports whether o takes a single operand on its left rather than
// one on each side. The parser needs the distinction before asking for binding
// powers, because a postfix operator has no right-hand side to parse.
func (o Operator) IsPostfix() bool {
	return o == OperatorPercent
}

// String returns the operator as a one-character string. An invalid operator
// renders as whatever rune it holds, including the replacement character.
func (o Operator) String() string {
	return string(rune(o))
}
