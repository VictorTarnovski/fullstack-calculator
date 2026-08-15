import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError, evaluate } from './api'

function jsonResponse(status: number, body: unknown) {
  return {
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(JSON.stringify(body)),
  }
}

describe('evaluate', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('sends the expression as a plain-text request body', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { data: { result: 10 } }) as Response)

    await evaluate('7+3')

    expect(fetch).toHaveBeenCalledWith(
      '/evaluations',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'text/plain; charset=utf-8' },
        body: '7+3',
      }),
    )
  })

  it('forwards an abort signal to fetch', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { data: { result: 10 } }) as Response)
    const controller = new AbortController()

    await evaluate('7+3', controller.signal)

    expect(fetch).toHaveBeenCalledWith('/evaluations', expect.objectContaining({ signal: controller.signal }))
  })

  it('resolves with the numeric result on success', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { data: { result: 42 } }) as Response)

    await expect(evaluate('40+2')).resolves.toBe(42)
  })

  it('throws an ApiError using the problem detail field when present', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(400, { type: 'about:blank', title: 'Bad Request', status: 400, detail: 'division by zero' }) as Response,
    )

    await expect(evaluate('1÷0')).rejects.toMatchObject({
      name: 'ApiError',
      message: 'division by zero',
      status: 400,
    })
  })

  it('falls back to the title when detail is missing', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(422, { type: 'about:blank', title: 'Unprocessable Entity', status: 422 }) as Response,
    )

    await expect(evaluate('1+')).rejects.toMatchObject({ message: 'Unprocessable Entity', status: 422 })
  })

  it('falls back to a generic message when the body has neither field', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      status: 502,
      text: () => Promise.resolve(''),
    } as Response)

    await expect(evaluate('1+1')).rejects.toMatchObject({ message: 'Request failed with status 502', status: 502 })
  })

  it('falls back to a generic message when the error body is not JSON', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      status: 502,
      text: () => Promise.resolve('<html>Bad Gateway</html>'),
    } as Response)

    await expect(evaluate('1+1')).rejects.toMatchObject({ message: 'Request failed with status 502', status: 502 })
  })

  it('rejects with ApiError when the success body has no numeric result', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { data: {} }) as Response)

    await expect(evaluate('7+3')).rejects.toMatchObject({
      name: 'ApiError',
      message: 'Malformed response from the server',
    })
  })

  it('is an instance of ApiError and Error', async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(500, {}) as Response)

    try {
      await evaluate('7+3')
      expect.unreachable('evaluate should have thrown')
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(Error)
    }
  })
})
