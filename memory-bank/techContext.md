# techContext.md — go-artisan-services

## Stack (go.mod, Go 1.26.10)

- Gin v1.12.0 + gzip ; PostgreSQL via **pgx/v5 v5.13.0** (pgxpool natif, `sql_package: pgx/v5`) ; requêtes compilées par **sqlc v1.27.0** ; migrations **goose v3.20.0** ; Redis client v7.4.1 ; JWT v4.5.2 ; bcrypt `golang.org/x/crypto` ; swag v1.16.6 ; maroto v2 (PDF) ; testify.
- Outils CLI épinglés dans le Makefile (`install_sqlc`, `install_goose`, `install_swag` → `.cache/tools/`, gitignored, retéléchargés au premier make après nettoyage).

## Stack frontend (app/)

- Vite 8 + React 19 (TypeScript strict, `tsc --noEmit` dans `pnpm build`) ; Tailwind CSS 4 (`@tailwindcss/vite`, `tsc --noEmit` + Vite) ; tests vitest ; pnpm 11.20.0 épinglé (`packageManager` → corepack, pas d'installation globale), Node ≥ 22.12.
- `app/pnpm-workspace.yaml` : `allowBuilds` (esbuild approuvé — pnpm 11 exige l'approbation des scripts de build, sinon `pnpm install` échoue avec `ERR_PNPM_IGNORED_BUILDS`) + `minimumReleaseAgeExclude`.
- Storybook 10.6.8 (`@storybook/react-vite`, addons **docs / a11y / themes**) : `app/.storybook/main.ts` (`defineMain`, CSF) + `preview.tsx` (import `../src/index.css`, `withThemeByClassName` light `''` / dark `'dark'`) ; stories CSF colocalisées (`src/**/*.stories.tsx`) + pages MDX ; palettes `app/src/design-system/` lues à l'exécution via `getComputedStyle` (contrat jetons = variables CSS de `index.css`, pont « Open Design » futur, zéro rework).
- Workflow frontend : `pnpm lint` (flat config couvrant `.storybook/`), `pnpm build`, `pnpm build-storybook` (smoke CI), `storybook doctor` (triage).

## Schéma (db/migrations, goose)

- `000001_baseline.sql` : `user` + `article` — PK uuid `gen_random_uuid()`, timestamps `timestamptz`, trigger `set_updated_at()`, `UNIQUE(email)`. PostgreSQL ≥ 13 requis.
- `000002_multi_tenant.sql` : rôles cluster idempotents (DO blocks) `app_user` / `app_admin BYPASSRLS` (+ mots de passe dev documentés dedans), tables `tenant` / `admin_user` / `admin_audit` (sans FK tenant, sans RLS), `tenant_id uuid NOT NULL` + RLS `FORCE` (policy `NULLIF(current_setting('app.tenant_id', true), '')::uuid`), FK composite anti cross-tenant `article(tenant_id, user_id)`, grants.

## Variables d'environnement (.env, jamais commité)

`ENV`, `PORT`, `SSL`, `DB_*` (rôle `app_user`), `DB_ADMIN_*` (rôle `app_admin`), `REDIS_HOST`/`REDIS_PASSWORD`, `REDIS_DB` (index Redis : défaut 1 = dev ; suites Go = 2 via `TestMain` ; cycle e2e = 3 via préfixe make/webServer — jamais posée dans `.env`), 3 secrets ≥ 32 octets distincts : `ACCESS_SECRET`, `REFRESH_SECRET`, `ADMIN_ACCESS_SECRET` (`config.ValidateSigningSecrets`).

## Workflow dev

```bash
make run | make test                 # test = -tags=all ./... (PG+Redis requis ; la cible source .env elle-même)
make migrate_up                      # superuser la 1re fois (crée les rôles) : make migrate_up DB_USER=postgres DB_PASS=...
make generate_sqlc                   # après TOUTE modif de db/queries/ ou db/migrations/ — commit du généré (CI: git diff --exit-code internal/db)
make generate_docs                   # après toute annotation Swagger
go build -mod=readonly . && go vet -mod=readonly ./...
go test -mod=readonly -count=1 -race -tags=all . ./config ./models ./controllers ./middleware ./internal/db ./tests
```

- Garde-fous CI : `! grep` sur les imports legacy quotés (`sqlx`/`lib/pq`), diff du code sqlc généré, smoke test des routes, `govulncheck`.
- `./invoice` (test PDF) exige un poppler réel : Windows natif OK (winget `oschwartz10612.Poppler` ; la cible `test` masque le pdftotext xpdf embarqué de Git for Windows) ; en WSL il faut installer `poppler-utils`.

## Environnement local (WSL Debian)

- `wsl.exe -d Debian -- bash -lc '<cmd>'` pour du simple ; scripts ponctuels dans `.kilo/*.sh` (quotes imbriquées cassées par wsl.exe — nettoyer les scripts après usage).
- Stack dev dépendances : conteneurs Docker (`docker-compose.yml` à la racine, projet `name: artisan-dev`) — PostgreSQL `postgres:18-alpine` sur `127.0.0.1:15432` (superuser postgres/postgres trust) et Redis `redis:8-alpine` sur `127.0.0.1:16379` (sans password) ; volumes nommés `pgdata`/`redisdata`, images identiques dans le compose, la CI (`ci.yml`) et le testutil éphémère ; démarrage via les cibles make (`make services_up` → prérequis de `run`/`test`/`e2e`, jamais des migrations ; contrat : `migrate_up` superuser crée le schéma). API sur 9001, Vite 5173 ; répartition des index Redis : index 1 = sessions dev, index 2 = suites Go, index 3 = cycle e2e (`REDIS_DB`) (l'ancien Redis 6379 était partagé : DB 1 = cache teable) ; bases PostgreSQL : dev `golang_gin_db`, jetable `artisan_test` (suites Go, recréée puis migrée par chaque `make test`), `artisan_e2e` (cycle e2e).
