# ADR-004: Expression Grouping for Display

## Status

Accepted

## Context

Mathematical expressions can contain long numbers (e.g., `1234567 + 89`). When displayed on a screen, long numbers are difficult to read without digit grouping (thousand separators).

However, the backend's lexer expects expressions without separators. Separators in the request would be interpreted as invalid tokens and rejected.

The display and the transmittable expression are different representations of the same data.

## Decision

Separate the expression format for display from the format transmitted to the backend:

- **Transmittable**: The raw expression without separators (e.g., `1234567+89`)
- **Display**: The same expression with digit grouping applied (e.g., `1,234,567+89`)

The `forDisplay()` function formats expressions for display only. The original expression is always sent to the backend. The separator is a comma, and the grouping size is 3 digits.

```typescript
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
```

## Rationale

**Improved Readability**: Long numbers are easier to read with visual grouping, reducing eye strain and transcription errors.

**Backend Compatibility**: Separators never reach the backend, so the lexer doesn't need to handle or strip them.

**Locale Independence**: The grouping uses a fixed comma separator and groups of 3, rather than following the viewer's locale. This is intentional: the grammar specifies `.` as the decimal separator, so a locale using `.` for grouping would create ambiguity (is `1.234` one-point-two-three-four or one-thousand-two-hundred-thirty-four?).

**Calculation Context**: The calculation display should be unambiguous and consistent, not adapted to the viewer's regional settings.

## Implementation

The implementation groups each "atom" (number literal) within the expression:

```typescript
function groupAtom(atom: string): string {
  const pointIndex = atom.indexOf('.')
  const integerPart = pointIndex === -1 ? atom : atom.slice(0, pointIndex)
  const fraction = pointIndex === -1 ? '' : atom.slice(pointIndex)

  // Exponential notation remains ungrouped
  if (!/^\d+$/.test(integerPart)) return atom

  let grouped = ''
  for (let i = 0; i < integerPart.length; i++) {
    if (i > 0 && (integerPart.length - i) % 3 === 0) {
      grouped += ','
    }
    grouped += integerPart[i]
  }

  return grouped + fraction
}
```

The display is computed whenever the expression changes:

```typescript
const display = forDisplay(readout)
const readoutRef = useFitText<HTMLOutputElement>(display)
```

The actual `readout` state (used for transmission) remains ungrouped.

## Consequences

**Positive:**
- Improved readability of long expressions
- No impact on backend implementation
- Clear separation between display and transmission representations
- Graceful handling of exponential notation (which is not grouped)

**Negative:**
- Display differs from the underlying representation, which can confuse users if they think they're seeing the raw expression
- Results in exponential notation (e.g., `1e+38`) cannot be grouped and remain visually noisy
- Requires maintaining two parallel representations during rendering

## Notes

Exponential notation is detected by checking if the integer part contains non-digit characters (like `e`, `+`, `-`). When exponential notation is present, the atom is returned ungrouped, since separators inside `1e+38` would be nonsense.

The `forDisplay()` function iterates character-by-character rather than splitting on operators, because that allows it to handle all operator types (even if new ones are added) without updating the implementation.

Results returned from the backend may contain exponential notation for very large or very small numbers. These results are converted to expression text via `toExpressionText()` and then can be used in new expressions. The `isSubmittable()` function will reject them if they contain exponential notation, keeping focus on what the keypad can actually emit.
