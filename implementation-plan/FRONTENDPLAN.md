# Frontend Implementation Plan

## Goal

Build the finance dashboard as a React/Vite application in the Nx monorepo, using Clerk for identity, React Query for server state, React Router for navigation, React Hook Form and Zod for forms, and the shared OpenAPI contract for API communication.

The frontend is a client of the Gin API. Authorization and plan enforcement remain backend responsibilities; frontend guards provide clear navigation and feedback only.

## Current Baseline

The current frontend is a minimal Vite React app with these useful dependencies already installed:

- Clerk React for authentication.
- React Router for routes.
- TanStack React Query for server state.
- Axios for HTTP.
- React Hook Form and Zod for forms and validation.
- Shared `@moneyflow/openapi` and `@moneyflow/zod` packages.
- Tailwind CSS, Lucide icons, and Sonner notifications.

## Target Frontend Structure

```text
apps/frontend/src/
  app/
    App.tsx
    router.tsx
    providers.tsx
  api/
    client.ts
    errors.ts
    queryKeys.ts
  components/
    layout/
    feedback/
    forms/
    data-display/
  config/
  features/
    auth/
    dashboard/
    categories/
    transactions/
    budgets/
    imports/
    reports/
    subscription/
  hooks/
  lib/
  pages/
  styles/
```

Organize by feature once product work begins. Keep reusable UI primitives separate from finance-specific components.

## Phase 0: Freeze Contract and Development Baseline

1. Confirm the API base URL and `/api/v1` prefix.
2. Confirm the Clerk publishable key and environment variable names.
3. Confirm the OpenAPI and Zod package build order.
4. Run the current frontend lint, typecheck, build, and test commands.
5. Add a frontend smoke test for app boot and API error rendering.
6. Define expected UI states for loading, empty, validation error, unauthorized, forbidden, not found, plan limit, server failure, and retry.

**Exit gate:** the frontend starts with documented environment variables and can render a controlled mock API success and failure state.

## Phase 1: Add Nx Frontend Project Configuration

1. Register `apps/frontend` as an Nx project.
2. Add targets for:
   - `dev`;
   - `build`;
   - `test`;
   - `lint`;
   - `typecheck`;
   - `format`;
   - `preview`.
3. Declare dependencies on `packages/openapi` and `packages/zod`.
4. Configure Nx caching for build, test, lint, typecheck, and format outputs.
5. Use Nx targets for frontend orchestration while preserving the existing Vite commands.
6. Verify that a change to a shared contract package marks the frontend affected.

**Exit gate:** `nx run frontend:dev`, `nx run frontend:typecheck`, and `nx run frontend:build` work from a clean checkout.

## Phase 2: App Shell and Providers

1. Create the application provider tree:
   - Clerk provider;
   - React Query client provider;
   - router provider;
   - toast/notification provider;
   - theme and application configuration providers if needed.
2. Create public routes for sign-in and sign-up.
3. Create protected routes for the authenticated application.
4. Add a route-level loading screen while Clerk initializes.
5. Add a not-found route and a global error boundary.
6. Add a responsive application shell with navigation, page title, account menu, and sign-out action.
7. Keep the first authenticated route as a useful dashboard placeholder rather than leaving `Hello World` in the app.

**Exit gate:** a user can sign in, reach a protected route, refresh the page, sign out, and receive a useful error page for an unknown route.

## Phase 3: API Client and Server State

1. Create one Axios client with the API base URL and JSON defaults.
2. Attach the Clerk bearer token through a request interceptor or request helper.
3. Preserve the request ID returned by the API for diagnostics.
4. Normalize API errors into a typed frontend error:
   - validation;
   - authentication;
   - forbidden;
   - not found;
   - conflict;
   - plan limit;
   - rate limit;
   - server error.
5. Generate or validate TypeScript types from OpenAPI.
6. Use Zod schemas at API/form boundaries where runtime validation is needed.
7. Create stable React Query key factories by resource and filter.
8. Add shared query defaults, retry rules, invalidation helpers, and pagination helpers.
9. Add a small MSW or equivalent mock layer for UI tests so features can be developed before every backend endpoint is complete.

**Exit gate:** a protected query can load typed data, display a typed API error, retry safely, and invalidate related data after a mutation.

## Phase 4: Design System and Shared UI

1. Define typography, color tokens, spacing, borders, focus states, and responsive breakpoints in the existing styling system.
2. Build accessible primitives for buttons, inputs, selects, dialogs, tables, badges, tabs, pagination, alerts, skeletons, and empty states.
3. Use Lucide icons inside icon actions and provide tooltips for unfamiliar icons.
4. Ensure keyboard navigation, visible focus, labels, error descriptions, and sufficient contrast.
5. Create consistent loading, empty, permission, and plan-limit states.
6. Add responsive layouts for dense tables and mobile transaction entry.
7. Avoid duplicating API or business rules in presentational components.

**Exit gate:** shared components are used by at least one real feature and pass lint, typecheck, and accessibility-focused tests.

## Phase 5: Authentication and Account Experience

1. Configure Clerk sign-in, sign-up, session loading, and sign-out.
2. Add protected route handling and redirect behavior.
3. Add an account/profile menu.
4. Add subscription plan display and a link to subscription settings.
5. Handle expired sessions and `401` responses by returning the user to authentication without losing useful context.
6. Avoid exposing private provider credentials or server-only configuration in the frontend.

**Exit gate:** authentication works on a clean browser session and protected API calls include the current Clerk token.

## Phase 6: Dashboard and Categories

### Dashboard

1. Add summary cards for current-month income, expenses, remaining budget, and transaction count.
2. Add category and monthly trend visualizations from API summary data.
3. Group multi-currency values clearly rather than combining incompatible currencies.
4. Add loading, empty, stale, and retry states.

### Categories

1. Add category list with income/expense type.
2. Add create and edit forms using React Hook Form and Zod.
3. Add delete confirmation.
4. Display a useful conflict message when a category is referenced by transactions.
5. Display plan-limit state when the user cannot create another category.
6. Invalidate dashboard and transaction queries after category mutations.

**Exit gate:** an authenticated user can create, edit, and delete valid categories and sees correct server errors.

## Phase 7: Transactions

1. Add paginated transaction list with server-side filters for date range, category, and type.
2. Add transaction create form with integer minor-unit handling and explicit currency.
3. Add edit and delete actions with confirmation for destructive changes.
4. Add accessible date and currency inputs.
5. Add optimistic UI only for operations where rollback behavior is well-defined; otherwise invalidate after success.
6. Add transaction summary views by month and category.
7. Add receipt upload flow:
   - request a presigned upload URL;
   - upload directly to storage;
   - save or refresh receipt metadata;
   - show upload progress and failure retry.
8. Add pagination, filter reset, URL-synchronized filters, and mobile-friendly list rendering.

**Exit gate:** a user can create, filter, edit, delete, and inspect transactions without combining different currencies incorrectly.

## Phase 8: Budgets and Shared Access

1. Add budget list with monthly limit, current usage, percentage used, and exceeded state.
2. Add budget create/edit forms.
3. Add category selection and currency validation.
4. Add member management for owners.
5. Display member and owner permissions clearly.
6. Prevent member-only users from seeing owner controls.
7. Add ownership transfer confirmation flows.
8. Handle backend `403`, hidden `404`, and plan-limit responses.
9. Refresh budget summaries after transaction mutations.

**Exit gate:** owners and members see only the actions permitted by the API, and budget usage updates after relevant transaction changes.

## Phase 9: Subscription and Plan-Limited UX

1. Add current subscription page with plan, status, and renewal information.
2. Add plan comparison using the backend-provided limits as the source of truth.
3. Add checkout and customer portal actions.
4. Handle provider redirects and pending subscription state.
5. Add reusable plan-limit prompt for `402 plan_limit_exceeded`.
6. Never hide a protected action solely based on stale frontend plan data; the API response remains authoritative.

**Exit gate:** users can view plan status, start billing actions, and understand why a protected operation is unavailable.

## Phase 10: Asynchronous Imports and Reports

### CSV Imports

1. Add CSV file selection and client-side size/type checks.
2. Submit the file using the backend upload contract.
3. Show `202` processing state immediately.
4. Poll import status with React Query or use a future push mechanism.
5. Show processed, successful, and failed row counts.
6. Render per-row failure details and provide retry guidance.
7. Apply plan limits from the API response.

### Reports

1. Add report request form for format and date range.
2. Show pending and processing states.
3. Poll report status with bounded retry behavior.
4. Show download action only when the report is ready.
5. Open the expiring download URL safely and handle expiration.
6. Display failed report generation with a retry action.

**Exit gate:** import and report workflows remain understandable during asynchronous processing and recover cleanly from failure.

## Phase 11: Notifications and Settings

1. Add notification preferences if supported by the API.
2. Add budget-alert and monthly-summary preference controls.
3. Add timezone and default-currency settings.
4. Add account export or support actions when the plan permits them.
5. Preserve unsaved form state only where it improves recovery and does not conflict with server state.

**Exit gate:** settings persist through the API and show server validation or permission errors without losing the rest of the form.

## Phase 12: Testing and Quality Gates

1. Unit test API error normalization, query keys, formatting, money display, and permission helpers.
2. Component test forms, validation, loading states, empty states, dialogs, tables, and plan-limit messaging.
3. Test authenticated and unauthenticated route behavior.
4. Add browser-level smoke tests for:
   - sign-in and protected navigation;
   - category creation;
   - transaction creation;
   - budget display;
   - CSV import status;
   - report download.
5. Test mobile and desktop layouts for overlapping content, truncated text, and unusable controls.
6. Run Nx affected lint, typecheck, test, and build checks on every change.
7. Run full frontend checks on release branches.

## Frontend Definition of Done

- Nx owns frontend development, build, lint, typecheck, test, and CI targets.
- Clerk-protected routes and authenticated API calls work reliably.
- API types and runtime validation come from the shared OpenAPI/Zod contract.
- Core category, transaction, budget, subscription, import, and report workflows are implemented.
- Every workflow includes loading, empty, validation, permission, plan-limit, failure, and retry states.
- Currency and money values are displayed without precision loss or incompatible aggregation.
- The UI is keyboard accessible and responsive on desktop and mobile.
- A clean checkout passes the frontend Nx targets and browser smoke tests.
