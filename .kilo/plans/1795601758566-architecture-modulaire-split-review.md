# Scission du plan « architecture modulaire » et cycle de review EGM

> Ce document est le **plan de suivi** : il n'est PAS une cible de review. Les cibles egm-review sont les deux fichiers scindés, listés dans les tâches 3 et 5 ci-dessous.

## Contexte

Le plan unique `.kilo/plans/1795597264251-architecture-modulaire-par-domaine.md` (268 lignes, ~30 Ko) faisait échouer le cycle egm-review : « Response reached its output limit and may be incomplete ». Causes : le Stage 6 exige deux paraphrases verbatim (sortie ≈ 2× la taille du document), les Stages 5/8 lisent le fichier complet + contexte, les Stages 7/9 tournent in-thread à 3 personas chacun.

Par ailleurs, une review précédente a laissé un état exploitable : `.kilo/plans/goldfish_review_report.md` (racine de `.kilo/plans/`) — 16 findings (A1-A2 Blockers, B1-B10 Majors, C1-C4 Nits), **tous classés PROPOSE FIX**, « en attente — validation des décisions A2/B10 requise avant génération des modifications détaillées (partie 2) ». Le rapport consolidé référencé (`1795597264251-consolidated-review-report.md`) est **perdu** — seuls le tableau récapitulatif et les sections de décision A2/B10 subsistent. Aucun des 16 correctifs n'est appliqué dans le document (vérifié : typos C1 présentes, `invoice/` toujours « optionnel », critère §7 inchangé). **État résorbé le 2026-11-25 : les 16 correctifs sont appliqués (tâches 1-2 ci-dessous marquées [Fait]).**

## État — exécuté pendant cette session (fichiers plan uniquement, aucune réécriture)

- [x] `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-design.md` : design (lignes originales 1-116 + 261-268), ~135 lignes — la paraphrase verbatim du Stage 6 redevient faisable.
- [x] `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-plan-phases.md` : plan d'exécution (lignes originales 118-259) sous en-tête `## Detailed Implementation Plan` — le Step 3 d'egm-review découvre automatiquement la frontière sémantique → DetailedRange = tout le document ; le preset « Implementation plan only » ne passe **pas** par le Stage 6 (aucune paraphrase).
- [x] Original remplacé par un stub de redirection avec mapping des lignes originales (aucun contenu perdu).
- [x] Liens croisés en vrais liens Markdown : le mécanisme « Linked Files » (Step 1.3 d'egm-review) ajoute automatiquement le document frère au GoldfishContext de chaque review.
- [x] Vérifié par grep : aucune référence au chemin du plan hors `.kilo/plans/` (memory-bank, AGENTS.md, code propres).

## Mode d'exécution des reviews — isolation Goldfish (s'applique aux tâches 3-5)

Le skill egm-review impose déjà ce mode (source de vérité : repo kilo, `execution_instructions.md` § « Subagent Delegation ») — rappelé ici noir sur blanc car son non-respect est la cause racine des troncatures constatées :

- **Chaque persona de review = un subagent isolé** (outil `Task` de Kilo), contexte frais sans historique de conversation. Jamais une review menée dans le contexte principal. Si l'environnement ne peut pas spawner de subagents : **s'arrêter net** — « a review in the same context is not a Goldfish review ».
- **Dispatch parallèle** là où le skill le prévoit, un subagent par persona : Stage 5 (`egm-zoom-out` + `egm-inventive-catalyst`), Stage 6 (`egm-comprehension`), Stage 7 (`egm-design`/`egm-structural`/`egm-style`), Stage 8 (`egm-zoom-out`), Stage 9 (`egm-implementation-readiness`/`egm-codebase-details`/`egm-style`). L'orchestrateur consolide ensuite les findings dans le rapport de stage.
- **Une fenêtre de contexte par cycle** : la review design (tâche 3) et la review plan-phases (tâche 5) se déroulent chacune dans une **session Kilo dédiée** — jamais les deux à la suite dans la même conversation. L'accumulation (documents × personas × standards × rapports) est précisément ce qui sature la fenêtre et la limite de sortie.
- La résolution (egm-resolve, tâches 4 et 6) reste dans le contexte principal de chaque cycle (rôle Elephant) : c'est le seul endroit où les findings s'appliquent aux fichiers.

## Tâches (dans l'ordre — 1 et 2 terminées le 2026-11-25)

1. **[Fait 2026-11-25] Valider les 2 décisions en attente** de `goldfish_review_report.md` — tranchées en gates interactifs : A2 → Option 1 « Tolérance transitoire explicite », B10 → OUI :
   - **A2** (fenêtre transitoire racine↔modules, Phases 3-6) — recommandation du rapport : Option 1 « Tolérance transitoire explicite » (imports dans les deux sens pour le code en migration, interdiction vérifiée en sortie de Phase 7 par décompte `git ls-files`, invariants absolus préservés) ;
   - **B10** (déplacement `invoice/` en Phase 8) — recommandation du rapport : OUI (`git mv invoice internal/modules/invoice`, aucun impact CI/Makefile).
2. **[Fait 2026-11-25] Générer et appliquer la « partie 2 »** : les 16 correctifs PROPOSE FIX générés depuis le tableau récapitulatif du `goldfish_review_report.md` + les décisions A2/B10 validées, puis appliqués en une passe aux deux fichiers scindés (rapport marqué `Status : resolved`). Cibles : les **deux** fichiers scindés (chaque finding peut toucher l'un ou l'autre — p.ex. A1/B1/B9→design, B2-B8/C2-C4→plan-phases, B9/C1→les deux) ; remapper les ancrages de lignes via le mapping noté dans l'en-tête de chaque fichier scindé. Les typos C1 (platforme, RésOLU, sécoupe, Bougeera, « ancienne monde »…) sont volontairement restées telles quelles pendant la scission : les corriger ici. Outil : egm-resolve (gates interactifs) ou un agent d'implémentation mandaté — ce sont des éditions de documents plan uniquement.
3. **[Fait 2026-11-25] Review de certification — design** : egm-review, TargetFile = `…-par-domaine-design.md`, Step = `feature_spec` (Stages 5→6→7→8 avec checkpoints) ; GoldfishContext : le document plan-phases (auto-ajouté par le lien) ; **artefacts dans un sous-répertoire dédié** (convention `1795597264251-design-review/`).
4. **[Fait 2026-11-25] Résoudre les findings design** — terminé dans la même session via deux passes egm-resolve (gate Stage 7 : interactive breakdown, 25/25 items acceptés ; gate Stage 8 : 5/5 fixés + re-audit de certification 0 nouveau finding). Rapports : `1795597264251-design-review/{zoom_out,comprehension,critic,goldfish,intermediate_zoom_out}_report.md` + `user_directives_2026-11-25.md` (4 directives architecturales utilisateur intégrées : interdiction `*pgdb.Queries` inter-modules + UoW `pgdb.WithinTx`, interfaces consumer-side `TokenService`/`TenantProvisioner`, soft-consume strict du refresh (STRING + état encodé, grâce 5 min, politique de rejeu tranchée, rollout/rollback borné), dette `binding.Validator` tracée). Le design est certifié (re-audit Stage 8 : 0 nouveau finding).
5. **[Fait 2026-11-26] Review de certification — plan** : egm-review, TargetFile = `…-par-domaine-plan-phases.md`, Step = `implementation` (Step 3 confirme DetailedRange = tout le doc via l'en-tête `Detailed Implementation Plan` ; Stage 9 : implementation_readiness + codebase_details + style) ; GoldfishContext : le document design (obligatoire — la cartographie §1-§4 et les invariants sont requis par codebase_details) ; artefacts dans `1795597264251-phases-review/`.
6. **[Fait 2026-11-26] Résoudre les findings** (gate Stage 9 → egm-resolve), puis le plan est prêt pour exécution (Phase 0 d'abord). — Rapports : `1795597264251-phases-review/{style,implementation_readiness,codebase_details}_findings.md` + `consolidated_report.md` ; résolution : `goldfish_review_report.md` (`Status: resolved`, auto accept — 22 PROPOSE FIX appliqués, 1 DECLINE intégré, 0 CONTEST/ESCALATE) ; plan-phases marqué certifié.

## Nettoyage optionnel (hors périmètre immédiat)

- La racine de `.kilo/plans/` mélange des rapports de plusieurs cycles (p.ex. `consolidated_review_report.md` vise `1795428465642-login-simple.md`, `zoom_out_report.md`/`goldfish_review_report.md` visent le plan scindé) : archiver chaque lot dans son sous-répertoire `<id>-review/` pour éviter que les prochains cycles les écrasent.
- Supprimer le stub de redirection une fois les reviews terminées si les références historiques n'importent plus.

## Hors périmètre

- Modification des personas/standards du repo kilo (verbosité des rapports) : seulement si la scission s'avère encore insuffisante.
- Mises à jour itératives du plan pendant l'exécution des phases : egm-lite (convention AGENTS.md) ; egm-review réservé aux jalons.
- Aucune montée de version de dépendance pendant la migration (déjà actée dans le plan).
