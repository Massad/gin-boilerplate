# Rapport — Étape 9 : Test de préparation à l'implémentation (DetailedRange, lignes 81..125)

*   **Relecteurs** : egm-implementation-readiness (Tech Lead), egm-codebase-details (auditeur conformité), egm-style (copiste technique)
*   **Standard** : `resources/document_standards/implementation-readiness-standard.md`
*   **Date** : 2026-11-23

## egm-implementation-readiness — Zero-Decision Rule

1. **Blocker** — `tests/article_test.go` (plan, ligne 92)
   Le plan indique « migration vers testify/suite, fixtures YAML en SetupTest, factories pour les cas dynamiques, require » sans détailler : quelles fonctions deviennent des méthodes de suite, où placer `SetupSuite`/`SetupTest`, comment remplacer les variables de package partagées, quelles assertions transformer, ni comment garantir l'ordre d'exécution identique. L'implémenteur doit prendre des décisions architecturales (structure de la suite, gestion d'état, sélection fixtures/factories). Guidance : spécifier une struct `ArticleSuite` (embed `suite.Suite`), déplacer les variables de package en champs, charger les fixtures dans `SetupSuite`, réinitialiser dans `SetupTest`, convertir chaque test en méthode, remplacer `assert`/`t.Fatal` par `require` aux points d'arrêt, garantir l'ordre (ordre de déclaration ou suite explicite), nettoyer dans `TearDownSuite`.

2. **Major** — `go.mod`/`go.sum` (plan, ligne 85) : aucune version épinglée pour `gofakeit/v7`, `testcontainers-go`, `testfixtures/v3`, `pressly/goose/v3`. Violation de la règle zéro décision. Épingler des versions précises ou définir la méthode d'obtention.

3. **Major** — `Makefile` (plan, ligne 86) : `MOCKERY_VERSION` sans valeur, règles absentes. Fournir version + texte complet des règles `install_mockery`/`generate_mocks` (pattern GOBIN `.cache/tools/` du Makefile existant).

4. **Major** — `.mockery.yaml` (plan, ligne 87) : « uniquement les interfaces existantes » sans les lister. Lister explicitement (cf. egm-codebase-details : seules `DBTX` et `Querier` existent, dans `internal/db`).

5. **Major** — `internal/testutil/factory.go` (plan, ligne 88) : signatures des factories non définies (retour, champs générés, usage gofakeit). Documenter chaque signature + seed basé sur `t.Name()`.

6. **Major** — `internal/testutil/fixtures.go` (plan, ligne 89) : tables à purger et mode de chargement non spécifiés. Documenter les tables (`tenant`, `"user"`, `admin_user`, `article`), le répertoire `tests/fixtures/`, l'appel testfixtures via le pool admin, la purge.

7. **Major** — `internal/testutil/postgres.go` (plan, ligne 90) : version d'image Postgres, mode d'exécution des migrations goose en bibliothèque et logique de détection Docker non spécifiées. Fournir constantes épinglées + mécanisme de skip.

8. **Major** — `tests/fixtures/*.yml` (plan, ligne 91) : contenu attendu non fourni. Documenter les colonnes réelles par table et les valeurs de départ (celles aujourd'hui en inline dans `tests/article_test.go`), mots de passe hachés bcrypt.

9. **Major** — `main_test.go` (plan, ligne 93) : le fichier n'utilise pas `assert` mais `t.Fatal`/`t.Skip` — « assert → require » inapplicable tel quel. Décision à trancher : soit conserver tel quel (défaut = arrêt, équivalent require), soit migration documentée vers require + gofakeit.

10. **Major** — `config/secrets_test.go` (ligne 94), `models/tenant_test.go` (ligne 95), `internal/db/pool_test.go` (ligne 96) : même inapplicabilité — aucun `assert` dans ces fichiers. Clarifier l'intention exacte par fichier.

11. **Major** — `controllers/auth_test.go` (ligne 97), `middleware/auth_test.go` (ligne 98) : « require, factories » trop vague — quelles lignes, quels appels, quel seeding. Patch concret requis.

12. **Major** — `AGENTS.md` (ligne 99) et `memory-bank/*` (ligne 100) : contenu exact des ajouts non fourni. Fournir le texte/paragraphes à insérer.

13. **Major** — Verification Plan (lignes 102-109) : la liste de paquets unitaires (`./config ./models ./invoice`) omet d'autres paquets à tests unitaires. Élargir ou basculer sur `./...`.

14. **Major** — Risks (lignes 111-117) : la migration de `tests/article_test.go` vers suite n'a qu'une attente prudente, sans mesure concrète. Détailler une stratégie : suite parallèle vérifiée contre l'exist avant bascule, ou flag de compilation progressif.

15. Constat négatif : les marqueurs Create/Modify sont présents ; les flags `-race`, `-tags=all`, `-mod=readonly` et les rôles non-superuser sont bien couverts.

## egm-codebase-details — Vérification code vs document

1. **Major** (Contradicts the Code) — lignes 12, 45, 94-98 du plan : cinq des six cibles de retrofit (`config/secrets_test.go`, `models/tenant_test.go`, `internal/db/pool_test.go`, `controllers/auth_test.go`, `middleware/auth_test.go`) n'utilisent NI `assert` NI testify — uniquement `t.Fatal`/`t.Fatalf`/`t.Errorf` (déjà fail-fast). Seul `tests/article_test.go` utilise `assert` (import l.27). Le préambule « les tests actuels utilisent `assert` seul » est inexact ; suivre le plan à la lettre amène l'implémenteur à chercher des `assert` inexistants.

2. **Major** (Missing Logic Detail) — `.mockery.yaml` (ligne 87) : vérification du code — `internal/db` a bien `DBTX` (`internal/db/db.go:14`) et `Querier` (`internal/db/querier.go:13`) ; `db` (Redis) n'a **aucune interface** (variables/fonctions de package) ; `models` n'a **aucune interface** (seules des assertions de type dans `models/auth.go:142,310`). La cible mockery est inefficace pour `db` et `models` sans que le plan l'anticipe.

3. **Nit** (Unstated Assumption) — skip testcontainers (lignes 43, 90) : `testcontainers-go` n'expose pas de fonction `IsDockerDaemonAvailable` ; l'API réelle est `testcontainers.IsDockerAvailable(ctx)` ou la gestion d'erreur au démarrage du conteneur. Mécanisme exact à documenter.

4. **Nit** (Unstated Assumption) — `pressly/goose/v3` en bibliothèque (ligne 85) vs `GOOSE_VERSION := v3.20.0` du Makefile (CLI) : l'alignement de versions CLI/bibliothèque n'est pas précisé. Risque de divergence.

5. Constat négatif (Undocumented Edge Cases) : aucun — les risques l.113-116 couvrent correctement ordre séquentiel, RLS, pool admin, seed gofakeit, non-fuite des deps.
6. Constat négatif (Non-Existent References) : aucun — toutes les références à l'état actuel existent (`go.mod:15` testify ; Makefile pattern ; `tests/article_test.go` 976 lignes, état partagé l.120-144 ; les six fichiers de test ; `memory-bank/activeContext.md` et `progress.md` ; `pgdb.TenantStore().Pool`).

## egm-style — Style (DetailedRange)

Nits : « épinglée » → « version figée » (l.86, 90, 95) ; « opt-in » → « optionnel » (l.108) ; « seed déterministe » → « graine déterministe » (l.115) ; « garde-fous » → « sécurités » (l.111) ; phrase longue à scinder (l.113) ; jugements subjectifs non étayés « stable » (l.105) et « propre » (l.108).
Constat négatif : hyperbole, redondance, caveats non calibrés — aucun constat.
