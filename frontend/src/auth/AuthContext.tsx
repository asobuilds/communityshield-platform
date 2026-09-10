import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { api, tokenStore, UNAUTHORIZED_EVENT } from '@/lib/apiClient'
import type { LoginResponse, ProfileResponse, Role, User } from '@/types/api'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

interface AuthContextValue {
  user: User | null
  role: Role | null
  status: AuthStatus
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [status, setStatus] = useState<AuthStatus>('loading')

  const logout = useCallback(() => {
    tokenStore.clear()
    setUser(null)
    setStatus('anonymous')
  }, [])

  // Restore a session on boot if a token exists.
  useEffect(() => {
    let cancelled = false

    async function restore() {
      if (!tokenStore.get()) {
        setStatus('anonymous')
        return
      }
      try {
        const { user: fresh } = await api.get<ProfileResponse>('/auth/profile')
        if (!cancelled) {
          setUser(fresh)
          setStatus('authenticated')
        }
      } catch {
        if (!cancelled) logout()
      }
    }

    void restore()
    return () => {
      cancelled = true
    }
  }, [logout])

  // Any 401 from the API client tears the session down.
  useEffect(() => {
    const onUnauthorized = () => logout()
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
  }, [logout])

  const login = useCallback(async (email: string, password: string) => {
    // Anonymous: sign-in must not send a stale token, and a rejected login must
    // not clear an existing session (see api.postAnonymous).
    const result = await api.postAnonymous<LoginResponse>('/auth/login', { email, password })
    tokenStore.set(result.token)
    // Login returns a partial user; fetch the full profile for status/timestamps.
    try {
      const { user: fresh } = await api.get<ProfileResponse>('/auth/profile')
      setUser(fresh)
    } catch {
      setUser(result.user as User)
    }
    setStatus('authenticated')
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({ user, role: user?.role ?? null, status, login, logout }),
    [user, status, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within <AuthProvider>')
  return ctx
}
