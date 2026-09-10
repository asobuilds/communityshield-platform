import { useState, type FormEvent } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { Shield } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Field, Input } from '@/components/ui/Field'
import { useToast } from '@/components/ui/Toast'
import { useAuth } from '@/auth/AuthContext'
import { homePathForRole } from '@/auth/RequireRole'
import { ApiError } from '@/lib/apiClient'
import { MOCK_ACCOUNTS, USE_MOCKS } from '@/mocks/config'

/**
 * Sign-in. When mocks are enabled (`VITE_USE_MOCKS=true`) the seeded demo
 * accounts are offered as one-tap fills so the workflows can be exercised
 * without a running API.
 */
export function LoginPage() {
  const { login, status, role } = useAuth()
  const location = useLocation()
  const { notify } = useToast()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (status === 'authenticated') {
    const from = (location.state as { from?: string } | null)?.from
    return <Navigate to={from || homePathForRole(role)} replace />
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await login(email.trim(), password)
      notify('Signed in.', 'success')
      // AuthContext flips `status`, and the redirect above takes over.
    } catch (cause) {
      setError(
        cause instanceof ApiError
          ? cause.message
          : 'Could not sign in. Check your details and try again.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="grid min-h-screen bg-base lg:grid-cols-2">
      {/* Brand panel */}
      <div className="hidden flex-col justify-between bg-surface p-10 lg:flex">
        <div className="flex items-center gap-2">
          <Shield className="size-7 text-signal" aria-hidden />
          <span className="text-base font-bold tracking-wide text-ink">COMMUNITYSHIELD</span>
        </div>
        <div>
          <h1 className="max-w-md text-3xl font-semibold leading-tight text-ink">
            Every report answered. Every case accounted for.
          </h1>
          <p className="mt-3 max-w-md text-sm text-ink-muted">
            CommunityShield connects citizens, security units and administrators on one operational
            record — from the first SOS to the final report.
          </p>
        </div>
        <p className="text-xs text-ink-faint">
          Authorised personnel only. Activity on this system is recorded.
        </p>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <div className="mb-6 flex items-center gap-2 lg:hidden">
            <Shield className="size-6 text-signal" aria-hidden />
            <span className="text-sm font-bold tracking-wide text-ink">COMMUNITYSHIELD</span>
          </div>

          <h2 className="text-xl font-semibold text-ink">Sign in</h2>
          <p className="mt-1 text-sm text-ink-muted">
            Use your service email address and password.
          </p>

          <form className="mt-6 flex flex-col gap-4" onSubmit={onSubmit} noValidate>
            <Field label="Email" required>
              {(props) => (
                <Input
                  {...props}
                  type="email"
                  autoComplete="email"
                  autoFocus
                  required
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                />
              )}
            </Field>

            <Field label="Password" required error={error ?? undefined}>
              {(props) => (
                <Input
                  {...props}
                  type="password"
                  autoComplete="current-password"
                  required
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
              )}
            </Field>

            <Button type="submit" variant="primary" size="lg" block loading={submitting}>
              Sign in
            </Button>
          </form>

          {USE_MOCKS ? (
            <div className="mt-8 rounded-panel border border-border-hi bg-surface-hi p-3">
              <p className="text-xs font-semibold text-ink">Demo accounts</p>
              <p className="mt-0.5 text-[11px] text-ink-muted">
                Running on mocks (<code>VITE_USE_MOCKS</code>) — tap a role to fill the form.
              </p>
              <div className="mt-2 flex flex-wrap gap-2">
                {MOCK_ACCOUNTS.map((account) => (
                  <Button
                    key={account.email}
                    size="sm"
                    variant="secondary"
                    onClick={() => {
                      setEmail(account.email)
                      setPassword(account.password)
                      setError(null)
                    }}
                  >
                    {account.role}
                  </Button>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}
