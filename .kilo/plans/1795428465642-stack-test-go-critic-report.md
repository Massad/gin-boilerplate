# Rapport — Étape 7 : Revue critique (FeatureRange, lignes 1..80)

*   **Relecteurs** : egm-design (critique architecture), egm-structural (structure), egm-style (style)
*   **Standard** : `resources/document_standards/feature-spec-standard.md`
*   **Date** : 2026-11-23
*   **Vérification Elephant** : le constat factuel n°2 (egm-design) a été croisé avec le code — CONFIRMÉ : seul `tests/article_test.go` importe `testify/assert` ; les autres fichiers de test utilisent le `testing` natif (`t.Fatal`/`t.Errorf`).

## egm-design — Design Critic (échelle d'origine conservée entre parenthèses)

1. **Major** — lignes 57-71 (« Alternatives Considered »)
   *Les rejets sont attribués à « Rejected by author » sans trade-offs techniques ou de périmètre concrets, sans lien vers issue/discussion/PR. Cela viole la règle des alternatives du standard.*
   Recommandation : réécrire chaque justification en explicitant les inconvénients techniques (complexité de maintenance, surcharge, couplage) ; lier toute décision s'appuyant sur un mainteneur.

2. **Major** — lignes 14-16 (« The Problem »)
   *L'affirmation « les tests actuels utilisent `assert` seul » est inexacte : plusieurs fichiers (`controllers/auth_test.go`, `middleware/auth_test.go`, `internal/db/pool_test.go`, `config/secrets_test.go`, `models/tenant_test.go`, `main_test.go`) n'utilisent pas testify du tout mais le paquet `testing` natif (`t.Fatal`, `t.Errorf`). La mauvaise caractérisation pourrait conduire à un retrofit inadapté.*
   Recommandation : reformuler — l'état actuel est un mélange (1 fichier partiellement testify, les autres en `testing` natif) ; l'objectif devient l'unification sur `require`/`suite` + factories/fixtures partout. **CONFIRMÉ par vérification croisée Elephant.**

3. **Major** (Moderate) — `internal/testutil/` (fichiers proposés)
   *Le plan suppose que `internal/testutil` reste réservé aux tests sans vérifier qu'aucun import de production ne pourra le référencer ; une fuite polluerait le binaire.*
   Recommandation : grep de `internal/testutil` hors `*_test.go` dans la CI ou au docteur de révision ; le garde-fou « legacy drivers » de la CI peut être étendu.

4. **Major** (Moderate) — lignes 94, 114 (fixtures via pool admin, RLS)
   *Le plan ne précise pas comment les tests éprouveront la RLS via le pool tenant une fois les données seedées par le pool admin — risque de masquer un dysfonctionnement RLS.*
   Recommandation : conserver explicitement les tests RLS fail-closed existants (`tests/article_test.go` lignes ~658-711 : insert cross-tenant 42501, 0 ligne sans contexte) et les nommer comme garants dans le plan. *(Nuance Elephant : la suite existante contient déjà ces tests ; le retrofit doit les préserver — le plan le dit partiellement ligne 92/114 mais sans les nommer.)*

5. **Major** (Moderate) — ligne 43 (skip testcontainers)
   *Le mécanisme de skip si Docker est indisponible n'est pas spécifié (runtime ? tag de build ?), risquant des échecs imprévisibles sans Docker.*
   Recommandation : préciser la technique (ex. `testcontainers` ping Docker au setup, `t.Skip` sinon) et son comportement en CI.

6. **Nit** (Minor) — lignes 99-100 (AGENTS.md, memory-bank)
   *Mises à jour documentaires sans mécanisme de vérification — risque de dérive doc/réalité.*
   Recommandation : livrer ces mises à jour dans le même commit que le code ; pas de test dédié nécessaire à ce stade.

## egm-structural — Structural (portée : lignes 1..80 uniquement)

1. **Blocker** — entre lignes 33 et 37 : section « Critical User Journeys (CUJs) » absente du document (aucune section CUJ n'existe nulle part dans le fichier).
2. **Blocker** — section « Detailed Implementation Plan » absente de la plage auditée. *(Annotation Elephant : faux positif de découpage — la section existe à la ligne 81, hors FeatureRange ; le relecteur n'avait que les lignes 1..80. À ignorer.)*
3. **Blocker** — section « Verification Plan » absente de la plage auditée. *(Annotation Elephant : faux positif de découpage — existe à la ligne 102. À ignorer.)*
4. Aucune violation de hiérarchie de titres, aucune violation de mise en forme Markdown, aucun écart de conformité de liens (Info, constat négatif).

## egm-style — Style Copy Editor

Majeurs :
1. **Major** — lignes 20, 40, 43 : « épinglé » (jargon interne de version figée) jamais expliqué pour un lecteur externe.
2. **Major** — ligne 35 : « 1re » informel + formulation lourde de l'Open Question.
3. **Major** — ligne 43 : « pattern skip si Docker indisponible » — formulation non idiomatique à réécrire.
4. **Major** (CALIBRATE) — ligne 16 : généralisation « le moindre jeu de données exige du code manuel et chaque nouvelle feature duplique ce socle » — exagère l'état actuel, à calibrer sur la tendance réelle.
5. **Major** (CALIBRATE) — ligne 35 : jugement présenté comme fait (« relève d'un refactoring à valider séparément ») — à rattacher à un critère de décision.
6. **Major** (CALIBRATE) — ligne 69 : affirmation absolue « le SQL brut resterait du code manuel à dupliquer » — à calibrer.

Nits : « conventions consolidées » (l.12) vague ; « socle homogène » (l.61) métaphorique ; « rien à ajouter pour lui, seulement à harmoniser » redondant (l.12) ; « cf. » (l.20) et « comme sqlc » (l.40) implicites ; « là-bas » (l.29) familier ; « Opt-in » (l.43) anglicisme ; redondance « Étendre la couverture … couvert » (l.29) ; « du code manuel à dupliquer » (l.69) approximatif.

## Synthèse de l'étape

Constats valides à traiter : inexactitude factuelle du Problem statement (n°2 design, confirmée), justifications d'alternatives sans trade-offs (n°1 design), absence réelle de section CUJ (n°1 structural), jargon non défini (style), mécanisme de skip non spécifié (n°5 design), garde RLS à nommer explicitement (n°4 design), risque d'import testutil (n°3 design). Deux faux positifs de découpage à écarter (DIP, Verification Plan hors plage).
