/** Every successful response carries its payload under the same key. */
export type Envelope<T> = {
  data: T
}

/** RFC 9457 problem detail, the body of every failed response. */
type ProblemDetail = {
  type: string
  title: string
  status: number
  detail?: string
  instance?: string
}

export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

const BASE_URL = import.meta.env.VITE_API_URL ?? ''

/* detail is the more specific field and is present on every classified error;
   a proxy error page or dropped connection leaves both missing, hence the
   fallback. */
function messageFrom(body: unknown, status: number): string {
  const problem = body as Partial<ProblemDetail> | null

  for (const field of [problem?.detail, problem?.title] as const) {
    if (typeof field === 'string' && field.trim()) {
      return field
    }
  }

  return `Request failed with status ${status}`
}

/**
 * evaluate sends the expression as the raw request body rather than a field in
 * a JSON object, because the backend parses straight from the request stream.
 * The charset is stated explicitly: the operators are Unicode math characters,
 * so any other encoding reaches the lexer as unrecognized runes.
 */
export async function evaluate(expression: string, signal?: AbortSignal): Promise<number> {
  const response = await fetch(`${BASE_URL}/evaluations`, {
    method: 'POST',
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    body: expression,
    signal,
  })

  const text = await response.text()

  let body: unknown = null
  try {
    body = text.trim() ? JSON.parse(text) : null
  } catch {
    body = null
  }

  if (!response.ok) {
    throw new ApiError(messageFrom(body, response.status), response.status)
  }

  const result = (body as Envelope<{ result: number }> | null)?.data?.result

  if (typeof result !== 'number') {
    throw new ApiError('Malformed response from the server', response.status)
  }

  return result
}
