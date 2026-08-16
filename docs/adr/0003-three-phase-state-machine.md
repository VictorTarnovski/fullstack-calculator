# ADR-003: Three-Phase State Machine

## Status

Accepted

## Context

The calculator needs to handle different user interactions depending on what state it's in. For example:
- When a result is on screen and the user presses a digit, should it start a new expression or extend the result?
- When an API request is in flight, should key presses be processed or ignored?
- When the user is typing, what happens when they press an operator after a result?

Different interactions require different handling based on the current state, so a state machine provides clarity about valid transitions and behavior.

## Decision

Implement a three-phase state machine:

- **`editing`**: User is building an expression by pressing keys. All key presses are processed normally.
- **`result`**: Backend returned a value. An operator key carries the result forward to build a new expression, while a digit key starts a fresh expression from zero.
- **`pending`**: Request is in flight to evaluate the expression. All key presses are ignored.

```
editing
  ├─ [operator] → remains in editing
  ├─ [digit] → remains in editing
  ├─ [=] (if valid) → sends request, transitions to pending
  ├─ [backspace] → remains in editing
  └─ [clear] → remains in editing

result
  ├─ [operator] → expression = last_result + operator, transitions to editing
  ├─ [digit] → expression = "0" + digit, transitions to editing
  ├─ [backspace] → expression becomes editing state
  └─ [clear] → remains in result

pending
  ├─ [any key] → ignored, remains pending
  └─ [response] → transitions to result with new expression
      [error] → transitions to editing with original expression
```

## Rationale

**Clear Semantics**: Each phase has explicit behavior. Users understand what happens when they press a key in any given state.

**Prevents Double-Submission**: The `pending` phase prevents multiple simultaneous API requests from the same keystroke.

**Familiar UX**: The result carry-forward behavior (operator extends result, digit starts fresh) matches the behavior of physical calculators, which users expect.

**Error Recovery**: When an error occurs, the expression is preserved in `editing` so users can correct it without retyping.

**Type Safety**: A union type prevents accidentally using an undefined state and ensures all transitions are explicit.

## Implementation

The phase is tracked as part of component state in `src/components/Calculator.tsx`:

```typescript
type Phase = 'editing' | 'result' | 'pending'

export default function Calculator() {
  const [readout, setReadout] = useState(EMPTY)
  const [phase, setPhase] = useState<Phase>('editing')

  function baseFor(kind: KeyKind): string {
    if (phase === 'result') return kind === 'atom' ? EMPTY : readout
    return readout
  }

  function append(key: CalculatorKey) {
    const base = baseFor(key.kind)
    const next = base === EMPTY && key.kind === 'atom' && key.label !== '.'
      ? key.label
      : appendToExpression(base, key.label)
    setReadout(next)
    setPhase('editing')
  }

  async function submit() {
    setPhase('pending')
    try {
      const result = await evaluate(expression)
      setReadout(toExpressionText(result))
      setPhase('result')
    } catch (error) {
      setReadout(expression)
      setPhase('editing')
    }
  }

  function press(key: CalculatorKey) {
    if (phase === 'pending') return
    ACTIONS[key.action](key)
  }
}
```

## Consequences

**Positive:**
- Clear, predictable behavior across all input scenarios
- Prevents UI glitches from rapid key presses
- Familiar calculator behavior that users expect
- Easy to test each phase's behavior independently

**Negative:**
- Requires careful state management to avoid bugs
- The `baseFor()` function adds complexity to the append logic
- Testing requires verifying all phase transitions

## Notes

The `baseFor()` function is the core of the result carry-forward behavior. It returns:
- `EMPTY` ("0") if a digit is pressed after a result
- The current `readout` if an operator is pressed after a result

This allows a result to be used as the left operand of the next operator while starting fresh with the right operand.

The error handling preserves the original expression that failed, allowing users to correct and resubmit it without losing their work.
