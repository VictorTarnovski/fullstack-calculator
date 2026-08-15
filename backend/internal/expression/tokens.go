package expression

import "fmt"

type tokenKind int

const (
	tokenKindUnknown tokenKind = iota
	tokenKindEOF
	tokenKindAtom
	tokenKindOperator
)

func (k tokenKind) String() string {
	switch k {
	case tokenKindEOF:
		return "EOF"
	case tokenKindAtom:
		return "ATOM"
	case tokenKindOperator:
		return "OPERATOR"
	default:
		return "UNKNOWN"
	}
}

var tokenEOF = newToken(tokenKindEOF, tokenKindEOF.String())

type token struct {
	kind    tokenKind
	literal string
}

func newToken(kind tokenKind, literal string) token {
	return token{
		kind:    kind,
		literal: literal,
	}
}

func (t token) String() string {
	return fmt.Sprintf("%s(%s)", t.kind, t.literal)
}
