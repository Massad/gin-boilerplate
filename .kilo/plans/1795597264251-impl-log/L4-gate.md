# Gate L4 — Module `tenant` + UoW plateforme `pgdb.WithinTx`/`pgdb.InTx`

**Date** : 2026-11-26 · **Mode** : branche `refactor/modular-phase-4` (fourchue de `8a5acd8` = `modular_archi` + L0 + L1 + L2 + L3 + handover L4 — séquence L4↪L3) · commit code `dc45bbb` (git reconnaît les renames `models/tenant.go ⇒ internal/modules/tenant/service.go` 59 %, `models/tenant_test.go ⇒ service_test.go` 99 %) + commit docs ci-présent · Plan-cible : plan-phases l.43-47 (Phase 4) + lots l.50 · Handover : `.kilo/plans/1795627363688-handover-l4.md` (message de reprise) · **Verdict : PASS** (CI distante au push ; validateur technique de couche 3 = checklist ci-dessous rejouée en lecture, push/merge opérateur)

## Commandes jouées + résultats (toutes dans WSL)

1. `go build -mod=readonly ./...` → **BUILD_OK** ; `go vet -mod=readonly ./...` → **VET_OK** ; `go vet -mod=readonly -tags=all ./...` → **VET_TAGS_OK** (0 finding) ; `gofmt -l` → vide sur les paquets touchés. **Correctif outillage consigné** : `internal/app/router.go` portait un écart gofmt **préexistant** (ordre imports `validate`/`redisdb`) — corrigé mécaniquement (`gofmt -w`, 0 comportement).
2. Tests purs (race, sans tag, sans services) : `go test -mod=readonly -count=1 -race ./internal/modules/tenant` → **ok** (suite slug/réservés déplacée au byte) ; `./internal/testutil` → **ok** (factories requalifiées, var locale renommée pour lever le masquage du paquet).
3. `pool_test.go` (`-tags=all`, `.env` sourcé comme le Makefile) : `TestPostgresDSNPasswords` 4/4 + **3 tests `WithinTx` verts** (`TestWithinTxCommitsAfterFnSuccess`, `TestWithinTxRollsBackAfterFnError`, `TestWithinTxNeverSetsSessionWideTenantSetting`) ; re-verts dans `make test`.
4. `make test` complet (artisan_test recréée, migrations superuser, Redis index 2) → **13/13 paquets `ok`, 0 FAIL** (rc=0) : `cmd/api`, `cmd/e2edb`, `controllers` (TestMain seul), `internal/modules/article`, `internal/modules/auth`, **`internal/modules/tenant`** (nouveau), `internal/platform/{config,db,httpx,redisdb}`, `internal/testutil`, `invoice`, `tests` — **`TestAPISurface` PASS sans éditer la liste attendue** (14 routes) ; assertions de `tests/` au byte près (**`git diff -- tests/` = 0** : register, reserved-slug 406, duplicate-email 406, suspendu 403, admin tenants/status+audit, matrice unified login intégralement verts) ; `models/` : plus aucun fichier de test (déplacement attendu).
5. Miroir CI `-race -tags=all` : `./internal/platform/db` → **ok** ; `./tests` → **ok** (19,3 s).
6. `make generate_docs` ×2 → **blobs identiques à chaque passe** (sha256 `docs.go` `a73deb90…`, `swagger.json` `8d90ae3f…`, `swagger.yaml` `dbddcb67…`) : diff = definitions/$ref renommées `models.AllTenantResponse` → `tenant.AllTenantResponse` **et** `models.Tenant` → `tenant.Tenant` (le `$ref` imbriqué de la liste suit mécaniquement le déplacement du type — même classe de requalification d'annotations que L3), **0 clé de route modifiée** (`route_lines_changed=0`) ; l'import blank de `controllers/admin.go` est résolu par swag (`Generating tenant.AllTenantResponse`/`tenant.Tenant` au log) ; sqlc/mocks : **0 code généré modifié** (`git diff --stat internal/platform/db` = `pool.go` +28 / `pool_test.go` +140 uniquement) ; `make e2e` non joué (Phases 1 et 7 seulement, plan-phases §3.9).

## Mecanismes appliqués (déplacements documentés)

- **`pgdb.InTx`/`pgdb.WithinTx` — spec zéro-décision recopiée à l'identique** (plan-phases l.44) : `type InTx interface { Queries() *Queries }` ; adaptateur interne `withinTx` ne portant que les queries liées au tx ; `func (s Store) WithinTx(ctx, fn func(tx InTx) error) error` — **wrapper strict de `Store.InTx`** (Begin/Rollback/Commit identiques, **aucun `set_config`** : les opérations in-Tx du register ciblent le pool admin BYPASSRLS sans contexte tenant).
- **Binding des repos** : `tenantRepo.CreateInTx` exécute `tx.Queries().CreateTenant(ctx, …)` (via `createTenantInTx`, boucle 99 tentatives inchangée : 23505 → suffixes `-2…-99`, `ErrInvalidSlug`/`ErrSlugUnavailable` inchangées), `userRepo.CreateUserOwner` — en Phase 4, `models/user.go` Register (code racine exécutant, fenêtre A2) exécute `tx.Queries().CreateUserOwner(ctx, …)` **sur le même `InTx`** transmis par `WithinTx` (descend en `userRepo.CreateOwnerTx` en L5) ; `models/admin.go` SetTenantStatus exécute `tx.Queries().InsertAdminAudit(ctx, …)` **sur le même `InTx`** que `tenantRepo.UpdateStatusTx` (ordres UpdateTenantStatus → InsertAdminAudit inchangés, même Tx ; migration complète admin en Phase 6).
- **Port side consommateur** : `TenantProvisioner` déclaré dans `models/user.go` (`CreateInTx(ctx, pgdb.InTx, name) (tenant.Tenant, error)` — retourne le type provider comme le prescrit le design, import racine→module toléré en A2), `TenantReader` déclaré dans `models/admin.go` (`List` + `UpdateStatusTx`) — ports de la liste certifiée, implémentés par le module tenant, injectés au composition root.
- **Suppressions sans retour** : var singleton `tenantModel`, `TenantModel.Create`/`SetStatus` — **0 appelant en production** (vérifié grep : le backoffice passe par `AdminModel`, le register par le port) — pas de code mort re-créé ; `TenantModel.All` migré en `repository.List`.
- **`reservedSlugs`** — commentaire complété pointant vers la liste figée de routes / le mécanisme `TestAPISurface` (P0.3), comme exigé au P4.4.
- **Écart documenté** : `pool_test.go` écrit désormais des lignes `tenant` ciblées (slug préfixé `pooltest-within-tx-`, nettoyées par `t.Cleanup`) sur la base configurée (dev `artisan`) — dérive de la propriété « pool_test lecture seule » ; AGENTS.md mis à jour dans le commit docs. Le sonde de « jamais SET session-wide » utilise un pool `MaxConns=1` + `current_setting('app.tenant_id', true)` avec **contrôle de sensibilité** (un `set_config('…', …, false)` session-wide réellement écrit est d'abord vu puis effacé par la sonde).
- **Environnement** : `.env` absent du worktree L4 (le seed ne l'avait pas fourchu) — copié depuis le worktree L3 ; `make generate_docs` a résolu le renommage sans toucher aux blobs généraux.

## Checklist de validation (protocole par lot, couche 3)

1. **Aucun `*pgdb.Queries` ne franchit une frontière de module** — module `tenant` : les méthodes portent `pgdb.InTx`/`pgdb.Store` ; `tx.Queries()` résolu **à l'intérieur** du module (`repository.go` CreateInTx/UpdateStatusTx, `service.go` createTenantInTx) et dans le code racine exécutant A2 (`models/user.go`, `models/admin.go`) ; module `auth` : 0 référence pgdb en production (grep, seul `refresh_ownership_test.go` touche les pools pour épinglage — état L3 inchangé) ✓
2. **Interfaces cross-module côté consommateur** — uniquement `TenantProvisioner` (dans `models/user.go`, L5 le déplacera) et `TenantReader` (dans `models/admin.go`, L6 le déplacera) ; aucune abstraction nouvelle hors liste certifiée ; `dto.go`/`handler.go` non créés (aucune route, aucun form — conforme P4) ✓
3. **Transactions UoW/WithTenant** — nouveau `WithinTx` (aucun `set_config`, preuve par sonde) ; `WithTenant`/`InTx` inchangés ; zéro `SET` session-wide ; registre cible = pool admin BYPASSRLS sans contexte tenant ✓
4. **DI explicite** — `tenant.NewRepository(pgdb.AdminStore())`, `models.NewUserModel(svc, tenants)`, `models.NewAdminModel(tenants)`, `controllers.NewAdminController(svc, models.AdminModel)` : câblés dans `internal/app` ; **la var préexistante `adminModel` est supprimée** (remplacée par l'injection) ; la var préexistante `userModel` (Refresh/OneWithTenant uniquement) reste — dette A2 documentée L3, disparaît en L7 ; aucune var de package nouvelle ✓
5. **Mutation surface globale tracée** — aucune : `binding.Validator` inchangé (dette P1.7(a) intacte), mode gin inchangé, `Deps` inchangé ✓
6. **Réponses byte-identiques au contrat actuel** — suite `tests/` 0 diff, `TestAPISurface` vert sans édition, register/login/suspend/unified/admin flows PASS ; docs = renames de definitions uniquement, 0 route ✓
7. **Rollback borné et documenté** — lot en working tree avant commits : `git checkout -- . && git clean -fd internal/modules/tenant` (+ laisser `.env`) ramène à `8a5acd8` ; post-commits : revert séquentiel docs puis code ✓
8. **Redis — propriétaire unique** — le lot ne touche aucune clé Redis ; index 1/2/3 inchangés ; `internal/platform/*` + `cmd/e2edb` : **0** import `internal/modules/*` (grep) ; seul `internal/modules/auth` accède aux clés de session, inchangé ✓

## Inventaire des tests du lot (exigence qualitative — aucun seuil chiffré)

- **Déplacés au byte** : `TestSlugify`, `TestValidSlug`, `TestValidSlugReserved` (+ helper `filler`) : `models/tenant_test.go` → `internal/modules/tenant/service_test.go` (paquet `tenant`, blancs identiques, race sans tag).
- **Ajoutés** (`pool_test.go`, tag `all`) : `TestWithinTxCommitsAfterFnSuccess` (insert lisible après commit, via `GetTenantByID`), `TestWithinTxRollsBackAfterFnError` (erreur propagée + `UNIQUE(slug)` re-créable hors tx = insert annulé), `TestWithinTxNeverSetsSessionWideTenantSetting` (pool 1 conn, sonde `current_setting`, contrôle de sensibilité `set_config(...,false)` visible puis effacé, assert `""` après `WithinTx` **et** après contrôle `WithTenant`) ; helpers `withinTxTestStore`/`withinTxSlug` (skip sans env/admin / 28P01-28000), `errWithinTxProbe`.
- **Requalifiés mécaniquement** : `internal/testutil/factory.go` (`NewTenantFixture` → `tenant.Tenant`/`tenant.Slugify`/`tenant.ValidSlug`) + `factory_test.go` (var locale `tenant` → `fixture`, import `models` supprimé devenu inutile) ; `controllers/user.go` (var locale `tenant` → `createdTenant` — masquage du paquet).
- **Exécution verte consignée** : `make test` 13/13 `ok` ; tests purs `ok` en `-race` ; pool_test 4/4 verts ×2 (focused + make test) ; miroir CI `-race -tags=all` verts (db, tests).

## Dettes consignées (fenêtre transitoire A2 — disparaissent aux Phases 5-7)

- `models/user.go` Register et `models/admin.go` SetTenantStatus obtiennent `tx.Queries()` en code racine exécutant (orchestration in-Tx en place, conforme à « sur le même `InTx` ») — descend en `userRepo.CreateOwnerTx` (L5) et dans le service admin (L6).
- Ports `TenantProvisioner`/`TenantReader` déclarés à la racine — migreront avec leurs modules consommateurs (L5/L6).
- Var `userModel` (Refresh/OneWithTenant) + `os.Getenv("REFRESH_SECRET")` du parse JWT — dettes L3 intactes, sorties en L7.
- `docs/` : 2 definitions + `$ref` ré-addressées `models.AllTenantResponse|models.Tenant` → `tenant.*` (équiv. déplacements d'annotations, 0 route).

## Critères de lot et verdict

| Critère L4 (lots l.50 / plan-phases l.43-47) | État |
|---|---|
| `pgdb.WithinTx` + interface étroite `pgdb.InTx` dans `pool.go` manuscrit (spec zéro-décision) | ✓ (wrapper strict `Store.InTx`, aucun `set_config` ; +28 l.) |
| Couverture `pool_test.go` : commit / rollback / jamais `SET` session-wide | ✓ (3 tests + sonde de sensibilité, verts ×2) |
| `tenant/{service,repository,model}.go` (boucle 99 tentatives identique, pool admin) | ✓ (renames reconnus 59 % ; erreurs discriminées inchangées) |
| Tests purs slug/réservés déplacés | ✓ (3 tests au byte, race verts) |
| `make test` vert ; suite `tests/` au byte ; `TestAPISurface` vert sans édition | ✓ 13/13 `ok`, `git diff tests/` = 0 |
| Wiring review — aucun `*pgdb.Queries` en port ; interfaces côté consommateur ; pas de nouveau singleton | ✓ (checklist 1-4 ; `adminModel` supprimée) |
| Test contrat `reservedSlugs` (P0.3) toujours vert + commentaire pointeur | ✓ (`TestRegisterReservedSlug` 406 PASS ; commentaire P0.3 ajouté) |
| `make generate_docs` diff = déplacements d'annotations uniquement, idempotent ×2 | ✓ (renames definitions + `$ref`, 0 route) |
| CI verte sur la branche | ⏳ au push (opérateur) |

**Verdict : PASS** — commits exécutés après APPROUVÉ (code `dc45bbb`, docs ci-présent) ; push/merge opérateur selon CI. Le lot suivant (**L5** : module `user` — login/register, port `TokenService`, implémente `UserAuthenticator`/`UserReader`, register porté par `WithinTx`/`TenantProvisioner`) exige ce gate PASS.
