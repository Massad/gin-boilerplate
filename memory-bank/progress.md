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

## Limites connues (assumées, documentées)

- `./invoice` : test PDF local impossible en WSL (poppler absent) → CI only.
- Port 9000 occupé localement (conteneur teable) → smoke test CI only.
- Tenant suspendu : accès résiduel ≤ TTL access 15 min (trade-off accepté).
- Tout le travail du plan est en working tree **non commité**.

## Historique des étapes (plan multi-tenant)

1–10 : implémentation T1–T13 (schéma/RRLS, env/secrets, pools, tenant model, auth+Redis, user, articles, forms, middleware, models admin + contrôleurs + routage). 11 : CI (app_user/app_admin) + Swagger. 12 : E7 nettoyage sqlx/lib/pq + garde-fous + docs. Terminé le 2026-11-21.

## Backlog déferré (hors périmètre accepté)

Invitations/membres + routes `RequirePermission`, dashboard admin complet, SPA PWA, domaines custom, refresh httpOnly, multi-appartenance, facturation, persistance invoices.
