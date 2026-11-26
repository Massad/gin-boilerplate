# Règles d'architecture Go — modulaire par domaine

> Source unique : `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-design.md` (design acté, review EGM, décisions 2026-11-25).
> **Statut** : l'architecture modulaire n'est pas encore implantée. Ces règles décrivent la cible de la migration : appliquer chaque règle au moment de sa phase, ne jamais présumer la structure cible dans le code actuel. Les invariants « multi-tenant » ci-dessous s'appliquent déjà et doivent être préservés pendant TOUTE la migration.

## Layout cible

| Chemin | Rôle |
|---|---|
| `cmd/api/main.go` | Entrypoint : dotenv, flags, secrets, 2 pools pgx, Redis, `app.NewRouter`, run |
| `internal/app/router.go` | Composition root : `NewRouter(deps)` — unique source de la table de routes + SPA/static/swagger/NoRoute ; partagé par `cmd/api` ET la suite d'intégration |
| `internal/platform/config/` | Validation des secrets (`ValidateSigningSecrets`) |
| `internal/platform/redisdb/` | Client Redis, `RedisDBIndex` |
| `internal/platform/db/` | Code sqlc généré + `pool.go` manuscrit (package `db`, alias `pgdb`) |
| `internal/platform/middleware/` | CORS, RequestID uniquement (purs Gin) |
| `internal/platform/validate/` | `DefaultValidator`, `Translate` |
| `internal/platform/httpx/` | Responders, mapping pgx→HTTP, `ParseUUIDParam` |
| `internal/platform/mocks/` | Mocks mockery commités |
| `internal/modules/<domaine>/` | `handler.go`, `service.go`, `repository.go`, `dto.go`, `model.go`, `*_test.go` |
| `internal/modules/article/` | CRUD article (pilote de la migration) |
| `internal/modules/auth/` | Sessions/JWT, login unifié, refresh, logouts, middlewares `TokenAuth`/`AdminAuth`/`RequirePermission`, ports ; documente les clés de contexte (userID/tenantID/role) |
| `internal/modules/user/` | Identité, register (Tx tenant+owner), login ; implémente `UserAuthenticator` |
| `internal/modules/tenant/` | Slug + provisioning (+ données backoffice) |
| `internal/modules/admin/` | Backoffice, audit ; implémente `AdminAuthenticator` |
| `internal/modules/invoice/` | Lib PDF sans route HTTP ni DB (futur billing) |
| `db/migrations/`, `db/queries/` | Sources SQL uniquement (goose + sqlc) — jamais de Go racine après migration |

## Règles de dépendance (absolues)

1. `platform` n'importe JAMAIS `modules` — invariant absolu, pendant et après la migration.
2. `modules → modules` : EXCLUSIVEMENT via des interfaces définies côté CONSOMMATEUR (valable pour les ports d'orchestration et les services techniques). Jamais d'interface cross-module définie côté fournisseur.
3. `*pgdb.Queries` ne franchit JAMAIS une frontière de module (ni paramètre de port, ni retour). Échanges in-Tx inter-modules : Unit of Work `pgdb.WithinTx` + interface plateforme étroite `pgdb.InTx`.
4. Au terme de la migration, seuls `internal/app` et `cmd/api` importent tous les modules ; la fenêtre transitoire (Phases 3-6, décision A2) tolère les imports croisés racine↔modules dans les deux sens, uniquement pour le code en cours de migration — refermée en Phase 7 (critère de sortie : décompte `git ls-files` = 0).
5. Middlewares `TokenAuth()`/`AdminAuth()`/`RequirePermission` vivent dans le module `auth` (politique de domaine : claims, RBAC) — jamais dans `platform`.
6. Interdiction de dupliquer le wiring des routes : toute suite d'intégration pilote `app.NewRouter` (la copie manuelle type `SetupRouter` est prohibée).
7. Déplacement d'annotations Swagger ⇒ `make generate_docs` + commit de `docs/` dans la même phase.

## Règles par module

- Un module = `handler.go` (HTTP) + `service.go` (métier) + `repository.go` (données) + `dto.go` (bind) + `model.go` + tests colocalisés ; chaque phase laisse `go build ./...`, `go vet ./...`, `make test` au vert.
- Aucune logique métier dans les handlers : parse JWT, consommation du refresh, création de tokens, logout appartiennent au service du module.
- DI explicite : constructeurs injectant leurs dépendances ; pas de var de package ni de singleton (anti-patterns hérités type `authModel`/`GetRedis()` prohibés).
- Interfaces définies côté consommateur UNIQUEMENT là où un fake de test est utile (`ArticleRepository`, `TenantProvisioner`, ports `user`/`admin` vers `TokenService`) ; pas d'interface si l'implémentation est unique.
- Le module `auth` définit les ports `UserAuthenticator`/`AdminAuthenticator` ; `user`/`admin` les implémentent au câblage dans `internal/app` et déclarent eux-mêmes leur port `TokenService` (implémentation `*auth.Service` injectée).
- `RequirePermission` + la map rôle→permissions (module `auth`) restent l'UNIQUE point de politique RBAC ; les futures invitations/membres se borneront à câbler `RequirePermission(...)` dans `internal/app`.
- Register : orchestration par le module `user` via Unit of Work — `TenantProvisioner.CreateInTx` + `CreateOwnerTx` exposés en interfaces portant `pgdb.InTx` ; tenant + owner créés dans UNE seule transaction.

## Invariants multi-tenant (à préserver pendant TOUTE la migration)

- RLS échec fermé : contexte tenant réglé UNIQUEMENT par `WithTenant` (transaction, `set_config(..., true)`) — jamais de `SET` session-wide sur un pool partagé.
- Pool admin (`app_admin BYPASSRLS`, `pgdb.AdminStore()`) : login, register, refresh, backoffice. Pool tenant (`app_user`, `pgdb.TenantStore()` + `WithTenant`) : tout le métier.
- Login unifié (module `auth`) : bcrypt à coût constant contre un hash factice ; pas d'énumération ; un email superadmin ne recouvre jamais un email user (user trouvé + mauvais mdp → 406, sans tentative admin).
- SQL uniquement dans `db/queries/*.sql` → sqlc ; zéro SQL concaténé en Go.
- Erreurs pgx : `ErrNoRows`→404 ; `23505`→406/409 ; `23503`→409 ; `42501`→500 (RLS) ; jamais de texte SQL brut au client.
- Tests : tout nouveau paquet de test d'intégration exige le `TestMain` épinglant `REDIS_DB=2` + `DB_NAME=artisan_test` ; secrets de test via `t.Setenv` uniquement ; rôles non-superuser obligatoires (42501 sans contexte tenant = comportement attendu).

## Contrat HTTP & soft-consume du refresh (décision 7)

- Contrat HTTP inchangé AU BYTE PRÈS pendant toute la migration ; formes `gin.H{"message": ...}` actuelles conservées, casse mixte `"Message"`/`"message"` incluse (homogénéiser seulement si les tests sont mis à jour dans la même phase).
- Chaque phase : pas de changement de comportement HTTP ; smoke CI vert à chaque branche ; `make e2e` rejoué aux Phases 1 et 7 (seules couches touchant entrypoint/assemblage du routeur).
- Soft-consume : jamais de `DEL` immédiat — le record STRING (adressé par l'UUID du refresh JWT ; le token n'est jamais stocké, en clair ni haché) porte l'état dans sa valeur, jamais l'inverse : actif `userID:tenantID:role` (format inchangé, aucune migration de schéma) → consommé `consumed:<delivered|recovered>:userID:tenantID:role`, TTL = grâce fixe 5 min (min(grâce, TTL restant)) ; atomicité Lua inchangée ; préfixe userID de la valeur comparé avant toute consommation (ownership, `TestRefreshOnlyConsumesMatchingOwner` déplacé avec le module).
- Politique de rejeu : `delivered=1` → 401 (forme du chemin invalide) ; `delivered=0` pendant la grâce → UNE re-émission `recovered` (200, forme happy path ; état encodé dans la valeur du record, jamais dans la réponse HTTP), sinon 401 ; crash (`delivered` inconnu) → échec fermé.
- Le sweep du module `auth` (ticker ~5 min) n'est JAMAIS la dépendance de purge : observabilité seulement ; l'éviction par TTL est primaire. Tests dédiés exigés pour le carve-out : rejeu, grâce, recovery unique, fails-closed, sweep, unicité de la recovery.

## Interdits

- Montée de version go-redis v7→v9 ou jwt/v4→v5 pendant la migration — follow-up dédié après Phase 8.
- Tout changement de comportement HTTP (backlog explicite : CORS configurable, refresh httpOnly, câblage RBAC membres/invitations, normalisation des payloads d'erreur, homogénéisation de la casse des clés).
