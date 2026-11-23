# Rapport consolidé de revue — Plan « Stack de test Go »

*   **Cible** : `.kilo/plans/1795428465642-stack-test-go.md` (125 lignes)
*   **Mode** : flux complet sans interruption (étapes 5→9), Goldfish en contextes frais
*   **Rapports sources** : `…zoom-out-report.md`, `…comprehension-report.md`, `…critic-report.md`, `…intermediate-zoom-out-report.md`, `…implementation-readiness-report.md`
*   **Date** : 2026-11-23

## Synthèse exécutive

Le plan est aligné avec les conventions du dépôt (outils épinglés, RLS, rôles non-superuser, flags de test) et propose un périmètre cohérent. Mais la revue révèle **une erreur factuelle sur l'état actuel des tests** qui invalide la moitié des postes de retrofit tels qu'écrits, **une cible mockery partiellement vide** (deux des trois packages n'ont aucune interface), et **plusieurs violations de la règle zéro décision** (versions non épinglées, suite testify non spécifiée, fixtures sans contenu). Deux des trois « Blockers » structurels du relecteur structural sont des faux positifs de découpage (sections existantes hors FeatureRange) et sont écartés.

## Actions consolidées par priorité

### 🔴 Bloqueurs

**A1 — Corriger l'énoncé du problème et le périmètre de retrofit (erreur factuelle confirmée par vérification croisée du code).**
Réalité : seul `tests/article_test.go` utilise `testify/assert` (l.27, mêlé à `t.Fatal`) ; les cinq autres cibles (`config/secrets_test.go`, `models/tenant_test.go`, `internal/db/pool_test.go`, `controllers/auth_test.go`, `middleware/auth_test.go`) n'utilisent ni `assert` ni testify — uniquement le `testing` natif (`t.Fatal`/`t.Errorf`, déjà fail-fast). La phrase « les tests actuels utilisent `assert` seul » (l.16) et les postes « assert → require » (l.94-98) sont inapplicables tels quels. Réécrire le Problem statement sur le mélange réel, et redéfinir le retrofit par fichier : soit migration documentée `t.Fatal` → `require` + factories, soit conservation explicite (décision à inscrire dans le plan, pas laissée à l'implémenteur).

**A2 — Spécifier la migration de `tests/article_test.go` vers `testify/suite` (règle zéro décision).**
À documenter : struct de suite (embed `suite.Suite`), variables de package (l.120-144) → champs, `SetupSuite`/`SetupTest`/`TearDownSuite`, conversion des fonctions de test en méthodes, garantie d'ordre d'exécution (l'ordre alphabétique actuel des fonctions Go n'est pas l'ordre des méthodes de suite), remplacement `assert`/`t.Fatal` → `require`. Stratégie d'atténuation concrète : suite parallèle vérifiée contre l'existante avant bascule, ou conservation du style variables de package avec adoption fixtures/factories uniquement.

**A3 — Ajouter la section CUJs (template obligatoire) ou justifier explicitement son absence.**
Pour ce plan d'infrastructure, des parcours développeur conviennent : « checkout frais → `make test` vert avec fixtures » ; « `make generate_mocks` idempotent (relance → diff vide) » ; « sans Docker → skip propre ».

### 🟡 Majeurs

**A4 — Épingler toutes les versions et fournir les règles complètes.**
`gofakeit/v7`, `testcontainers-go`, `testfixtures/v3`, `pressly/goose/v3` (à aligner sur `GOOSE_VERSION := v3.20.0` du CLI Makefile), `MOCKERY_VERSION` + texte exact des règles `install_mockery`/`generate_mocks` sur le pattern GOBIN `.cache/tools/` existant.

**A5 — Corriger le périmètre mockery.**
Seules `DBTX` (`internal/db/db.go:14`) et `Querier` (`internal/db/querier.go:13`) sont des interfaces mockables ; `db` (Redis) et `models` n'en déclarent aucune. Restreindre `.mockery.yaml` au package `internal/db` avec les deux interfaces nommées, et documenter la config exacte.

**A6 — Spécifier factories, chargeur de fixtures et contenu YAML.**
Signatures des quatre factories + seed gofakeit par `t.Name()` ; tables purgées et mode de chargement (pool admin) ; contenu des fixtures (colonnes réelles des tables, valeurs extraites des données inline actuelles de `tests/article_test.go`, mots de passe bcrypt). Note de dépendance : le contenu de `admin_user.yml` dépend de la colonne `name` issue de la migration du plan login — préciser la clause correspondante.

**A7 — Spécifier le mécanisme de skip Docker et l'image Postgres.**
`testcontainers-go` n'expose pas `IsDockerDaemonAvailable` ; utiliser la gestion d'erreur au démarrage du conteneur ou `testcontainers.IsDockerAvailable(ctx)` (selon l'API de la version épinglée) + `t.Skip`. Image Postgres épinglée à documenter.

**A8 — Ajouter un garde anti-fuite pour `internal/testutil`.**
Étendre le garde-fou CI (pattern « legacy drivers ») ou ajouter une vérification : aucun import de `internal/testutil` hors `*_test.go`.

**A9 — Expliciter les trade-offs techniques dans « Alternatives Considered ».**
Chaque rejet doit citer des inconvénients concrets (complexité, maintenance, dépendance), pas seulement « Rejected by author ».

**A10 — Définir les termes internes (auto-contenance).**
« RLS » (définir inline : Row-Level Security, filtrage PostgreSQL par ligne), symboles de l'Open Question (`pgx.Tx`, `InTx`, `pgdb.TenantStore()`/`AdminStore()` — un mot chacun), « épinglé » (« version figée »), rôle des fichiers `AGENTS.md` et `memory-bank/`.

**A11 — Élargir le Verification Plan.**
La liste `./config ./models ./invoice` omet des paquets à tests unitaires ; élargir ou préciser un passage `./...` (hors tag `all` pour les unitaires).

### 🟢 Mineurs

**A12 — Nommer les tests RLS existants comme garants préservés** (`tests/article_test.go` l.~658-711 : insert cross-tenant → 42501, 0 ligne sans contexte) dans le poste 8 et la section Risques.
**A13 — Déplacer les détails bas niveau du Technical Plan (noms de fichiers helpers, signatures) vers le plan détaillé** pour respecter l'altitude architecturale.
**A14 — Nettoyage stylistique groupé** : « cf. » → « voir » ; « 1re » → « première » ; « opt-in » → « optionnel » ; « là-bas » (l.29) reformuler ; redondance « Étendre la couverture … couvert » ; CALIBRATE : généralisations l.16, l.35, l.69 ; jugements subjectifs « stable », « propre » ; phrase longue l.113.
**A15 — Mises à jour AGENTS.md / memory-bank : joindre le contenu cible** (paragraphes à insérer) et livrer dans le même commit que le code.

### Écartés (faux positifs de découpage)

Les « Blockers » structurels « Detailed Implementation Plan absent » et « Verification Plan absente » : ces sections existent aux lignes 81 et 102, hors du FeatureRange (1..80) audité par le relecteur structural. Écartés par l'Elephant après vérification.

## Matrice de traçabilité

| Action | Constat(s) sources | Relecteur(s) | Étape |
|---|---|---|---|
| A1 | Constat n°2 (design) ; constat n°1 (codebase-details) ; n°9-12 (impl-readiness) ; confirmé par grep Elephant | design, codebase-details, impl-readiness | 7, 9 |
| A2 | Constat n°1 (impl-readiness) ; risque l.113 du plan ; n°14 (impl-readiness) | impl-readiness, design | 7, 9 |
| A3 | Constat n°1 (structural) | structural | 7 |
| A4 | n°2-3 (impl-readiness) ; n°4 (codebase-details) | impl-readiness, codebase-details | 9 |
| A5 | n°4 (impl-readiness) ; n°2 (codebase-details) | impl-readiness, codebase-details | 9 |
| A6 | n°5-8 (impl-readiness) | impl-readiness | 9 |
| A7 | n°7 (impl-readiness) ; n°3 (codebase-details) ; n°5 (design) | impl-readiness, codebase-details, design | 7, 9 |
| A8 | n°3 (design) ; n°2 (zoom-out intermédiaire) | design, zoom-out | 7, 8 |
| A9 | n°1 (design) ; n°3 (zoom-out intermédiaire) ; n°1 (zoom-out ét. 5) | design, zoom-out | 7, 8 |
| A10 | n°1-2, 5 (comprehension) ; n°1 (style) | comprehension, style | 6, 7 |
| A11 | n°13 (impl-readiness) | impl-readiness | 9 |
| A12 | n°4 (design) ; n°4 (zoom-out intermédiaire) | design, zoom-out | 7, 8 |
| A13 | n°1 (zoom-out ét. 5) ; n°1 (zoom-out intermédiaire) | zoom-out | 5, 8 |
| A14 | Constats style (étapes 7 et 9) | style | 7, 9 |
| A15 | n°12 (impl-readiness) ; n°5 (comprehension) ; n°6 (design) | impl-readiness, comprehension, design | 6, 7, 9 |

## Recommandation

Corriger A1-A3 avant toute implémentation (le plan contredit le code sur un fait vérifiable et laisse des décisions structurantes à l'implémenteur) ; appliquer A4-A11 en même passe ; A12-A15 opportuns.
