# progress.md — go-artisan-services

> Mise à jour à chaque session ; tracer les changements et décisions majeures.

## Ce qui fonctionne (validé par build/vet/tests `-race`)

- [x] Socle pgx/v5 (pgxpool) + sqlc + goose — plus aucune dépendance sqlx/lib/pq (garde-fou CI en place)
- [x] Migrations goose 000001 (uuid/timestamptz) + 000002 (rôles idempotents, tables tenant/admin, RLS FORCE échec fermé, grants) — idempotentes, rejouables
- [x] 2 pools : `TenantStore` (app_user, RLS) / `AdminStore` (app_admin BYPASSRLS) + `WithTenant` (`set_config` transactionnel)
- [x] Auth tenant (claims tenant_id/role, Redis `userID:tenantID:role`, index `user_sessions:*`, Lua refresh adapté, pool admin pour login/register/refresh)
- [x] Auth admin (token unique ~8h, secret/préfixe/pool séparés, rejet croisé, pas de refresh admin)
- [x] Articles tenant-aware via `WithTenant` (backlog RBAC `RequirePermission` posé, non câblé — YAGNI)
- [x] Backoffice admin (login, tenants, toggle status + `admin_audit` atomique) + CLI `-create-admin` idempotent (testé)
- [x] Principe « rien de public » : lookup `/v1/public/:slug` et routes invoice démo retirés, Swagger seulement hors production ; fallback SPA `r.NoRoute` (404 sur `/v1*` et non-GET)
- [x] Tests : unitaires purs + intégration tag `all` (7 packages), isolation croisée, RLS fail-closed (0 ligne sans `set_config`), insert cross-tenant 42501, un seul owner, backoffice — CI en non-superuser
- [x] CI : migrate_up superuser → tests en `app_user`/`app_admin`, sqlc-diff, legacy-drivers guard, smoke, govulncheck
- [x] Docs : Swagger, README, AGENTS.md, memory bank
- [x] Frontend `app/` (2026-11-22) : Storybook 10.6.8 (react-vite, docs/a11y/themes), harness tokens `src/design-system/` (palettes runtime + MDX), gap lint réparé (flat config), `.storybook` dans tsconfig, job CI `frontend` — `pnpm lint` / `pnpm build` / `pnpm build-storybook` verts, `storybook doctor` OK

## Limites connues (assumées, documentées)

- `./invoice` : sous Windows natif, poppler installé (winget `oschwartz10612.Poppler`) et cible `test` masque le xpdf de Git for Windows → tests PDF verts en local ; en WSL, exiger l'installation de `poppler-utils` (non installé).
- Port 9000 occupé localement (conteneur teable) → smoke test CI only.
- Tenant suspendu : accès résiduel ≤ TTL access 15 min (trade-off accepté).
- Tout le travail du plan est en working tree **non commité**.
- Frontend : validation 100 % locale (lint/build/storybook build/doctor) — le job `frontend` n'a pas encore tourné sur GitHub.

## Historique des étapes (plan multi-tenant)

1–10 : implémentation T1–T13 (schéma/RRLS, env/secrets, pools, tenant model, auth+Redis, user, articles, forms, middleware, models admin + contrôleurs + routage). 11 : CI (app_user/app_admin) + Swagger. 12 : E7 nettoyage sqlx/lib/pq + garde-fous + docs. Terminé le 2026-11-21.
13 : install Storybook dans `app/` (plan `1795339024940-install-storybook-front.md`) — groundwork design-system, 2026-11-22.
14 : correction de l'environnement de test Windows (2026-11-22) — `.env` déquoté + `REDIS_PASSWORD` vidé, purge des données de test orphelines, Poppler 26.09 via winget, cible `test` du Makefile auto-source `.env` et masque sous MSYS/MINGW le pdftotext xpdf embarqué de Git for Windows ; `make test` vert (8 packages, zéro SKIP). README.md + AGENTS.md alignés. Non commité.
15 : stack de test intégrée (2026-11-23, plan `1795428465642-stack-test-go.md`) — mockery v2.53.8 épinglé (mocks commités, génération idempotente), gofakeit/v7 + testfixtures/v3 + testcontainers-go + goose v3.20.0 en go.mod, helpers `internal/testutil/` (factories seedées, chargeur de fixtures, Postgres éphémère), fixtures YAML `tests/fixtures/`, retrofit `require` des 7 fichiers de test, garde CI « fuite helpers ». Suite `all` verte avec fixtures/factories ; testcontainers SKIP sans Docker. Non commité (convention : commits quand l'utilisateur le décide).

## Backlog déferré (hors périmètre accepté)

Invitations/membres + routes `RequirePermission`, dashboard admin complet, SPA PWA, domaines custom, refresh httpOnly, multi-appartenance, facturation, persistance invoices.
