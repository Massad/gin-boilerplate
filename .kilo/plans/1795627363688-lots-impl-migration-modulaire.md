# Stage 9 — Exécution de la migration modulaire par lots validés

> **Ce plan ne re-spécifie pas les phases.** La source de vérité opérationnelle est le [plan-phases certifié](1795597264251-architecture-modulaire-par-domaine-plan-phases.md) (Phases 0-8, étapes P0.x/P1.x, critères de sortie invariables). Le présent document est le **contrat d'exécution** : structure des lots (L0-L8 = Phases 0-8), protocole de validation (gates + agent de validation technique), garde-fous et suivi (statut « Fait » + liens vers artefacts). Toute évolution du contenu d'une phase pendant l'exécution passe par **egm-lite** et édite le plan-phases — jamais ce document.

## Artefacts d'entrée

- Design certifié : [1795597264251-architecture-modulaire-par-domaine-design.md](1795597264251-architecture-modulaire-par-domaine-design.md) — Stage 7 : 25/25 résolus ; Stage 8 : 5/5 + re-audit 0 finding.
- Plan d'exécution par phases : [1795597264251-architecture-modulaire-par-domaine-plan-phases.md](1795597264251-architecture-modulaire-par-domaine-plan-phases.md).
- Directives architecturales du 2026-11-25 : [user_directives_2026-11-25.md](1795597264251-design-review/user_directives_2026-11-25.md).
- Suivi EGM : [1795601758566-architecture-modulaire-split-review.md](1795601758566-architecture-modulaire-split-review.md) (tâches 5-6 → gate G-1 ci-dessous).
- **État du dépôt vérifié le 2026-11-26** : branche `modular_archi` — **baseline officielle de la migration** (point de fourche et cible de merge des branches `refactor/modular-phase-N`, cf. « Structure des lots ») ; aucun commit de migration, Phase 0 non démarrée (smoke `/v1/invoice` toujours présent `ci.yml:139` ; pas de `tests/api_surface_test.go` ; pas de `internal/platform/`, `internal/app/`, `internal/modules/`, `cmd/api/`). Arbre de travail propre : seul `.kilo/plans/1795597264251-phases-review/` est non suivi (rapport `style_findings.md` du G-1 en cours). Point de départ : **G-1 puis L0**.

## Décisions actées pour ce Stage (gate utilisateur, 2026-11-25)

1. **G-1 avant tout code** : la review de certification du plan-phases (split-review tâche 5) est un **gate obligatoire** avant L0 — le design est certifié, le plan-phases n'a pas encore subi son propre cycle.
2. **Réconciliations des contraintes Stage 9 avec le design certifié** (le design post-Stage-8 gouverne, conformément à « respecter strictement les décisions validées en Stage 7/8 ») :
   - « Hash Redis (`consumed_at`, `delivered`, `token_hash`) » → **supersédé par la résolution R2-1 (Stage 8)** : le record reste la **chaîne STRING existante** adressée par l'UUID (jti) du refresh JWT — le token n'est jamais stocké, en clair ni haché (l'intention « hash » de la directive est satisfaite par l'adressage jti) ; état encodé dans la valeur (`consumed:<delivered|recovered>:userID:tenantID:role`) ; grâce **fixe 5 min** (min(grâce, TTL restant)) ; TTL = éviction primaire ; sweep in-process = observabilité + purge défensive.
   - « Scripts de migration (DB/Redis) » → **aucune migration de schéma** (décision 7 certifiée, « aucune migration goose ») : les livrables réels du lot L3 sont le script Lua soft-consume (atomicité + ownership inchangés), l'encodage d'état dans la valeur, le rollout/rollback borné (records actifs au format inchangé, binaire pré-Phase-3 compatible) et les tests dédiés. Aucun nouveau fichier goose n'est attendu.
   - `docs/test.md` (Phase 8 : « suppression non commitée ») : vraisemblablement déjà résolu par le commit `581d770` (« docs(test): remove test documentation ») — **à faire confirmer par G-1** (drift) et corriger dans le plan-phases.
   - Ancrages plan-phases vérifiés ce jour (échantillon, tous conformes) : `main.go:16`, `models/user.go:122`, `middleware/auth.go:10`, `tests/article_test.go:34-80`, `ci.yml:139`. L'audit complet d'ancrage relève du persona codebase-details de G-1.
   - **`pgdb.WithinTx` / `pgdb.InTx` n'existent pas** dans `internal/db/pool.go` (seuls `Store.InTx` et `Store.WithTenant`) : leur création est un livrable plateforme du **lot L4** (premier consommateur : `TenantProvisioner.CreateInTx`), conformément au principe « pas de code mort sans appelant » (déjà appliqué à httpx, P1.6). G-1 doit exiger une spécification **zéro-décision** de l'interface étroite (comment les repos concrets lient leurs requêtes sqlc au tx sans exposer `*pgdb.Queries`).

## Gate G-1 — Certification du plan-phases (AVANT L0)

Exécuter exactement les tâches 5-6 du [split-review](1795601758566-architecture-modulaire-split-review.md), dans une **session Kilo dédiée** (jamais dans la foulée d'un autre cycle) :

1. egm-review, TargetFile = `…-architecture-modulaire-par-domaine-plan-phases.md`, Step = `implementation` (l'en-tête `## Detailed Implementation Plan` fait de tout le document le DetailedRange) ; GoldfishContext = le document design (obligatoire : cartographie §1-§4 et invariants requis par codebase_details).
2. Stage 9 en dispatch parallèle, **un subagent isolé par persona** (contexte frais, lecture seule) : `egm-implementation-readiness` + `egm-codebase-details` + `egm-style` (personas et standards chargés depuis le repo kilo : `\\wsl.localhost\Debian\home\sylvaindeb\projects\kilo\`) ; rapports en français dans `1795597264251-phases-review/`, **nommage standardisé du contrat** : un fichier par persona `{persona}_findings.md` (`implementation_readiness_findings.md`, `codebase_details_findings.md`, `style_findings.md` — ce dernier déjà livré) ; synthèse consolidée du stage dans `consolidated_report.md` ; la passe egm-resolve écrit son `goldfish_review_report.md` dans le même répertoire. Ce nommage, imposé par le présent contrat, prévaut sur les noms par défaut du skill egm-review (`implementation_readiness_report.md`, `consolidated_review_report.md`) — le synthesizer écrivant « to the report file the caller names » (`agents/synthesizer.md`, repo kilo).
3. Points à faire auditer explicitement (alimentés par les vérifications de cette session) :
   - spécification zéro-décision de `pgdb.InTx`/`pgdb.WithinTx` (cf. ci-dessus) ;
   - sweep d'ancrage complet : chaque fichier/numéro de ligne des étapes P0.x-P1.x et des phases 2-8 contre le code réel ;
   - fraîcheur des items Phase 8 (`docs/test.md` vs commit `581d770` ; artefacts racine `gin-boilerplate`/`e2edb.exe`/`tmp_line.txt` toujours présents ?) ;
   - cohérence de la puce P3 (soft-consume) avec le design décision 7 post-pass-2 ;
   - opérationnalisation du critère « aucune phase ne baisse la couverture » : acter l'exigence qualitative — chaque rapport `L<N>-gate.md` consigne l'inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte) ; aucun seuil chiffré (le dépôt n'instrumente pas la couverture — en inventer un serait une exigence nouvelle) ;
   - spécification zéro-décision de l'acquisition du client Redis par le repository sessions du module `auth` (Phase 3) : constructeur injecté (recommandé, aligné sur `.kilo/rules/go-architecture-modulaire.md` qui prohibe les singletons `GetRedis()`) vs tolérance du singleton pendant la fenêtre + dette tracée — texte prêt : audit [1795684415717-audit-lots-impl-migration.md](1795684415717-audit-lots-impl-migration.md), correctif C8 ;
   - propriétaire unique Redis : acter que seul le module `auth` accède en production aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — vérifié à l'état initial : seul `models/auth.go` consomme `GetRedis()` hors tests ; règle reprise en gate (checklist item 8).
4. Résolution des findings : egm-resolve (gates interactifs) — les éditions portent sur le plan-phases.
5. **Critère de sortie G-1** : 0 Major bloquant l'exécution (nits tolérés), éditions appliquées et revérifiées ; en-tête du plan-phases marqué « certifié » ; split-review tâches 5-6 marquées `[Fait]`.

## Structure des lots (L0-L8 = Phases 0-8)

Règles communes : un lot = une branche `refactor/modular-phase-N` **fourchue depuis `modular_archi` (baseline officielle) et re-mergée dans `modular_archi` uniquement si CI verte** ; **séquence stricte — aucun code n'est généré avant que le gate du lot précédent soit PASS** ; le pas-à-pas de chaque lot est la section « Phase N » du plan-phases (+ « Étapes détaillées » pour L0/L1) ; discipline de commits du plan-phases (un commit par étape, jamais `git mv` et édition de contenu dans le même commit, `go build` + `go vet` après chaque étape, `make test` en fin de lot, jamais de suite en parallèle dans le même checkout).

| Lot | Périmètre (domaine/module) | Livrables mesurables | Gate technique (complète les critères de sortie du plan-phases) |
|---|---|---|---|
| **L0** | Filet de sécurité (transverse, aucun déplacement de code) | smoke `ci.yml` corrigé ; baseline verte consignée ; `tests/api_surface_test.go` (contrat bidirectionnel des 14 routes) ; discipline branches posée | Pré-flight : committer l'état en cours des fichiers docs/plans (modifications et non-suivis sous `.kilo/plans/` — jamais `.env`) ; `make services_up` puis `make migrate_up` ; poppler dans le PATH ; `pnpm -C app exec playwright install chromium` (une fois). Gate : build+vet+`make test` verts ; `TestAPISurface` vert **sans éditer la liste attendue** ; régénérations idempotentes (sqlc/mocks/docs → diffs vides) |
| **L1** | Socle `internal/platform/` + `cmd/api` + `internal/app` (transverse, comportement inchangé au byte près) | 9 `git mv` + 1 création ; Makefile/CI/Playwright/AGENTS à jour ; `app.NewRouter` épingle le vrai câblage (`SetupRouter` disparaît) ; test d'équivalence `gin.Routes()` | build+vet+`make test`+`make e2e` verts ; équivalence routeur vide ; `TestAPISurface` vert ; régénérations idempotentes ; CI verte sur la branche |
| **L2** | Module pilote `article` + création `internal/platform/httpx` | `internal/modules/article/{handler,service,repository,dto,model}.go` + `service_test.go` pur (fake repository : `ErrNotFound`→404, 0-ligne update/delete) + `httpx` + `httpx_test.go` ; **suite `tests/` inchangée** | `make test` vert (assertions non modifiées) ; wiring review (interface `ArticleRepository` côté service ; helpers identité locaux au handler) ; `make generate_docs` → diff = déplacements d'annotations uniquement |
| **L3** | Module `auth` (sessions/JWT, middlewares, logouts, **soft-consume**) | `auth/{service,repository,middleware,handler,dto,model}.go` ; script Lua soft-consume (grâce 5 min, états `consumed:<delivered\|recovered>`) ; sweep in-process ; tests dédiés : rejeu / grâce / recovery **unique** / fails-closed / sweep / ownership (`TestRefreshOnlyConsumesMatchingOwner` déplacé) ; 7 tests middleware + `controllers/auth_test.go` migrés | `make test` vert ; matrice `login_test.go` inchangée ; seul carve-out = « rejeu pendant grâce » sous tests dédiés ; wiring review (struct `Secrets` injecté, constructeurs paramétrés `TokenAuth(svc)`) |
| **L4** | Module `tenant` + **UoW plateforme** | `tenant/{service,repository,model}.go` ; tests purs slug/réservés déplacés ; **création `pgdb.WithinTx` + interface étroite `pgdb.InTx`** dans `pool.go` manuscrit (+ couverture `pool_test.go`) | tests purs verts ; `make test` vert ; wiring review (**aucun `*pgdb.Queries` ne franchit la frontière de module**) ; test contrat `reservedSlugs` (P0.3) toujours vert |
| **L5** | Module `user` | `user/{handler,service,repository,dto,model}.go` ; register orchestré par `pgdb.WithinTx` + port `TenantProvisioner.CreateInTx` ; `service_test.go` pur (email dupliqué via `ErrNoRows`, slug réservé/épuisé) ; implémente `UserAuthenticator`/`UserReader` | `make test` vert (register/login inchangés, `ErrTenantSuspended` 403 avant bcrypt) ; wiring review (UoW : aucun `q *pgdb.Queries` en paramètre de port) |
| **L6** | Module `admin` | `admin/{handler,service,repository,dto}.go` ; audit **dans la même Tx** que `SetTenantStatus` (ordre prouvé par fake) ; CLI `-create-admin` via le service | `make test` vert (suite admin + `cmd/api/main_test.go`) ; wiring review (port `TenantReader` consommateur ; port `TokenService` déclaré par admin) |
| **L7** | Module `auth` — collage final (login unifié + refresh) | Refresh descendu dans `auth/service.go` ; `UnifiedLogin` (bcrypt factice + `DummyHashComparisons` déplacé) ; ports `UserAuthenticator`/`AdminAuthenticator`/`UserReader` implémentés par user/admin et **injectés dans `internal/app`** ; répertoires racine `controllers/`, `models/`, `forms/`, `middleware/` supprimés | build+vet+`make test`+`make e2e` verts ; **décompte `git ls-files` sur les 4 répertoires racine = 0** ; matrice `login_test.go` inchangée (406 à coût constant, 403 suspendu, champ `kind`) |
| **L8** | Nettoyage + règles projet (transverse) | `invoice/` → `internal/modules/invoice/` ; artefacts racine supprimés ; **`binding.Validator` relocalisé hors `NewRouter`** (sortie de dette tracée, directive 4) ; AGENTS/README/memory-bank réécrits au format module ; garde CI anti-imports racine | `make test` vert ; `go test ./internal/modules/invoice` vert (poppler) ; garde CI active ; docs conventions à jour |

## Protocole de validation par lot (gates)

Chaque lot est exécuté par une **session d'implémentation mandatée** (agent capable de modifier le code), dans l'ordre strict. Le gate se joue en trois couches, dans cet ordre :

**1. Batterie machine** (jouée par la session d'implémentation, sorties consignées) :
`go build -mod=readonly ./...` → `go vet -mod=readonly ./...` → tests des paquets touchés (`go test -mod=readonly -count=1 ./<pkg>`) → `make test` complet (L1 et L7 : `+ make e2e`) → régénérations idempotentes (`make generate_sqlc` / `make generate_mocks` / `make generate_docs` → diffs vides).

**2. Contrat HTTP** : `TestAPISurface` vert **sans édition de la liste attendue** ; assertions de `tests/` inchangées (sauf déplacements mécaniques listés par le plan-phases) ; tout écart = drift de surface API → **priorité absolue, stop-the-line**.

**3. Agent de validation technique (wiring & contrat)** — dispatch d'un **subagent frais en lecture seule** après chaque lot ; il rejoue la lecture du diff du lot contre la checklist certifiée :
1. Aucun `*pgdb.Queries` ne franchit une frontière de module (ni paramètre, ni retour, ni champ de struct échangé) ;
2. Interfaces cross-module **exclusivement côté consommateur**, limitées à la liste certifiée : ports `UserAuthenticator`/`AdminAuthenticator`/`UserReader` (définis dans auth, L3) ; port `TokenService` (déclaré par user L5 / admin L6, implémentation `*auth.Service` injectée dans `internal/app`) ; `TenantProvisioner` (déclaré par user, L5) ; `TenantReader` (déclaré par admin, L6) — **toute abstraction nouvelle hors design = FAIL** (amendement egm-lite requis d'abord) ;
3. Transactions : uniquement `pgdb.WithinTx`/`pgdb.InTx` (UoW) ou `WithTenant` (RLS) ; jamais de `SET` session-wide ; `set_config(..., true)` en transaction seul ;
4. Injection explicite dans `internal/app`/constructeurs ; aucun nouveau singleton, var de package ou état global ;
5. **Toute mutation de surface API globale (`binding.Validator`, mode gin, etc.) explicitement tracée** dans le rapport de gate + dette consignée (tracabilité directive 4) ;
6. Réponses HTTP byte-identiques au contrat actuel (unique carve-out acté : rejeu pendant grâce, L3) ;
7. **Rollback du lot borné et documenté** : branche, dernier commit vert, procédure de revert (reversible avant la phase concernée).
8. **Redis — propriétaire unique** : seul le module `auth` accède en production aux clés de session (records STRING adressés par l'UUID du JWT, index SET `user_sessions:{userID}`, préfixe `admin:`, script Lua soft-consume) ; aucun autre module sous `internal/modules/` n'importe `internal/platform/redisdb` ni ne lit/écrit ces clés — `cmd/api` initialise le client (`InitRedis`), `cmd/e2edb` purge l'index (CLI infra), les tests y accèdent directement pour leurs nettoyages/assertions ; nommage, TTL et formats inchangés (grâce 5 min sur les records consommés, éviction TTL primaire, sweep dans `auth`).

**Sortie du gate** : rapport `.kilo/plans/1795597264251-impl-log/L<N>-gate.md` (répertoire créé au premier gate) — commandes jouées + résultats, checklist 1-8 avec preuves (diffs, sorties de tests, liens commits), **inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte — opérationnalisation du critère « aucune phase ne baisse la couverture » : exigence qualitative, aucun seuil chiffré)**, verdict **PASS/FAIL**.

### En cas d'échec

- **Isolation stricte par lot** : corriger DANS le lot, rejouer la batterie complète ; jamais commencer le lot suivant avec un gate rouge ; jamais de commit correctif en avance — revert au dernier commit vert puis re-diagnostic (règle plan-phases « Risques transverses et rollback »).
- `TestAPISurface` FAIL (route perdue ou ajoutée) = stop-the-line (drift de surface API = tâche prioritaire).
- Doute de conception découvert en cours de lot → STOP ; amendement du plan-phases via **egm-lite** ; re-gate. Pas de pattern improvisé.

### Après PASS (protocole de suivi)

1. Plan-phases : marqueur `[Fait YYYY-MM-DD]` sur la phase + liens vers les artefacts (branche, plage de commits, `L<N>-gate.md`).
2. `memory-bank/` (`activeContext`, `progress`) + `AGENTS.md` si chemins/commandes changent (convention dépôt, chaque phase).
3. Le lot suivant est autorisé.

### Reprise / état courant

L'état vit dans trois lieux : les marqueurs `[Fait]` du plan-phases, les rapports `1795597264251-impl-log/`, et les branches `refactor/modular-phase-N`. **Une session de reprise ne démarre jamais un lot dont le prédécesseur n'est pas `[Fait]` + gate PASS.**

## Garde-fous transverses (rappel exécutoire)

- Interdiction absolue des raccourcis `*pgdb.Queries` inter-domaines — le validator les **refuse** en gate (checklist 1).
- Fenêtre transitoire A2 (L3-L6 uniquement) : imports racine↔`internal/modules/*` tolérés **dans les deux sens, pour le code en migration seulement** ; invariants absolus pendant toute la migration : `internal/platform/*` et `cmd/e2edb` n'importent **jamais** `internal/modules/*` ; la fenêtre est refermée vérifiablement en L7 (décompte `git ls-files` = 0).
- Pas de dérive de patterns : les gabarits/ports sont ceux du design ; toute nouveauté passe par amendement.
- Erreurs isolées par lot ; drifts de surface API prioritaires ; rollback borné, documenté, réversible (L3 : rollback pré-Phase-3 borné aux consommations en grâce — re-login équivalent à une consommation normale).
- Jamais `make test` / `make e2e` / `go test` en parallèle dans le même checkout ; secrets de test uniquement via `t.Setenv` ; `.env` jamais commité.

## Ordre d'exécution (résumé)

**G-1** (session dédiée) → L0 → L1 → L2 → L3 → L4 → L5 → L6 → L7 → L8 — chaque flèche = gate PASS obligatoire. Fin de migration = critères de sortie L8 + backlog certifié (montées de versions, RBAC câblé, CORS configurable, refresh httpOnly, normalisation des payloads d'erreur) repris post-Phase 8 hors périmètre. **Clôture — validation post-refactor** : egm-drift-report (skill du repo kilo) sur le design et le plan-phases contre le code migré, rapports datés dans `.kilo/plans/` (flux EGM du dépôt : brainstorm → draft → review → resolve → drift-report).
