import { cva, type VariantProps } from 'class-variance-authority'
import { useEffect, useState } from 'react'

import { toast } from 'sonner'

import { useFitText } from '@/hooks/use-fit-text'
import { evaluate } from '@/lib/api'
import {
  append as appendToExpression,
  backspace as backspaceExpression,
  forDisplay,
  isSubmittable,
  toExpressionText,
} from '@/lib/expression'

/* Each kind declares its size exactly once, so no two font-size utilities land
   on the same element and compete.

   Sized in cqw against the keypad, not vw against the viewport: the keypad's
   width is capped by height as well (58dvh), so on a landscape phone it is far
   narrower than the window. Measured against vw the labels would outgrow the
   keys holding them, and each row would take its height from the type rather
   than the key — which is what pushed the bottom row off the screen. */
const KEY_TYPE_SIZE = 'text-[clamp(1rem,9.4cqw,2.125rem)]'
const OPERATOR_TYPE_SIZE = 'text-[clamp(1.25rem,12.8cqw,2.875rem)]'

const keyVariants = cva(
  [
    'flex touch-manipulation items-center justify-center rounded-full',
    'font-normal leading-none select-none transition-colors duration-100',
    'focus-visible:outline-2 focus-visible:outline-offset-2',
    'focus-visible:outline-focus-ring',
  ],
  {
    variants: {
      /* Named for the backend's token kinds: an atom key contributes to a
         number literal, an operator key to an OPERATOR token. Function keys
         act on the expression itself and never reach the lexer. */
      kind: {
        atom: [
          'bg-key-atom text-key-atom-label active:bg-key-atom-pressed',
          KEY_TYPE_SIZE,
        ],
        function: [
          'bg-key-function text-key-function-label',
          'active:bg-key-function-pressed',
          KEY_TYPE_SIZE,
        ],
        operator: [
          'bg-key-operator text-key-operator-label',
          'active:bg-key-operator-pressed',
          OPERATOR_TYPE_SIZE,
        ],
      },
      span: {
        one: 'aspect-square',
        two: 'col-span-2',
      },
    },
    defaultVariants: {
      span: 'one',
    },
  },
)

/* Centring a flex item centres its line box, not the glyph inside it, so every
   label lands low by the gap between the baseline and the box centre. These are
   measured ink offsets, applied to the label rather than the button so the
   circle itself never moves. */
const keyLabelVariants = cva('block', {
  variants: {
    optical: {
      /** Cap-height glyphs resting on the baseline: digits, AC, %. */
      cap: 'translate-y-[-0.06em]',
      /** Math symbols, whose smaller ink sits well below cap height. */
      math: 'translate-y-[-0.12em]',
      /** Left on the baseline: a period belongs low, not floating mid-circle. */
      baseline: null,
      /* The erase glyph sits at cap height, but is the one label whose box
         being centred does not centre what you see: it is a wedge tapering to
         a point on the left against a full-height right edge, so its ink
         weight lands 0.083em right of its box. That is its measured centroid,
         which every other glyph here already has within 0.02em of zero. */
      erase: 'translate-x-[-0.083em] translate-y-[-0.06em]',
    },
    span: {
      one: null,
      /* Half the pill minus half the 12px gap is exactly one key wide, and
         mr-auto pins it left, centring the glyph over the column of 1 / 4 / 7
         above. */
      two: 'mr-auto w-[calc(50%-0.375rem)] text-center',
    },
  },
  defaultVariants: {
    span: 'one',
  },
})

type KeyVariants = VariantProps<typeof keyVariants>
type KeyLabelVariants = VariantProps<typeof keyLabelVariants>

type KeyKind = NonNullable<KeyVariants['kind']>
type Optical = NonNullable<KeyLabelVariants['optical']>

/** The nudge each kind takes unless a key overrides it. */
const OPTICAL_BY_KIND: Record<KeyKind, Optical> = {
  atom: 'cap',
  function: 'cap',
  operator: 'math',
}

/* How an 'append' key joins what is already there depends on its kind, which is
   why action and kind stay separate fields. */
type KeyAction = 'append' | 'backspace' | 'clear' | 'evaluate'

interface CalculatorKey {
  /** Glyph as it appears on the key, and what an 'append' key contributes. */
  label: string
  /** Accessible name, since most glyphs read poorly to a screen reader. */
  name: string
  kind: KeyKind
  action: KeyAction
  /* KeyboardEvent.key values that press this key. The numpad needs no entries
     of its own: it reports the same characters as the top row. */
  keystrokes?: string[]
  span?: KeyVariants['span']
  /** Set only where the glyph's ink sits differently from its kind. */
  optical?: Optical
}

/** Key sequence in reading order: five rows of four, left to right. */
const KEYS: CalculatorKey[] = [
  /* No keystroke: clearing everything is too destructive to sit under a key
     that could be hit by accident. */
  { label: 'AC', name: 'All clear', kind: 'function', action: 'clear' },
  {
    label: '⌫',
    name: 'Backspace',
    kind: 'function',
    action: 'backspace',
    keystrokes: ['Backspace'],
    optical: 'erase',
  },
  { label: '%', name: 'Percent', kind: 'function', action: 'append', keystrokes: ['%'] },
  /* The ASCII operators a keyboard actually has, mapped onto the Unicode ones
     the grammar accepts: typing '/' puts '÷' in the readout. */
  { label: '÷', name: 'Divide', kind: 'operator', action: 'append', keystrokes: ['/'] },

  { label: '7', name: 'Seven', kind: 'atom', action: 'append', keystrokes: ['7'] },
  { label: '8', name: 'Eight', kind: 'atom', action: 'append', keystrokes: ['8'] },
  { label: '9', name: 'Nine', kind: 'atom', action: 'append', keystrokes: ['9'] },
  { label: '×', name: 'Multiply', kind: 'operator', action: 'append', keystrokes: ['*'] },

  { label: '4', name: 'Four', kind: 'atom', action: 'append', keystrokes: ['4'] },
  { label: '5', name: 'Five', kind: 'atom', action: 'append', keystrokes: ['5'] },
  { label: '6', name: 'Six', kind: 'atom', action: 'append', keystrokes: ['6'] },
  { label: '−', name: 'Subtract', kind: 'operator', action: 'append', keystrokes: ['-'] },

  { label: '1', name: 'One', kind: 'atom', action: 'append', keystrokes: ['1'] },
  { label: '2', name: 'Two', kind: 'atom', action: 'append', keystrokes: ['2'] },
  { label: '3', name: 'Three', kind: 'atom', action: 'append', keystrokes: ['3'] },
  { label: '+', name: 'Add', kind: 'operator', action: 'append', keystrokes: ['+'] },

  {
    label: '0',
    name: 'Zero',
    kind: 'atom',
    action: 'append',
    keystrokes: ['0'],
    span: 'two',
  },
  {
    label: '.',
    name: 'Decimal point',
    kind: 'atom',
    action: 'append',
    keystrokes: ['.'],
    optical: 'baseline',
  },
  {
    label: '=',
    name: 'Equals',
    kind: 'operator',
    action: 'evaluate',
    keystrokes: ['Enter', '='],
  },
]

const KEY_BY_KEYSTROKE = new Map(
  KEYS.flatMap((key) => (key.keystrokes ?? []).map((stroke) => [stroke, key] as const)),
)

type Phase =
  /** An expression being built. The only phase the readout is editable in. */
  | 'editing'
  /** A value returned by the backend. A digit starts over; an operator carries
      the value into a new expression, the way a calculator is expected to. */
  | 'result'
  /** A request is in flight; further presses are ignored until it settles. */
  | 'pending'

const EMPTY = '0'

export default function Calculator() {
  /* readout holds the expression as the grammar accepts it; the grouped form is
     only ever read, never sent. */
  const [readout, setReadout] = useState(EMPTY)
  const [phase, setPhase] = useState<Phase>('editing')
  const display = forDisplay(readout)
  const readoutRef = useFitText<HTMLOutputElement>(display)

  /* An operator carries a result forward; a digit after one starts fresh. */
  function baseFor(kind: KeyKind): string {
    if (phase === 'result') return kind === 'atom' ? EMPTY : readout
    return readout
  }

  function append(key: CalculatorKey) {
    const base = baseFor(key.kind)

    /* The leading zero is a placeholder rather than typed input, so a digit
       replaces it outright. Everything else goes through the grammar rules. */
    const next =
      base === EMPTY && key.kind === 'atom' && key.label !== '.'
        ? key.label
        : appendToExpression(base, key.label)

    setReadout(next)
    setPhase('editing')
  }

  function clear() {
    setReadout(EMPTY)
    setPhase('editing')
  }

  /* Deleting from a result turns it back into something being typed. */
  function erase() {
    const shortened = backspaceExpression(readout)

    setReadout(shortened === '' ? EMPTY : shortened)
    setPhase('editing')
  }

  /* Only the errors the keypad cannot prevent reach here: division by zero, and
     anything unmapped. Both go to a toast so the expression stays on screen to
     be corrected rather than retyped. */
  async function submit() {
    const expression = readout

    setPhase('pending')

    try {
      const result = await evaluate(expression)
      setReadout(toExpressionText(result))
      setPhase('result')
    } catch (error) {
      setReadout(expression)
      setPhase('editing')
      toast.error(error instanceof Error ? error.message : 'Something went wrong')
    }
  }

  /* No dependency array: the listener is replaced after every render and always
     closes over the current readout, which is cheaper than the memoisation a
     stable handler would need to stay correct. */
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      /* A modifier means the keystroke belongs to the browser — Ctrl+R is a
         reload, not a subtraction. */
      if (event.ctrlKey || event.metaKey || event.altKey) return

      const key = KEY_BY_KEYSTROKE.get(event.key)
      if (!key) return

      /* Claims the keystroke: '/' opens quick-find in Firefox, and Enter would
         otherwise also activate the focused key, firing two presses for one. */
      event.preventDefault()

      press(key)
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  })

  /* A Record rather than a switch, so adding to KeyAction fails to compile
     until it has a handler here instead of falling through in silence. */
  const ACTIONS: Record<KeyAction, (key: CalculatorKey) => void> = {
    append,
    backspace: erase,
    clear,
    /* An incomplete expression is never sent: the backend would only answer
       with a problem detail restating what the keypad already knows. */
    evaluate: () => {
      if (isSubmittable(readout)) void submit()
    },
  }

  function press(key: CalculatorKey) {
    if (phase === 'pending') return

    ACTIONS[key.action](key)
  }

  return (
    /* Edge to edge on a phone held upright, a bounded panel as soon as there is
       width for it, so the keypad is held by something instead of floating in
       an open field. */
    <div className="bg-surface border-surface-edge flex min-h-dvh w-full flex-col items-center panel:h-[var(--panel-height)] panel:max-w-[var(--panel-width)] panel:min-h-0 panel:rounded-[2.5rem] panel:border panel:shadow-panel">
      {/* Width is capped so keys never grow past a comfortable size, and tracks
          viewport height so all five rows stay on screen in short windows. */}
      <div className="flex w-full max-w-[var(--keypad-width)] flex-1 flex-col px-3 pt-[max(1rem,env(safe-area-inset-top))] pb-[max(0.75rem,env(safe-area-inset-bottom))] panel:pb-[min(1.75rem,4dvh)]">
        {/* The keypad sits at the bottom; the display hangs just above it. */}
        <div className="flex-1" />

        {/* The type size is a ceiling, not a fixed size: useFitText scales it
            down so a long expression stays whole. The dvh term is what keeps a
            landscape phone's bottom row on screen — sized on width alone, the
            readout would hold its full 5rem in a viewport 375px tall. */}
        <output
          ref={readoutRef}
          aria-live="polite"
          aria-busy={phase === 'pending'}
          className="text-readout block overflow-hidden pr-1 pb-2 text-right break-all text-[clamp(2rem,min(20vw,14dvh),5rem)] leading-tight font-light tracking-tight tabular-nums"
        >
          {display}
        </output>

        {/* The container the key type is measured against. */}
        <div className="@container grid grid-cols-4 gap-3">
          {KEYS.map((key) => (
            <button
              key={key.label}
              type="button"
              aria-label={key.name}
              onClick={() => press(key)}
              className={keyVariants({ kind: key.kind, span: key.span })}
            >
              <span
                className={keyLabelVariants({
                  optical: key.optical ?? OPTICAL_BY_KIND[key.kind],
                  span: key.span,
                })}
              >
                {key.label}
              </span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
