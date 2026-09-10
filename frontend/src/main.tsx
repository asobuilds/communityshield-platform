import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { USE_MOCKS } from '@/mocks/config'
import './index.css'

async function bootstrap() {
  // Install the mock API *before* rendering — the app must not fire a request
  // before the adapter is in place.
  if (USE_MOCKS) {
    const { installMockApi } = await import('./mocks/install')
    installMockApi()
  }

  const { App } = await import('./App')

  const container = document.getElementById('root')
  if (!container) throw new Error('Root element #root not found')

  createRoot(container).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

void bootstrap()
