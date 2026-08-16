# ADR-006: Optical Adjustments for Typography

## Status

Accepted

## Context

When text is centered within a container using flexbox, the text's line box (not the glyph itself) is centered. This causes glyphs to appear visually misaligned because of the gap between the baseline and the box's center.

Different glyph types have different optical characteristics:
- **Cap-height glyphs** (digits, letters like "A", "C"): Rest on the baseline, need to move up
- **Math symbols** (÷, ×, −): Their ink sits below cap height, need a different adjustment
- **Decimal point** (.): Belongs on the baseline, shouldn't move
- **Erase symbol** (⌫): An asymmetric wedge, needs both horizontal and vertical adjustment

Without adjustments, key labels appear to rest low in their buttons, making them look misaligned and unprofessional.

## Decision

Apply measured optical adjustments to each glyph using CSS `translate()` transforms. These moves are applied to the label text, not the button, so the button's visual boundary never shifts.

Adjustments are grouped by glyph type:

```typescript
const keyLabelVariants = cva('block', {
  variants: {
    optical: {
      cap: 'translate-y-[-0.06em]',        // Cap-height glyphs
      math: 'translate-y-[-0.12em]',       // Math symbols
      baseline: null,                       // Decimal point
      erase: 'translate-x-[-0.083em] translate-y-[-0.06em]',  // ⌫ symbol
    },
  },
})
```

Each key declares its optical type, with a default based on kind:

```typescript
const OPTICAL_BY_KIND: Record<KeyKind, Optical> = {
  atom: 'cap',
  function: 'cap',
  operator: 'math',
}
```

Most keys use the default, but specific glyphs override it:

```typescript
const KEYS: CalculatorKey[] = [
  { label: '⌫', name: 'Backspace', optical: 'erase', ... },
  { label: '.', name: 'Decimal point', optical: 'baseline', ... },
  // All others use the default for their kind
]
```

## Rationale

**Optical Correctness**: Glyphs appear properly centered, matching the designer's intent and professional standards.

**Measured Accuracy**: Adjustments are measured ink offsets, not guess-work. The values were determined by measuring the actual centroid of each glyph relative to its box center.

**Font Independence**: Using `em` units ensures adjustments scale with the glyph size. If font size changes, adjustments scale automatically.

**Semantic Structure**: Grouping by optical type (rather than individual glyphs) allows multiple glyphs to share the same adjustment, reducing duplication.

**Professional Appearance**: Proper optical alignment is a hallmark of carefully designed interfaces and elevates the perceived quality.

## Implementation

Adjustments are defined using `class-variance-authority` (CVA) variants:

```typescript
const keyLabelVariants = cva('block', {
  variants: {
    optical: {
      cap: 'translate-y-[-0.06em]',
      math: 'translate-y-[-0.12em]',
      baseline: null,
      erase: 'translate-x-[-0.083em] translate-y-[-0.06em]',
    },
  },
})
```

In the render, the correct variant is selected based on the key:

```jsx
<span
  className={keyLabelVariants({
    optical: key.optical ?? OPTICAL_BY_KIND[key.kind],
    span: key.span,
  })}
>
  {key.label}
</span>
```

Keys with standard glyphs don't specify optical adjustments and inherit the default for their kind:

```typescript
{ label: '7', name: 'Seven', kind: 'atom', action: 'append' },
// Uses the default: optical: 'cap'
```

Only glyphs that deviate from their kind's default specify an override:

```typescript
{ label: '÷', name: 'Divide', kind: 'operator', action: 'append' },
// Uses the default: optical: 'math'

{ label: '.', name: 'Decimal point', kind: 'atom', action: 'append', optical: 'baseline' },
// Overrides to: optical: 'baseline'
```

## Consequences

**Positive:**
- Glyphs appear properly centered and professionally aligned
- Adjustments scale with font size automatically
- Maintainable through semantic grouping (math, cap, etc.)
- No JavaScript required; pure CSS rendering

**Negative:**
- Requires careful measurement to determine correct offsets
- Adjustments may need tweaking if the font family changes
- The `em` unit adjustment values appear arbitrary without context (requires documentation)
- Font hinting and rasterization differences across browsers can slightly affect alignment

## Notes

The measurements are in `em` units, which are relative to the element's computed font size. For a 2rem button label, a `-0.06em` adjustment moves the text up by `2rem * 0.06 ≈ 0.12rem`.

The erase symbol (⌫) is the most complex adjustment because it's an asymmetric glyph: a full-height right edge with a point tapering to the left. Its ink weight (the visual centroid) lands about `0.083em` to the right of its box, requiring both horizontal and vertical compensation.

If the font family changes in the future, these adjustments may need recalibration. The measurements are specific to the current font stack (Geist variable, via `@fontsource-variable/geist`). A different font's metrics would require remeasurement.

The `baseline` variant is used for the decimal point because its ink sits at the baseline by design, and centering it vertically would move it too high, appearing unnatural.
