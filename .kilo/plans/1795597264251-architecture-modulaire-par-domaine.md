# Migration incrémentale vers une architecture modulaire par domaine

> Document scindé le 2026-11-25 en deux fichiers pour rendre le cycle de review EGM exécutable : le fichier unique de 268 lignes dépassait la limite de sortie d'une réponse lors des étapes de review (notamment le Stage 6 — paraphrases verbatim — et les Stages 5/7/8/9). Ce fichier est désormais un point de redirection ; aucun contenu n'a été perdu.

- **Design** : [1795597264251-architecture-modulaire-par-domaine-design.md](1795597264251-architecture-modulaire-par-domaine-design.md) — lignes 1-116 et 261-268 de l'original (objectif, décisions structurantes actées, cartographie, invariants, domaines, couplages, layout cible, questions ouvertes). Review : egm-review, Step `feature_spec` (Stages 5-8).
- **Plan d'exécution** : [1795597264251-architecture-modulaire-par-domaine-plan-phases.md](1795597264251-architecture-modulaire-par-domaine-plan-phases.md) — lignes 118-259 de l'original (feuille de route Phases 0-8, proposition prête à l'emploi Phases 0-1, critères de sortie). Review : egm-review, Step `implementation` (Step 3 → DetailedRange = tout le document via l'en-tête `## Detailed Implementation Plan` ; Stage 9).

Contenu préservé à l'identique ; seuls ajustés : titres des fichiers, références croisées « §6 » → « PROPOSITION PRÊTE À L'EMPLOI », en-tête `## Detailed Implementation Plan`, notes d'en-tête de scission et mapping des lignes originales.

Les rapports de review antérieurs qui ciblent ce chemin (`zoom_out_report.md`, `goldfish_review_report.md`, racine de `.kilo/plans/`) ancrent leurs findings dans la numérotation de lignes de l'original — utiliser le mapping ci-dessus pour les reporter sur les fichiers scindés.

> **2026-11-25** : les 16 correctifs du `goldfish_review_report.md` (partie 2, décisions A2/B10 validées) sont **appliqués** sur les deux fichiers scindés ; le rapport est `Status : resolved`. La section « PROPOSITION PRÊTE À L'EMPLOI » du plan-phases s'intitule désormais « Étapes détaillées — Phase 0 et Phase 1 ».
