# Media retention and bucket policy

Applies to the `nativityguard-media` Supabase Storage bucket.

Free tier: 1 GB storage, 2 GB bandwidth/month, 50 MB max file size.
Supabase speaks S3 protocol. Migration to Cloudflare R2 later is a
five env var change — no code change.

## Storage backends

The backend has one storage interface with two backends:

| Condition | Backend | Notes |
|---|---|---|
| `MEDIA_ENDPOINT` (and `MEDIA_ACCESS_KEY_ID`, `MEDIA_SECRET_ACCESS_KEY`, `MEDIA_BUCKET`) set | Supabase Storage via S3 API | production |
| `MEDIA_ENDPOINT` absent | local disk under `STORAGE_ROOT` (default `./storage`) | local development |

The key layout is identical in both modes:

```
<category>/<yyyy>/<mm>/<sha256>.<ext>
```

Content addressing is unchanged, so keys written by the legacy local disk
backend and keys written to Supabase use the same naming rule. Nothing about the
existing `avatarPath`, `coverPath`, `hash` or `size` response fields changes.

## Public vs private prefixes

One bucket, two access tiers. The split is enforced in code in
`backend/services/r2_service.go` (`IsPublicStorageKey` / `IsPrivateStorageKey`).

**Public** — served from `MEDIA_PUBLIC_BASE` (the Supabase public object URL):

| Prefix | Content |
|---|---|
| `avatars/` | user avatars |
| `covers/` | user cover images |
| `units/` | unit logos and cover images |
| `news/` | news images |
| `community/` | community post images |

**Private** — never exposed as a public URL. Reachable only through the
authenticated `GET /api/files/:category/:hash` handler, which checks access and
then redirects to a presigned GET URL with a 10 minute TTL:

| Prefix | Content |
|---|---|
| `evidence/` | case evidence (photos, video, audio, documents) |
| `gov_ids/` | submitted government identity documents |
| `voice_notes/` | case voice notes |

Any prefix that is not on the public list is treated as private. A new category
therefore cannot leak by accident — it has to be added to the public list
deliberately.

`Evidence.FileURL` holds the object **key**, not a public URL.

## Retention targets

| Asset | Hot | Then |
|---|---|---|
| Evidence | 5 years | transition to cold storage |
| Video | 2 years | archive |
| Avatar / cover | life of the account | delete 30 days after account deletion |
| News / community media | indefinite | deleted only when flagged |

Evidence and video retention are enforced with Supabase lifecycle rules. Avatar and
cover retention is enforced by the application: `DeleteAvatar` / `DeleteCover`
(`backend/handlers/user_media_handler.go`) remove the object immediately, and a
future wave adds the 30 day grace-period job for account deletion.

## Supabase lifecycle rules

Apply these manually in the Supabase dashboard under
**Storage → nativityguard-media → Settings → Object lifecycle rules**. The backend
documents them; it does not create them.

Rule 1 — evidence, 5 year hot then cold:

```json
{
  "rules": [
    {
      "id": "evidence-cold-after-5y",
      "enabled": true,
      "prefix": "evidence/",
      "actions": [
        { "type": "TransitionStorageClass", "storageClass": "INFREQUENT_ACCESS" }
      ],
      "deleteAfterDays": 1825
    }
  ]
}
```

Rule 2 — video, 2 year hot then archived:

```json
{
  "rules": [
    {
      "id": "video-archive-after-2y",
      "enabled": true,
      "prefix": "evidence/",
      "suffixes": [".mp4", ".webm", ".mov"],
      "actions": [
        { "type": "TransitionStorageClass", "storageClass": "GLACIER" }
      ],
      "deleteAfterDays": 730
    }
  ]
}
```

If the dashboard rejects the `INFREQUENT_ACCESS` / `GLACIER` storage classes on
the free plan, apply the delete-only form, which is accepted everywhere:

```json
{
  "rules": [
    { "id": "evidence-expire-5y", "enabled": true, "prefix": "evidence/", "deleteAfterDays": 1825 },
    { "id": "video-expire-2y", "enabled": true, "prefix": "evidence/", "suffixes": [".mp4", ".webm", ".mov"], "deleteAfterDays": 730 }
  ]
}
```

News and community media get no rule: they are retained indefinitely unless an
administrator flags them.

## Validation gaps

The presigned upload flow has one known gap, accepted for this wave.

`middleware.UploadValidationMiddleware` (`backend/middleware/upload_validation_middleware.go`)
sniffs magic bytes, but it only runs on `multipart/form-data` requests. The
presign flow sends JSON and the client then PUTs the bytes straight to Supabase, so
the server never sees the payload.

What the flow does check:

- `POST /api/evidence/case/:caseId/presign` validates `sizeBytes` against the
  100 MB evidence cap and validates `contentType` against the evidence
  allowlist. It does **not** check the content.
- `POST /api/evidence/case/:caseId/confirm` issues a `HeadObject` and verifies
  the stored size matches the declared size and the stored content type matches
  the declared content type. It does **not** verify the bytes.

Not verified today: magic-byte / content sniffing, and sha256 verification of the
uploaded payload against the digest supplied in the presign request. The digest
is carried as Supabase object metadata and used as the key stem, so a mismatch
would only be caught by recomputing the hash server-side.

The single-step `POST /api/evidence/case/:caseId/file` endpoint keeps the full
magic-byte validation, because the server handles those bytes. Clients that care
about content validation should prefer it, or a client-side check should run
before the PUT.

Post-upload scanning (antivirus, content-type re-verification, hash
recomputation) is a future wave.

## Local disk fallback

With `MEDIA_*` unset the backend behaves exactly as before this change:

- `Save` writes to `STORAGE_ROOT/<category>/<yyyy>/<mm>/<sha256>.<ext>` and
  deduplicates by hash.
- `GET /api/files/:category/:hash` resolves with
  `filepath.Glob(STORAGE_ROOT/<category>/*/*/<hash>.*)` and streams the bytes
  from disk instead of redirecting. There is no public/private redirect in this
  mode, so local development serves every category through the same
  access-checked handler.
- `Delete` and `Exists` operate on the filesystem.

`MEDIA_PUBLIC_BASE` is only consulted in Supabase mode. If it is empty or missing, public
file requests return `500 MEDIA_PUBLIC_BASE is not configured` rather than leaking
a raw Supabase endpoint.
