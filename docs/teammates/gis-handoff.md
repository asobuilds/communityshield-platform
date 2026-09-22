# GIS / Map Handoff — Wave 8

This is a reconciliation doc. You have not pushed your local GIS work
yet, so this section reflects only what is in the repository today. Use
it to align before your first push so we don't debug interface mismatches
later.

## 1. What map work already exists in the repo

Map stack present in `frontend/`:
- `leaflet ^1.9.4`, `react-leaflet ^5.0.0`, `@types/leaflet ^1.9.15`
  (package.json)
- `src/index.css:2` imports `leaflet/dist/leaflet.css`
- `src/components/map/MapView.tsx` — the current "one map" (view + pick
  modes; OSM tiles at lines 207–208; tile-failure fallback to list at
  lines 181–191)
- `src/components/map/IncidentMap.tsx` — earlier map component
- `src/pages/MapPage.tsx` — registered at route `/map` (App.tsx:97)
- `src/components/map/recovery/*` — mid-refactor stage snapshots
  (IncidentMap.stage1.tsx, stage3, SAFE-BACKUP). Treat as WIP.

No Mapbox, MapLibre, Google Maps, deck.gl, or react-map-gl present.

## 2. What the backend exposes for maps (align before your first push)

Every location-bearing endpoint, with JSON field names. No GeoJSON, no
WKT, no PostGIS anywhere — coordinates are plain float64 numeric columns.

Public (no JWT):
- `GET /public/units` → `latitude`, `longitude`
- `GET /public/cases` → `latitude`, `longitude`, `location`,
  `gisLatitude`, `gisLongitude`

Authenticated (Bearer token required):
- `GET /units` → `latitude`, `longitude`
- `GET /units/nearby?lat=&lng=&radius=` → units with `latitude`,
  `longitude`
- `GET /units/by-location` → units with `latitude`, `longitude`
- `GET /units/:id` → `latitude`, `longitude`
- `GET /cases` → `latitude`, `longitude`, `location`, `gisLatitude`,
  `gisLongitude`
- `GET /cases/:id` → same as above
- `GET /sos`, `GET /sos/:id`, `GET /sos/my` → `latitude`, `longitude`
- `GET /suspects/:id` → `latitude`, `longitude`, `location`
- `GET /suspects/:id/sightings` → `latitude`, `longitude`, `location`
- `GET /alerts` → `latitude`, `longitude` (CommunityAlert)
- `POST /location` → updates viewer position; returns `latitude`,
  `longitude`
- `POST /ai/analyze-location` → returns `latitude`, `longitude`
- `POST /ai/map-insights`, `POST /ai/predict-hotspots` → AI endpoints
  adjacent to maps; confirm response shape before consuming

Key JSON field names to consume verbatim: `latitude`, `longitude`
(float64), `gisLatitude`, `gisLongitude` (Case only), `location`
(string). AI responses also expose `viewLat` / `viewLng` on some paths.

## 3. What the backend guarantees

- Storage: coordinates are `numeric` / `float64` columns on
  `Case`, `Unit`, `SOSAlert`, `Suspect`, `UserLocation`, `Community`,
  `CommunityAlert`, `VideoAnalytics`. No spatial index, no geometry
  type. Distance queries are computed in code, not in the DB.
- Auth model: public endpoints above need no token but are field-scoped
  (only public-flagged rows). Everything else requires
  `Authorization: Bearer <token>`.
- Rate limiting on map / location endpoints: none today. (See §4.)

## 4. What's still in motion

- Wave 8g will add rate limiting on map-heavy endpoints. Coordinate
  tile-server choice and query shape before that lands so we can size
  limits appropriately.
- Wave 8h may tighten which location fields are public
  (`Case.gisLatitude`/`gisLongitude` and unit `lat`/`lng` visibility
  may change; SOS and suspect sighting coords may be scoped). Align
  before committing field decisions into your tile pipelines.
- The frontend map code is mid-refactor. Payload shapes may shift.

## 5. Where your creative freedom lives

These are entirely your call and not specified by the backend:
- Tile provider (Mapbox, MapLibre, Google, self-hosted)
- Rendering library (not locked to the current Leaflet scaffolding)
- Marker clustering, heatmaps, choropleths, layer architecture
- Offline tile strategy, zoom behavior, interaction patterns
- How you visualize the 20k-case dataset (see §7)

## 6. Where you must conform

- Consume the exact JSON field names above. Do not invent fields, do not
  expect GeoJSON or WKT.
- Respect the public vs authenticated split. Do not call authed
  endpoints as public.
- Default tile source is OpenStreetMap (`MapView.tsx:207–208`). If you
  swap providers, keep the tile-failure → list fallback contract
  (`MapView.tsx:181–191`).

## 7. How to test against the real thing

A local seeder produces 20,000 cases with randomized lat/lng spread
(±10° across both axes). It is a psql script, not a Go program:

  $env:PGPASSWORD = "testpass"
  psql -h 127.0.0.1 -U postgres -d wardguard_test -f backend/cmd/seedtest/seed.sql

Two companion scripts:
- `backend/cmd/seedtest/add_indexes.sql` — the four perf indexes
- `backend/cmd/seedtest/measure1.sql` — the EXPLAIN queries used to
  validate

Frontend env for real-data eyeballing:
  VITE_API_URL=http://localhost:8080
  VITE_USE_MOCKS=false
Then visit `/map` and plot cases + units from real rows.
