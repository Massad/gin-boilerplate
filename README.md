![alt tag](https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png)

[![License](https://img.shields.io/github/license/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/LICENSE) [![GitHub release (latest by date)](https://img.shields.io/github/v/release/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/releases) [![Go Version](https://img.shields.io/github/go-mod/go-version/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-PostgreSQL--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-Redis--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod)

# Golang Gin Boilerplate v3

The fastest way to deploy a RESTful API with [Gin Framework](https://github.com/gin-gonic/gin/) — structured with **PostgreSQL** (multi-tenant Row-Level Security), **JWT** authentication stored in **Redis**, and ready to build on.

## What's Included

- [pgx](https://github.com/jackc/pgx): PostgreSQL driver and connection pooling (pgxpool, per-tenant and platform pools)
- [sqlc](https://github.com/sqlc-dev/sqlc): Type-safe Go code generated from versioned SQL queries
- [goose](https://github.com/pressly/goose): Versioned database migrations
- [jwt-go](https://github.com/golang-jwt/jwt): JSON Web Tokens (JWT) middleware
- [go-redis](https://github.com/go-redis/redis): Redis client
- [maroto](https://github.com/johnfercher/maroto): Pure Go PDF generation
- Built-in **CORS**, **RequestID**, and **Auth** middleware (tenant and admin tokens)
- Built-in **custom form validators** with reusable error translation
- **Multi-tenant isolation** — shared schema + `tenant_id` enforced by fail-closed PostgreSQL RLS
- **Admin backoffice foundation** — separate superadmin login, tenant status control, audit trail
- **Swagger API documentation**
- SSL support
- Go Modules

## Getting Started

### Prerequisites

- Go 1.26.10+
- Docker with Compose ≥ v2.17 (recommended) — the dev dependency stack starts
  with `make services_up`. If `docker ps` refuses permission:
  `sudo usermod -aG docker <user>` then log in again.
- A native PostgreSQL 13+ (uuid primary keys use `gen_random_uuid()`) and
  Redis also work, as long as `DB_PORT` and `REDIS_HOST` are re-pointed in
  `.env` (e.g. `5432` and `127.0.0.1:6379`).

### Setup

Clone the repository:

```bash
git clone https://github.com/Massad/gin-boilerplate.git
cd gin-boilerplate
```

Install dependencies:

```bash
go mod download
```

Set up your environment:

```bash
cp .env_rename_me .env
# Edit .env with your database credentials and signing keys
```

Generate each signing key separately with `openssl rand -hex 32`, then paste
results into `ACCESS_SECRET`, `REFRESH_SECRET`, and `ADMIN_ACCESS_SECRET`
(the superadmin access tokens) in your local `.env`. Keep that file private.
The application refuses to start with blank keys, keys shorter than 32 bytes,
or identical keys. Redis authentication uses
`REDIS_PASSWORD` when configured on your Redis server. The `REDIS_DB`
variable selects the Redis index: default 1 hosts the dev sessions, the Go
suites pin index 2 through each package's `TestMain`, and the e2e cycle runs
on index 3 — never set it in `.env`: the Makefile exports `.env` to every
recipe, so `make run` would follow it.

Start the dev dependency stack (Docker images `postgres:18-alpine` and
`redis:8-alpine`, published loopback-only on `127.0.0.1:15432` /
`127.0.0.1:16379`):

```bash
make services_up
```

Run the database migrations. The first run needs a PostgreSQL superuser
because the migrations themselves create the `app_user` and `app_admin`
application roles (the application must never connect as the superuser) —
override `DB_USER`/`DB_PASS` just for this command. The Makefile reads the
same `.env` as the application (`-include .env`), so no manual export of the
database variables is needed:

```bash
make migrate_up DB_USER=postgres DB_PASS='<superuser password>'
```

`make` must be available on Windows (for example `choco install make`);
under WSL, run these commands inside the distribution. Starting the
containers does not replace the first migration: `services_up` boots an
empty cluster, and `make migrate_up` creates the schema.

Migrations are versioned SQL files in `db/migrations/`, managed with
[goose](https://github.com/pressly/goose). All application queries live in
`db/queries/*.sql` and are compiled to type-safe Go code with
[sqlc](https://github.com/sqlc-dev/sqlc).

- `app_user`: non-superuser role constrained by Row-Level Security
- `app_admin`: bypasses RLS, used only for platform-level operations
  (login, registration, admin backoffice)
- In production, set the application role passwords with
  `ALTER ROLE ... PASSWORD`

To leave the Docker dev stack and go back to native services: stop it with
`docker compose down` (add `-v` to also wipe the data), remove the
`services_up` prerequisites from the Makefile (git), and re-point `.env` to
the native instance (`DB_PORT=5432`, `REDIS_HOST=127.0.0.1:6379`).

Create the superadmin account (CLI bootstrap, no HTTP route involved):

```bash
go run . -create-admin -email you@example.com -password 'a long random password'
```

### Running

```bash
make run
```

Or directly:

```bash
go run .
```

### Building

```bash
go build -v -o gin-boilerplate .
./gin-boilerplate
```

### Testing

Invoice PDF content tests use `pdftotext` to verify the rendered invoice number,
line items, and totals — install a real poppler first:

- macOS: `brew install poppler`
- Ubuntu: `sudo apt-get install poppler-utils`
- Windows: `winget install --id oschwartz10612.Poppler` (open a new terminal
  afterwards so the updated PATH is picked up)

Then run without database services:

```bash
go test -mod=readonly -count=1 -race -v ./invoice
```

The full integration suite also requires running PostgreSQL and Redis.
`make test` starts the Docker dev stack first (prerequisite `services_up`)
and sources `.env` automatically before the tools run, so no manual export
is needed. On Windows, make also masks Git-for-Windows' bundled
`pdftotext` from the recipe PATH so the suite resolves the real poppler.

```bash
make test
```

Use a dedicated test database and Redis instance. The connection parser tests
require PostgreSQL trust authentication because they exercise different password
values. CI provisions disposable services for these checks. The dev stack
isolates sessions by Redis index: the dev app runs on index 1 (default), the
Go suites pin index 2 through each package's `TestMain`, and the e2e cycle
runs on index 3 (`REDIS_DB=3`) — the test purge cycles never evict the dev
sessions. `make test` also recreates the disposable `artisan_test` database
(migrations included), and the suites pin it via each package's `TestMain`:
the dev PostgreSQL database is never written by the tests.

### SSL (Optional)

The Docker dev stack runs without TLS: keep `SSL=FALSE` and the API answers
on `http://localhost:9001`.

To enable SSL, set `SSL=TRUE` in `.env` and generate certificates:

```bash
mkdir cert/
sh generate-certificate.sh
```

To disable SSL, set `SSL=FALSE` in `.env`.

## API Endpoints

### User
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/user/register` | No | Register a new user (creates the tenant and its owner) |
| POST | `/v1/user/login` | No | Login and receive JWT tokens with the tenant details |
| GET | `/v1/user/logout` | Bearer | Logout (invalidates token) |

### Article
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/article` | Bearer | Create an article |
| GET | `/v1/articles` | Bearer | Get all articles of the current tenant member |
| GET | `/v1/article/:id` | Bearer | Get one article |
| PUT | `/v1/article/:id` | Bearer | Update an article |
| DELETE | `/v1/article/:id` | Bearer | Delete an article |

### Auth
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/token/refresh` | No | Refresh access and refresh tokens |

### Admin
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/admin/login` | No | Superadmin login, returns an admin access token |
| GET | `/v1/admin/logout` | Admin | Logout (invalidates token) |
| GET | `/v1/admin/tenants` | Admin | List tenants |
| POST | `/v1/admin/tenants/:id/status` | Admin | Suspend or activate a tenant |

There is no public API: every data route requires a tenant (`Bearer`) or
admin token. Only the auth endpoints above (`register`, `login`,
`token/refresh`, `admin/login`) and the SPA shell are reachable without
identification.

## Swagger Docs

Generate and view API documentation:

```bash
make generate_docs
make run
```

Then open: `https://localhost:9001/swagger/index.html`

The Swagger UI is served only outside `ENV=PRODUCTION`; the generated
`docs/` spec stays available for tooling.

## Frontend (`app/`)

The `app/` directory is the SPA workspace (Vite + React 19 + TypeScript +
Tailwind CSS 4). Node 22.12+ and pnpm 11+ are required (versions pinned in
`package.json`; corepack supplies the exact pnpm):

```bash
cd app
pnpm install
pnpm dev            # Vite app — http://localhost:5173
pnpm storybook      # Storybook — http://localhost:6006
pnpm build          # tsc --noEmit + Vite build
pnpm lint           # ESLint (flat config)
pnpm build-storybook # Static Storybook build (storybook-static/)
```

Stories follow CSF and are colocated (`src/**/*.stories.tsx`). The
`src/design-system/` palette stories and the "Design System" docs pages are
generated at runtime from the CSS token contract in `src/index.css`
(`:root` oklch variables + the `.dark` class), so the Storybook theme toolbar
exercises the app's real dark-mode mechanism. See `app/README.md` for
details.

CI runs an independent `frontend` job (frozen-lockfile install, lint, build,
Storybook build smoke) next to the Go pipeline.

## End-to-End Testing (Playwright)

The E2E suite lives in `app/e2e/` and runs against a disposable database
(`artisan_e2e`) recreated at every cycle. The Go binary serves the built SPA
(`app/dist`) on a single origin (port 9100).

### Full Cycle

```bash
make e2e
```

This recreates the test database, runs migrations, seeds the superadmin and
owner tenant, builds the SPA, and executes the Playwright specs.

### Headed (Visual) Mode

By default Playwright runs headless. To watch the browser execute tests:

```bash
# From the app/ directory
pnpm exec playwright test --headed
```

Or use the Playwright UI for step-by-step debugging, traces, and watch mode:

```bash
pnpm exec playwright test --ui
```

> **WSL note**: On Windows 11, WSLg provides a Wayland compositor so headed
> Chromium windows appear on the Windows desktop automatically. On WSL 2
> without WSLg, you need an X server (e.g., VcXsrv) and `$DISPLAY` set.

### Prerequisites

- One-time browser install: `pnpm -C app exec playwright install chromium`
- PostgreSQL + Redis running
- `.env` present at the repo root
- `app/dist` built (done automatically by `make e2e`)

## Trusted Proxies & CORS

By default, `SetTrustedProxies(nil)` is configured — Gin will not trust any proxy headers. If you deploy behind a reverse proxy (nginx, CloudFlare, etc.), set your trusted proxy IPs:

```go
r.SetTrustedProxies([]string{"192.168.1.0/24", "10.0.0.0/8"})
```

CORS is configured in `middleware/cors.go`. Update the `Access-Control-Allow-Origin` header for your domain in production.

## Authentication

This boilerplate uses **Bearer Token** authentication:

1. **Login** returns an `access_token` (15 min) and `refresh_token` (7 days)
2. Include the access token in requests: `Authorization: Bearer <access_token>`
3. When the access token expires, use `/v1/token/refresh` with the refresh token to get new tokens
4. Both tokens are stored in Redis. Logout invalidates the presented access token;
   refresh tokens remain valid until rotation or expiry.

Refresh rotation checks the stored Redis owner and consumes the refresh token
atomically. If you restrict Redis commands with an ACL, allow `EVAL` and
`EVALSHA` for this operation.

Admin sessions use a single access token signed with the separate
`ADMIN_ACCESS_SECRET` (no admin refresh endpoint). It lives ~8 hours and is
revoked immediately on logout.

## Multi-Tenancy

Every user belongs to a tenant created at registration (`company` becomes a
unique slug). Data isolation is enforced by PostgreSQL Row-Level Security on
a shared schema with a `tenant_id` column — fail-closed policies that return
zero rows when no tenant context is set — never by application-side filtering
alone. The tenant always comes from the signed JWT, never from a client
header or hostname. Platform operations (login, registration, admin
backoffice) run through a dedicated role that bypasses RLS, while all
business queries run through the constrained role inside a tenant-scoped
transaction. Integration tests run as non-superuser roles, so the policies
are exercised for real.

## Project Structure

```
controllers/    HTTP handlers
models/         Database structs and business services (sqlc-generated queries)
forms/          Request validation structs and reusable error translation
middleware/     CORS, RequestID, JWT auth (tenant + admin), RBAC helper
internal/db/    sqlc-generated code + pgxpool bootstrap (tenant & admin pools)
db/queries/     SQL sources for sqlc — all application queries live here
db/migrations/  Versioned goose migrations
db/             Redis initialization
config/         Signing secrets validation
invoice/        HTML/PDF invoice generation library (no routes)
docs/           Swagger documentation (auto-generated)
public/         Static files and HTML templates
tests/          Integration tests
app/            Frontend workspace (Vite + React 19, Storybook)
```

## Adding a New Resource

1. Write the SQL queries in `db/queries/*.sql` and run `make generate_sqlc` (commit the regenerated `internal/db/` code)
2. Add a model in `models/` calling the generated queries through the tenant or admin store
3. Create form structs in `forms/` with binding tags and a `ValidationMessages` map
4. Create a controller in `controllers/` — use `forms.Translate(err, messages)` for validation errors
5. Register routes in `main.go` and run `make generate_docs` to update Swagger

## Previous Versions

- **v2.0** — JWT authentication with Redis (current architecture)
- **v1.x** — Session/cookie authentication ([v1 branch](https://github.com/Massad/gin-boilerplate/tree/v1-session-cookies-auth))

## License

MIT
