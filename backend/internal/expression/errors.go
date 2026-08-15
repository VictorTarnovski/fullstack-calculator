package expression

import "errors"

// Sentinel errors returned by this package. Every error is wrapped with
// positional or operand context before being returned, so callers MUST match
// them with errors.Is rather than ==.
var (
	// ErrUnrecognizedToken reports a rune that is neither whitespace, a digit,
	// nor one of the operators in this package. ASCII lookalikes such as '*'
	// and '-' land here.
	ErrUnrecognizedToken = errors.New("expression: unrecognized token")

	// ErrUnexpectedToken reports a token that is valid on its own but is not
	// allowed at its position, such as an operator with no left operand.
	ErrUnexpectedToken = errors.New("expression: unexpected token")

	// ErrInvalidNumber reports a malformed numeric literal: a trailing decimal
	// point, or more than one decimal point.
	ErrInvalidNumber = errors.New("expression: invalid number")

	// ErrUnknownOperator reports an operator rune with no defined behaviour.
	// Reaching it means the lexer and the evaluator disagree on the operator
	// set.
	ErrUnknownOperator = errors.New("expression: unknown operator")

	// ErrDivisionByZero reports a division whose right operand evaluated to
	// zero. Division returns this instead of a float infinity.
	ErrDivisionByZero = errors.New("expression: division by zero")

	// ErrOperandCount reports an operation built with a number of operands
	// other than two.
	ErrOperandCount = errors.New("expression: operation requires exactly two operands")
)
