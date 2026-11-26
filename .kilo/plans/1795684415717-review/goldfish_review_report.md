# Goldfish Annotated Review Report — Vérification des 5 points critiques (post-audit `1795684415717`)

**Status: resolved** — auto accept (gate egm-resolve, 2026-11-26) : C1-C7 appliqués au lots plan en un batch ; C8 routé vers G-1 (point d'audit ancré). 0 CONTEST / 0 ESCALATE / 0 DECLINE.

**Target:** `.kilo/plans/1795627363688-lots-impl-migration-modulaire.md` (fichier entier) ; finding 4 uniquement : `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-plan-phases.md` (Phase 3, Risques, l.30 — édition routée par G-1, jamais directe)
**Mode:** egm-resolve — relecture intégrale des cibles le 2026-11-26 (constat pré-application) : **aucun correctif de l'audit n'était appliqué** (l.11 « branche `login` », l.39 sans baseline, G-1 point 2 sans nommage, point 3 limité à 4 bullets, checklist 1-7, l.71 « checklist 1-7 », l.99 sans clôture post-L8, plan-phases l.30 inchangé).
**FindingsList:** les 5 points de vérification (prompt 2026-11-26). Correctifs de référence : `.kilo/plans/1795684415717-audit-lots-impl-migration.md` (C1-C8). Aucun ESCALATE, aucun CONTEST, aucun DECLINE — 5 findings PROPOSE FIX, tous Major.

## Prioritized Findings & AI Suggestions

### [Issue 1] [Point 1 — Baseline de branche] - Severity: Major
*   **Affected Scope:** l.11, l.39 (`1795627363688-lots-impl-migration-modulaire.md`)
*   **Discrepancy Description:** le plan ne référence pas `modular_archi` comme branche cible ; l'état du dépôt consigné est périmé (« branche `login` » — branche inexistante depuis le renommage ; « arbre non propre : `AGENTS.md` modifié » — aujourd'hui committé).
*   **Rationale:** la ligne « État du dépôt » est le point d'entrée des sessions de reprise ; les branches `refactor/modular-phase-N` (l.39) n'ont ni point de fourche ni cible de merge nommés.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** vérifié ce jour : `git branch --list` → `master`, `modular_archi` (courante) ; `login` n'existe plus ; `modular_archi` : 0 occurrence dans le lots plan. Aucune invention : la branche réelle est constatée, le reste suit la discipline existante (« merge uniquement si CI verte », plan-phases l.69/102).
*   ***Proposed Edit:*** correctifs **C1 + C2** de l'audit (§ Étape 2) :

    *   Location: l.11 (« Artefacts d'entrée », ligne « État du dépôt »)
    *   Change: remplacer par :
        > `- **État du dépôt vérifié le 2026-11-26** : branche `modular_archi` — **baseline officielle de la migration** (point de fourche et cible de merge des branches `refactor/modular-phase-N`, cf. « Structure des lots ») ; aucun commit de migration, Phase 0 non démarrée (smoke `/v1/invoice` toujours présent `ci.yml:139` ; pas de `tests/api_surface_test.go` ; pas de `internal/platform/`, `internal/app/`, `internal/modules/`, `cmd/api/`). Arbre de travail propre : seul `.kilo/plans/1795597264251-phases-review/` est non suivi (rapport `style_findings.md` du G-1 en cours). Point de départ : **G-1 puis L0**.`

    *   Location: l.39 (« Structure des lots », règles communes)
    *   Change: `un lot = une branche `refactor/modular-phase-N`` → `un lot = une branche `refactor/modular-phase-N` **fourchue depuis `modular_archi` (baseline officielle) et re-mergée dans `modular_archi` uniquement si CI verte**`

### [Issue 2] [Point 2 — Chemins de sortie standardisés] - Severity: Major
*   **Affected Scope:** l.28 (Gate G-1, point 2), l.71 (« Sortie du gate »)
*   **Discrepancy Description:** pas de convention unique pour les rapports G-1 : le répertoire `1795597264251-phases-review/` est épinglé (l.28) mais aucun nom de fichier ; trois conventions cohabitent (skill egm-review : `implementation_readiness_report.md`/`goldfish_review_report.md`/`consolidated_review_report.md` ; cycle design-review : `{stage}_report.md` ; existant : `style_findings.md`) ; `consolidated_report.md` n'apparaît nulle part. Seule la famille des gates de lots est standardisée (l.71 : `1795597264251-impl-log/L<N>-gate.md`).
*   **Rationale:** sans nommage épinglé, chaque cycle produira ses rapports sous des noms divergents (déjà le cas : design-review vs phases-review), et la synthèse consolidée du stage n'a pas de chemin garanti.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** fondé sur le skill lui-même : `agents/synthesizer.md` (repo kilo) écrit « to the report file the caller names » — le contrat d'exécution EST l'autorité de nommage ; le répertoire des artefacts G-1 est déjà fixé par le split-review (tâche 5 : « artefacts dans `1795597264251-phases-review/` »).
*   ***Proposed Edit:*** correctif **C3** (l.28, ajout du nommage) + **C6** (l.71, alignement « checklist 1-8 » + répertoire créé au premier gate) — textes exacts dans l'audit (§ Étape 2) :

    *   Location: l.28, fin du point 2 du G-1
    *   Change: compléter « rapports en français dans `1795597264251-phases-review/`. » par :
        > `rapports en français dans `1795597264251-phases-review/`, **nommage standardisé du contrat** : un fichier par persona `{persona}_findings.md` (`implementation_readiness_findings.md`, `codebase_details_findings.md`, `style_findings.md` — ce dernier déjà livré) ; synthèse consolidée du stage dans `consolidated_report.md` ; la passe egm-resolve écrit son `goldfish_review_report.md` dans le même répertoire. Ce nommage, imposé par le présent contrat, prévaut sur les noms par défaut du skill egm-review (`implementation_readiness_report.md`, `consolidated_review_report.md`) — le synthesizer écrivant « to the report file the caller names » (`agents/synthesizer.md`, repo kilo).`

### [Issue 3] [Point 3 — Stratégie de test opérationnelle] - Severity: Major
*   **Affected Scope:** l.71 (contenu du rapport de gate), l.99 (« Ordre d'exécution ») ; plan-phases l.164-165 (critères de sortie)
*   **Discrepancy Description:** le critère « aucune phase ne baisse la couverture » (plan-phases l.164) n'est pas opérationnalisé — aucun inventaire de tests exigé dans `L<N>-gate.md` (l.71 le décrit sans inventaire) ; les fixtures isolées sont en revanche explicitement définies (plan-phases l.165 : `TestMain` `REDIS_DB=2` + `DB_NAME=artisan_test`, index 1/2/3, `t.Setenv`) ; aucun calendrier de validation post-refactor après L8 (l.99 s'arrête au backlog certifié).
*   **Rationale:** « ne baisse pas la couverture » sans consignation n'est ni vérifiable ni opposable en gate ; la clôture de migration n'a aucune passe document/code.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** sans inventer d'exigence nouvelle : le dépôt n'instrumente pas la couverture (aucun `%` dans Makefile/CI) — l'opérationnalisation qualitative (inventaire des fonctions déplacées/ajoutées dans le rapport de gate) est la seule alignée sur l'existant ; la validation post-refactor s'appuie sur le skill `egm-drift-report` déjà présent dans le flux EGM du dépôt (AGENTS.md : brainstorm → draft → review → resolve → drift-report).
*   ***Proposed Edit:*** correctifs **C6** (inventaire dans le rapport de gate) + **C7** (clôture egm-drift-report post-L8) — textes exacts dans l'audit (§ Étape 2) :
    *   Location: l.71
    *   Change: `**Sortie du gate** : rapport `.kilo/plans/1795597264251-impl-log/L<N>-gate.md` (répertoire créé au premier gate) — commandes jouées + résultats, checklist 1-8 avec preuves (diffs, sorties de tests, liens commits), **inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte — opérationnalisation du critère « aucune phase ne baisse la couverture » : exigence qualitative, aucun seuil chiffré)**, verdict **PASS/FAIL**.`
    *   Location: l.99, en fin de paragraphe
    *   Change: ajouter : `**Clôture — validation post-refactor** : egm-drift-report (skill du repo kilo) sur le design et le plan-phases contre le code migré, rapports datés dans `.kilo/plans/` (flux EGM du dépôt : brainstorm → draft → review → resolve → drift-report).`

### [Issue 4] [Point 4 — Spécifications Redis/Cache ratifiées] - Severity: Major
*   **Affected Scope:** checklist l.63-69 (items 1-7, aucun Redis) ; plan-phases l.30 (Phase 3, Risques — rien sur l'acquisition du client) ; l.46/L3 et l.17 du lots plan + plan-phases l.26 (TTL/nommage/cohérence : présents et complets)
*   **Discrepancy Description:** TTL/invalidation/cohérence inter-domaines sont actés au byte près (soft-consume, grâce 5 min, formats inchangés, rollback borné) ; mais (i) le pattern d'acquisition du client Redis (DI vs singleton) n'est spécifié nulle part — information introuvable constatée : le plan-phases injecte `Secrets` (l.30) mais pas le client ; la cible `.kilo/rules/go-architecture-modulaire.md` prohibe les singletons (`GetRedis()` cité) tandis que P1.2 en conserve les exports ; (ii) le propriétaire unique (module `auth`) n'est ni acté ni vérifié en gate.
*   **Rationale:** après découpe, rien n'empêche un module d'importer `internal/platform/redisdb` et de toucher les clés de session d'un autre domaine ; le wiring du repository sessions est indéterminé.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** fait vérifié : grep `GetRedis()` sur `*.go` → seul `models/auth.go` consomme en production (le reste = tests) ; la résolution recommandée (constructeur `NewRepository(client *redis.Client)`, lecture du singleton confinée au composition root pendant la fenêtre, dette tracée) est alignée sur la cible `.kilo/rules` — aucune invention.
*   ***Proposed Edit:*** correctifs **C5** (checklist item 8 du lots plan) + **C8** (plan-phases l.30 — **routé par G-1** : l'en-tête du lots plan (l.3) impose que toute évolution du contenu d'une phase passe par egm-lite/G-1 et édite le plan-phases ; texte prêt, jamais application directe) + **C4** (point d'audit G-1 correspondant) — textes exacts dans l'audit (§ Étape 2) :
    *   Location: lots plan, après l'item 7 de la checklist (l.69)
    *   Change: ajouter l'item 8 : `**Redis — propriétaire unique** : seul le module `auth` accède en production aux clés de session (records STRING adressés par l'UUID du JWT, index SET `user_sessions:{userID}`, préfixe `admin:`, script Lua soft-consume) ; aucun autre module sous `internal/modules/` n'importe `internal/platform/redisdb` ni ne lit/écrit ces clés — `cmd/api` initialise le client (`InitRedis`), `cmd/e2edb` purge l'index (CLI infra), les tests y accèdent directement pour leurs nettoyages/assertions ; nommage, TTL et formats inchangés (grâce 5 min sur les records consommés, éviction TTL primaire, sweep dans `auth`).`
    *   Location: plan-phases l.30 (Risques Phase 3) — via G-1 : ajouter après « injecter un `Secrets` struct au constructeur (DI explicite, testable `t.Setenv`→variables) » : `**acquisition du client Redis : `NewRepository(client *redis.Client)` — le client est résolu au câblage dans `internal/app` (lecture du singleton `redisdb.GetRedis()` confinée au composition root pendant la fenêtre transitoire ; `InitRedis` reste à l'entrypoint `cmd/api`), dette tracée pour sortie post-Phase 8 (cible `.kilo/rules/go-architecture-modulaire.md` : singletons prohibés) ; seul le module `auth` accède aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — import de `internal/platform/redisdb` interdit aux autres modules (`cmd/api`/`cmd/e2edb`/tests autorisés : wiring, purge infra, nettoyages)**`

### [Issue 5] [Point 5 — Validation des gates (lacunes tests/Redis ratifiées ?)] - Severity: Major
*   **Affected Scope:** l.29-33 (G-1, points d'audit : 4 items), l.63-69 (checklist), l.71 (« checklist 1-7 »)
*   **Discrepancy Description:** les lacunes identifiées (couverture non opérationnalisée, Redis propriétaire unique/acquisition) ne sont ratifiées dans aucun gate : le G-1 point 3 ne compte que 4 points d'audit (InTx/WithinTx, sweep d'ancrage, fraîcheur Phase 8, cohérence P3) ; la checklist ne compte aucun item Redis ; le rapport de gate exige la « checklist 1-7 » sans inventaire de tests.
*   **Rationale:** sans ratification dans les gates, les lacunes ne seront ni auditées par G-1 ni vérifiées lot par lot.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** le mécanisme de ratification existe déjà dans le plan : les points d'audit du G-1 sont « alimentés par les vérifications de la session » (l.29) et la checklist est le support du validator — il s'étend les deux, sans nouveau mécanisme.
*   ***Proposed Edit:*** correctifs **C4** (G-1 point 3 : +3 points d'audit) + **C5** (checklist item 8) + **C6** (« checklist 1-8 » + inventaire) — textes exacts dans l'audit (§ Étape 2) :
    *   Location: l.33 (dernier bullet du point 3 du G-1)
    *   Change: remplacer `cohérence de la puce P3 (soft-consume) avec le design décision 7 post-pass-2.` par :
        > `cohérence de la puce P3 (soft-consume) avec le design décision 7 post-pass-2 ;
        > opérationnalisation du critère « aucune phase ne baisse la couverture » : acter l'exigence qualitative — chaque rapport `L<N>-gate.md` consigne l'inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte) ; aucun seuil chiffré (le dépôt n'instrumente pas la couverture — en inventer un serait une exigence nouvelle) ;
        > spécification zéro-décision de l'acquisition du client Redis par le repository sessions du module `auth` (Phase 3) : constructeur injecté (recommandé, aligné sur `.kilo/rules/go-architecture-modulaire.md` qui prohibe les singletons `GetRedis()`) vs tolérance du singleton pendant la fenêtre + dette tracée — texte prêt : cf. C8 ;
        > propriétaire unique Redis : acter que seul le module `auth` accède en production aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — vérifié à l'état initial : seul `models/auth.go` consomme `GetRedis()` hors tests ; règle reprise en gate (checklist item 8).`

## Synthèse

| Point | Verdict | Correctifs | Cible |
|---|---|---|---|
| 1 — Baseline `modular_archi` | ❌ | C1 + C2 | lots plan (l.11, l.39) |
| 2 — Chemins de sortie standardisés | ⚠️ (impl-log ✅ ; G-1 ❌) | C3 + C6 | lots plan (l.28, l.71) |
| 3 — Stratégie de test opérationnelle | ⚠️ (fixtures ✅ ; couverture/ post-refactor ❌) | C6 + C7 | lots plan (l.71, l.99) |
| 4 — Redis/Cache ratifiés | ⚠️ (TTL/clés ✅ ; acquisition/propriétaire ❌) | C5 + C8 (+C4) | lots plan (l.69) + plan-phases (l.30, via G-1) |
| 5 — Ratification dans les gates | ❌ | C4 + C5 + C6 | lots plan (l.33, l.69, l.71) |

Application : C1-C7 en une passe sur le lots plan (fichiers plan uniquement) ; C8 réservé à la passe egm-resolve de G-1 (gouvernance l.3 du lots plan). Vérifications post-application : grep `modular_archi` → 2 occurrences ; grep `consolidated_report.md` → 1 ; checklist 1-8 cohérente ; points d'audit G-1 = 7.
