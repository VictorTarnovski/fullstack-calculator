import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import Calculator from './Calculator'

vi.mock('@/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api')>()
  return { ...actual, evaluate: vi.fn() }
})

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

import { ApiError, evaluate } from '@/lib/api'
import { toast } from 'sonner'

function readout() {
  return screen.getByRole('status')
}

/** The exact display text, since toHaveTextContent matches substrings and
    would let e.g. "105" satisfy an assertion that expects exactly "5". */
function display() {
  return readout().textContent
}

function press(name: string) {
  fireEvent.click(screen.getByRole('button', { name }))
}

describe('Calculator', () => {
  beforeEach(() => {
    vi.mocked(evaluate).mockReset()
    vi.mocked(toast.error).mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('starts showing the placeholder zero', () => {
    render(<Calculator />)

    expect(display()).toBe('0')
  })

  it('builds an expression from digit presses, replacing the placeholder', () => {
    render(<Calculator />)

    press('Seven')
    press('Eight')

    expect(display()).toBe('78')
  })

  it('groups digits for display as they are typed', () => {
    render(<Calculator />)

    for (const name of ['One', 'Two', 'Three', 'Four', 'Five', 'Six', 'Seven']) {
      press(name)
    }

    expect(display()).toBe('1,234,567')
  })

  it('appends an operator after a finished number', () => {
    render(<Calculator />)

    press('Seven')
    press('Add')

    expect(display()).toBe('7+')
  })

  it('replaces a trailing operator instead of stacking a second one', () => {
    render(<Calculator />)

    press('Seven')
    press('Add')
    press('Subtract')

    expect(display()).toBe('7−')
  })

  it('supports percent and decimal keys', () => {
    render(<Calculator />)

    press('One')
    press('Decimal point')
    press('Five')
    press('Percent')

    expect(display()).toBe('1.5%')
  })

  it('clears back to the placeholder on All clear', () => {
    render(<Calculator />)

    press('One')
    press('Two')
    press('All clear')

    expect(display()).toBe('0')
  })

  it('erases one character at a time and lands back on the placeholder', () => {
    render(<Calculator />)

    press('One')
    press('Two')
    press('Backspace')

    expect(display()).toBe('1')

    press('Backspace')

    expect(display()).toBe('0')
  })

  it('supports typing through the keyboard, including ASCII operator aliases', () => {
    render(<Calculator />)

    fireEvent.keyDown(window, { key: '7' })
    fireEvent.keyDown(window, { key: '/' })
    fireEvent.keyDown(window, { key: '3' })

    expect(display()).toBe('7÷3')
  })

  it('ignores keystrokes held with a modifier', () => {
    render(<Calculator />)

    fireEvent.keyDown(window, { key: '7', ctrlKey: true })

    expect(display()).toBe('0')
  })

  it('does not submit an incomplete expression', async () => {
    render(<Calculator />)

    press('One')
    press('Add')
    press('Equals')

    await waitFor(() => expect(evaluate).not.toHaveBeenCalled())
    expect(display()).toBe('1+')
  })

  it('submits a complete expression and shows the result', async () => {
    vi.mocked(evaluate).mockResolvedValue(10)
    render(<Calculator />)

    press('Seven')
    press('Add')
    press('Three')
    press('Equals')

    expect(evaluate).toHaveBeenCalledWith('7+3')
    await waitFor(() => expect(display()).toBe('10'))
  })

  it('starts a fresh expression when a digit follows a result', async () => {
    vi.mocked(evaluate).mockResolvedValue(10)
    render(<Calculator />)

    press('Seven')
    press('Add')
    press('Three')
    press('Equals')
    await waitFor(() => expect(display()).toBe('10'))

    press('Five')

    expect(display()).toBe('5')
  })

  it('carries a result forward when an operator follows it', async () => {
    vi.mocked(evaluate).mockResolvedValue(10)
    render(<Calculator />)

    press('Seven')
    press('Add')
    press('Three')
    press('Equals')
    await waitFor(() => expect(display()).toBe('10'))

    press('Add')

    expect(display()).toBe('10+')
  })

  it('shows a toast and keeps the expression editable when the backend rejects it', async () => {
    vi.mocked(evaluate).mockRejectedValue(new ApiError('division by zero', 422))
    render(<Calculator />)

    press('One')
    press('Divide')
    press('Zero')
    press('Equals')

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('division by zero'))
    expect(display()).toBe('1÷0')
  })

  it('ignores presses while a request is pending', async () => {
    let resolveEvaluate: (value: number) => void = () => {}
    vi.mocked(evaluate).mockReturnValue(
      new Promise((resolve) => {
        resolveEvaluate = resolve
      }),
    )
    render(<Calculator />)

    press('Seven')
    press('Add')
    press('Three')
    press('Equals')

    expect(readout()).toHaveAttribute('aria-busy', 'true')

    press('Nine')
    expect(display()).toBe('7+3')

    resolveEvaluate(10)
    await waitFor(() => expect(readout()).toHaveAttribute('aria-busy', 'false'))
  })

  it('supports pointer interaction end to end via userEvent', async () => {
    const user = userEvent.setup()
    vi.mocked(evaluate).mockResolvedValue(4)
    render(<Calculator />)

    await user.click(screen.getByRole('button', { name: 'Two' }))
    await user.click(screen.getByRole('button', { name: 'Add' }))
    await user.click(screen.getByRole('button', { name: 'Two' }))
    await user.click(screen.getByRole('button', { name: 'Equals' }))

    await waitFor(() => expect(display()).toBe('4'))
  })
})
