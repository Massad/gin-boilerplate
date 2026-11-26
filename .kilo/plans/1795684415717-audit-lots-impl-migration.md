# Audit du plan de migration modulaire (Stage 9) — constats & correctifs prêts à appliquer

> **Demande** : revoir `.kilo/plans/1795627363688-lots-impl-migration-modulaire.md` et ses sous-fichiers sur 5 points (frontières/DAG, isolation tenant, stratégie de test, Redis/cache, chemins de sortie & baseline), marquer chaque point ✅/⚠️/❌, puis corriger/enrichir les ⚠️/❌ par des diffs prêts à appliquer.
> **Nature du travail** : révision de documents plan uniquement — **aucun code source**. Règle « ne rien inventer » respectée : chaque correctif s'appuie sur la documentation existante (design certifié, plan-phases, skill egm-review du repo kilo, `.kilo/rules/go-architecture-modulaire.md`, code vérifié à la lecture) ; les informations introuvables sont signalées avant proposition.
> **Fichiers audités** (2026-11-26) : lots plan (contrat d'exécution L0-L8), design certifié `1795597264251-architecture-modulaire-par-domaine-design.md`, plan-phases `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`, split-review `1795601758566-architecture-modulaire-split-review.md`, `1795597264251-phases-review/style_findings.md` (G-1 en cours : persona style livré, findings non résolus, 2 personas en attente), skill egm-review (`skills/egm-review/SKILL.md` du repo kilo), état réel du dépôt (branche `modular_archi` courante, arbre propre sauf `1795597264251-phases-review/` non suivi, smoke `/v1/invoice` toujours présent `ci.yml:139`, `GetRedis()` consommé en production uniquement par `models/auth.go`).

## Étape 1 — Vérification & confirmation des lacunes

| # | Point | Verdict | Évidence (fichier · emplacement) |
|---|---|---|---|
| 1 | Frontières de domaines & DAG | ✅ Présent & clair | Bounded contexts : design §2 (tableau domaine → sources/tables/routes pour article/auth/user/tenant/admin/invoice) + §2-bis (pôles super-admin vs tenant : pools, tables, sessions, isolation) + §4 (layout cible, rôle par répertoire). Règles de dépendance : design §4 — `platform` n'importe jamais `modules` (invariant absolu) ; `modules → modules` uniquement via interfaces définies côté consommateur ; `*pgdb.Queries` jamais de frontière (décision 6 : UoW `pgdb.WithinTx` + interface étroite `pgdb.InTx`) ; fenêtre transitoire A2 (Phases 3-6) refermée vérifiablement en Phase 7 (`git ls-files` = 0) ; `internal/app` + `cmd/api` seuls importeurs de tout à terme. Communication inter-domaines : 2 familles nommées (ports d'orchestration `UserAuthenticator`/`AdminAuthenticator`/`UserReader` ; services techniques `TokenService`/`TenantProvisioner`/`TenantReader`) + UoW in-Tx ; liste verrouillée par la checklist gate item 2 (« toute abstraction nouvelle hors design = FAIL »). Pas de diagramme visuel, mais règles complètes et vérifiables mécaniquement (checklist + décompte). |
| 2 | Isolation Tenant | ✅ Présent & clair | Propagation : tenant issu UNIQUEMENT du claim JWT signé ; contexte RLS posé exclusivement par `WithTenant` (`set_config(..., true)` en transaction) — checklist gate item 3 interdit `SET` session-wide ; clés de contexte (`userID`/`tenantID`/`role`) posées par `TokenAuth` (module auth), helpers locaux aux handlers. Pools DB : `app_admin BYPASSRLS` (plateforme : login, register, refresh, backoffice) vs `app_user` + RLS échec fermé (métier) — design invariants + §2-bis. Isolation Redis : préfixes `admin:` vs sessions user, secrets de signature séparés, index `REDIS_DB` 1/2/3 épinglés (TestMain + critères de sortie). Fuites : filets nommés (`TestCrossTenantIsolation`, `TestRLSFailClosed`, `TestAdminTokenOnTenantRoute`, `TestTokenAuthRejectsAdminToken`, `TestOneOwnerConstraint`) ; rôle re-fraîchi depuis la base à chaque refresh (`OneWithTenant`). |
| 3 | Stratégie de test de migration | ⚠️ Présent mais partiel | Contract tests ✅ : `TestAPISurface` bidirectionnel des 14 routes, liste figée, interdiction d'éditer (P0.3) + suite `tests/` inchangée (contrat de bout en bout via `app.NewRouter`) + `service_test.go` pur par module (cas exigés listés par phase). Fixtures isolées ✅ : base jetable `artisan_test` recréée par `make test`, `TestMain` épinglant `REDIS_DB=2` + `DB_NAME=artisan_test`, e2e `artisan_e2e` index 3, secrets via `t.Setenv` uniquement. **Lacune 1** : le critère « aucune phase ne baisse la couverture » n'est pas opérationnalisé — aucun inventaire ni mesure consignés (aucun seuil chiffré n'existe par ailleurs dans le dépôt, qui n'instrumente pas la couverture). **Lacune 2** : aucune validation post-refactor document/code après L8 — le skill `egm-drift-report` (flux EGM du dépôt : brainstorm → draft → review → resolve → drift-report) n'est pas schedulé en clôture. |
| 4 | Restructuration Redis/Cache | ⚠️ Présent mais partiel | Nommage des clés ✅ : records STRING adressés par l'UUID du JWT (access 15 min / refresh 7 j), index SET `user_sessions:{userID}`, préfixe `admin:` (8 h) — formats inchangés, aucune migration de schéma ; conforme au code (`models/auth.go:119-127,260,302-360`). TTL ✅ : grâce fixe 5 min (min(grâce, TTL restant)) sur les records consommés, éviction TTL primaire, sweep in-process ~5 min (observabilité + purge défensive, invariant de course). **Lacune 3** : la propriété « Redis = domaine à propriétaire unique (module auth) » est un fait vérifié (grep : seul `models/auth.go` consomme `GetRedis()` hors tests) mais n'est écrite dans aucun plan, et aucune règle de gate ne l'interdit aux autres modules après découpe. **Lacune 4** : l'acquisition du client Redis par le repository sessions du module `auth` n'est spécifiée nulle part — la cible (`.kilo/rules/go-architecture-modulaire.md`) prohibe les singletons (`GetRedis()` cité), P1.2 garde les exports, ni constructeur injecté ni tolérance+dettes ne sont actés. |
| 5 | Chemins de sortie & Baseline | ❌ Absent ou insuffisant | Rapports de gate de lots ✅ : `.kilo/plans/1795597264251-impl-log/L<N>-gate.md` (contenu et verdict PASS/FAIL définis — le répertoire n'existe pas encore, création au premier gate). **Lacune 5** : rapports G-1 sans nommage standard — trois conventions cohabitent (skill egm-review du repo kilo : `implementation_readiness_report.md`/`goldfish_review_report.md`/`consolidated_review_report.md` ; cycle design-review : `{stage}_report.md` ; existant dans `1795597264251-phases-review/` : `style_findings.md`) ; la synthèse consolidée du stage n'a ni nom ni chemin épinglés. **Lacune 6** : la branche `modular_archi` (branche réelle, courante) n'apparaît dans aucun plan ; la ligne « État du dépôt » du lots plan est périmée (« branche `login` » — renommée depuis ; « arbre non propre : `AGENTS.md` modifié » — aujourd'hui committé, seul `1795597264251-phases-review/` non suivi) ; les branches `refactor/modular-phase-N` n'ont ni point de fourche ni cible de merge nommés. |

**Confirmation concise (≤10 lignes)** :
1. ✅ Frontières & DAG : complets, explicites, vérifiables en gate (design §2/§4 + checklist 1-2) — confirmé, aucune modification requise.
2. ✅ Isolation tenant : propagation RLS, pools BYPASSRLS/RLS, préfixes Redis, filets anti-fuite nommés — confirmé, aucune modification requise.
3. ⚠️ Tests : contract tests + fixtures solides ; couverture minimale non opérationnalisée (aucun inventaire exigé dans les rapports) et validation post-refactor non schedulée.
4. ⚠️ Redis : clés/TTL/cohérence spécifiés au byte près ; propriétaire unique non acté ni vérifié en gate ; acquisition du client (constructeur vs singleton) non tranchée.
5. ❌ Sorties & baseline : nommage G-1 non standard, `consolidated_report.md` sans chemin épinglé, `modular_archi` absente de tous les plans, état du dépôt périmé.

## Étape 2 — Correctifs (diffs prêts à appliquer)

**Gouvernance d'application** (respect des règles posées par les plans eux-mêmes) :
- **C1-C7** : éditions du **lots plan** `1795627363688-lots-impl-migration-modulaire.md` (contrat d'exécution — hors périmètre de certification G-1, qui cible le plan-phases). Application directe en une passe.
- **C8** : édition du **plan-phases** — **jamais en direct** : l'en-tête du lots plan impose « toute évolution du contenu d'une phase passe par egm-lite et édite le plan-phases » ; le texte prêt ci-dessous alimente le point d'audit C4 et sera appliqué par la passe egm-resolve de G-1.
- Les findings style F1-F14 de `style_findings.md` (7 Major, 8 Nit, non résolus) restent du ressort de G-1 — hors périmètre de cet audit.

### C1 — État du dépôt + baseline `modular_archi` (lots plan, § « Artefacts d'entrée »)

À remplacer :

```markdown
- **État du dépôt vérifié le 2026-11-25** : branche `login`, aucun commit de migration, Phase 0 non démarrée (smoke `/v1/invoice` toujours présent `ci.yml:139` ; pas de `tests/api_surface_test.go` ; pas de `internal/platform/`, `internal/app/`, `internal/modules/`, `cmd/api/`). Arbre de travail non propre : `AGENTS.md` modifié + fichiers `.kilo/plans/` non suivis. Point de départ : **G-1 puis L0**.
```

Par :

```markdown
- **État du dépôt vérifié le 2026-11-26** : branche `modular_archi` — **baseline officielle de la migration** (point de fourche et cible de merge des branches `refactor/modular-phase-N`, cf. « Structure des lots ») ; aucun commit de migration, Phase 0 non démarrée (smoke `/v1/invoice` toujours présent `ci.yml:139` ; pas de `tests/api_surface_test.go` ; pas de `internal/platform/`, `internal/app/`, `internal/modules/`, `cmd/api/`). Arbre de travail propre : seul `.kilo/plans/1795597264251-phases-review/` est non suivi (rapport `style_findings.md` du G-1 en cours). Point de départ : **G-1 puis L0**.
```

### C2 — Fourche/merge des lots sur la baseline (lots plan, § « Structure des lots », règles communes)

À remplacer :

```markdown
Règles communes : un lot = une branche `refactor/modular-phase-N` ; **séquence stricte — aucun code n'est généré avant que le gate du lot précédent soit PASS** ;
```

Par :

```markdown
Règles communes : un lot = une branche `refactor/modular-phase-N` **fourchue depuis `modular_archi` (baseline officielle) et re-mergée dans `modular_archi` uniquement si CI verte** ; **séquence stricte — aucun code n'est généré avant que le gate du lot précédent soit PASS** ;
```

### C3 — Nommage standardisé des rapports G-1 (lots plan, § « Gate G-1 », point 2)

À remplacer :

```markdown
2. Stage 9 en dispatch parallèle, **un subagent isolé par persona** (contexte frais, lecture seule) : `egm-implementation-readiness` + `egm-codebase-details` + `egm-style` (personas et standards chargés depuis le repo kilo : `\\wsl.localhost\Debian\home\sylvaindeb\projects\kilo\`) ; rapports en français dans `1795597264251-phases-review/`.
```

Par :

```markdown
2. Stage 9 en dispatch parallèle, **un subagent isolé par persona** (contexte frais, lecture seule) : `egm-implementation-readiness` + `egm-codebase-details` + `egm-style` (personas et standards chargés depuis le repo kilo : `\\wsl.localhost\Debian\home\sylvaindeb\projects\kilo\`) ; rapports en français dans `1795597264251-phases-review/`, **nommage standardisé du contrat** : un fichier par persona `{persona}_findings.md` (`implementation_readiness_findings.md`, `codebase_details_findings.md`, `style_findings.md` — ce dernier déjà livré) ; synthèse consolidée du stage dans `consolidated_report.md` ; la passe egm-resolve écrit son `goldfish_review_report.md` dans le même répertoire. Ce nommage, imposé par le présent contrat, prévaut sur les noms par défaut du skill egm-review (`implementation_readiness_report.md`, `consolidated_review_report.md`) — le synthesizer écrivant « to the report file the caller names » (`agents/synthesizer.md`, repo kilo).
```

### C4 — Points d'audit G-1 supplémentaires (lots plan, § « Gate G-1 », point 3)

À remplacer :

```markdown
    - cohérence de la puce P3 (soft-consume) avec le design décision 7 post-pass-2.
```

Par :

```markdown
    - cohérence de la puce P3 (soft-consume) avec le design décision 7 post-pass-2 ;
    - opérationnalisation du critère « aucune phase ne baisse la couverture » : acter l'exigence qualitative — chaque rapport `L<N>-gate.md` consigne l'inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte) ; aucun seuil chiffré (le dépôt n'instrumente pas la couverture — en inventer un serait une exigence nouvelle) ;
    - spécification zéro-décision de l'acquisition du client Redis par le repository sessions du module `auth` (Phase 3) : constructeur injecté (recommandé, aligné sur `.kilo/rules/go-architecture-modulaire.md` qui prohibe les singletons `GetRedis()`) vs tolérance du singleton pendant la fenêtre + dette tracée — texte prêt : cf. C8 ;
    - propriétaire unique Redis : acter que seul le module `auth` accède en production aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — vérifié à l'état initial : seul `models/auth.go` consomme `GetRedis()` hors tests ; règle reprise en gate (checklist item 8).
```

### C5 — Checklist gate : item 8 « Redis — propriétaire unique » (lots plan, § « Protocole de validation par lot », couche 3)

À remplacer :

```markdown
7. **Rollback du lot borné et documenté** : branche, dernier commit vert, procédure de revert (reversible avant la phase concernée).
```

Par :

```markdown
7. **Rollback du lot borné et documenté** : branche, dernier commit vert, procédure de revert (reversible avant la phase concernée).
8. **Redis — propriétaire unique** : seul le module `auth` accède en production aux clés de session (records STRING adressés par l'UUID du JWT, index SET `user_sessions:{userID}`, préfixe `admin:`, script Lua soft-consume) ; aucun autre module sous `internal/modules/` n'importe `internal/platform/redisdb` ni ne lit/écrit ces clés — `cmd/api` initialise le client (`InitRedis`), `cmd/e2edb` purge l'index (CLI infra), les tests y accèdent directement pour leurs nettoyages/assertions ; nommage, TTL et formats inchangés (grâce 5 min sur les records consommés, éviction TTL primaire, sweep dans `auth`).
```

### C6 — Rapport de gate : répertoire + inventaire de tests (lots plan, § « Sortie du gate »)

À remplacer :

```markdown
**Sortie du gate** : rapport `.kilo/plans/1795597264251-impl-log/L<N>-gate.md` — commandes jouées + résultats, checklist 1-7 avec preuves (diffs, sorties de tests, liens commits), verdict **PASS/FAIL**.
```

Par :

```markdown
**Sortie du gate** : rapport `.kilo/plans/1795597264251-impl-log/L<N>-gate.md` (répertoire créé au premier gate) — commandes jouées + résultats, checklist 1-8 avec preuves (diffs, sorties de tests, liens commits), **inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte — opérationnalisation du critère « aucune phase ne baisse la couverture » : exigence qualitative, aucun seuil chiffré)**, verdict **PASS/FAIL**.
```

### C7 — Clôture : validation post-refactor (lots plan, § « Ordre d'exécution »)

À remplacer :

```markdown
**G-1** (session dédiée) → L0 → L1 → L2 → L3 → L4 → L5 → L6 → L7 → L8 — chaque flèche = gate PASS obligatoire. Fin de migration = critères de sortie L8 + backlog certifié (montées de versions, RBAC câblé, CORS configurable, refresh httpOnly, normalisation des payloads d'erreur) repris post-Phase 8 hors périmètre.
```

Par :

```markdown
**G-1** (session dédiée) → L0 → L1 → L2 → L3 → L4 → L5 → L6 → L7 → L8 — chaque flèche = gate PASS obligatoire. Fin de migration = critères de sortie L8 + backlog certifié (montées de versions, RBAC câblé, CORS configurable, refresh httpOnly, normalisation des payloads d'erreur) repris post-Phase 8 hors périmètre. **Clôture — validation post-refactor** : egm-drift-report (skill du repo kilo) sur le design et le plan-phases contre le code migré, rapports datés dans `.kilo/plans/` (flux EGM du dépôt : brainstorm → draft → review → resolve → drift-report).
```

### C8 — Plan-phases, Phase 3 : acquisition du client Redis (texte prêt — à appliquer par la passe egm-resolve de G-1, jamais en direct)

**Information introuvable constatée** : ni le design ni le plan-phases ne spécifient comment le repository sessions du module `auth` obtient son client Redis ; la cible `.kilo/rules/go-architecture-modulaire.md` prohibe les singletons (`GetRedis()` cité nommément) tandis que P1.2 conserve les exports `redisdb`. Résolution recommandée (alignée sur la cible, sans exigence nouvelle) : constructeur injecté, lecture du singleton confinée au composition root pendant la fenêtre transitoire, dette tracée.

Dans `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`, Phase 3, puce **Risques** (l.30), à remplacer :

```markdown
- **Risques** : sémantique du script Lua (ownership) — couverte par `TestRefreshOnlyConsumesMatchingOwner` (le déplacer au module) ; secrets lus via `os.Getenv` dans le service — injecter un `Secrets` struct au constructeur (DI explicite, testable `t.Setenv`→variables) ; middlewares deviennent des constructeurs paramétrés (`TokenAuth(svc)`) → routes mises à jour uniquement dans `internal/app` + `tests/` (via app.NewRouter, rien à changer).
```

Par :

```markdown
- **Risques** : sémantique du script Lua (ownership) — couverte par `TestRefreshOnlyConsumesMatchingOwner` (le déplacer au module) ; secrets lus via `os.Getenv` dans le service — injecter un `Secrets` struct au constructeur (DI explicite, testable `t.Setenv`→variables) ; **acquisition du client Redis : `NewRepository(client *redis.Client)` — le client est résolu au câblage dans `internal/app` (lecture du singleton `redisdb.GetRedis()` confinée au composition root pendant la fenêtre transitoire ; `InitRedis` reste à l'entrypoint `cmd/api`), dette tracée pour sortie post-Phase 8 (cible `.kilo/rules/go-architecture-modulaire.md` : singletons prohibés) ; seul le module `auth` accède aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — import de `internal/platform/redisdb` interdit aux autres modules (`cmd/api`/`cmd/e2edb`/tests autorisés : wiring, purge infra, nettoyages)** ; middlewares deviennent des constructeurs paramétrés (`TokenAuth(svc)`) → routes mises à jour uniquement dans `internal/app` + `tests/` (via app.NewRouter, rien à changer).
```

## Application & validation

1. **Ordre** : C1→C7 en une passe sur le lots plan (recherche/remplacement exact des extraits ci-dessus, structure et titres inchangés) ; C8 réservé à G-1 (point d'audit C4 + texte prêt).
2. **Vérifications post-application** : grep `modular_archi` → 2 occurrences (Artefacts d'entrée + Structure des lots) ; grep `consolidated_report.md` → 1 occurrence (G-1 point 2) ; checklist numérotée 1-8 cohérente entre la couche 3 et « Sortie du gate » ; les 4 points d'audit initiaux de G-1 deviennent 7.
3. **Périmètre strict** : aucun code source, aucun autre fichier plan (design certifié et split-review intouchés ; `style_findings.md` et ses F1-F14 restent à G-1).
4. **Critère de succès de l'audit** : points 3, 4, 5 passent ⚠️/❌ → couverts (couverture opérationnalisée, Redis propriétaire unique + acquisition tranchée en recommendation, sorties standardisées, baseline `modular_archi` notée explicitement) ; points 1-2 documentés comme confirmés sans modification.

## Statut d'application (egm-resolve — auto accept, 2026-11-26)

- **C1-C7 appliqués** au lots plan `1795627363688-lots-impl-migration-modulaire.md` en un batch ; rapport `1795684415717-review/goldfish_review_report.md` marqué `Status: resolved` (5 findings PROPOSE FIX, 0 CONTEST / 0 ESCALATE / 0 DECLINE).
- **C8 NON appliqué — réservé à la session G-1 dédiée** (gouvernance : en-tête du lots plan — les évolutions du contenu d'une phase passent par G-1/egm-lite et éditent le plan-phases) ; le point d'audit correspondant est ancré dans G-1 (C4 appliqué, lots plan l.35). **Aucun agent d'implémentation ne doit appliquer C8 en direct.**
- **Vérifications post-application (relecture intégrale du lots plan, 103 lignes)** : `modular_archi` = 2 occurrences (l.11, l.42) ; `consolidated_report.md` épinglé (G-1 point 2, l.28) ; checklist 1-8 cohérente couche 3 ↔ « Sortie du gate » (l.73, l.75) ; points d'audit G-1 = 7 (l.30-36) ; structure Markdown intacte.
- **Résiduel découvert en re-vérification (non traité, hors batch accepté)** : ligne L0 pré-flight du tableau des lots (l.46) cite encore « `AGENTS.md` modifié » — état périmé depuis les commits `232eeaa`/`99c9302` ; l'intention (committer les fichiers docs/plans non suivis avant L0) reste valide, seuls les exemples datent. Correction d'une ligne à la prochaine passe, ou signalement à G-1.
