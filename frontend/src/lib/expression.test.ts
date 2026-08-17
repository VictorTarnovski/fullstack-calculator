import { describe, expect, it } from 'vitest'

import {
  append,
  backspace,
  forDisplay,
  isOperator,
  isSubmittable,
  toExpressionText,
} from './expression'

describe('isOperator', () => {
  it.each(['+', '−', '×', '÷'])('accepts %s', (label) => {
    expect(isOperator(label)).toBe(true)
  })

  it.each(['5', '.', '%', '-', '*', '/', ''])('rejects %s', (label) => {
    expect(isOperator(label)).toBe(false)
  })
})

describe('append', () => {
  it('adds a digit to the expression', () => {
    expect(append('7', '8')).toBe('78')
  })

  it('adds an operator after a finished number', () => {
    expect(append('7', '+')).toBe('7+')
  })

  it('replaces a trailing operator instead of stacking them', () => {
    expect(append('7+', '−')).toBe('7−')
  })

  it('drops a dangling decimal point before an operator', () => {
    expect(append('7.', '×')).toBe('7×')
  })

  it('appends percent after a finished number', () => {
    expect(append('5', '%')).toBe('5%')
  })

  it('refuses percent after an infix operator', () => {
    expect(append('5+', '%')).toBe('5+')
  })

  it('refuses percent after a dangling decimal point', () => {
    expect(append('5.', '%')).toBe('5.')
  })

  it('allows percent after another percent', () => {
    expect(append('5%', '%')).toBe('5%%')
  })

  it('adds a decimal point to a number with none yet', () => {
    expect(append('7', '.')).toBe('7.')
  })

  it('refuses a second decimal point in the same operand', () => {
    expect(append('1.5', '.')).toBe('1.5')
  })

  it('allows a decimal point in a fresh operand after an operator', () => {
    expect(append('1+2', '.')).toBe('1+2.')
  })

  it('refuses a digit right after percent', () => {
    expect(append('5%', '3')).toBe('5%')
  })

  it('refuses a decimal point right after percent', () => {
    expect(append('5%', '.')).toBe('5%')
  })

  it('allows an operator right after percent', () => {
    expect(append('5%', '+')).toBe('5%+')
  })
})

describe('backspace', () => {
  it('removes the last character', () => {
    expect(backspace('123')).toBe('12')
  })

  it('removes the last character of an expression with an operator', () => {
    expect(backspace('1+2')).toBe('1+')
  })

  it('returns empty when nothing is left', () => {
    expect(backspace('1')).toBe('')
  })

  it('returns empty on an already-empty expression', () => {
    expect(backspace('')).toBe('')
  })

  it('collapses a bare sign left behind by deleting into a negative result', () => {
    expect(backspace('−5')).toBe('')
  })
})

describe('isSubmittable', () => {
  it('accepts a complete binary expression', () => {
    expect(isSubmittable('1+2')).toBe(true)
  })

  it('accepts a lone number', () => {
    expect(isSubmittable('42')).toBe(true)
  })

  it('accepts a trailing percent', () => {
    expect(isSubmittable('5%')).toBe(true)
  })

  it('accepts a chained percent', () => {
    expect(isSubmittable('5%%')).toBe(true)
  })

  it('rejects the empty expression', () => {
    expect(isSubmittable('')).toBe(false)
  })

  it('rejects a leading operator', () => {
    expect(isSubmittable('+1')).toBe(false)
  })

  it('rejects a leading decimal point', () => {
    expect(isSubmittable('.1')).toBe(false)
  })

  it('rejects a trailing operator', () => {
    expect(isSubmittable('1+')).toBe(false)
  })

  it('rejects a trailing decimal point', () => {
    expect(isSubmittable('1.')).toBe(false)
  })

  it('rejects text outside the keypad alphabet, such as exponential notation', () => {
    expect(isSubmittable('1e-21')).toBe(false)
  })
})

describe('toExpressionText', () => {
  it('renders a positive number as-is', () => {
    expect(toExpressionText(5)).toBe('5')
  })

  it('renders a negative number with the Unicode minus sign', () => {
    expect(toExpressionText(-5)).toBe('−5')
  })

  it('renders a decimal result', () => {
    expect(toExpressionText(1.5)).toBe('1.5')
  })

  it('only replaces the leading sign, leaving an exponent hyphen alone', () => {
    expect(toExpressionText(-1e21)).toBe('−1e+21')
  })
})

describe('forDisplay', () => {
  it('groups a large integer', () => {
    expect(forDisplay('1234567')).toBe('1,234,567')
  })

  it('groups the integer part of a decimal, leaving the fraction alone', () => {
    expect(forDisplay('1234567.89')).toBe('1,234,567.89')
  })

  it('groups every operand in a full expression', () => {
    expect(forDisplay('1000+234567')).toBe('1,000+234,567')
  })

  it('leaves small numbers ungrouped', () => {
    expect(forDisplay('123')).toBe('123')
  })

  it('leaves a percent operand grouped up to the percent sign', () => {
    expect(forDisplay('1000%')).toBe('1,000%')
  })

  it('leaves exponential notation unformatted', () => {
    expect(forDisplay('1e+38')).toBe('1e+38')
  })
})
