# Rapport — Étape 6 : Test de compréhension

*   **Cible** : `.kilo/plans/1795428465642-stack-test-go.md` (FeatureRange 1..80, DetailedRange 81..125)
*   **Relecteur** : egm-comprehension (lecteur naïf, document seul, sans liste de contexte)
*   **Date** : 2026-11-23

## Paraphrase

**Ce que ce document propose**

Le document décrit l'ajout de quatre outils de test Go (mockery, gofakeit/v7, testcontainers-go, testfixtures/v3) et le retrofit des tests existants afin d'adopter des conventions consolidées : utilisation systématique de `require`/`suite` de testify, fabrication de données de test avec gofakeit, chargement de fixtures déclaratives YAML via testfixtures pour la suite d'intégration, tout en préservant les flags actuels (`-race`, `-tags=all`, `-mod=readonly`) et les connexions en rôles non-superuser.

**Comment fonctionne le système existant, selon ce document**

Actuellement les tests s'appuient uniquement sur `assert`, utilisent des données inline et partagent un état via des variables de package (`tests/article_test.go`). Aucun mock, aucune factory, aucune fixture déclarative ni dépendance éphémère n'est présent ; chaque nouvelle fonctionnalité oblige à recréer manuellement le même socle de données, ce qui entraîne duplication et fragilité.

## Constats

1. **Sévérité** : Major
   **Portée affectée** : lignes 30 et 114 (mention de *RLS*)
   **Description de l'écart & justification** : l'acronyme *RLS* (Row-Level Security) apparaît sans définition ni explication. Un lecteur ignorant ce mécanisme spécifique à PostgreSQL ne comprendra pas pourquoi le chargement des fixtures doit passer par le pool admin ni quels risques de contournement sont évoqués.

2. **Sévérité** : Major
   **Portée affectée** : ligne 35 (Open Questions) — termes *pgx.Tx*, *InTx*, *pgdb.TenantStore()*, *pgdb.AdminStore()*
   **Description de l'écart & justification** : ces symboles font référence à des composants internes du projet (couche sqlc/pgx, gestion de transactions, stores tenant/admin) mais ne sont jamais définis ni liés à une documentation. Leur utilisation dans une question ouverte suppose une connaissance préalable de l'architecture interne, ce qui contrevient à l'exigence d'auto-contenance.

3. **Sévérité** : Nit
   **Portée affectée** : ligne 40 — référence au motif *« cf. swag/sqlc/goose »* pour l'installation de mockery
   **Description de l'écart & justification** : l'expression évoque un schéma d'installation utilisé ailleurs dans le dépôt, sans lien ni bref descriptif. Un lecteur ne connaissant pas ces outils ne sait pas comment reproduire le procédé.

4. **Sévérité** : Nit
   **Portée affectée** : ligne 43 — dépendance *pressly/goose/v3* pour les migrations en bibliothèque
   **Description de l'écart & justification** : la bibliothèque *pressly/goose/v3* est mentionnée comme nécessaire, mais son rôle n'est pas expliqué. Un néophyte ignore pourquoi cette dépendance apparaît seulement dans ce contexte.

5. **Sévérité** : Major
   **Portée affectée** : lignes 99-100 — fichiers *AGENTS.md*, *memory-bank/activeContext.md*, *memory-bank/progress.md*
   **Description de l'écart & justification** : ces fichiers spécifiques au dépôt sont listés comme devant être modifiés, mais leur fonction dans le projet n'est jamais décrite. Sans explication, le lecteur ne sait pas quel impact ces changements auraient sur la configuration ou le suivi de progression.

## Note de synthèse

La paraphrase est fidèle à l'intention de l'auteur (vérifiée en contexte Elephant). Les constats sont des manques d'auto-contenance du document, pas des erreurs de fond.
