import { Navigate, useLocation } from 'react-router-dom'
import type { ReactNode } from 'react'
import { useAuth } from './AuthContext'
import type { Role } from '@/types/api'
import { FullPageSpinner } from '@/components/ui/States'

/**
 * Route guard. Redirects anonymous users to login (preserving the intended
 * destination) and authenticated users who lack the required role to their own
 * home. Role is the coarse gate; server-side unit/resource checks still apply.
 */
export function RequireRole({
  roles,
  children,
}: {
  roles: Role[]
  children: ReactNode
}) {
  const { status, role } = useAuth()
  const location = useLocation()

  if (status === 'loading') return <FullPageSpinner label="Restoring session…" />
  if (status === 'anonymous') {
    return <Navigate to="/auth/login" replace state={{ from: location.pathname }} />
  }
  if (role && !roles.includes(role)) {
    return <Navigate to={homePathForRole(role)} replace />
  }
  return <>{children}</>
}

export function homePathForRole(role: Role | null | undefined): string {
  switch (role) {
    case 'officer':
      return '/officer/queue'
    case 'unit_admin':
      // Land on the screen that works today, not the one that is still a stub.
      return '/admin/cases'
    case 'super_admin':
      return '/super/overview'
    case 'citizen':
    default:
      return '/'
  }
}
