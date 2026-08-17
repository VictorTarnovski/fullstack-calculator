import { render } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useFitText } from './use-fit-text'

type Layout = { client: number; scroll: number }

const layouts = new WeakMap<Element, Layout>()

function setLayout(element: Element, layout: Layout) {
  layouts.set(element, layout)
}

let resizeCallbacks: Array<() => void> = []

class MockResizeObserver {
  constructor(callback: () => void) {
    resizeCallbacks.push(callback)
  }

  observe() {}
  unobserve() {}
  disconnect() {}
}

function triggerResize() {
  for (const callback of resizeCallbacks) callback()
}

function Readout({ content }: { content: string }) {
  const ref = useFitText<HTMLOutputElement>(content)
  return <output ref={ref}>{content}</output>
}

describe('useFitText', () => {
  beforeEach(() => {
    resizeCallbacks = []
    vi.stubGlobal('ResizeObserver', MockResizeObserver)

    Object.defineProperty(HTMLElement.prototype, 'clientWidth', {
      configurable: true,
      get() {
        return layouts.get(this)?.client ?? 0
      },
    })
    Object.defineProperty(HTMLElement.prototype, 'scrollWidth', {
      configurable: true,
      get() {
        return layouts.get(this)?.scroll ?? 0
      },
    })
    vi.spyOn(window, 'getComputedStyle').mockReturnValue({ fontSize: '32px' } as CSSStyleDeclaration)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('leaves the font size alone when the content already fits', () => {
    const { container } = render(<Readout content="42" />)
    const output = container.querySelector('output') as HTMLElement
    setLayout(output, { client: 300, scroll: 100 })

    triggerResize()

    expect(output.style.fontSize).toBe('')
    expect(output.style.whiteSpace).toBe('nowrap')
  })

  it('shrinks the font size proportionally when the content overflows', () => {
    const { container } = render(<Readout content="1234567890" />)
    const output = container.querySelector('output') as HTMLElement
    setLayout(output, { client: 100, scroll: 400 })

    triggerResize()

    /* One proportional step: 32px * (100/400) = 8px, floored at MIN_FONT_SIZE_PX (16px). */
    expect(output.style.fontSize).toBe('16px')
  })

  it('wraps instead of shrinking below the minimum font size', () => {
    const { container } = render(<Readout content="a very long expression" />)
    const output = container.querySelector('output') as HTMLElement
    /* scrollWidth never drops below client width no matter how much the
       stubbed measurement "shrinks" the font, so the element keeps
       overflowing through both fitting passes. */
    setLayout(output, { client: 100, scroll: 1000 })

    triggerResize()

    expect(output.style.whiteSpace).toBe('normal')
  })

  it('does nothing when the element has not been laid out yet', () => {
    const { container } = render(<Readout content="42" />)
    const output = container.querySelector('output') as HTMLElement
    setLayout(output, { client: 0, scroll: 500 })

    triggerResize()

    expect(output.style.fontSize).toBe('')
  })

  it('refits when the content changes', () => {
    const { container, rerender } = render(<Readout content="1" />)
    const output = container.querySelector('output') as HTMLElement
    setLayout(output, { client: 100, scroll: 400 })

    rerender(<Readout content="1234567890" />)

    expect(output.style.fontSize).toBe('16px')
  })
})
