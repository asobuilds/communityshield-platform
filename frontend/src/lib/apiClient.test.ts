import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError, api, tokenStore } from './apiClient'

/**
 * The API client is the seam between the UI and every backend call, so its
 * error normalisation and auth-header behaviour are pinned here.
 */

function jsonResponse(body: unknown, init: { status?: number; statusText?: string } = {}) {
  const status = init.status ?? 200
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: init.statusText ?? '',
    text: async () => JSON.stringify(body),
    json: async () => body,
  } as unknown as Response
}

let store: Record<string, string>

beforeEach(() => {
  store = {}
  // tokenStore prefers localStorage; give it one so the token path is exercised.
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value
    },
    removeItem: (key: string) => {
      delete store[key]
    },
  })
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('api client', () => {
  it('requests the versioned API path and parses JSON', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ cases: [] }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await api.get<{ cases: unknown[] }>('/cases')

    expect(result).toEqual({ cases: [] })
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(String(fetchMock.mock.calls[0][0])).toMatch(/\/api\/v1\/cases$/)
  })

  it('attaches the bearer token when a session exists', async () => {
    tokenStore.set('token-123')
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ cases: [] }))
    vi.stubGlobal('fetch', fetchMock)

    await api.get('/cases')

    const init = fetchMock.mock.calls[0][1] as RequestInit
    expect((init.headers as Record<string, string>).Authorization).toBe('Bearer token-123')
  })

  it('omits the auth header for anonymous requests', async () => {
    tokenStore.set('token-123')
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ token: 'x' }))
    vi.stubGlobal('fetch', fetchMock)

    await api.postAnonymous('/auth/login', { email: 'a@b.c', password: 'x' })

    const init = fetchMock.mock.calls[0][1] as RequestInit
    expect((init.headers as Record<string, string>).Authorization).toBeUndefined()
  })

  it('keeps the session when an anonymous request is rejected', async () => {
    // A failed sign-in must not log the user out of a session they already have.
    tokenStore.set('token-123')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse({ error: 'invalid credentials' }, { status: 401 })),
    )

    await expect(
      api.postAnonymous('/auth/login', { email: 'a@b.c', password: 'wrong' }),
    ).rejects.toMatchObject({ status: 401 })
    expect(tokenStore.get()).toBe('token-123')
  })

  it('normalises an error body into ApiError with its status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse({ error: 'case not found' }, { status: 404, statusText: 'Not Found' }),
      ),
    )

    await expect(api.get('/cases/does-not-exist')).rejects.toMatchObject({
      name: 'ApiError',
      status: 404,
      message: 'case not found',
    })
  })

  it('clears the session on 401', async () => {
    tokenStore.set('token-123')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse({ error: 'nope' }, { status: 401 })),
    )

    await expect(api.get('/cases')).rejects.toBeInstanceOf(ApiError)
    expect(tokenStore.get()).toBeNull()
  })

  it('reports an unreachable API as a network failure (status 0)', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('failed to fetch')))

    const error = await api.get('/cases').catch((cause: unknown) => cause)
    expect(ApiError.isNetwork(error)).toBe(true)
    expect((error as ApiError).status).toBe(0)
  })

  it('serialises a request body and sets the content type', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ message: 'ok' }))
    vi.stubGlobal('fetch', fetchMock)

    await api.post('/cases/1/progress', { action: 'note', description: 'checked' })

    const init = fetchMock.mock.calls[0][1] as RequestInit
    expect(init.method).toBe('POST')
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json')
    expect(JSON.parse(String(init.body))).toEqual({ action: 'note', description: 'checked' })
  })
})
