import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Bell, Briefcase, Users, Loader2 } from 'lucide-react'
import { Link } from 'react-router-dom'
import { api } from '@/lib/apiClient'
import { AvatarUpload } from '@/components/ui/AvatarUpload'
import { Card, CardBody, CardHeader } from '@/components/ui/Card'
import { Field, Input } from '@/components/ui/Field'
import { cn } from '@/lib/cn'
import type { User as UserType } from '@/types/api'

function useProfile() {
  return useQuery({
    queryKey: ['profile'],
    queryFn: () => api.get<UserType>('/auth/profile'),
    staleTime: 60_000,
  })
}

export function ProfilePage() {
  const queryClient = useQueryClient()
  const { data: profile, isLoading, error } = useProfile()

  const [form, setForm] = useState({
    firstName: '',
    lastName: '',
    phone: '',
  })

  // Keep form in sync with profile data
  if (profile) {
    setForm((prev) => ({
      ...prev,
      firstName: profile.firstName ?? '',
      lastName: profile.lastName ?? '',
      phone: profile.phone ?? '',
    }))
  }

  const handleChange = (key: keyof typeof form, value: string) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const handleAvatarUploaded = (newPath: string) => {
    queryClient.setQueryData<UserType>(['profile'], (old) => old ? { ...old, photoUrl: newPath } : undefined)
  }

  const handleAvatarDeleted = () => {
    queryClient.setQueryData<UserType>(['profile'], (old) => old ? { ...old, photoUrl: undefined } : undefined)
  }

  if (isLoading) {
    return (
      <div className="mx-auto max-w-xl p-6">
        <div className="flex items-center justify-center h-64">
          <Loader2 className="size-8 text-signal animate-spin" aria-hidden />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="mx-auto max-w-xl p-6">
        <div className="rounded-lg border border-emergency/30 bg-emergency/10 p-4 text-emergency">
          Could not load profile. Please try again.
        </div>
      </div>
    )
  }

  const avatarUrl = profile?.photoUrl ?? undefined

  return (
    <div className="mx-auto max-w-xl p-6">
      <h1 className="text-2xl font-bold text-ink mb-6">Profile</h1>

      <Card>
        <CardHeader title="Avatar" />
        <CardBody className="flex flex-col items-center gap-4">
          <AvatarUpload
            currentUrl={avatarUrl}
            size={120}
            onUploaded={handleAvatarUploaded}
            onDeleted={handleAvatarDeleted}
          />
          <p className="text-xs text-ink-muted text-center max-w-xs">
            Click the avatar to change it. Maximum 5 MB · JPEG, PNG, WebP, GIF
          </p>
        </CardBody>
      </Card>

      <Card className="mt-4">
        <CardHeader title="Personal details" />
        <CardBody className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="First name">
              {({ id, ...aria }) => (
                <Input
                  id={id}
                  {...aria}
                  value={form.firstName}
                  onChange={(e) => handleChange('firstName', e.target.value)}
                  disabled // Read-only: no PUT /users/me endpoint
                />
              )}
            </Field>
            <Field label="Last name">
              {({ id, ...aria }) => (
                <Input
                  id={id}
                  {...aria}
                  value={form.lastName}
                  onChange={(e) => handleChange('lastName', e.target.value)}
                  disabled
                />
              )}
            </Field>
          </div>
          <Field label="Contact phone">
            {({ id, ...aria }) => (
              <Input
                id={id}
                {...aria}
                type="tel"
                value={form.phone}
                onChange={(e) => handleChange('phone', e.target.value)}
                disabled
              />
            )}
          </Field>

          <p className="text-xs text-ink-muted">
            Name and phone are read-only — the profile update endpoint is not yet available.
          </p>

          {/* Save button would go here when PUT /users/me exists
          <Button
            variant="primary"
            icon={<Save className="size-4" aria-hidden />}
            loading={updateProfile.isPending}
            onClick={() => void updateProfile.mutate(form)}
            disabled
          >
            Save changes
          </Button>
          */}
        </CardBody>
      </Card>

      <Card className="mt-4">
        <CardHeader title="Quick links" />
        <CardBody className="flex flex-col gap-2">
          <Link to="/notifications" className={cn('flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-surface-hi transition-colors')}>
            <Bell className="size-5 text-ink-muted shrink-0" aria-hidden />
            <span className="text-sm text-ink">Notifications</span>
          </Link>
          <Link to="/cases" className={cn('flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-surface-hi transition-colors')}>
            <Briefcase className="size-5 text-ink-muted shrink-0" aria-hidden />
            <span className="text-sm text-ink">My cases</span>
          </Link>
          <Link to="/units" className={cn('flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-surface-hi transition-colors')}>
            <Users className="size-5 text-ink-muted shrink-0" aria-hidden />
            <span className="text-sm text-ink">My units</span>
          </Link>
        </CardBody>
      </Card>
    </div>
  )
}