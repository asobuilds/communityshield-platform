import type { ReactNode } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import {
  BarChart3,
  FileText,
  FolderKanban,
  LayoutDashboard,
  LogOut,
  Map as MapIcon,
  Megaphone,
  Shield,
  ShieldAlert,
  UserCircle,
  Users,
} from 'lucide-react'
import { cn } from '@/lib/cn'
import { useAuth } from '@/auth/AuthContext'
import { fullName, initials } from '@/lib/format'
import { NotificationBell } from './NotificationBell'
import type { Role } from '@/types/api'

interface NavItem {
  to: string
  label: string
  icon: ReactNode
  /** Not built yet — rendered disabled with an explanation, never a dead link. */
  soon?: boolean
}

const NAV: Record<Role, NavItem[]> = {
  citizen: [
    { to: '/', label: 'Home', icon: <LayoutDashboard className="size-4" /> },
    { to: '/report', label: 'Report', icon: <FileText className="size-4" />, soon: true },
    { to: '/track', label: 'Track', icon: <FolderKanban className="size-4" />, soon: true },
    { to: '/alerts', label: 'Alerts', icon: <Megaphone className="size-4" />, soon: true },
    { to: '/map', label: 'Safety map', icon: <MapIcon className="size-4" /> },
  ],
  officer: [
    { to: '/officer/queue', label: 'Case queue', icon: <FolderKanban className="size-4" /> },
    { to: '/map', label: 'Operations map', icon: <MapIcon className="size-4" /> },
    { to: '/officer/comms', label: 'Comms', icon: <Users className="size-4" />, soon: true },
  ],
  unit_admin: [
    { to: '/admin/cases', label: 'Case review', icon: <FolderKanban className="size-4" /> },
    { to: '/admin/overview', label: 'Overview', icon: <LayoutDashboard className="size-4" />, soon: true },
    { to: '/admin/officers', label: 'Officers', icon: <Users className="size-4" />, soon: true },
    { to: '/admin/analytics', label: 'Analytics', icon: <BarChart3 className="size-4" />, soon: true },
    { to: '/map', label: 'Operations map', icon: <MapIcon className="size-4" /> },
  ],
  super_admin: [
    { to: '/admin/cases', label: 'Case review', icon: <FolderKanban className="size-4" /> },
    { to: '/super/overview', label: 'Governance', icon: <ShieldAlert className="size-4" />, soon: true },
    { to: '/super/users', label: 'Users', icon: <Users className="size-4" />, soon: true },
    { to: '/super/audit', label: 'Audit', icon: <FileText className="size-4" />, soon: true },
    { to: '/super/analytics', label: 'Analytics', icon: <BarChart3 className="size-4" />, soon: true },
    { to: '/map', label: 'Operations map', icon: <MapIcon className="size-4" /> },
  ],
}

const ROLE_LABEL: Record<Role, string> = {
  citizen: 'Citizen',
  officer: 'Officer',
  unit_admin: 'Unit admin',
  super_admin: 'Super admin',
}

function NavItems({ items, variant }: { items: NavItem[]; variant: 'sidebar' | 'bottom' }) {
  return (
    <>
      {items.map((item) => {
        if (item.soon) {
          return (
            <span
              key={item.to}
              aria-disabled="true"
              title="Coming in a later milestone"
              className={cn(
                'flex cursor-not-allowed items-center gap-3 rounded-lg text-ink-faint',
                variant === 'sidebar' ? 'px-3 py-2 text-sm' : 'flex-col gap-1 px-2 py-1 text-[10px]',
              )}
            >
              {item.icon}
              <span className="truncate">{item.label}</span>
              {variant === 'sidebar' ? (
                <span className="ml-auto rounded bg-surface-hi px-1.5 py-0.5 text-[9px] uppercase tracking-wide">
                  Soon
                </span>
              ) : null}
            </span>
          )
        }
        return (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              cn(
                'flex items-center rounded-lg transition-colors',
                variant === 'sidebar' ? 'gap-3 px-3 py-2 text-sm' : 'flex-col gap-1 px-2 py-1 text-[10px]',
                isActive
                  ? 'bg-signal/10 text-signal'
                  : 'text-ink-muted hover:bg-surface-hi hover:text-ink',
              )
            }
          >
            {item.icon}
            <span className="truncate">{item.label}</span>
          </NavLink>
        )
      })}
    </>
  )
}

/**
 * Application shell: desktop sidebar + top bar, mobile bottom nav.
 * Role-aware, and honest about what isn't built yet.
 */
export function AppShell() {
  const { user, role, logout } = useAuth()
  const navigate = useNavigate()
  const items = role ? NAV[role] : []

  function handleLogout() {
    logout()
    navigate('/auth/login', { replace: true })
  }

  return (
    <div className="flex min-h-screen bg-base">
      {/* Desktop sidebar */}
      <aside className="hidden w-60 shrink-0 flex-col border-r border-border bg-surface md:flex">
        <div className="flex items-center gap-2 border-b border-border px-4 py-4">
          <Shield className="size-6 text-signal" aria-hidden />
          <div>
            <p className="text-sm font-bold tracking-wide text-ink">COMMUNITYSHIELD</p>
            <p className="text-[11px] text-ink-muted">{role ? ROLE_LABEL[role] : ''} console</p>
          </div>
        </div>

        <nav className="flex flex-1 flex-col gap-1 p-3" aria-label="Primary">
          <NavItems items={items} variant="sidebar" />
        </nav>

        <div className="border-t border-border p-3">
          <div className="flex items-center gap-2">
            <span className="grid size-8 shrink-0 place-items-center rounded-full bg-surface-hi text-xs font-semibold text-ink">
              {initials(user?.firstName, user?.lastName)}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-medium text-ink">
                {fullName(user?.firstName, user?.lastName)}
              </p>
              <p className="truncate text-[10px] text-ink-muted">{user?.email}</p>
            </div>
            <button
              type="button"
              onClick={handleLogout}
              aria-label="Log out"
              title="Log out"
              className="rounded-md p-1.5 text-ink-muted transition-colors hover:bg-surface-hi hover:text-ink"
            >
              <LogOut className="size-4" aria-hidden />
            </button>
          </div>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        {/* Top bar */}
        <header className="sticky top-0 z-30 flex items-center justify-between gap-3 border-b border-border bg-base/90 px-4 py-3 backdrop-blur">
          <div className="flex items-center gap-2 md:hidden">
            <Shield className="size-5 text-signal" aria-hidden />
            <span className="text-sm font-bold tracking-wide text-ink">COMMUNITYSHIELD</span>
          </div>
          <div className="hidden md:block" />
          <div className="flex items-center gap-2">
            <NotificationBell />
            <span className="grid size-9 place-items-center rounded-lg border border-border-hi bg-surface-hi text-ink-muted md:hidden">
              <UserCircle className="size-4" aria-hidden />
            </span>
          </div>
        </header>

        <main className="min-w-0 flex-1 pb-20 md:pb-0">
          <Outlet />
        </main>

        {/* Mobile bottom nav */}
        <nav
          aria-label="Primary"
          className="fixed inset-x-0 bottom-0 z-30 flex items-center justify-around border-t border-border bg-surface/95 px-1 py-1.5 backdrop-blur md:hidden"
        >
          <NavItems items={items} variant="bottom" />
        </nav>
      </div>
    </div>
  )
}
