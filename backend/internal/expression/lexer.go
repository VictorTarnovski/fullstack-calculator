package expression

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"unicode"
)

type lexer struct {
	rd     *bufio.Reader
	tokens []token
	pos    int
}

func newLexer(rd io.Reader) (*lexer, error) {
	lx := &lexer{
		rd:     bufio.NewReader(rd),
		tokens: make([]token, 0),
	}

	for {
		tok, err := lx.lex()
		if err != nil {
			return nil, err
		}

		if tok == tokenEOF {
			break
		}

		lx.tokens = append(lx.tokens, tok)
	}

	slices.Reverse(lx.tokens)

	return lx, nil
}

func (lx *lexer) readRune() (rune, error) {
	r, _, err := lx.rd.ReadRune()
	if err != nil {
		return 0, err
	}

	lx.pos++

	return r, nil
}

func (lx *lexer) unreadRune() error {
	if err := lx.rd.UnreadRune(); err != nil {
		return err
	}

	lx.pos--

	return nil
}

func (lx *lexer) lex() (token, error) {
	for {
		r, err := lx.readRune()
		if err != nil {
			if err == io.EOF {
				return tokenEOF, nil
			}

			return token{}, err
		}

		if unicode.IsSpace(r) {
			continue
		}

		if Operator(r).IsValid() {
			return newToken(tokenKindOperator, string(r)), nil
		}

		if unicode.IsDigit(r) {
			literal, err := lx.lexNumber(r)
			if err != nil {
				return token{}, err
			}

			return newToken(tokenKindAtom, literal), nil
		}

		return token{}, fmt.Errorf("%w %q at position %d", ErrUnrecognizedToken, r, lx.pos)
	}
}

// lexNumber consumes the remaining runes of a number whose first digit has
// already been read. The accepted grammar is strict: digits, optionally
// followed by a single '.' that must itself be followed by at least one digit.
func (lx *lexer) lexNumber(first rune) (string, error) {
	digits := []rune{first}
	hasDot := false

	for {
		r, err := lx.readRune()
		if err != nil {
			if err == io.EOF {
				break
			}

			return "", err
		}

		if unicode.IsDigit(r) {
			digits = append(digits, r)
			continue
		}

		if r == '.' {
			if hasDot {
				return "", fmt.Errorf(
					"%w: multiple decimal points in %q at position %d",
					ErrInvalidNumber, string(append(digits, r)), lx.pos,
				)
			}

			hasDot = true
			digits = append(digits, r)

			continue
		}

		if err := lx.unreadRune(); err != nil {
			return "", err
		}

		break
	}

	if digits[len(digits)-1] == '.' {
		return "", fmt.Errorf(
			"%w: trailing decimal point in %q at position %d",
			ErrInvalidNumber, string(digits), lx.pos,
		)
	}

	return string(digits), nil
}

func (lx *lexer) peek() token {
	if len(lx.tokens) == 0 {
		return tokenEOF
	}

	return lx.tokens[len(lx.tokens)-1]
}

func (lx *lexer) next() token {
	if len(lx.tokens) == 0 {
		return tokenEOF
	}

	tok := lx.tokens[len(lx.tokens)-1]
	lx.tokens = lx.tokens[:len(lx.tokens)-1]

	return tok
}
