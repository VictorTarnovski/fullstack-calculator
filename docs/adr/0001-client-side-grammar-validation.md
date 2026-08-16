# ADR-001: Client-Side Expression Grammar Validation

## Status

Accepted

## Context

The calculator's frontend needs to build expressions that will be evaluated by a backend API. The backend has specific grammar rules that expressions must follow:
- Infix operators require operands on both sides
- Numbers can contain at most one decimal point
- Percent is a postfix operator
- Specific Unicode characters are required for operators

Sending invalid expressions to the backend results in unnecessary network requests and forces users to see error messages while retyping.

## Decision

Implement a comprehensive grammar validator on the frontend that mirrors the backend's grammar rules. This validator prevents the keypad from emitting invalid expressions entirely.

The validator implements three key rules:

1. **Operator Isolation**: Two operators in a row are invalid; when detected, the new operator replaces the old
2. **Decimal Point Isolation**: Numbers contain at most one decimal point; numbers may not end on one
3. **Percent Closure**: Percent is postfix and closes a number; nothing can extend it afterwards

These three rules are sufficient because the keypad can only emit:
- Digits
- A decimal point
- Four infix operators (+, −, ×, ÷)
- Percent (%)

Any valid sequence of these tokens that satisfies all three rules matches the full accepted grammar.

## Rationale

**Reducing Backend Load**: Invalid expressions never reach the backend, reducing unnecessary API calls and server resource usage.

**Immediate Feedback**: Users get instant visual feedback when they attempt invalid operations, without waiting for a network request.

**Offline Capability**: The grammar validation layer makes it possible to build a fully functional offline experience, with the backend becoming optional.

**Separation of Concerns**: The frontend validates that expressions are structurally sound; the backend validates evaluation correctness (division by zero, overflow, etc.).

## Implementation

The validator is implemented in `src/lib/expression.ts`:

```typescript
export function append(expression: string, label: string): string {
  if (isOperator(label)) {
    // Drop trailing decimal point before operator
    const base = lastCharacter(expression) === '.' ? expression.slice(0, -1) : expression
    // Replace consecutive operators
    return isOperator(lastCharacter(base)) ? base.slice(0, -1) + label : base + label
  }

  if (label === '%') {
    // Require finished number on left
    return (last === '' || isOperator(last) || last === '.') ? expression : expression + label
  }

  // Percent closes the number; nothing extends it
  if (lastCharacter(expression) === '%') {
    return expression
  }

  // Only one decimal point per number
  if (label === '.' && currentOperand(expression).includes('.')) {
    return expression
  }

  return expression + label
}

export function isSubmittable(expression: string): boolean {
  // Additional checks for expression completeness
  const first = expression[0]
  const last = lastCharacter(expression)

  if (endsOperand(first) || first === '.') return false
  return !isOperator(last) && last !== '.'
}
```

## Consequences

**Positive:**
- Reduced backend load from invalid requests
- Better user experience with immediate feedback
- Clearer separation between structural validation (frontend) and semantic validation (backend)
- Foundation for offline-capable features

**Negative:**
- Increased frontend complexity
- Grammar changes require updates in two places (frontend and backend)
- Grammar inconsistency between frontend and backend could cause confusion

## Notes

The `currentOperand()` function extracts the number being typed (everything after the last operator). This allows the validator to apply rules like "one decimal point per number" correctly, even in multi-operand expressions.

The backspace operation includes a special case: deleting from a negative result (which would leave a bare operator) collapses the expression to empty, since the grammar has no unary minus operator.
