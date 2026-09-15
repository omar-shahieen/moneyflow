# Cloudflare R2 Integration Plan (v2 — Payment CSV Hardened)

## Overview
- **Upload Flow**: Client gets presigned upload URL → uploads directly to R2 → backend processes from R2
- **Processing**: Async via asynq background jobs
- **Fallback**: Keep `LocalStorage` as fallback via env var toggle

### Best practices folded into this version
- Immutable, date-partitioned object keys (no overwrites)
- SHA-256 checksum computed and verified for every uploaded file
- Idempotent processing keyed on checksum (prevents double-counting on retry/re-upload)
- Tightly scoped R2 token (write-only to `payments/` prefix) and short presign expiry
- Private bucket only — no public URLs for payment data
- Schema/structure validation before rows are parsed and processed
- Object metadata tagging (status, checksum, source, uploader) for queryability without opening files
- Retention/lifecycle policy and access logging called out as required infra steps

---

## Files to Create/Modify

### 1. `internal/config/config.go` — Add R2 config
```go
type R2Config struct {
    AccountID      string `koanf:"account_id"`
    AccessKeyID    string `koanf:"access_key_id"`
    SecretAccessKey string `koanf:"secret_access_key"`
    BucketName     string `koanf:"bucket_name"`
    PublicURL      string `koanf:"public_url"`      // leave unset for payments — bucket stays private
    PresignExpiry  int    `koanf:"presign_expiry"`   // minutes, default 15 (keep short for payment data)
    KeyPrefix      string `koanf:"key_prefix"`       // e.g. "payments" — enforced root prefix for this token's scope
}

type StorageConfig struct {
    Provider string             `koanf:"provider"` // "r2" or "local"
    Local    LocalStorageConfig `koanf:"local"`
    R2       R2Config           `koanf:"r2"`
}

type LocalStorageConfig struct {
    BasePath string `koanf:"base_path"`
}
```
Add `Storage StorageConfig` to the main `Config` struct. Env vars: `MONEYFLOW_STORAGE_PROVIDER`, `MONEYFLOW_STORAGE_R2_ACCOUNT_ID`, etc.

> **Best practice:** the R2 API token backing these credentials should be created scoped to *write* access on the `payments/` prefix only (not bucket-wide, not read/delete for the upload path). Create a separate read-scoped token for the worker that downloads/processes files.

---

### 2. `internal/ports/storage.go` — Extend interface
Add methods for direct upload, download, and integrity verification:
```go
type Storage interface {
    GenerateUploadURL(key string, contentType string, expiry time.Duration) (string, error)
    GenerateDownloadURL(key string, expiry time.Duration) (string, error)
    Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) error // NEW
    Download(ctx context.Context, key string) ([]byte, error)                                                  // NEW
    Head(ctx context.Context, key string) (*ObjectInfo, error)                                                  // NEW — checksum/metadata lookup
    Delete(key string) error
    Exists(key string) (bool, error)
}

type ObjectInfo struct {
    ETag        string
    ChecksumSHA256 string
    Metadata    map[string]string
    Size        int64
}
```

---

### 3. `internal/storage/r2.go` — NEW: R2 implementation
Implement `ports.Storage` using `github.com/aws/aws-sdk-go-v2`:
- `R2Storage` struct with R2 client, bucket name, presign client
- `GenerateUploadURL` → presigned PUT URL, **scoped to the exact key** (no wildcard prefixes), expiry from config (default 15 min)
- `GenerateDownloadURL` → presigned GET URL, short expiry, used only internally/by finance tooling — never returned to end users directly for payment files
- `Upload` → `PutObject` with `ChecksumSHA256` set (R2/S3 supports checksum-on-write) and object metadata (`status`, `uploader-id`, `source`)
- `Download` → `GetObject`, verify returned checksum matches stored metadata before returning bytes
- `Head` → `HeadObject`, returns checksum + metadata for status/idempotency checks without downloading the file
- `Delete` → `DeleteObject` (should only ever be invoked by the retention/lifecycle job, not ad hoc application code)
- `Exists` → `HeadObject`

---

### 4. `internal/storage/local.go` — Add missing methods
Add `Upload`, `Download`, and `Head` to `LocalStorage` (compute/store SHA-256 as a sidecar `.sha256` file or extended attribute) to keep dev/local parity with the integrity guarantees R2 gives in production.

---

### 5. `internal/service/services.go` — Config-driven storage
```go
func NewServices(...) (*Services, error) {
    var store ports.Storage
    switch cfg.Storage.Provider {
    case "r2":
        store = storage.NewR2Storage(cfg.Storage.R2, logger)
    default:
        store = storage.NewLocalStorage(cfg.Storage.Local.BasePath)
    }
    // ... inject `store` into services
}
```
Change `Storage` field type from `*storage.LocalStorage` to `ports.Storage`.

---

### 6. `internal/lib/job/import_tasks.go` — NEW: Import job definition
```go
const TaskProcessImport = "import:process"

type ProcessImportPayload struct {
    ImportID     string `json:"import_id"`
    UserID       string `json:"user_id"`
    StorageKey   string `json:"storage_key"`    // R2 key of uploaded CSV
    ChecksumSHA256 string `json:"checksum_sha256"` // set by handler after HeadObject confirms upload; used for idempotency
}
```

---

### 7. `internal/lib/job/handlers.go` — Register import handler
Register `TaskProcessImport` in `Start()`. Pass the import service reference via `InitHandlers`.

> **Best practice — idempotency:** before processing, the handler should check whether an import with the same `checksum_sha256` has already reached a terminal state (`completed`/`failed`). If so, short-circuit and skip reprocessing rather than re-ingesting the same payment rows. This protects against duplicate task delivery (asynq's at-least-once semantics) and accidental re-uploads of the same file.

---

### 8. `internal/service/import.go` — Refactor for R2 + async
- `CreateImport` → accepts `storageKey` parameter (the R2 key of the uploaded CSV)
- `ConfirmUpload` (NEW) → called after client finishes uploading; does a `Head()` on the object to confirm it exists, capture its checksum/size, reject if content-type or size is unexpected, then flips import status from `pending` to `queued` and enqueues the job. **Never trust the client's claim that upload succeeded — verify server-side.**
- `ProcessImport`:
  1. `storage.Head()` → re-verify checksum matches what was recorded at confirm-time (defends against tampering/corruption between confirm and processing)
  2. `storage.Download()` from R2
  3. **Schema validation pass first** — check header row, required columns (amount, currency, transaction id, date, etc.), row count vs. expected `TotalRows`. Reject the whole batch to a `failed` status with a clear reason if the schema doesn't match, before any row-level processing.
  4. Parse and process rows individually; collect `FailedRows` for bad individual rows (existing behavior) without failing the whole batch
  5. On completion, update object metadata (`status=processed`) via a metadata-only `CopyObject` or a tracking record in Postgres — **do not delete or mutate the original CSV**. If a corrected version is ever needed, write a new object.
- The import service needs a `ports.Storage` reference.

---

### 9. `internal/handler/import.go` — Presigned URL flow
- `Create`:
  - Client sends metadata only (`total_rows`, optionally a client-computed checksum for later cross-check)
  - Backend creates import record (`status: pending`)
  - Backend generates the storage key (see key scheme below) and a presigned upload URL
  - Backend returns `{ import_id, upload_url, expires_at }`
- `ConfirmUpload` (NEW endpoint, e.g. `POST /imports/{id}/confirm`):
  - Client calls this after the direct-to-R2 upload finishes
  - Backend does the `Head()` verification described above, then enqueues `import:process`
  - This is the point where a fake/incomplete upload is caught, rather than trusting the client
- Client uploads CSV directly to `upload_url` (R2), then calls confirm
- Background job picks up and processes

---

### 10. `internal/model/imports/imports.dto.go` — Update DTOs
```go
type ImportCreateRequest struct {
    TotalRows int `json:"total_rows" binding:"required,min=1"`
}

type ImportCreateResponse struct {
    ImportID  string    `json:"import_id"`
    UploadURL string    `json:"upload_url"`
    ExpiresAt time.Time `json:"expires_at"`
}

type ImportResponse struct {
    ID             string           `json:"id"`
    Status         string           `json:"status"` // pending|queued|processing|completed|failed
    TotalRows      int              `json:"total_rows"`
    SuccessRows    int              `json:"success_rows"`
    FailedRows     []ImportRowError `json:"failed_rows,omitempty"`
    CreatedAt      string           `json:"created_at"`
    StorageKey     string           `json:"storage_key,omitempty"`
    ChecksumSHA256 string           `json:"checksum_sha256,omitempty"` // NEW
}
```

---

### 11. `internal/model/imports/imports.go` — Add fields
```go
type Import struct {
    // ... existing fields
    StorageKey     string `json:"storage_key" db:"storage_key"`
    ChecksumSHA256 string `json:"checksum_sha256" db:"checksum_sha256"` // NEW — drives idempotency + integrity checks
}
```

---

### 12. DB Migration — Add columns + uniqueness constraint
```sql
ALTER TABLE imports ADD COLUMN storage_key VARCHAR(512);
ALTER TABLE imports ADD COLUMN checksum_sha256 VARCHAR(64);

-- Best practice: prevent the same file from being processed twice per user
CREATE UNIQUE INDEX idx_imports_user_checksum
  ON imports (user_id, checksum_sha256)
  WHERE checksum_sha256 IS NOT NULL;
```

---

### 13. Object key scheme (best practice)
Replace the flat `imports/{user_id}/{import_id}.csv` scheme with a **date-partitioned, immutable, checksum-traceable** layout:

```
payments/{env}/{user_id}/{yyyy}/{mm}/{dd}/{import_id}.csv
```

- `env` separates dev/staging/prod data inside a shared bucket if you don't want fully separate buckets.
- Date partitioning makes lifecycle rules (retention/archival) and bulk listing cheap.
- The `import_id` (not a raw filename) guarantees uniqueness and avoids leaking user-supplied filenames into the key; keep the original filename only in DB metadata.
- Objects are **never overwritten**. A re-upload for the same logical batch gets a new `import_id`; the checksum-based unique index (step 12) stops it from being double-processed.

---

## Execution Order

| Step | File | Action |
|------|------|--------|
| 1 | `config/config.go` | Add `StorageConfig`, `R2Config` (incl. `KeyPrefix`, short `PresignExpiry`) |
| 2 | `go.mod` | Add `github.com/aws/aws-sdk-go-v2` dependencies |
| 3 | `ports/storage.go` | Add `Upload`, `Download`, `Head`, `ObjectInfo` |
| 4 | `storage/r2.go` | **Create** R2 implementation with checksum-on-write |
| 5 | `storage/local.go` | Add `Upload`, `Download`, `Head` (sidecar checksum) |
| 6 | `model/imports/imports.go` | Add `StorageKey`, `ChecksumSHA256` fields |
| 7 | `model/imports/imports.dto.go` | Update DTOs; add `ConfirmUpload` request/response if needed |
| 8 | DB migration | Add columns + unique `(user_id, checksum_sha256)` index |
| 9 | `lib/job/import_tasks.go` | **Create** import task definition (incl. checksum) |
| 10 | `lib/job/handlers.go` | Register handler; add idempotency short-circuit |
| 11 | `service/import.go` | Add `ConfirmUpload`, schema validation pass, checksum re-verify |
| 12 | `handler/import.go` | Presigned URL flow + new confirm endpoint |
| 13 | `service/services.go` | Config-driven storage selection |
| 14 | `router/v1/import.go` | Add route for `POST /imports/{id}/confirm` |

---

## Infra steps (outside the codebase, do before/alongside deploy)
1. Create the R2 bucket as **private**; no public access, no public custom domain.
2. Create two API tokens: one write-only scoped to `payments/` for the upload path, one read-only for the processing worker.
3. Enable **Logpush** for R2 access logs (audit trail requirement for financial data).
4. Configure a **lifecycle rule** matching your retention policy (commonly 5–7 years for financial records) — archive/delete only after that window.
5. Manage the bucket, tokens (permissions only, not secret values), and lifecycle rules via Terraform/Wrangler config rather than the dashboard, so they're reviewable and reproducible.

---

## Environment Variables Needed
```
MONEYFLOW_STORAGE_PROVIDER=r2
MONEYFLOW_STORAGE_R2_ACCOUNT_ID=xxx
MONEYFLOW_STORAGE_R2_ACCESS_KEY_ID=xxx        # write-scoped token for the API service
MONEYFLOW_STORAGE_R2_SECRET_ACCESS_KEY=xxx
MONEYFLOW_STORAGE_R2_BUCKET_NAME=moneyflow
MONEYFLOW_STORAGE_R2_PRESIGN_EXPIRY=15
MONEYFLOW_STORAGE_R2_KEY_PREFIX=payments

# Separate read-scoped credentials for the async worker process, if it runs as a distinct service
MONEYFLOW_WORKER_R2_ACCESS_KEY_ID=yyy
MONEYFLOW_WORKER_R2_SECRET_ACCESS_KEY=yyy
```

---

## API Flow Change

### Before (Current)
```
POST /imports (multipart/form-data with CSV file)
→ Backend parses CSV in-memory
→ Backend processes rows synchronously
→ Returns import result
```

### After (R2 + Async, hardened)
```
POST /imports (JSON: { total_rows: 100 })
→ Backend creates import record (status: pending)
→ Backend generates immutable, date-partitioned key: payments/{env}/{user}/{yyyy}/{mm}/{dd}/{import_id}.csv
→ Backend generates short-lived presigned upload URL (15 min)
→ Returns { import_id, upload_url, expires_at }

Client uploads CSV to upload_url (R2)

POST /imports/{id}/confirm
→ Backend runs Head() on the object: confirms existence, captures checksum/size, validates content-type
→ Backend checks (user_id, checksum) uniqueness — rejects duplicate re-uploads early
→ Backend flips status pending → queued, enqueues import:process task

asynq worker (read-scoped R2 credentials)
→ Head() re-verify checksum
→ Download from R2
→ Schema validation (columns, row count) — reject whole batch on mismatch
→ Parse & process rows individually; collect per-row failures
→ Update import record with results; original CSV object is never modified or deleted
```

---

## Implementation Prompt (paste into Claude Code)

```
Implement the R2 payment-CSV import integration described in cloudflare-r2-integration-plan-v2.md.

Work through the "Execution Order" table in the plan, one step at a time, in order (steps 1–14).
For each step:
1. Show me the diff/new file content before moving to the next step.
2. Follow the exact interfaces, struct fields, and behavior described in the corresponding numbered
   section of the plan (e.g. step 4's R2 implementation must include checksum-on-write via
   PutObject's ChecksumSHA256, and Head() must return ObjectInfo with the stored checksum).
3. Preserve existing behavior for the LocalStorage fallback — it must implement the same extended
   ports.Storage interface (Upload/Download/Head) so `Provider=local` keeps working for dev.
4. Do not skip the idempotency logic (checksum-based unique index in step 12, and the short-circuit
   check in step 10's handler) — this is a hard requirement, not optional polish.
5. Do not skip server-side upload verification (the ConfirmUpload flow in steps 8–9) — never trust
   a client-reported "upload succeeded" for payment data.
6. After the code changes, generate the SQL migration file for step 12 as a separate file following
   this repo's existing migration naming convention (check the migrations directory for the pattern).
7. Add unit tests for: checksum verification on Head/Download, the idempotency short-circuit in the
   job handler, and schema validation rejecting a malformed CSV batch.
8. Do NOT implement the infra steps (bucket creation, API token scoping, Logpush, lifecycle rules) —
   flag them as a checklist for me to do manually in the Cloudflare dashboard / Terraform, per the
   "Infra steps" section of the plan.

Ask me before proceeding if anything in the existing codebase (e.g. current asynq handler
registration pattern, migration tool, or DTO validation library) conflicts with what the plan assumes.
```