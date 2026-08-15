// Package expression turns a stream of Unicode text into an evaluable arithmetic
// expression tree.
//
// Numbers are decimal literals of the form "digits[.digits]"; "1." and ".5" are
// rejected. Operators are the characters listed in the Operator constants, and
// multiplication and division bind tighter than addition and subtraction.
//
// Parsing and evaluation are separate steps, and both can fail:
//
//	expr, err := expression.New(strings.NewReader("1 + 2 × 3"))
//	if err != nil {
//	    return err
//	}
//
//	val, err := expr.Eval() // 7
//	if err != nil {
//	    return err
//	}
package expression

import (
	"fmt"
	"io"
	"strconv"
)

// Expression is a node in a parsed expression tree, as returned by New. Its
// implementations are unexported, so a tree can only be built by parsing.
type Expression interface {
	// Eval computes the value of this node and everything below it.
	Eval() (float32, error)

	// Stringer renders the node as an S-expression, which makes the tree shape
	// and operator precedence visible.
	fmt.Stringer
}

var (
	_ Expression = atomExpression{}
	_ Expression = binaryExpression{}
)

// atomExpression is a leaf node holding an unparsed numeric literal. The
// literal is kept as text so that String reproduces the input exactly, and is
// converted only on Eval.
type atomExpression struct {
	value string
}

func newAtomExpression(value string) atomExpression {
	return atomExpression{value: value}
}

// Eval returns ErrInvalidNumber if the literal is malformed, which cannot
// happen for a tree built by New because the lexer validates literals first.
func (a atomExpression) Eval() (float32, error) {
	val64, err := strconv.ParseFloat(a.value, 32)
	if err != nil {
		return 0, fmt.Errorf("%w %q: %w", ErrInvalidNumber, a.value, err)
	}

	return float32(val64), nil
}

func (a atomExpression) String() string {
	return a.value
}

// binaryExpression is an internal node applying an operator to exactly two
// operands. Build it with newBinaryExpression, which enforces that count;
// constructing it directly with a different number of operands makes Eval
// panic on an out-of-range index.
type binaryExpression struct {
	operator Operator
	operands []Expression
}

// newBinaryExpression returns ErrOperandCount unless exactly two operands
// are supplied.
func newBinaryExpression(op Operator, operands ...Expression) (binaryExpression, error) {
	if len(operands) != 2 {
		return binaryExpression{}, fmt.Errorf("%w, got %d", ErrOperandCount, len(operands))
	}

	return binaryExpression{
		operator: op,
		operands: operands,
	}, nil
}

// Eval evaluates both operands depth-first, then applies the operator. It
// propagates the first operand error unchanged, returns ErrDivisionByZero when
// dividing by a zero right operand rather than yielding a float infinity, and
// returns ErrUnknownOperator for an operator outside the supported set.
func (o binaryExpression) Eval() (float32, error) {
	left, err := o.operands[0].Eval()
	if err != nil {
		return 0, err
	}

	right, err := o.operands[1].Eval()
	if err != nil {
		return 0, err
	}

	switch o.operator {
	case OperatorPlus:
		return left + right, nil
	case OperatorMinus:
		return left - right, nil
	case OperatorMultiplication:
		return left * right, nil
	case OperatorDivision:
		if right == 0 {
			return 0, fmt.Errorf("%w: %s", ErrDivisionByZero, o)
		}

		return left / right, nil
	default:
		return 0, fmt.Errorf("%w %q", ErrUnknownOperator, o.operator)
	}
}

// String renders the node as an S-expression, so "1 + 2 × 3" becomes
// "(+ 1 (× 2 3))" and operator precedence is visible in the output.
func (o binaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", o.operator, o.operands[0], o.operands[1])
}

// New reads an arithmetic expression from rd and returns its parsed tree. It
// consumes rd to EOF, so the whole input must be a single expression.
// Evaluation is deferred: a successful parse still requires a call to Eval.
//
// It returns ErrUnrecognizedToken for a rune outside the accepted set,
// ErrInvalidNumber for a malformed literal, and ErrUnexpectedToken for a
// well-formed token in the wrong position, including on empty input.
func New(rd io.Reader) (Expression, error) {
	lx, err := newLexer(rd)
	if err != nil {
		return nil, err
	}

	return parseExpression(lx, 0.0)
}

func parseExpression(lx *lexer, minBindingPower float32) (Expression, error) {
	tok := lx.next()
	if tok.kind != tokenKindAtom {
		return nil, fmt.Errorf("%w: expected ATOM, got %s %q", ErrUnexpectedToken, tok.kind, tok.literal)
	}

	var leftHandSide Expression = newAtomExpression(tok.literal)

	for {
		peeked := lx.peek()

		if peeked.kind == tokenKindEOF {
			break
		}

		if peeked.kind != tokenKindOperator {
			return nil, fmt.Errorf("%w: expected OPERATOR, got %s %q", ErrUnexpectedToken, peeked.kind, peeked.literal)
		}

		op := Operator([]rune(peeked.literal)[0])

		leftBindingPower, rightBindingPower, err := infixBindingPower(op)
		if err != nil {
			return nil, err
		}

		if leftBindingPower < minBindingPower {
			break
		}

		lx.next()

		rightHandSide, err := parseExpression(lx, rightBindingPower)
		if err != nil {
			return nil, err
		}

		leftHandSide, err = newBinaryExpression(op, leftHandSide, rightHandSide)
		if err != nil {
			return nil, err
		}
	}

	return leftHandSide, nil
}

// infixBindingPower returns the left and right binding powers for op. The right
// power is the higher of the two, which makes operators left-associative.
func infixBindingPower(op Operator) (float32, float32, error) {
	switch op {
	case OperatorPlus, OperatorMinus:
		return 1.0, 1.1, nil
	case OperatorMultiplication, OperatorDivision:
		return 2.0, 2.1, nil
	default:
		return 0, 0, fmt.Errorf("%w %q", ErrUnknownOperator, op)
	}
}
