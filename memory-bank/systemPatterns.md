# systemPatterns.md — go-artisan-services

## Architecture

```
controllers/ (HTTP, Swagger annotations)  →  models/ (services métier)
        → pgdb (internal/db) : Store{Queries sqlc, Pool pgxpool, Role}
db/queries/*.sql --sqlc--> internal/db/*.sql.go (généré, commité, ne pas éditer)
db/migrations/*.sql (goose) ← schema source de sqlc
```

- main.go : route `/v1`, `db.InitRedis`, `pgdb.InitTenant` + `pgdb.InitAdmin` (+ `Close`), CLI `-create-admin` avant tout démarrage HTTP (`runCreateAdmin` testable), fallback SPA.

## Deux pools, deux rôles (défense en profondeur RLS)

- `TenantStore()` — pool `app_user` non-superuser : tout le métier, contraint à 100 % par RLS.
- `AdminStore()` — pool `app_admin BYPASSRLS` : opérations « plateforme » uniquement (login, register, refresh, backoffice admin) — contournement du problème œuf-poule (un SELECT user sans contexte tenant renverrait 0 ligne).

## Contexte tenant (pattern critique)

```go
pgdb.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error { … })
```
= transaction + `SELECT set_config('app.tenant_id', $1, true)` (auto-reset au commit/rollback) → pool-safe. Jamais de `SET` session-wide. Les requêtes sqlc tenant-scoped portent aussi `@tenant_id` (double défense : RLS + paramètre).

## Auth (models/auth.go)

- Tenant : access `authorized/access_uuid/user_id/tenant_id/role` + refresh `refresh_uuid/user_id/tenant_id` ; valeur Redis `userID:tenantID:role` ; index sessions `user_sessions:{userID}` (SADD/SREM/revocation, TTL aligné).
- Admin : un seul token ~8h (claims `admin_id`, `role:"superadmin"`), signé `ADMIN_ACCESS_SECRET`, Redis préfixé `admin:`, **pas de refresh admin** ; rejet croisé des tokens par direction.
- Refresh : consomme le token (Lua `ConsumeRefreshAuth` — match du 1ᵉʳ segment avant `:`), puis RELIT le user via le pool admin (rôle frais, existence, `tenant.status`) — jamais le rôle des claims.
- Middleware : `TokenAuth` set `userID/tenantID/role` (strings → `uuid.Parse` côté contrôleurs), `AdminAuth`, `RequirePermission` (map `rolePermissions`, posée non câblée).

## Conventions transverses

- Erreurs : `pgx.ErrNoRows` → 404 ; `pgconn.PgError` par Code (`23505` → 406/409, `23503` → 409, `42501` → 500) ; jamais de texte SQL brut au client ; validation via `forms.Translate(err, forms.XxxMessages)`.
- Réponses `gin.H{"message": ...}` ; JSON tenant embarqué au login (`user`, `tenant`, `token`) ; Register ne renvoie pas de tokens.
- Docs : godoc sur les exportés, pas de commentaires superflus ; Swagger régénéré via `make generate_docs` (swag v1.16.6) après toute annotation.

## Tests (pattern à respecter pour toute feature)

- Unitaires sans tag (config secrets, tenant slugify) + intégration sous `//go:build all` (httptest + `SetupRouter`, suite séquentielle à variables de package : `TestIntDB` init, `TestCleanUp` cleanup ciblé).
- La CI tourne en rôles non-superuser → la RLS est réellement prouvée ; une requête sans contexte renvoie 0 ligne et un insert cross-tenant lève 42501 : comportement **attendu**, ne jamais « corriger » au superuser.
