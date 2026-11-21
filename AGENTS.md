# AGENTS.md — go-artisan-services

API REST Go (Gin) avec PostgreSQL (pgx/v5 via pgxpool + requêtes sqlc, migrations goose), Redis (tokens JWT), Swagger et génération de PDF (maroto). Multi-tenant : schéma partagé + `tenant_id` + Row-Level Security échec fermé. Module : `github.com/Massad/gin-boilerplate`, Go 1.26.10.

## Commandes

- `make run` — lance l'app (`go run *.go`), nécessite `.env` chargé
- `make test` — suite complète : `go test -mod=readonly -count=1 -v -tags=all ./...` (nécessite PostgreSQL + Redis actifs + `.env` exporté : `set -a; . ./.env; set +a`)
- `make migrate_up` / `make migrate_down` — migrations goose (`db/migrations/`, épinglé dans le Makefile). JAMAIS de `psql < db/*.sql` direct. Le superuser n'est requis que pour migrer (la migration 000002 crée les rôles `app_user`/`app_admin`) ; l'app et les tests se connectent en rôles non-superuser
- `make generate_sqlc` — régénère `internal/db/` ; à rejouer après TOUTE modification de `db/queries/` ou `db/migrations/`, code généré commité (la CI le vérifie par `git diff --exit-code internal/db`)
- `go run *.go -create-admin -email <e> -password <p>` — bootstrap superadmin (CLI, aucune route HTTP) ; requiert le pool admin
- `go test -mod=readonly -count=1 -race -v ./invoice` — tests PDF sans base de données (requiert `pdftotext` / poppler)
- `go build -mod=readonly .` et `go vet -mod=readonly ./...` — build + vet (toujours avec `-mod=readonly`)
- `make generate_docs` — régénère la doc Swagger (swag épinglé v1.16.6, installé via `make install_swag`)
- CI (`.github/workflows/ci.yml`) : `go mod verify`, build, vet, vérif du code sqlc généré (`git diff --exit-code internal/db`), garde-fous « legacy drivers » (aucun import `sqlx`/`lib/pq`), tests avec `-race -tags=all`, smoke test des routes, `govulncheck ./...`. Migrations appliquées en superuser (`make migrate_up DB_USER=postgres DB_PASS=postgres`) puis tests exécutés en `app_user`/`app_admin` non-superuser — la RLS est réellement prouvée

## Structure

- `main.go` — point d'entrée, routes `/v1`, annotations Swagger principales, fallback SPA (`r.NoRoute`), CLI `-create-admin`
- `controllers/` — handlers HTTP (structs vides, méthodes avec annotations Swagger) ; `admin.go` (backoffice), `public.go` (lookup tenant public), `user.go` (auth), `article.go` (métier), `auth.go` (refresh)
- `models/` — structs DB et services métier ; appellent le code sqlc via `pgdb.TenantStore()` / `pgdb.AdminStore()` et `pgdb.WithTenant`
- `forms/` — structs de validation (tags `binding:`) + maps `XxxMessages`
- `middleware/` — CORS, RequestID, `TokenAuth()` (JWT tenant), `AdminAuth()` (JWT admin, secret séparé), `RequirePermission` (RBAC minimal owner/member)
- `internal/db/` — code GÉNÉRÉ sqlc (ne pas éditer à la main) + `pool.go` écrit à la main : deux pools pgx (tenant `app_user` / admin `app_admin BYPASSRLS`), `WithTenant`, `postgresDSN`
- `db/` — init Redis (`InitRedis`/`GetRedis`)
- `db/migrations/` — migrations goose versionnées (000001 baseline uuid/timestamptz, 000002 multi-tenant : rôles, tables tenant/admin, RLS, grants)
- `db/queries/` — sources SQL de sqlc (articles, users, tenants, admins) : TOUTE requête SQL vit ici
- `db/` racine : Redis seulement
- `config/` — validation des 3 secrets de signature (`ValidateSigningSecrets`)
- `invoice/` — exemple de facture HTML/PDF
- `docs/` — Swagger auto-généré (ne pas éditer à la main), `public/` — templates/statiques, `tests/` — tests d'intégration (tag `all`)

## Conventions

- Ajouter une ressource : requêtes dans `db/queries/*.sql` → `make generate_sqlc` → model dans `models/` (accès via `pgdb.TenantStore()`/`pgdb.AdminStore()`, métier tenant via `pgdb.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error)`) → form dans `forms/` (tags `binding:` + map `XxxMessages`) → controller dans `controllers/` → route dans `main.go` → `make generate_sqlc` (si requêtes) et `make generate_docs`
- Requêtes SQL uniquement dans `db/queries/*.sql` — aucune concaténation en Go (`fmt.Sprintf`, `+`), revu en review et refusé
- Pools : opérations « plateforme » (login, register, refresh, backoffice admin) → pool admin `pgdb.AdminStore()` (rôle `app_admin BYPASSRLS`) ; tout le métier tenant → pool `pgdb.TenantStore()` + `WithTenant` (contexte RLS réglé exclusivement par `set_config(..., true)` en transaction — jamais de `SET` session-wide sur un pool partagé)
- Multi-tenant : RLS échec fermé (`FORCE ROW LEVEL SECURITY`, policy `NULLIF(current_setting('app.tenant_id', true), '')::uuid`) ; le tenant vient UNIQUEMENT du claim JWT signé (jamais d'un header/hôte client) ; `tenant`/`admin_user`/`admin_audit` sans RLS (tables plateforme)
- Erreurs pgx : `pgx.ErrNoRows` → 404 ; `pgconn.PgError` par `Code` (`23505` unique → 406/409, `23503` FK → 409, `42501` RLS → 500) ; ne jamais renvoyer le texte SQL brut au client
- Erreurs de validation : toujours passer par `forms.Translate(err, forms.XxxMessages)` puis `c.AbortWithStatusJSON(status, gin.H{"message": ...})`
- Réponses JSON via `gin.H{"message": ...}` ; sessions JWT (Redis) gérées dans `models/auth.go` avec rotation atomique et index `user_sessions:*` ; tokens admin préfixés `admin:`, signés avec `ADMIN_ACCESS_SECRET`
- 401 = non authentifié ; 403 = rôle insuffisant (`RequirePermission`) ou tenant suspendu
- Commentaires godoc sur les types/fonctions exportés ; pas de commentaires superflus dans le code

## Tests

**Obligatoire : toute nouvelle feature doit être livrée avec ses tests, et les tests doivent être rejoués après chaque ajout/modification de code — le travail n'est pas terminé tant que `go vet` et la suite de tests passent.**

Deux niveaux :

- **Unitaires** (sans services, sans tag) : `config/secrets_test.go` (3 secrets ≥ 32 octets distincts), `models/tenant_test.go` (slugify/réservés, fonctions pures), `invoice/invoice_test.go` (requiert `pdftotext` / poppler) — exécutables via `go test -mod=readonly -count=1 -race ./<pkg>`
- **Intégration** (derrière le tag `all`) : `tests/article_test.go` (suite HTTP complète via `httptest` + `SetupRouter`, séquentielle : `TestIntDB` initialise, `TestCleanUp` nettoie, état partagé via variables de package ; isolation croisée, RLS fail-closed en SQL brut, rôles admin), `controllers/auth_test.go`, `middleware/auth_test.go`, `internal/db/pool_test.go` (parse DSN pgx, mots de passe espacés/quotés) — nécessitent PostgreSQL + Redis actifs + `.env` exporté (`set -a; . ./.env; set +a` puis `make test`)

Conventions de test :

- testify (`assert`), `httptest` + `gin.SetMode(gin.TestMode)`, `binding.Validator = new(forms.DefaultValidator)` pour les routes testées
- Secrets de test via `t.Setenv` (jamais de vraies clés), isolation via `t.Cleanup`, tests table-driven avec sous-tests quand pertinent
- Tests Redis : skip si `REDIS_HOST` absent ; utiliser une base/instance de test dédiée ; les tests de parse de connexion exigent l'authentification `trust` PostgreSQL
- Les tests se connectent en rôles non-superuser (`app_user`/`app_admin`) : une requête sans contexte tenant renvoie 0 ligne (RLS) et un insert cross-tenant lève 42501 — comportement attendu et testé, ne pas « corriger » en passant en superuser

Règles obligatoires :

- Nouvelle feature → écrire les tests en même temps que le code : cas unitaires sans dépendances quand possible + cas d'intégration dans la suite `all`
- Après chaque ajout/modification de code : `go vet -mod=readonly ./...`, puis les tests du/des package(s) touché(s), puis `make test` complet
- Ne jamais terminer un travail avec des tests en échec ou sautés par erreur ; si l'environnement manque (PostgreSQL/Redis indisponibles), le signaler explicitement

## Environnement & sécurité

- Copier `.env_rename_me` vers `.env` ; jamais commit `.env` ni de clés
- `ACCESS_SECRET`, `REFRESH_SECRET` et `ADMIN_ACCESS_SECRET` : ≥ 32 octets, générés indépendamment (`openssl rand -hex 32`) et tous différents — l'app refuse de démarrer sinon
- `DB_USER`/`DB_PASS` : rôle applicatif `app_user` NON-superuser (RLS active) ; `DB_ADMIN_USER`/`DB_ADMIN_PASS` : rôle plateforme `app_admin` `BYPASSRLS` — les deux rôles sont créés par la migration 000002 elle-même ; mots de passe dev documentés dans la migration, la prod via `ALTER ROLE ... PASSWORD`
- Variables principales : `PORT`, `ENV` (PRODUCTION → mode release), `SSL` (TRUE/FALSE), `DB_*`, `DB_ADMIN_*`, `REDIS_HOST`/`REDIS_PASSWORD`
- Schéma DB : migrations goose dans `db/migrations/` (`make migrate_up`) ; PK uuid (`gen_random_uuid()`, PostgreSQL ≥ 13), timestamps `timestamptz`, `updated_at` via trigger `set_updated_at()`
- Toujours `-mod=readonly` sur les commandes go ; vérifier `go vet` et `make test` avant de terminer

## Memory bank

Contexte projet persistant dans `memory-bank/` (`projectbrief`, `productContext`, `systemPatterns`, `techContext`, `activeContext` — à mettre à jour à chaque session —, `progress`).
