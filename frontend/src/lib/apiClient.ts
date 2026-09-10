/**
 * Typed fetch client for the CommunityShield API.
 *
 * - Injects `Authorization: Bearer <token>` when a session exists.
 * - Normalises errors into `ApiError`.
 * - On 401 it clears the session and emits `cs:unauthorized` so AuthContext can
 *   route the user back to login (see src/auth/AuthContext.tsx).
 */

const TOKEN_KEY = 'cs.token'

function resolveBase(): string {
  const raw = (import.meta.env.VITE_API_URL ?? '').trim()
  if (!raw) return '/api/v1'
  const origin = raw.replace(/\/+$/, '')
  return origin.endsWith('/api/v1') ? origin : `${origin}/api/v1`
}

export const API_BASE = resolveBase()

export const UNAUTHORIZED_EVENT = 'cs:unauthorized'

export class ApiError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(status: number, message: string, body?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }

  /** True when the failure is a connectivity problem rather than an HTTP error. */
  static isNetwork(error: unknown): boolean {
    return error instanceof ApiError && error.status === 0
  }
}

export const tokenStore = {
  get(): string | null {
    try {
      return localStorage.getItem(TOKEN_KEY)
    } catch {
      return null
    }
  },
  set(token: string): void {
    try {
      localStorage.setItem(TOKEN_KEY, token)
    } catch {
      /* storage unavailable (private mode) — session stays in memory only */
    }
  },
  clear(): void {
    try {
      localStorage.removeItem(TOKEN_KEY)
    } catch {
      /* ignore */
    }
  },
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  signal?: AbortSignal
  /** Skip auth header + 401 handling (used by login). */
  anonymous?: boolean
}

async function extractError(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string; message?: string }
    return data.error ?? data.message ?? response.statusText
  } catch {
    return response.statusText || `Request failed (${response.status})`
  }
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, signal, anonymous = false } = options

  const headers: Record<string, string> = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  const token = tokenStore.get()
  if (token && !anonymous) headers.Authorization = `Bearer ${token}`

  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      signal,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') throw cause
    throw new ApiError(0, 'Network unavailable — check your connection.', cause)
  }

  if (response.status === 401 && !anonymous) {
    tokenStore.clear()
    // Guarded so the client also works outside a DOM (tests, future SSR).
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent(UNAUTHORIZED_EVENT))
    }
    throw new ApiError(401, await extractError(response))
  }

  if (!response.ok) {
    throw new ApiError(response.status, await extractError(response))
  }

  if (response.status === 204) return undefined as T

  const text = await response.text()
  if (!text) return undefined as T
  return JSON.parse(text) as T
}

/**
 * `postAnonymous` exists for sign-in: a login request must not carry a stale
 * bearer token, and a 401 from *bad credentials* must not tear down a session
 * the user already has — only an authenticated request can do that.
 */
export const api = {
  get: <T>(path: string, signal?: AbortSignal) => request<T>(path, { method: 'GET', signal }),
  post: <T>(path: string, body?: unknown) => request<T>(path, { method: 'POST', body }),
  postAnonymous: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body, anonymous: true }),
  put: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PATCH', body }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
