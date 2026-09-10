import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from '@/auth/AuthContext'
import { RequireRole, homePathForRole } from '@/auth/RequireRole'
import { AppShell } from '@/components/layout/AppShell'
import { ToastProvider } from '@/components/ui/Toast'
import { queryClient } from '@/lib/queryClient'
import { LoginPage } from '@/pages/auth/LoginPage'
import { CitizenHomePage } from '@/pages/citizen/CitizenHomePage'
import { OfficerQueuePage } from '@/pages/officer/OfficerQueuePage'
import { OfficerCasePage } from '@/pages/officer/OfficerCasePage'
import { MapPage } from '@/pages/MapPage'
import { AdminCaseQueuePage } from '@/pages/admin/AdminCaseQueuePage'
import { AdminCaseReviewPage } from '@/pages/admin/AdminCaseReviewPage'
import { ComingSoonPage } from '@/pages/ComingSoonPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import type { Role } from '@/types/api'

const ALL_ROLES: Role[] = ['citizen', 'officer', 'unit_admin', 'super_admin']
const STAFF: Role[] = ['officer', 'unit_admin', 'super_admin']
const ADMIN_ROLES: Role[] = ['unit_admin', 'super_admin']

/** Signed-in area: session required, role-aware chrome, nested routes. */
function ProtectedShell() {
  return (
    <RequireRole roles={ALL_ROLES}>
      <AppShell />
    </RequireRole>
  )
}

/** `/` means different things to different roles. */
function HomeRoute() {
  const { role } = useAuth()
  if (role && role !== 'citizen') return <Navigate to={homePathForRole(role)} replace />
  return <CitizenHomePage />
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <BrowserRouter>
          <AuthProvider>
            <Routes>
              <Route path="/auth/login" element={<LoginPage />} />

              <Route element={<ProtectedShell />}>
                <Route index element={<HomeRoute />} />

                {/* Officer workspace — the operational core (staff only). */}
                <Route
                  path="/officer/queue"
                  element={
                    <RequireRole roles={STAFF}>
                      <OfficerQueuePage />
                    </RequireRole>
                  }
                />
                <Route
                  path="/officer/cases/:id"
                  element={
                    <RequireRole roles={STAFF}>
                      <OfficerCasePage />
                    </RequireRole>
                  }
                />

                {/* Shared spatial view, used by every role. */}
                <Route path="/map" element={<MapPage />} />

                {/* Unit administration — triage, assignment and case review. */}
                <Route
                  path="/admin/cases"
                  element={
                    <RequireRole roles={ADMIN_ROLES}>
                      <AdminCaseQueuePage />
                    </RequireRole>
                  }
                />
                <Route
                  path="/admin/cases/:id"
                  element={
                    <RequireRole roles={ADMIN_ROLES}>
                      <AdminCaseReviewPage />
                    </RequireRole>
                  }
                />

                {/* Scheduled, not yet built — named rather than 404'd. */}
                <Route
                  path="/admin/*"
                  element={
                    <RequireRole roles={['unit_admin', 'super_admin']}>
                      <ComingSoonPage
                        title="Unit administration"
                        description="Officer management, unit analytics and settings for unit administrators are designed but not yet implemented. Case review and assignment are available."
                        milestone="M5"
                      />
                    </RequireRole>
                  }
                />
                <Route
                  path="/super/*"
                  element={
                    <RequireRole roles={['super_admin']}>
                      <ComingSoonPage
                        title="Platform governance"
                        description="Cross-unit oversight, user administration and the audit trail for super administrators are designed but not yet implemented."
                        milestone="M7"
                      />
                    </RequireRole>
                  }
                />

                <Route path="*" element={<NotFoundPage />} />
              </Route>
            </Routes>
          </AuthProvider>
        </BrowserRouter>
      </ToastProvider>
    </QueryClientProvider>
  )
}
