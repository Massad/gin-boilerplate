# Gate L0 — Filet de sécurité (transverse)

**Date** : 2026-11-26 · **Branche** : `refactor/modular-phase-0` (fourchue de `modular_archi` @ `07d81c4`, baseline officielle) · **Commits du lot** : `30ae675` (P0.1, auteur utilisateur), `959f50e` (correctif outillage mockery), `67c9e77` (P0.3 TestAPISurface) · **Verdict** : **PASS** (CI branche : poussée requise, voir ci-dessous)

## Commandes jouées + résultats (Windows UNC → WSL Debian : build/vet/tests/make exécutés dans WSL)

1. `go version` : go1.27.4 linux/amd64 (WSL) — le UNC `\\wsl.localhost\...` refuse les locks du toolchain Go côté Windows (« RLock go.mod : Fonction incorrecte ») → batterie exécutée via `wsl.exe` dans le checkout natif.
2. `go build -mod=readonly ./...` → **BUILD_OK** ; `go vet -mod=readonly ./...` → **VET_OK** (+ `go vet -tags=all ./tests/` pour le paquet taggé).
3. `make test` (services up, `artisan_test` recréée par `cmd/e2edb`, migrations goose, `FLUSHDB` index 2) → **vert, 2 runs consécutifs** puis **3e run final** incluant `TestAPISurface` : 11/11 paquets `ok` (`github.com/Massad/gin-boilerplate` racine, `cmd/e2edb`, `config`, `controllers`, `db`, `internal/db`, `internal/testutil`, `invoice`, `middleware`, `models`, `tests`), 0 FAIL, 0 panic.
4. Régénérations idempotentes :
   - `make generate_sqlc` (v1.30.0) → `git status` : `internal/db` **aucune modification** ✓
   - `make generate_mocks` (v2.53.8) → `internal/mocks` **aucune modification** ✓ (requis : voir « Anomalies traitées »)
   - `make generate_docs` (swag v1.16.6) → blobs `docs/{docs.go,swagger.json,swagger.yaml}` **identiques à HEAD** : `fc02d5c`/`f2df02b`/`f7d51ff` ✓
5. `go test -mod=readonly -count=1 -tags=all ./tests/ -run TestAPISurface` → **PASS** (`--- PASS: TestAPISurface` dans le run complet `make test` également).

## Anomalies traitées (contexte du lot, hors périmètre planifié)

- **P0.1 préexécuté par l'utilisateur** (commit `30ae675` sur la branche du lot, conforme) : re-add périmé de `/v1/invoice*` retrouvé en working tree (sauvegarde d'éditeur obsolète) → `git restore` (blob vérifié égal à HEAD `6301204`), puis fantôme `M` lié EOL (autocrlf=true, w/crlf vs i/lf) résolu par normalisation LF + `git add` du blob identique.
- **`make generate_mocks` cassé** (bloquant P0.2) — deux causes corrigées dans `959f50e` :
  1. `.mockery.yaml` sans `version: 2` — le mock pinned v2.53.8 refuse la config (`'version' cannot parse value as 'bool'`) ;
  2. collision viper/env : le Makefile exporte toutes ses variables (`export` global, source unique `.env`), dont `MOCKERY_VERSION` → mockery lit `MOCKERY_VERSION=v2.53.8` comme override de la clé config `version`. Renommées `MOCKERY_CLI_VERSION`/`MOCKERY_CLI_BIN`. (CI n'exécute pas `generate_mocks` — jamais détectée.)
- `TestAPISurface` : comparaison paths×méthodes ↔ liste figée ; note : 12 chemins portent les 14 paires (méthodes multiples sur `/v1/article/{id}`) — égalité bidirectionnelle stricte, sans compte redondant.

## Checklist de validation (protocole par lot, couche 3)

1. **Aucun `*pgdb.Queries` franchit une frontière de module** — N/A (aucun déplacement de code ; diff du lot : `.github/workflows/ci.yml`, `.mockery.yaml`, `Makefile`, `tests/api_surface_test.go` uniquement) ✓
2. **Interfaces cross-module côté consommateur, liste certifiée** — aucune abstraction nouvelle (0) ✓
3. **Transactions UoW/WithTenant** — aucune nouvelle transaction (aucun changement) ✓
4. **DI explicite, aucun singleton/var de package** — aucun code Go de production modifié ✓
5. **Mutation surface globale tracée** — aucune (`binding.Validator` intact) ✓
6. **Réponses byte-identiques** — aucun handler modifié (sortie `make test` : assertions `tests/` inchangées, matrice login 406/403/200 conforme) ✓
7. **Rollback borné et documenté** — branche `refactor/modular-phase-0`, dernier commit vert `67c9e77` ; revert = retour à `modular_archi` (`07d81c4`) ou revert séquentiel des 3 commits ✓
8. **Redis — propriétaire unique** — inchangé et préservé : seul `models/auth.go` accède aux clés de session (vérifié G-1, aucun diff sur `models/`) ✓

## Inventaire des tests du lot (exigence qualitative — aucun seuil chiffré)

- **Ajouté** : `TestAPISurface` (`tests/api_surface_test.go`, tag `all` — egalité bidirectionnelle exacte des 14 routes `/v1` entre `docs/swagger.json` committé et liste figée ; lecture seule du JSON, sans requête applicative, 0.00s)
- **Déplacé** : 0 · **Modifié** : 0 (assertions `tests/` inchangées, intégralité des suites existantes verdes)
- **Exécution verte consignée** : `make test` final 11/11 paquets `ok` (détails dans la trace du run ; journal consigné en session, non commité)

## Critères de lot et verdict

| Critère L0 | État |
|---|---|
| smoke `ci.yml` corrigé | ✓ `30ae675` (`/v1/invoice*` + fragments retirés, routes 14 vérifiées `main.go:153-181`) |
| baseline verte consignée | ✓ ce rapport + `memory-bank/progress.md` |
| `tests/api_surface_test.go` (contrat bidirectionnel 14 routes) | ✓ `67c9e77`, vert sans éditer la liste attendue |
| discipline branches posée | ✓ `refactor/modular-phase-0` fourchue de `modular_archi`, commits par étape |
| build+vet+`make test` verts | ✓ (voir § Commandes) |
| `TestAPISurface` vert sans éditer la liste attendue | ✓ |
| régénérations idempotentes | ✓ sqlc/mocks/docs → diffs vides (après correctif outillage `959f50e`) |
| CI verte sur la branche | ⏳ **poussée requise** (opérateur) — l'exécution locale est verte ; la CI distant validera au push |

**Verdict : PASS** (réserve mineure : validation CI au push, action opérateur hors session).
