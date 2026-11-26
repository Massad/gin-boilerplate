# Gate L2 — Module pilote « article » + `internal/platform/httpx`

**Date** : 2026-11-26 · **Mode** : branche `refactor/modular-phase-2` (fourchue du HEAD courant `660a5e0` = `modular_archi` + L0 + L1 — séquence L2↪L1) · commit code `55f256c` (git reconnaît les renames `forms/article.go ⇒ internal/modules/article/dto.go` 76 % et `controllers/article.go ⇒ handler.go` 51 %) + commit docs ci-présent · Plan-cible : plan-phases l.18-23 (Phase 2) · Handover : `.kilo/plans/1795627363688-handover-l2.md` · **Verdict : PASS** (CI distante au push ; validateur technique de couche 3 = checklist ci-dessous rejouée en lecture, push/merge opérateur)

## Commandes jouées + résultats (toutes dans WSL)

1. `go build -mod=readonly ./...` → **BUILD_OK** ; `go vet -mod=readonly ./...` → **VET_OK** ; `go vet -mod=readonly -tags=all` sur `tests/`, `cmd/api/`, `middleware/`, `controllers/`, `internal/testutil/`, `internal/modules/article/`, `internal/platform/httpx/` → **VET_TAGS_OK** ; `gofmt -l` → vide sur l'arbre Go.
2. Tests purs (race, sans services) : `go test -mod=readonly -count=1 -race ./internal/modules/article ./internal/platform/httpx ./internal/testutil` → **3/3 `ok`**.
3. `make test` complet (artisan_test recréée, migrations superuser, Redis index 2) → **13/13 paquets `ok`, 0 FAIL** : `cmd/api`, `cmd/e2edb`, `controllers`, `internal/modules/article` (nouveau), `internal/platform/config`, `internal/platform/db`, `internal/platform/httpx` (nouveau), `internal/platform/redisdb`, `internal/testutil`, `invoice`, `middleware`, `models`, `tests` — **`TestAPISurface` PASS sans éditer la liste attendue** (14 routes) ; assertions de `tests/` au byte près (diff de `tests/article_test.go` = 1 import + 4 requalifications `forms.CreateArticleForm` → `article.CreateArticleForm`, aucune logique d'assertion modifiée).
4. `make generate_docs` ×2 → **blobs identiques à chaque passe** (sha256 doc `70395567…`, `swagger.json` `d4997d66…`, `swagger.yaml` `26389b17…`) ; diff vs HEAD = **déplacements d'annotations uniquement** : refs `$ref` re-écrites `forms.*`/`models.*` → `article.*` (30 refs), clés de définitions déplacées (contenus identiques), retrait du bloc définition `forms.ArticleResponse` (C16 ci-dessous) ; **aucune clé de chemin/route modifiée** (vérifié par filtre sur le diff : 0 ligne `"/…"` changée).
5. `make e2e` non joué (non requis pour L2 — Phases 1 et 7 seulement, plan-phases §3.9). sqlc/mocks : aucune entrée modifiée (SQL/`sqlc.yaml`/`.mockery.yaml` intacts) — garde CI `git diff --exit-code internal/platform/db` inchangée et satisfaite.

## Arbitrage C16 en cours de lot (hors certification, appliqué et documenté)

- **Collision de nom** : `forms/article.go` portait un second type `ArticleResponse` (`id int64` + `message`), **zéro appelant Go** — uniquement des refs Swagger `@Success/@Failure` des handlers article. Le plan certifié envoie `forms/article.go` → `dto.go` (types module-locaux, C15) : les deux `ArticleResponse` sont en collision dans le paquet `article` (illégal), et aucune lecture du plan ne l'évite.
- **Résolution (validée utilisateur en livrable)** : type retiré (canon « pas de code mort sans appelant », P1.6) ; les 9 refs `@Failure/@Success` ré-addressées vers `article.ArticleResponse` (règle unique : le catalogue `/v1/article*` pointe intégralement vers `article.*`). Effet docs limité au retrait de la définition + réécriture des refs — aucun contrat HTTP touché (le test de surface lit les chemins, pas les définitions).
- **Détails connexes consignés** : `getUserID`/`getTenantID` **recopiés** localement au handler (d'après `controllers/user.go:19-29`), `getRole` **non recopié** (sans appelant) ; les deux accesseurs racine restent dans `controllers/user.go` (désormais sans appelant — suppression avec le répertoire en Phase 7, hors périmètre L2) ; `internal/testutil/factory_test.go` **inchangé** (types inférés — contrairement à l'attente du handover) ; `httpx.MapPgxError` n'a pas d'appelant production en L2 (la sémantique d'erreur article est conservée : `ErrNotFound`→404, tout autre → 406/404 — tout re-mapping changerait le contrat byte-identique) ; exercé par ses tests purs, adoption à la normalisation des erreurs (post-Phase 8, design §6).

## Checklist de validation (protocole par lot, couche 3)

1. **Aucun `*pgdb.Queries` franchit une frontière de module** — le port `ArticleRepository` ne porte que des types du module (`CreateArticleForm`, `ArticleResponse`, `Result`) ; les `*pgdb.Queries` restent des closures `WithTenant` intra-`repository.go` (jamais paramètre/retour) ✓
2. **Interfaces cross-module côté consommateur** — uniquement `ArticleRepository` (défini dans `service.go`, le consommateur, certifié plan-phases l.20) ; repository concret non-exporté + `NewRepository` ; aucune autre abstraction nouvelle ✓
3. **Transactions UoW/WithTenant** — uniquement `pgdb.WithTenant` (RLS via `set_config(..., true)` transactionnel, inchangé) ; aucun `SET` session-wide ; aucun UoW attendu en L2 ✓
4. **DI explicite** — `article.NewRepository(store pgdb.Store)`, `article.NewService(port)`, `article.NewHandler(*Service)` câblés dans `internal/app` ; le pool est résolu une fois au composition root (`pgdb.TenantStore()`, pattern canon Phase 3 : lecture du singleton confinée à `internal/app`) ; aucun nouveau var de package/singleton (sentinelle `ErrNotFound` + map `ArticleMessages` = constantes de domaine) ✓
5. **Mutation surface globale tracée** — aucune : `binding.Validator` inchangé (dette P1.7(a) intacte), aucun nouveau global ✓
6. **Réponses byte-identiques** — casse `"Message"`/`"message"` recopiée à l'identique (3 chemins `"Message"` bruts ; chemins `{"message":…}` via `httpx.AbortMessage`/`JSONMessage`, même forme `gin.H`) ; `TestAPISurface` vert sans édition ; suite `tests/` assertions au byte ✓
7. **Rollback borné et documenté** — lot en working tree avant commits ; `git restore . && git clean -fd internal/modules internal/platform/httpx` ramène à `660a5e0` ; post-commit : revert séquentiel des 2 commits ✓
8. **Redis — propriétaire unique** — L2 ne touche aucun fichier Redis ; `models/auth.go` reste le seul consommateur de `redisdb.GetRedis()` en prod ; aucun import `internal/platform/redisdb` sous `internal/modules/` ou `httpx` (grep) ✓

## Inventaire des tests du lot (exigence qualitative — aucun seuil chiffré)

- **Déplacés sans modification de logique** : `controllers/article.go` → `internal/modules/article/handler.go` (annotations déplacées, helpers recopiés) ; `models/article.go` → `service.go`/`repository.go` (corps verbatim, récepteurs `pgdb.TenantStore()` → `r.store`) ; `forms/article.go` → `dto.go` (+ C16).
- **Modifiés (mécaniques)** : `tests/article_test.go` (import + 4 requalifications de type uniquement), `internal/testutil/factory.go` (type de retour + littéral).
- **Ajoutés (purs, sans services)** : `internal/modules/article/service_test.go` — fake du port (`ErrNotFound`≡0-ligne modélisé) : `TestCreateArticle` (créé 200 + propagation clés TokenAuth/contexte + ids, erreur repo → 406, binding → 406 traduit), `TestListArticles` (enveloppe `results` ronde, erreur → 406 `"Message"`), `TestGetArticle` (200 + port contractuel, erreur → 404 `"Message"`, paramètre invalide → 404), `TestUpdateArticle` (200, 0-ligne → 404, erreur → 406, binding, paramètre invalide), `TestDeleteArticle` (200, 0-ligne → 404, erreur → 406, paramètre invalide) — 5 tests / 18 sous-tests ; `internal/platform/httpx/httpx_test.go` — `TestMapPgxError` (8 cas table-driven), `TestParseUUIDParam` (3 cas sur route réelle), `TestAbortMessage`, `TestJSONMessageDoesNotAbort` — 4 tests.
- **Repository** : couvert par l'intégration (convention certifiée — `repository_test.go` non requis).
- **Exécution verte consignée** : `make test` 13/13 `ok` (détails § Commandes) ; tests purs ×3 `ok` en `-race`.

## Critères de lot et verdict

| Critère L2 (lots l.48 / plan-phases l.19-23) | État |
|---|---|
| `internal/modules/article/{handler,service,repository,dto,model}.go` | ✓ (+ port `ArticleRepository` côté service) |
| `service_test.go` pur (fake repository : `ErrNotFound`→404, update/delete 0-ligne) | ✓ (5 tests / 18 sous-tests, fake sans DB/Redis) |
| `httpx` + `httpx_test.go` purs (respond/pgxerr/params) | ✓ (canon P1.6 — 4 tests, table-driven) |
| Suite `tests/` inchangée (assertions) | ✓ (4 requalifications de type, 0 assertion modifiée) |
| `make test` vert | ✓ 13/13 paquets `ok` |
| Wiring review — `pgdb.TenantStore()` lu au composition root, injecté au module | ✓ (pattern canon Phase 3) |
| `make generate_docs` diff = déplacements d'annotations uniquement | ✓ (C16 documenté, routage inchangé, idempotence ×2) |
| CI verte sur la branche | ⏳ au push (opérateur) |

**Verdict : PASS** — commits exécutés après APPROUVÉ (code `55f256c`, docs ci-présent) ; push/merge opérateur selon CI.
