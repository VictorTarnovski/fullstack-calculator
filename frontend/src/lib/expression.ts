/**
 * The grammar rules a keypad has to know to avoid building an expression the
 * backend will reject. Not a second parser: it decides what may follow what,
 * never what an expression evaluates to.
 *
 * Three rules cover everything the keypad can get wrong, because it can only
 * emit digits, a decimal point, the four infix operators and percent:
 *
 *   1. an infix operator needs an operand on both sides, so two in a row is
 *      invalid
 *   2. a number may hold at most one decimal point, and may not end on one
 *   3. percent is postfix: it needs a finished number on its left, and nothing
 *      can extend that number afterwards
 *
 * Anything obeying all three is `number% (operator number%)*`, exactly the
 * accepted grammar — so precedence, evaluation, what percent means against each
 * operator, and division by zero all stay on the server.
 */

/* The Unicode math characters, matching the Operator constants in
   internal/expression. The backend rejects the ASCII lookalikes '*' and '-', so
   the keypad must never emit one. */
const OPERATORS = ['+', '−', '×', '÷'] as const

export type Operator = (typeof OPERATORS)[number]

const DECIMAL_POINT = '.'

/* Postfix, so it is not in OPERATORS: it takes an operand on its left only, and
   every rule keyed off "is an operator" means the infix ones. */
const PERCENT = '%'

/** U+2212, not the ASCII hyphen JavaScript writes a negative number with. */
const MINUS: Operator = '−'

/* Grouping is a comma and the decimal point stays a period rather than
   following the viewer's locale: the grammar fixes '.' as the decimal
   separator, so a locale grouping with '.' would render "1.234" for both
   one-thousand-two-hundred-thirty-four and one-point-two-three-four. */
const GROUP_SEPARATOR = ','
const GROUP_SIZE = 3

/** Every character a key can contribute, which is also every character the
    lexer accepts. */
const KEYPAD_CHARACTERS = new RegExp(`^[0-9${DECIMAL_POINT}${PERCENT}${OPERATORS.join('')}]+$`)

export function isOperator(label: string): label is Operator {
  return (OPERATORS as readonly string[]).includes(label)
}

/** Anything that ends the number before it. */
function endsOperand(character: string): boolean {
  return isOperator(character) || character === PERCENT
}

/** The number currently being typed: everything after the last operator. */
function currentOperand(expression: string): string {
  let start = 0

  for (let i = 0; i < expression.length; i++) {
    if (endsOperand(expression[i])) {
      start = i + 1
    }
  }

  return expression.slice(start)
}

function lastCharacter(expression: string): string {
  return expression.slice(-1)
}

/**
 * append adds a key's label to the expression, dropping the press instead when
 * it would produce something unparseable.
 */
export function append(expression: string, label: string): string {
  if (isOperator(label)) {
    /* A number ending in a decimal point is unfinished, so the point is dropped
       rather than stranded before the operator: "7." + "×" is "7×". */
    const base = lastCharacter(expression) === DECIMAL_POINT ? expression.slice(0, -1) : expression

    /* Two operators in a row have no operand between them, so the new one
       replaces the old: "7+" then "×" is "7×". */
    return isOperator(lastCharacter(base)) ? base.slice(0, -1) + label : base + label
  }

  if (label === PERCENT) {
    const last = lastCharacter(expression)

    /* Percent needs a finished number on its left, so it is refused after an
       infix operator or a dangling point. After another percent it is allowed:
       "5%%" is a hundredth of a hundredth, which the grammar parses. */
    return last === '' || isOperator(last) || last === DECIMAL_POINT ? expression : expression + label
  }

  /* Nothing extends a number that percent has already closed: "5%3" would put
     two operands side by side. */
  if (lastCharacter(expression) === PERCENT) {
    return expression
  }

  /* Only the current operand matters: the point in "1.5+2" is not the one being
     typed. */
  if (label === DECIMAL_POINT && currentOperand(expression).includes(DECIMAL_POINT)) {
    return expression
  }

  return expression + label
}

/**
 * backspace removes the last character, returning an empty string when nothing
 * usable is left. Deleting into a negative result leaves a bare sign, which is
 * not a number and cannot become one — this grammar has no unary minus — so it
 * collapses rather than stranding the readout on a lone operator.
 */
export function backspace(expression: string): string {
  const shortened = expression.slice(0, -1)

  return isOperator(shortened) ? '' : shortened
}

/**
 * isSubmittable reports whether the expression is complete enough to send. The
 * entry rules in append keep the middle well-formed, leaving the two ends and
 * the alphabet.
 *
 * The alphabet check is what holds for text the keypad did not type: a result
 * can come back in exponential notation, and "1e-21" carries characters the
 * lexer would reject.
 */
export function isSubmittable(expression: string): boolean {
  if (!KEYPAD_CHARACTERS.test(expression)) return false

  const first = expression[0]
  const last = lastCharacter(expression)

  /* A leading operator has no left operand — there is no unary minus here — and
     percent has nothing to be a percentage of. */
  if (endsOperand(first) || first === DECIMAL_POINT) return false

  /* A trailing infix operator has no right operand, but a trailing percent is
     complete: it took its operand from the left. */
  return !isOperator(last) && last !== DECIMAL_POINT
}

/**
 * toExpressionText renders a result as text the readout can keep building on,
 * swapping JavaScript's ASCII sign for the U+2212 the subtract key carries.
 * Only a leading sign is replaced; a hyphen inside an exponent is left for
 * isSubmittable to reject.
 */
export function toExpressionText(value: number): string {
  return String(value).replace(/^-/, MINUS)
}

/** Group the integer part of a single number: "1234567.89" → "1,234,567.89". */
function groupAtom(atom: string): string {
  const pointIndex = atom.indexOf(DECIMAL_POINT)
  const integerPart = pointIndex === -1 ? atom : atom.slice(0, pointIndex)
  const fraction = pointIndex === -1 ? '' : atom.slice(pointIndex)

  /* Exponential notation reaches here from a result, and separators inside
     "1e+38" would be nonsense. */
  if (!/^\d+$/.test(integerPart)) return atom

  let grouped = ''

  for (let i = 0; i < integerPart.length; i++) {
    if (i > 0 && (integerPart.length - i) % GROUP_SIZE === 0) {
      grouped += GROUP_SEPARATOR
    }

    grouped += integerPart[i]
  }

  return grouped + fraction
}

/**
 * forDisplay groups the digits of every number in an expression, for the screen
 * only: separators are not part of the grammar and would be rejected as
 * unrecognized tokens, so what gets sent is always the unformatted string.
 */
export function forDisplay(expression: string): string {
  let display = ''
  let atom = ''

  for (const character of expression) {
    if (endsOperand(character)) {
      display += groupAtom(atom) + character
      atom = ''
      continue
    }

    atom += character
  }

  return display + groupAtom(atom)
}
