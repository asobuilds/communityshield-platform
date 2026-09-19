# WardGuard — Frontend

React 19 + Vite + TypeScript client for WardGuard. See the repository root `README.md` for the full project overview and `frontReadme.md` / `frontagent.md` in this directory for the frontend feature map and design contract.

## Quick start

    npm install
    npm run dev

Mocks are on by default — every screen works with no backend required.

## Demo accounts

Sign in with any of these (password: `password`):

- `citizen@wardguard.ng`
- `officer@wardguard.ng`
- `admin@wardguard.ng`
- `super@wardguard.ng`

## Build

    npm run build

Requires Node 20+. The build runs `tsc --noEmit` followed by `vite build`.

## Structure

- `src/` — application source
- `src/components/` — UI primitives, layout, case, map, admin
- `src/hooks/` — React Query data hooks
- `src/lib/` — pure utilities and API client
- `src/mocks/` — in-browser mock API for offline development
- `src/types/` — hand-written API contract matching the Go handlers

## Environment

`.env.development` sets `VITE_USE_MOCKS=true` by default. To develop against the real backend, create `.env.local`:

    VITE_USE_MOCKS=false
    VITE_API_URL=http://localhost:8080