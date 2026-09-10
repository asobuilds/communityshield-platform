/**
 * Mock configuration.
 *
 * Kept dependency-free (no MSW import) so that anything may safely read
 * `USE_MOCKS` / the demo accounts without pulling the mock worker into the
 * production bundle.
 *
 * The worker itself is dynamically imported in `src/main.tsx`, only when
 * `VITE_USE_MOCKS === 'true'`.
 */

export const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'

/** Deterministic pseudo-UUIDs so the demo data is stable between reloads. */
export function mockUid(prefix: string, index: number): string {
  return `${prefix}-0000-4000-8000-${String(index).padStart(12, '0')}`
}

export interface MockAccount {
  email: string
  password: string
  role: string
  userId: string
}

/** One-tap sign-in targets offered on the login screen while mocking. */
export const MOCK_ACCOUNTS: MockAccount[] = [
  { email: 'officer@shield.ng', password: 'password', role: 'Officer', userId: mockUid('aaaa1111', 1) },
  { email: 'admin@shield.ng', password: 'password', role: 'Unit admin', userId: mockUid('aaaa1111', 2) },
  { email: 'citizen@shield.ng', password: 'password', role: 'Citizen', userId: mockUid('aaaa1111', 3) },
  { email: 'super@shield.ng', password: 'password', role: 'Super admin', userId: mockUid('aaaa1111', 4) },
]

export function mockTokenFor(userId: string): string {
  return `mock.${userId}`
}

export function userIdFromToken(token: string | null): string | null {
  if (!token?.startsWith('mock.')) return null
  return token.slice('mock.'.length) || null
}
