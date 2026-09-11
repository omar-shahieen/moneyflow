# Backend Implementation Plan

## Goal

Build the finance API on Gin inside the Nx monorepo, migrate the current Echo backend without breaking the API contract, and implement the product as tested vertical slices.

The backend must preserve these boundaries:

```text
Gin handlers -> services/use cases -> repository and infrastructure interfaces
                                             -> Postgres, Redis, storage, email, billing
```

Services own authorization, plan enforcement, transactions, and idempotency. Gin handlers only translate HTTP requests and responses.

## Target Backend Shape

```text
apps/backend/
  cmd/api/main.go
  internal/
    config/
    database/
    domain/
      model/
      ports/
      errors.go
    handler/
    middleware/
    repository/
    router/
    service/
    worker/
    logger/
    validation/
  migrations/
```

Keep `apps/backend` initially to reduce migration risk. Rename it to `apps/api` only after the Gin migration is stable.

## Phase 0: Baseline and Contract Freeze

1. Record required versions and local dependencies: Bun, Node, Go, PostgreSQL, Redis, and Docker.
2. Run the existing backend tests, `go vet`, formatting checks, frontend checks, and package checks.
3. Record existing failures separately from migration failures.
4. Confirm the API prefix. The target is `/api/v1`; update the OpenAPI contract if the intended public prefix is different.
5. Freeze the response envelope, status codes, pagination shape, request ID behavior, auth claims, and health routes.
6. Add contract tests for:
   - health success;
   - unauthenticated access;
   - malformed and invalid requests;
   - not found and forbidden responses;
   - rate limiting;
   - successful JSON and no-content responses.

**Exit gate:** baseline commands are reproducible and contract tests detect changes to routes, status codes, or error responses.

## Phase 1: Add Nx Backend Targets

1. Add the backend as an Nx project without changing Go code.
2. Create targets for:
   - `dev`: `go run ./cmd/api`;
   - `build`: `go build ./cmd/api`;
   - `test`: `go test ./...`;
   - `test:race`: `go test -race ./...`;
   - `lint`: `go vet ./...` plus the repository Go linter;
   - `format`: `gofmt` check;
   - `clean`: remove generated binaries and test artifacts.
3. Add inputs and outputs so Nx can cache Go builds and tests safely.
4. Add the backend to `nx affected` and CI.
5. Keep Turborepo commands until Nx produces equivalent results in CI.

**Exit gate:** `nx run backend:test`, `nx run backend:build`, and `nx affected -t lint test build` work from a clean checkout.

## Phase 2: Extract Framework-Neutral Backend Code

1. Move request DTOs, response DTOs, validation types, and domain errors away from Echo imports.
2. Ensure services accept `context.Context`, typed inputs, and typed outputs only.
3. Define interfaces for repositories, object storage, queues, email, billing, and clocks.
4. Keep database models separate from HTTP DTOs.
5. Move error mapping rules into domain/application errors:
   - `ErrValidation`;
   - `ErrNotFound`;
   - `ErrForbidden`;
   - `ErrConflict`;
   - `ErrPlanLimitExceeded`.
6. Add service tests that do not import Echo or Gin.
7. Keep existing migrations, database helpers, and infrastructure implementations working.

**Exit gate:** service tests compile without a web framework dependency and existing API behavior remains unchanged.

## Phase 3: Build the Gin Composition Root

1. Add Gin to `go.mod` and create the API composition root.
2. Assemble config, logger, database, Redis, job client, services, handlers, and router in one startup path.
3. Add graceful shutdown for HTTP, database, Redis, and workers.
4. Add `/health` and system routes.
5. Register versioned routes under `/api/v1`.
6. Keep dependencies passed explicitly rather than using package globals.

**Exit gate:** the API starts with valid configuration, fails fast with missing required configuration, answers health checks, and shuts down cleanly.

## Phase 4: Migrate Core Middleware

Migrate one middleware at a time and retain contract tests after each change:

1. Panic recovery with safe production responses.
2. Request ID creation and response header propagation.
3. Structured request logging with method, route, status, latency, IP, user ID, and request ID.
4. CORS configuration.
5. Secure HTTP headers.
6. Rate limiting and rate-limit metrics.
7. Context enrichment.
8. New Relic request tracing and error reporting.
9. Clerk token/session verification and typed authenticated identity.

Do not pass `*gin.Context` into services. Store identity in the request context and expose it through a typed helper.

**Exit gate:** authentication, logging, tracing, rate limiting, and error responses behave the same under Gin as under the baseline tests.

## Phase 5: Replace Handler Helpers and Error Handling

1. Implement Gin request binding and validation helpers.
2. Implement standard JSON, file, and no-content response helpers.
3. Add one centralized error translator for validation, domain, database, and unknown errors.
4. Ensure internal error details are logged but not returned to clients.
5. Preserve field-level validation errors and the standard error envelope.
6. Add handler tests for every response class.

**Exit gate:** no handler manually constructs inconsistent error envelopes and all handler tests pass.

## Phase 6: Migrate Existing Infrastructure

Migrate and verify the existing capabilities in this order:

1. Configuration and environment validation.
2. PostgreSQL pool, migrations, transactions, and query tracing.
3. Redis connection and Asynq client/server.
4. Worker queues: `critical`, `default`, and `low`.
5. Resend client, templates, and welcome email task.
6. New Relic logging and database instrumentation.
7. OpenAPI/Swagger documentation.
8. Testcontainers and authenticated server test helpers.

For every job, define a deterministic idempotency key, retry policy, timeout, and failure behavior before enabling it.

**Exit gate:** all current non-domain features run through Gin and Echo can be removed from production code and `go.mod`.

## Phase 7: Product Vertical Slices

Implement each feature across migration, model, repository, service, handler, route, OpenAPI, job, and tests before starting the next feature.

### Slice 1: Accounts and Categories

1. Synchronize local users from Clerk identity.
2. Add category CRUD and income/expense validation.
3. Enforce ownership in every query.
4. Prevent deletion when transactions reference a category.
5. Enforce category plan limits in the service.

### Slice 2: Transactions

1. Add transaction schema and indexes.
2. Validate integer minor units and currency codes.
3. Implement CRUD with ownership checks.
4. Add date, category, type, and pagination filters.
5. Add SQL aggregation for monthly and category summaries.
6. Add multi-currency grouping.
7. Add transaction and authorization integration tests.

### Slice 3: Budgets and Membership

1. Add budgets and budget members.
2. Enforce one owner per budget.
3. Add member invitation/removal and ownership transfer rules.
4. Calculate usage and exceeded state in SQL.
5. Enforce owner plan limits.
6. Test cross-user access and hidden `404` behavior.

### Slice 4: Subscription and Plan Guard

1. Add subscription state and default free plan.
2. Implement reusable `PlanGuard` service.
3. Return `402 plan_limit_exceeded` consistently.
4. Add billing provider interfaces for checkout and portal.
5. Verify provider webhook signatures.
6. Store provider event IDs and process duplicate events safely.

### Slice 5: Recurring Transactions

1. Add recurring rule CRUD.
2. Add weekly and monthly schedule calculation.
3. Add occurrence uniqueness using rule ID plus occurrence date.
4. Create the Asynq generation task.
5. Test retries, missed runs, end dates, and duplicate prevention.

### Slice 6: Receipts

1. Add private object storage abstraction.
2. Validate receipt ownership and content metadata.
3. Return presigned upload URLs.
4. Never proxy receipt bytes through the API.
5. Add expiry and unauthorized access tests.

### Slice 7: CSV Imports

1. Create import record and enqueue endpoint returning `202`.
2. Add streaming CSV parsing in the worker.
3. Validate rows independently.
4. Store per-row failures and progress checkpoints.
5. Enforce upload row limits by plan.
6. Make retries resume safely.

### Slice 8: Reports

1. Create report records and enqueue endpoint returning `202`.
2. Generate CSV and PDF in workers.
3. Store files privately.
4. Add report status polling and expiring download URLs.
5. Enforce report history limits.
6. Test retries and failed generation.

### Slice 9: Notifications

1. Add budget-exceeded notifications.
2. Add monthly summary notifications.
3. Add notification deduplication keys.
4. Queue email delivery and retry transient failures.
5. Add delivery and suppression tests.

## Phase 8: Backend Hardening

1. Add cross-user authorization tests for every resource.
2. Add database constraint tests and migration tests.
3. Run race tests for shared services and worker code.
4. Load-test list, summary, import enqueue, and report enqueue endpoints.
5. Verify expensive operations are never performed inside HTTP handlers.
6. Add metrics for HTTP RED, queue depth, job retries, job failures, and database health.
7. Document backup/restore, secret rotation, provider webhook recovery, and incident procedures.
8. Run full Nx affected checks in pull requests and full backend verification on release branches.

## Backend Definition of Done

- Gin is the only production HTTP framework.
- Nx owns backend build, test, lint, format, dev, and CI targets.
- `/api/v1`, error envelopes, auth behavior, and status codes are contract-tested.
- Services contain authorization and plan enforcement.
- Jobs and webhooks are idempotent under retries.
- Postgres, Redis, storage, email, billing, and observability are replaceable through interfaces.
- Integration tests use real Postgres/Redis through Testcontainers where behavior depends on infrastructure.
- A clean checkout passes the backend Nx targets.
