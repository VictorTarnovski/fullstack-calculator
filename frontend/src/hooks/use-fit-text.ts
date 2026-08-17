import { useLayoutEffect, useRef } from 'react'

/** Below this the readout stops shrinking and wraps instead: a number too small
    to read is not a better answer than one on two lines. */
const MIN_FONT_SIZE_PX = 16

/**
 * useFitText shrinks an element's type until its content fits the width it has
 * been given, so a long expression is never cut off with an ellipsis.
 *
 * The size is written to the element's style rather than held in state, since a
 * measurement that fed a re-render would re-run the layout it just measured.
 * Clearing the style first lets the stylesheet keep owning the ceiling — the
 * hook only ever scales down from whatever size the CSS resolves to.
 */
export function useFitText<T extends HTMLElement>(content: string) {
  const ref = useRef<T>(null)

  useLayoutEffect(() => {
    const element = ref.current
    if (!element) return

    const fit = () => {
      element.style.fontSize = ''
      element.style.whiteSpace = 'nowrap'

      const available = element.clientWidth
      if (available === 0) return

      let size = Number.parseFloat(getComputedStyle(element).fontSize)

      /* Text width scales linearly with font size, so one proportional step
         lands within a pixel of fitting; the second absorbs its rounding. */
      for (let pass = 0; pass < 2 && element.scrollWidth > available; pass++) {
        size = Math.max(MIN_FONT_SIZE_PX, size * (available / element.scrollWidth))
        element.style.fontSize = `${size}px`
      }

      element.style.whiteSpace = element.scrollWidth > available ? 'normal' : 'nowrap'
    }

    fit()

    /* The ceiling is a vw-based clamp, so the fit is redone whenever the
       viewport changes. The parent is observed rather than the element:
       observing the element would see the height change that resizing its own
       type causes, and loop. */
    const observer = new ResizeObserver(fit)
    const parent = element.parentElement

    if (parent) observer.observe(parent)

    return () => observer.disconnect()
  }, [content])

  return ref
}
