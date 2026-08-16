# ADR-005: Container Queries for Responsive Design

## Status

Accepted

## Context

The calculator must scale properly across devices ranging from narrow phones in portrait to landscape tablets. The keypad needs to be appropriately sized in each context:

- On phones: Full-width, taking all available space
- On tablets: A bounded panel with predictable dimensions
- On landscape phones: The panel might be narrower than the viewport

Traditional media queries respond to viewport size, but this causes problems:
- The keypad's width is capped by height (56dvh), so on landscape phones it's far narrower than the viewport
- Sizing key labels relative to viewport width (`vw`) means they can outgrow the keys that hold them
- This can push the bottom row off the screen

## Decision

Use CSS Container Queries to size interactive elements relative to their container's width, not the viewport width. This allows key labels to scale correctly regardless of the viewport's aspect ratio.

Measured in Container Query Width (`cqw`), key labels scale from their CSS ceiling (`clamp()` max value) down to fit their container:

```css
const KEY_TYPE_SIZE = 'text-[clamp(1rem,9.4cqw,2.125rem)]'
const OPERATOR_TYPE_SIZE = 'text-[clamp(1.25rem,12.8cqw,2.875rem)]'
```

The container is the `@container` div holding the key grid:

```jsx
<div className="@container grid grid-cols-4 gap-3">
  {KEYS.map((key) => (
    <button className={keyVariants({ kind, span })}>{key.label}</button>
  ))}
</div>
```

## Rationale

**Correct Aspect Ratio Handling**: Container queries respond to the keypad's actual width, not the viewport's width. A landscape phone's keypad is constrained by height, making it much narrower than the viewport—a media query would over-size the labels.

**Decoupled Sizing**: Key label size depends on the keypad size, not arbitrary viewport breakpoints. Adding a sidebar or resizing panels doesn't require changing media queries.

**Nesting Safety**: If the calculator were ever embedded in a larger application or multiple times on a page, container queries would still scale correctly relative to their context.

**Graceful Fallbacks**: `clamp()` ensures labels never exceed a comfortable size (`2.125rem` for standard keys, `2.875rem` for operators) even if the container is very wide, and never shrink below `1rem`.

## Implementation

Container queries are defined in `src/index.css`:

```css
@container {
  /* Styles applied relative to container width */
}
```

The keyboard layout uses a container context:

```jsx
<div className="@container grid grid-cols-4 gap-3">
  {KEYS.map((key) => (
    <button
      className={keyVariants({ kind: key.kind, span: key.span })}
    >
      <span className={keyLabelVariants({...})}>{key.label}</span>
    </button>
  ))}
</div>
```

CSS clamping in `Calculator.tsx`:

```typescript
const KEY_TYPE_SIZE = 'text-[clamp(1rem,9.4cqw,2.125rem)]'
const OPERATOR_TYPE_SIZE = 'text-[clamp(1.25rem,12.8cqw,2.875rem)]'
```

## Consequences

**Positive:**
- Responsive sizing that works across all device aspect ratios
- No arbitrary breakpoint values to maintain
- Scales correctly in any layout context (embedded, multiple instances, etc.)
- Key labels never outgrow their buttons

**Negative:**
- Container queries are a newer CSS feature (limited browser support pre-2022)
- Debugging responsive behavior requires understanding container widths, not viewport widths
- `cqw` units are less intuitive than percentage or viewport-relative units
- Browser DevTools container query support is still developing

## Notes

The keypad's width is also capped by viewport height using CSS custom properties:

```css
--keypad-width: min(24rem, 56dvh);
```

This ensures that in landscape mode, the keypad doesn't become taller than the viewport allows. The container query measurements are made against this final width, ensuring consistency.

The `clamp()` function ensures graceful degradation: the middle value (the responsive size in `cqw`) only applies when it falls between the min and max. If the container is very small, `1rem` is used; if very large, `2.125rem` is used.

Modern browsers (Chrome 105+, Safari 16+, Firefox 110+) support container queries, covering the vast majority of users. Older browsers fall back to the max value (`2.125rem` or `2.875rem`), resulting in larger-than-optimal labels but still functional.
