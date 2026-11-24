# HANDOFF — Plan multi-tenant RLS : TERMINÉ (étapes 1–12 / 12)

Plan source : `C:\Users\schat\.local\share\kilo\plans\1795250166405-multi-tenant-postgres-rls.md`
Conventions : `AGENTS.md` (racine, mise à jour à l'étape 12). Toujours `-mod=readonly` sur les commandes go.

## Étape 12 (T18 / E7) — TERMINÉ

1. **Dépendances legacy** : aucun import `jmoiron/sqlx` ni `lib/pq` dans les `*.go` (seule mention restante = commentaire doc DSN de `internal/db/pool.go:70-72`, hors path d'import). `go.mod`/`go.sum` étaient déjà propres : `go mod edit -droprequire` × 2 + `go mod tidy` + `go mod verify` sont des no-ops (diff pre/post tidy vide, « all modules verified »). Attention : `go mod tidy` ne touche pas go.sum au-delà des entrées nécessaires ; les drops explicites restent documentés dans le plan.
2. **Garde-fou CI (régression)** — `.github/workflows/ci.yml`, étape `Check removed legacy drivers` ajoutée après `Verify generated sqlc code` :
   `! grep -rEn '"(github\.com/)?(jmoiron/sqlx|lib/pq)"' --include='*.go' .`
   **Écart volontaire vs la proposition du HANDOFF** (`grep -rn "jmoiron/sqlx\|lib/pq"` sans quotes) : le pattern simple aurait false-positive sur le commentaire `... previous lib/pq driver ...` de `internal/db/pool.go:71` (écrit à la main, pas généré). Le pattern ne matche que les littéraux d'import quotés — un commentaire en prose ne casse pas la CI. Sémantique testée localement (exit 0 sans match, non-zéro si re-introduction).
3. **AGENTS.md** réécrit : description stack (pgx/v5 pgxpool + sqlc + goose, RLS), commandes (`make migrate_up` JAMAIS `psql <` direct, `make generate_sqlc` + commit du généré + check CI `git diff --exit-code internal/db`, CLI `-create-admin`, secrets 3×), structure (`internal/db` généré + `pool.go` 2 pools, `db/migrations`, `db/queries`, `controllers/admin.go|public.go`, `models/admin.go|tenant.go`), conventions (requêtes uniquement dans `db/queries/*.sql` sans concat Go, plateforme → `AdminStore()` BYPASSRLS / métier → `TenantStore()` + `WithTenant` via `set_config(..., true)`, erreurs pgx 23505/23503/42501, RLS fail-closed + FORCE, 401/403), tests (unitaires sans tag + intégration tag `all` en rôles non-superuser, RLS prouvée = comportement attendu).
4. **README.md** mis à jour : stack pgx/sqlc/goose + multi-tenant RLS + backoffice admin dans « What's Included » ; secret admin générique (aucun mot de passe dev réel) ; remplacé `psql < db/database.sql` par `make migrate_up` avec superuser uniquement pour la première migration (les migrations créent les rôles) ; section superadmin CLI ; endpoints admin/public ; auth admin (1 token ~8h, pas de refresh) ; section « Multi-Tenancy » ; structure projet ; flux « Adding a New Resource » sqlc.
5. **Non fait (voulu)** : pas de commit (aucun demandé) ; `tests/`, `models/`, `internal/db/` (généré) non touchés ; `make generate_docs` inutile — aucune annotation Swagger modifiée à l'étape 12 (docs régénérées à l'étape 11).
6. **Validation complète** (WSL Debian, PG 15.19 + Redis actifs, `.env` exporté, Go 1.27.4) :
   - `go build -mod=readonly .` ✓, `go vet -mod=readonly ./...` ✓
   - `go test -mod=readonly -count=1 -race -tags=all ./config ./models ./controllers ./middleware ./internal/db ./tests` ✓ (integration suite `tests` 11.5s en `app_user`/`app_admin` : RLS réellement prouvée)
   - `go mod verify` ✓ ; simulateur du garde-fou CI ✓

## Conclusion du plan

**Aucune tâche restante.** Les 12 étapes (T1–T18, parties I + II) sont implémentées et validées : migration pgx/v5 + sqlc + goose achevée (E7 nettoyage inclus), multi-tenant RLS échec fermé prouvé en tests et en CI en rôles non-superuser, backoffice admin + audit, CLI superadmin, fallback SPA, docs (Swagger/AGENTS/README) et CI à jour. Tout le travail reste en working tree non commité (aucun commit demandé).

## Environnement d'exécution (référence)

- `wsl.exe -d Debian -- bash -lc '<cmd>'` pour les commandes simples ; scripts complexes dans `.kilo/*.sh` (écrire puis exécuter — inline nested quotes cassées par wsl.exe) ; `export PATH=$PATH:/usr/local/go/bin` (Go 1.27.4).
- Stack dev Docker (`docker-compose.yml`, projet `artisan-dev`) : PostgreSQL `postgres:18-alpine` `127.0.0.1:15432` (superuser `postgres`/`postgres` trust, uniquement pour `make migrate_up`), bases : dev `golang_gin_db` (goose version 3), jetables `artisan_test` (suites Go, provisionnée par chaque `make test`) / `artisan_e2e` (cycle e2e), Redis `redis:8-alpine` sans mot de passe `127.0.0.1:16379` (répartition des index Redis à jour à l'isolation REDIS_DB : 1 = sessions dev, 2 = suites Go via `TestMain`, 3 = cycle e2e via `REDIS_DB=3` ; la mention historique « DB 1 dédiée à ce projet » ci-dessous restait vraie pour l'index dev, cette section est encore utilisée comme référence), `.env` local sur `app_user`/`app_admin` (non commité).
- Exception connue : `./invoice` (test PDF) exige `pdftotext` (poppler), non installable sans sudo en WSL — la CI l'installe. Port 9000 occupé (conteneur teable) : pas de smoke local.

## Pièges pour une reprise éventuelle

- requêtes sqlc `:one ... RETURNING id` → retourne directement `uuid.UUID` (`.ID` est une méthode `func() uint32`)
- `internal/db` généré : ne pas éditer ; accesseurs `TenantStore()`/`AdminStore()` ; timestamptz → `time.Time` (overrides sqlc.yaml)
- contexte middleware : claims stockés en STRINGS puis `uuid.Parse` côté contrôleurs
- CI : migrations en superuser via override make (`DB_USER=postgres DB_PASS=postgres`), tests ensuite en `app_user` — ne pas inverser
- `admin_audit.tenant_id` sans FK (l'audit survit à la suppression d'un tenant) ; `tenant`/`admin_user`/`admin_audit` sans RLS
