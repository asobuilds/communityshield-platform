import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { Link } from 'react-router-dom'
import { AlertTriangle, LocateFixed, ShieldAlert } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input, Select, Textarea } from '@/components/ui/Field'
import { ErrorState, Skeleton } from '@/components/ui/States'
import { useMySos, useSendSos } from '@/hooks/useSos'
import { useUnits } from '@/hooks/useUnits'
import { MapView } from '@/components/map/MapView'
import { ApiError } from '@/lib/apiClient'
import { USE_MOCKS } from '@/mocks/config'
import { formatDateTime } from '@/lib/format'

const statusLabel = {
  pending: 'Pending',
  dispatched: 'Dispatched',
  resolved: 'Resolved',
  escalated: 'Escalated',
}

/** An explicit confirmation is required before any emergency request leaves the device. */
export function SosPage() {
  const [confirming, setConfirming] = useState(false)
  const confirmRef = useRef<HTMLDivElement>(null)
  const [locating, setLocating] = useState(false)
  const [locationError, setLocationError] = useState('')
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')
  const [contacts, setContacts] = useState('')
  const [medical, setMedical] = useState('')
  const [priority, setPriority] = useState<'high' | 'critical'>('high')
  const [unitId, setUnitId] = useState('')
  const [receipt, setReceipt] = useState<{ trackingId: string; status: string } | null>(null)
  const alerts = useMySos(USE_MOCKS)
  const units = useUnits()
  const send = useSendSos()

  const latitude = Number(lat)
  const longitude = Number(lng)
  const hasLocation = lat.trim() !== '' && lng.trim() !== '' && Number.isFinite(latitude) &&
    Number.isFinite(longitude) && Math.abs(latitude) <= 90 && Math.abs(longitude) <= 180 &&
    !(latitude === 0 && longitude === 0)

  useEffect(() => {
    if (confirming) confirmRef.current?.focus()
  }, [confirming])

  function locate() {
    if (!navigator.geolocation) {
      setLocationError('Location is unavailable on this device. Enter coordinates below.')
      return
    }
    setLocating(true)
    setLocationError('')
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setLat(String(coords.latitude))
        setLng(String(coords.longitude))
        setLocating(false)
      },
      () => {
        setLocationError('Location could not be obtained. Enter coordinates below.')
        setLocating(false)
      },
      { enableHighAccuracy: true, timeout: 10_000 },
    )
  }

  async function confirmSend() {
    if (!USE_MOCKS || !hasLocation || send.isPending) return
    try {
      const { alert } = await send.mutateAsync({
        latitude,
        longitude,
        priority,
        ...(unitId ? { unitId } : {}),
        ...(contacts.trim() ? { emergencyContacts: contacts.trim() } : {}),
        ...(medical.trim() ? { medicalInfo: medical.trim() } : {}),
      })
      setReceipt({ trackingId: alert.trackingId, status: alert.status })
      setConfirming(false)
      setContacts('')
      setMedical('')
    } catch {
      // Keep the confirmation and every field available for an explicit retry.
    }
  }

  function onDialogKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape' && !send.isPending) setConfirming(false)
    if (event.key !== 'Tab') return
    const buttons = Array.from(event.currentTarget.querySelectorAll<HTMLButtonElement>('button:not(:disabled)'))
    const first = buttons[0]
    const last = buttons[buttons.length - 1]
    if (!first || !last) return
    if (event.shiftKey && (document.activeElement === first || document.activeElement === event.currentTarget)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl space-y-5 p-4 pb-28 sm:p-6">
      <header>
        <h1 className="flex items-center gap-2 text-xl font-semibold text-ink">
          <ShieldAlert className="size-6 text-emergency" aria-hidden /> Emergency SOS
        </h1>
        <p className="mt-2 text-sm text-ink-muted">
          Prepare your location, then confirm before sending. Nothing is sent when you open this page.
        </p>
      </header>

      {!USE_MOCKS ? (
        <Card className="border-warn p-5" >
          <p className="text-sm font-semibold text-ink">SOS is not available in this environment</p>
          <p className="mt-2 text-sm text-ink-muted">The live emergency service has not been verified with this app. Use your local emergency contact channels for immediate help.</p>
        </Card>
      ) : null}

      {receipt ? (
        <Card className="border-ok/50 p-5">
          <span role="status" className="sr-only">Emergency request received</span>
          <p className="font-semibold text-ink">Emergency request received</p>
          <p className="mt-2 text-sm text-ink-muted">
            Tracking ID <strong className="text-ink">{receipt.trackingId}</strong> · {receipt.status}.
            Check its status below. A receipt does not mean a unit has been dispatched.
          </p>
        </Card>
      ) : null}

      <Card className="space-y-5 p-5">
        <div>
          <p className="text-sm font-medium text-ink">Where should help go?</p>
          <Button className="mt-3" loading={locating} icon={<LocateFixed className="size-4" />} onClick={locate}>
            Use my location
          </Button>
          {locationError ? <p className="mt-2 text-sm text-warn" role="alert">{locationError}</p> : null}
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            <label className="text-xs text-ink-muted">Latitude
              <Input type="number" step="any" value={lat} onChange={(e) => setLat(e.target.value)} placeholder="6.5244" />
            </label>
            <label className="text-xs text-ink-muted">Longitude
              <Input type="number" step="any" value={lng} onChange={(e) => setLng(e.target.value)} placeholder="3.3792" />
            </label>
          </div>
          <MapView
            mode="pick"
            height="35vh"
            className="mt-3"
            label="Choose the SOS location"
            pickLocation={hasLocation ? [latitude, longitude] : null}
            onPickLocation={(nextLat, nextLng) => { setLat(String(nextLat)); setLng(String(nextLng)) }}
          />
          {!hasLocation ? <p className="mt-2 text-xs text-ink-muted">Valid coordinates are required; 0,0 is not a usable location.</p> : null}
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <label className="text-xs text-ink-muted">Urgency
            <Select value={priority} onChange={(e) => setPriority(e.target.value as 'high' | 'critical')}>
              <option value="high">High</option><option value="critical">Critical</option>
            </Select>
          </label>
          <label className="text-xs text-ink-muted">Preferred responding unit (optional)
            <Select value={unitId} onChange={(e) => setUnitId(e.target.value)}>
              <option value="">Let the service choose</option>
              {(units.data ?? []).map((unit) => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
            </Select>
          </label>
        </div>
        <p className="text-xs text-ink-muted">A unit preference does not confirm dispatch.</p>
        <label className="block text-xs text-ink-muted">Emergency contact details (optional)
          <Input value={contacts} maxLength={250} onChange={(e) => setContacts(e.target.value)} placeholder="Name and phone number" />
        </label>
        <label className="block text-xs text-ink-muted">Medical information to share with responders (optional)
          <Textarea value={medical} maxLength={500} onChange={(e) => setMedical(e.target.value)} placeholder="Only what responders need to know" />
        </label>
        <p className="text-xs text-ink-muted">The optional details are sent only when you confirm; they are cleared from this form after success.</p>
        <Button variant="danger" size="lg" disabled={!USE_MOCKS || !hasLocation} onClick={() => setConfirming(true)} icon={<AlertTriangle className="size-5" />}>
          Prepare SOS
        </Button>
      </Card>

      <section aria-label="Your SOS history">
        <h2 className="mb-3 text-lg font-semibold text-ink">Your SOS history</h2>
        {!USE_MOCKS ? <Card className="p-5 text-sm text-ink-muted">SOS history is unavailable until the live service is connected.</Card> : alerts.isLoading ? <Skeleton className="h-24 w-full" /> : alerts.isError ? (
          <Card><ErrorState title="Could not load SOS history" description="Reconnect and try again." onRetry={() => void alerts.refetch()} /></Card>
        ) : !alerts.data?.length ? (
          <Card className="p-5 text-sm text-ink-muted">No SOS requests on record.</Card>
        ) : (
          <ul className="space-y-2">
            {alerts.data.map((alert) => (
              <li key={alert.id}>
                <Card className="flex flex-wrap items-center justify-between gap-2 p-4">
                  <div><p className="text-sm font-medium text-ink">{alert.trackingId}</p>
                    <p className="text-xs text-ink-muted">{formatDateTime(alert.createdAt)}</p></div>
                  <span className="text-sm font-semibold text-ink">{statusLabel[alert.status] ?? alert.status}</span>
                </Card>
              </li>
            ))}
          </ul>
        )}
        <p className="mt-3 text-xs text-ink-muted">Statuses refresh while this page is open. <Link to="/" className="text-signal">Back to reports</Link></p>
      </section>

      {confirming ? (
        <div className="fixed inset-0 z-50 grid place-items-center bg-void/90 p-4" role="presentation">
          <div ref={confirmRef} tabIndex={-1} onKeyDown={onDialogKeyDown} role="alertdialog" aria-modal="true" aria-labelledby="sos-confirm-title" aria-describedby="sos-confirm-description" className="w-full max-w-md rounded-panel border border-emergency bg-surface p-5 shadow-panel">
            <h2 id="sos-confirm-title" className="text-lg font-semibold text-ink">Send emergency SOS?</h2>
            <p id="sos-confirm-description" className="mt-2 text-sm text-ink-muted">Your location ({latitude.toFixed(5)}, {longitude.toFixed(5)}) and any details above will be sent. Confirm only if you need emergency help.</p>
            {send.isError ? <p className="mt-3 text-sm text-warn" role="alert">{ApiError.isNetwork(send.error) ? 'Delivery could not be confirmed. Check your SOS history before retrying to avoid sending twice.' : 'Request was not confirmed. Check SOS history before retrying.'}</p> : null}
            <div className="mt-5 flex flex-wrap gap-2">
              <Button variant="danger" loading={send.isPending} onClick={() => void confirmSend()}>Confirm and send SOS</Button>
              <Button disabled={send.isPending} onClick={() => setConfirming(false)}>Cancel</Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
