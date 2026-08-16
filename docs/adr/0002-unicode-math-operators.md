# ADR-002: Unicode Math Operators

## Status

Accepted

## Context

Mathematical expressions can use different character sets for operators:
- ASCII operators: `+`, `-`, `*`, `/`
- Unicode math operators: `+`, `−` (U+2212), `×` (U+00D7), `÷` (U+00F7)

The calculator must choose which character set to use consistently throughout the frontend and ensure the backend understands the same set.

## Decision

Use Unicode math operators exclusively:
- Addition: `+` (U+002B) — identical to ASCII
- Subtraction: `−` (U+2212) — minus sign, not hyphen
- Multiplication: `×` (U+00D7) — multiplication sign, not letter x or asterisk
- Division: `÷` (U+00F7) — obelus, not slash

The API explicitly specifies `charset=utf-8` in requests to ensure the backend lexer correctly interprets these Unicode characters.

Keyboard input maps ASCII operators to their Unicode equivalents:
- `/` key → `÷`
- `*` key → `×`
- `-` key → `−`
- `+` key → `+`

## Rationale

**Mathematical Correctness**: Unicode math operators are the standard used in mathematical typesetting and are visually distinct from programming operators.

**Accessibility**: The mathematical operators are easier to distinguish visually and have proper semantic meaning for screen readers.

**Backend Alignment**: The backend lexer is configured to accept Unicode characters. Using them from the frontend eliminates ambiguity about which characters are valid.

**Keyboard Mapping**: Mapping ASCII keyboard characters to Unicode operators makes the calculator familiar to keyboard users while maintaining proper rendering for display and transmission.

**Encoding Clarity**: Explicitly specifying `charset=utf-8` in the Content-Type header ensures the backend's stream parser correctly interprets multi-byte characters.

## Implementation

The operator set is defined in `src/lib/expression.ts`:

```typescript
const OPERATORS = ['+', '−', '×', '÷'] as const

export type Operator = (typeof OPERATORS)[number]

const MINUS: Operator = '−' // U+2212

export function isOperator(label: string): label is Operator {
  return (OPERATORS as readonly string[]).includes(label)
}
```

Keyboard mapping in `src/components/Calculator.tsx`:

```typescript
const KEYS: CalculatorKey[] = [
  { label: '÷', name: 'Divide', kind: 'operator', action: 'append', keystrokes: ['/'] },
  { label: '×', name: 'Multiply', kind: 'operator', action: 'append', keystrokes: ['*'] },
  { label: '−', name: 'Subtract', kind: 'operator', action: 'append', keystrokes: ['-'] },
  { label: '+', name: 'Add', kind: 'operator', action: 'append', keystrokes: ['+'] },
]
```

API client in `src/lib/api.ts`:

```typescript
const response = await fetch(`${BASE_URL}/evaluations`, {
  method: 'POST',
  headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  body: expression,
  signal,
})
```

## Consequences

**Positive:**
- Clear distinction between programming operators (ASCII) and mathematical operators (Unicode)
- Proper rendering across all platforms and fonts
- Semantic correctness for accessibility tools
- Eliminates ambiguity about operator interpretation

**Negative:**
- Users typing manually into a text field would need to know Unicode operators or use copy-paste
- Debugging requires understanding Unicode character codes (U+2212 vs `-`)
- Browser and font support must be verified (though modern browsers handle this uniformly)

## Notes

The display output from results uses `toExpressionText()` to convert JavaScript's ASCII negative sign (`-`) to the Unicode minus (`−`) when rendering results. This ensures results can be seamlessly incorporated into new expressions.

Only the leading negative sign is replaced; hyphens within exponential notation (like `1e-21`) are intentionally left unchanged since `isSubmittable()` will reject them as invalid keyboard input.
