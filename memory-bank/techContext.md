# techContext.md — go-artisan-services

## Stack (go.mod, Go 1.26.10)

- Gin v1.12.0 + gzip ; PostgreSQL via **pgx/v5 v5.13.0** (pgxpool natif, `sql_package: pgx/v5`) ; requêtes compilées par **sqlc v1.27.0** ; migrations **goose v3.20.0** ; Redis client v7.4.1 ; JWT v4.5.2 ; bcrypt `golang.org/x/crypto` ; swag v1.16.6 ; maroto v2 (PDF) ; testify.
- Outils CLI épinglés dans le Makefile (`install_sqlc`, `install_goose`, `install_swag` → `.cache/tools/`, gitignored, retéléchargés au premier make après nettoyage).

## Schéma (db/migrations, goose)

- `000001_baseline.sql` : `user` + `article` — PK uuid `gen_random_uuid()`, timestamps `timestamptz`, trigger `set_updated_at()`, `UNIQUE(email)`. PostgreSQL ≥ 13 requis.
- `000002_multi_tenant.sql` : rôles cluster idempotents (DO blocks) `app_user` / `app_admin BYPASSRLS` (+ mots de passe dev documentés dedans), tables `tenant` / `admin_user` / `admin_audit` (sans FK tenant, sans RLS), `tenant_id uuid NOT NULL` + RLS `FORCE` (policy `NULLIF(current_setting('app.tenant_id', true), '')::uuid`), FK composite anti cross-tenant `article(tenant_id, user_id)`, grants.

## Variables d'environnement (.env, jamais commité)

`ENV`, `PORT`, `SSL`, `DB_*` (rôle `app_user`), `DB_ADMIN_*` (rôle `app_admin`), `REDIS_HOST`/`REDIS_PASSWORD`, 3 secrets ≥ 32 octets distincts : `ACCESS_SECRET`, `REFRESH_SECRET`, `ADMIN_ACCESS_SECRET` (`config.ValidateSigningSecrets`).

## Workflow dev

```bash
make run | make test                 # test = -tags=all ./... (PG+Redis requis, .env exporté)
make migrate_up                      # superuser la 1re fois (crée les rôles) : make migrate_up DB_USER=postgres DB_PASS=...
make generate_sqlc                   # après TOUTE modif de db/queries/ ou db/migrations/ — commit du généré (CI: git diff --exit-code internal/db)
make generate_docs                   # après toute annotation Swagger
go build -mod=readonly . && go vet -mod=readonly ./...
go test -mod=readonly -count=1 -race -tags=all . ./config ./models ./controllers ./middleware ./internal/db ./tests
```

- Garde-fous CI : `! grep` sur les imports legacy quotés (`sqlx`/`lib/pq`), diff du code sqlc généré, smoke test des routes, `govulncheck`.
- Exception WSL : `./invoice` (test PDF) exige `pdftotext` (poppler), non installable localement → CI only.

## Environnement local (WSL Debian)

- `wsl.exe -d Debian -- bash -lc '<cmd>'` pour du simple ; scripts ponctuels dans `.kilo/*.sh` (quotes imbriquées cassées par wsl.exe — nettoyer les scripts après usage).
- PG 15.19 127.0.0.1:5432 (superuser postgres/postgres trust, uniquement pour migrate_up) ; Redis 127.0.0.1:6379 sans password (DB 1 = cache teable — cleanup de tests ciblé sur `user_sessions:*`/`admin:*`) ; port 9000 occupé (ne pas smoke-tester le binaire en local).
