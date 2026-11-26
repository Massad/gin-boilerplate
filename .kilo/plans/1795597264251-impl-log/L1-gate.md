# Gate L1 — Socle modulaire (`internal/platform`, `cmd/api`, `internal/app`)

**Date** : 2026-11-26 · **Mode** : branche `refactor/modular-phase-1` créée après validation utilisateur — commit code `3dc7869` (mouvements reconnus R97-R100 par git) + commit docs ; fourche `d7f6a22` = `modular_archi` + L0, car la séquence L1↪L0 exige les commits L0 non encore fusionnés dans la baseline · Plan-cible : plan-phases l.12-22 (P1.1-P1.8) · **Verdict : PASS** (CI distante au push)

## Commandes jouées + résultats (toutes dans WSL)

1. `go build -mod=readonly ./...` → **BUILD_OK** ; `go vet -mod=readonly ./...` + `-tags=all ./tests/ ./cmd/api/ ./middleware/ ./controllers/ ./internal/testutil/` → **VET_ALL_OK** ; `gofmt -l` → vide sur tout l'arbre Go.
2. `make test` (artisan_test recréée, migrations, Redis index 2) → **11/11 paquets `ok`, 0 FAIL** : `cmd/api` (nouveau), `cmd/e2edb`, `controllers`, `internal/platform/config`, `internal/platform/db`, `internal/platform/redisdb`, `internal/testutil`, `invoice`, `middleware`, `models`, `tests` (incl. `TestAPISurface` vert — 14 routes sans éditer la liste — et `TestRunCreateAdmin` sous `cmd/api`).
3. Test d'équivalence routeur (transitoire, supprimé au même commit) : **PASS** — tables de routes `/v1` de `SetupRouter` (14) ≡ `app.NewRouter` **à l'octet près** (méthode, chemin, nom de handler) ; extras hors-/v1 ≡ exactement les blocs main.go (GET `/`, `/assets/*filepath` + HEAD, `/public/*filepath` + HEAD, `/swagger/*any`) ; chaîne de middlewares = `gin.Logger`, `gin.CustomRecoveryWithWriter`, `middleware.CORS`, `middleware.RequestID`, `gzip` dans cet ordre.
4. Régénérations idempotentes (blobs vérifiés au hachage) : `generate_docs` (swag `-g cmd/api/main.go`) → `docs/{docs.go,swagger.json,swagger.yaml}` **identiques à HEAD** (`fc02d5c`/`f2df02b`/`f7d51ff`) ; `generate_mocks` → **stable sur 3 passes consécutives** (sha256 constants) ; `generate_sqlc` → sha256 constants sur re-passe (`internal/platform/db` non-tracké : la CI `git diff --exit-code internal/platform/db` s'appliquera après commit).
5. `make e2e` (webServer déplacé : `go run -mod=readonly ./cmd/api`) → **4 passed (5.1s)** — cycle complet (e2edb, migrations, seed CLI `./cmd/api -create-admin`, build pnpm, scénarios) vert.

## Dépassement planifié maîtrisé : l'outillage mockery

- Découverte : `make generate_mocks` (pin v2.53.8) est un **no-op silencieux depuis toujours** (aucune régénération réelle ; les mocks datent du 2026-11-23). Charnière : la clé de schéma `version: 2` se décode vers le champ **booléen `Version`** du flag CLI (décodage viper weak-typed : `2 != 0`) → `Run()` imprime le sémver et sort au check `--version` sans générer ; la collision `MOCKERY_VERSION` (export global du Makefile) produisait l'erreur `'version' cannot parse value as 'bool'` (string non booléen).
- Correctif retenu (2 fichiers, cohérent avec le pin) : suppression de `version: 2` de `.mockery.yaml` (les configs de mockery lui-même l'omettent) + renommage des variables Makefile (`MOCKERY_CLI_VERSION`/`MOCKERY_CLI_BIN`). Le correctif L0 (`959f50e`) a ainsi été **amendé conceptuellement** : sa partie `.mockery.yaml version: 2` est retirée dans ce lot (la partie Makefile reste). Vérifié : génération réelle DBTX/Querier + idempotence ×3.
- Binaire : la CI n'exécute pas `generate_mocks` (jamais détecté) ; le v2.53.8 installé via `go install` sous Go 1.27.4 fonctionne au delà de la disparition du flag (l'hypothèse x/tools/Go 1.27 a été testée avec le binaire officiel précompilé : mêmes symptômes avec clé `version: 2`, génération OK sans — la clé était la seule cause).

## Checklist de validation (protocole par lot, couche 3)

1. **Aucun `*pgdb.Queries` franchit une frontière de module** — le nouveau `internal/app/router.go` n'importe QUE `controllers/models/platform` ; `pgdb` reste confiné aux consommateurs racine (fenêtre A2) ✓
2. **Interfaces cross-module côté consommateur** — aucune abstraction nouvelle : 0 interface créée ✓
3. **Transactions UoW/WithTenant** — inchangées (`pgdb.WithTenant`/`Store.InTx` inchangés, aucune nouvelle Tx) ✓
4. **DI explicite** — `app.Deps{TemplateGlob, Env}` porteur unique des entrées du routeur ; `cmd/api` construit pools/Redis et appelle `app.NewRouter` ; aucun var de package nouveau (cible : `.kilo/rules/go-architecture-modulaire.md`) ✓
5. **Mutation surface globale tracée** — l'affectation globale `binding.Validator = new(validate.DefaultValidator)` reste recopiée dans `NewRouter` = **dette P1.7(a) canon** (l.140 du plan-phases, élimination post-Phase 8) ✓
6. **Réponses byte-identiques** — aucun handler modifié ; `TestAPISurface` vert sans éditer la liste ; équivalence chaîne middlewares + extras figurés ✓
7. **Rollback borné et documenté** — tout le lot est dans le working tree sur `refactor/modular-phase-0` ; `git restore . && git clean -fd` (ou checkout) ramène à `67c9e77`/`d7f6a22` ; après commit : revert séquentiel ✓
8. **Redis — propriétaire unique** — `models/auth.go` reste le seul consommateur de `redisdb.GetRedis()` en prod (vérifié : grep — `main.go` init à l'entrypoint, `cmd/e2edb` purge, tests) ✓

## Inventaire des tests du lot (exigence qualitative — aucun seuil chiffré)

- **Déplacés sans modification** : `config/secrets_test.go` → `internal/platform/config/`, `db/db_test.go` → `internal/platform/redisdb/redis_test.go` (package `redisdb` renommé), `internal/db/pool_test.go` → `internal/platform/db/`, `main_test.go` → `cmd/api/main_test.go`.
- **Modifiés** : imports uniquement (`controllers/auth_test.go`, `middleware/auth_test.go`, `tests/{article,login}_test.go`, `internal/testutil/{factory,postgres}_test.go`) ; `tests/article_test.go` : retrait de la duplication de routes (SetupRouter → `testRouter()` délégant à `app.NewRouter`) — la suite épingle désormais le câblage de production.
- **Ajouté** : 0 test nouveau (l'équivalence routeur preuve transitoire supprimée au commit, conformément au pas-à-pas).
- **Exécution verte consignée** : `make test` 11/11 `ok` (détails § Commandes) ; `make e2e` 4 passed.

## Critères de lot et verdict

| Critère L1 (plan-phases) | État |
|---|---|
| P1.1 config → `internal/platform/config` | ✓ (+ `secrets_test.go`) |
| P1.2 redisdb → `internal/platform/redisdb` (exportés inchangés) | ✓ (package `redisdb`, 7 sites d'import corrigés) |
| P1.3 sqlc → `internal/platform/db`, mocks → `internal/platform/mocks` (sqlc.yaml l.9, .mockery.yaml l.9, alias `pgdb` inchangé) | ✓ (11 sites + mocks régénérés) |
| P1.4 CORS/RequestID → `internal/platform/middleware` (package inchangé) | ✓ |
| P1.5 validate (validator.go + errors.go, exportés inchangés) | ✓ (7 sites `validate.Translate`, 3 cartes messages, `factory_test.go:86`) |
| P1.6 httpx reporté en Phase 2 | ✓ (rien créé — canon P1.6 du plan) |
| P1.7 `internal/app` (`Deps`, `NewRouter`, recopie verbatim) + `cmd/api/{main,main_test}.go` + test d'équivalence | ✓ (PASS, preuve ci-dessus) |
| P1.8 orchestration/docs (Makefile run/docs/seed, ci.yml build+sqlc-diff, playwright webServer, régénérations commées) | ✓ (docs blobs = HEAD, mock idempotence ×3, sqlc stable) |
| build+vet+`make test` verts | ✓ |
| équivalence routeur vide + `TestAPISurface` | ✓ |
| `make e2e` rejoué (webServer changé) | ✓ 4 passed |
| CI verte sur la branche | ⏳ au push (opérateur) |

**Verdict : PASS** — commits exécutés après APPROUVÉ (code `3dc7869`, docs ci-présent) ; push/merge opérateur selon CI.
