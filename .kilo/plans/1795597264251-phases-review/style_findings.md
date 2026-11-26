# Rapport Style Goldfish — `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`

**Mode** : REVIEW (document entier, lignes 1-168) · **Persona** : `egm-style` · **GoldfishContext** : design `1795597264251-architecture-modulaire-par-domaine-design.md` (consulté pour vérifier la calibration des caveats).

## Synthèse

Le document est dense, factuel et exécutable ; le fond technique n'est pas remis en cause. Les problèmes relevés sont de **forme** :

1. **Duplication** — le même contenu est dit 2 à 3 fois (fenêtre A2 ×3, carve-out soft-consume ×2, dette `binding.Validator` ×3, helpers `getUserID`/`httpx` ×2-3). Chaque doublon est une future source de dérive à chaque édition du plan.
2. **Mur de texte Phase 3** — le soft-consume (ligne 26) est un paragraphe unique d'environ 600 mots qui ré-encode la politique déjà fixée dans le design (décision 7), alors que le plan lui-même affirme « politique fixée dans le design ».
3. **Caveats globalement bien calibrés** — le rollback (l.69), la liste des filets (l.67) et « sans éditer la liste attendue » (l.163) nomment des décisions concrètes : `KEEP`. Trois formulations du paragraphe Phase 3 méritent `CALIBRATE`.

**Compte** : Blocker 0 · Major 7 · Nit 8 · Caveats : KEEP 5 (non reportés), CALIBRATE 3, DROP/RELOCATE 0.

**Verdict** : conforme pour exécution après déduplication ; aucun bloquant de lisibilité aussi sévère que le paragraphe Phase 3, qui seul justifie un passage majeur avant réexécution de la review.

---

## Findings

### F1 — Bandeaux d'introduction (l.3-4)
- **Sévérité** : Major
- **Périmètre** : blocs citation en tête de document (lignes 3-4)
- **Justification** : deux lignes de méta-légende (procédure de scission, numérotation des rapports de review antérieurs, étapes egm-review) adressées au cycle de review, pas à l'ingénieur qui exécute le plan. Redites presque à l'identique dans l'en-tête du design (« aucun contenu perdu », même liste de rapports). Pour un lecteur neuf, l'essentiel — où se trouve le design — est noyé dans la mécanique de scission.
- **Réécriture proposée** :
  > « Ce document est le plan d'exécution Phases 0-8 ; le design (objectif, décisions actées, invariants, layout cible) vit dans [1795597264251-…-design.md](…-design.md). »
- (**Éditorial** : déplacer la note « ancrages de lignes des rapports antérieurs = numérotation de l'original » dans le dossier `1795597264251-*-review/` — elle ne concerne pas les exécutants. **RELOCATE**.)

### F2 — Phase 3, paragraphe soft-consume (l.26)
- **Sévérité** : Major
- **Périmètre** : « Phase 3 — Module `auth` », puce Soft-consume (lignes 24-31, surtout l.26)
- **Justification** : paragraphe unique de ~600 mots mélangeant mécanisme, format, TTL, politique de rejeu, rollout/rollback, sweep, tests exigés, observabilité et carve-out, avec gras et parenthèses imbriquées. La politique de rejeu est **dupliquée du design (décision 7)** alors que la puce affirme deux fois « politique fixée dans le design » — soit le renvoi suffit, soit la recopie crée un risque de désynchronisation. Le carve-out final (« Carve-out de contrat assumé… ») est ensuite redit intégralement ligne 67.
- **Réécriture proposée** (structure, contenu inchangé) :
  > « **Soft-consume strict du refresh (design décision 7 — politique tranchée dans le design, non recopiée ici)** :
  > - Record : chaîne STRING inchangée adressée par l'UUID du refresh ; état encodé dans la valeur (actif `userID:tenantID:role`, consommé `consumed:<delivered|recovered>:<value>`) ; aucune migration de schéma.
  > - TTL : grâce fixe 5 min (min(grâce, TTL restant)) au lieu de DEL ; atomicité Lua et ownership inchangés.
  > - Rollout : rollback pré-Phase-3 borné aux consommations en grâce (re-login requis) ; éviction TTL primaire, sweep ~5 min d'observabilité.
  > - Tests exigés : rejeu / grâce / recovery / fails-closed / sweep / unicité de la recovery.
  > - Carve-out : seul le chemin « rejeu pendant grâce » diffère (cf. Risques transverses). »

### F3 — Répétition de la fenêtre transitoire A2 (l.28, l.68, l.167)
- **Sévérité** : Major
- **Périmètre** : Phase 3 puce « Tolérance transitoire » (l.28) ; « Risques transverses » puce « Fenêtre de câblage » (l.68) ; « Critères de sortie » dernier point (l.167)
- **Justification** : la même proposition (« imports racine↔modules tolérés dans les deux sens, Fenêtre refermée en Phase 7, vérifiée par décompte `git ls-files`, platform/cmd-e2edb n'importent jamais modules ») est répétée trois fois quasi verbatim, avec chaque fois la même parenthèse « (décision A2 — tolérance transitoire explicite ; définition complète et impact : §4 du design) ».
- **Réécriture proposée** : version complète une seule fois (Phase 3, l.28) ; l.68 devient « couverte par le critère de sortie (voir Phase 3, tolérance transitoire) » ; l.167 conserve uniquement le décompte opérationnel (la ligne 167 est déjà quasi la version complète — garder l.167, réduire l.28 et l.68, au choix, mais une seule occurrence complète).

### F4 — Carve-out soft-conume redit dans « Risques transverses » (l.67 ↔ l.26)
- **Sévérité** : Major
- **Périmètre** : « Risques transverses et rollback », premier point (l.67)
- **Justification** : la ligne 67 re-décrit le carve-out (« un jeton non délivré retrouvé pendant la grâce bénéficie de la recovery unique… ») déjà encodé ligne 26 et dans le design. Répétition synonymique inter-sections exacte.
- **Réécriture proposée** :
  > « **Contrat HTTP/SPA** : filets = test contrat P0.3, assertions `tests/` inchangées, smoke CI à chaque branche, `make e2e` aux Phases 1 et 7. **Unique carve-out acté** : soft-consume du refresh, Phase 3 (chemin « rejeu pendant grâce » — politique et tests dédiés au § Phase 3). »

### F5 — Duplication helpers contextuels + `httpx` différé (l.13, l.18, l.21 ↔ l.134-136)
- **Sévérité** : Major
- **Périmètre** : Phase 1 objet (l.13), Phase 2 Sources et Risques (l.18, l.21), étape P1.6 (l.134-136)
- **Justification** : la justification « getUserID/getTenantID restent locaux, clés posées par TokenAuth qui documente leur contrat, platform n'en duplique pas la source de vérité » est recopiée presque mot pour mot en l.21 et l.136. Le report de `httpx` en Phase 2 est dit trois fois (l.13 « httpx différé en Phase 2 », l.18 « premier appelant réel — reporté depuis P1.6 », l.134 « création différée en Phase 2 ») avec les deux justifications (« pas de code mort » / « premier appelant réel »).
- **Réécriture proposée** : P1.6 porte la décision et sa justification une fois (« `httpx` créé en Phase 2, au premier appelant réel — pas de code mort sans appelant ») ; l.13 et l.18 se contentent de « httpx : voir P1.6 ». La note sur les clés de contexte vit dans P1.6 ; l.21 se limite à « pas dans `httpx` (voir P1.6) ».

### F6 — Dette `binding.Validator` dite trois fois (l.60, l.131, l.140)
- **Sévérité** : Major
- **Périmètre** : Phase 8 (l.60), P1.5 (l.131), P1.7 « Dettes consignées (a) » (l.140)
- **Justification** : le même déplacement futur de `binding.Validator` (vers entrypoint ou `platform/validate`, élimination en sortie de migration) est décrit trois fois, dont deux avec la même mention « élimination requise à la sortie de migration ». Chaque occurrence risque de diverger indépendamment.
- **Réécriture proposée** : P1.7(a) est le point canonique (« dette consignée : relocalisation en sortie de migration, suivi Phase 8 ») ; l.131 se limite au comportement (« passage à `validate.DefaultValidator`, même comportement ») ; l.60 devient « traiter la dette P1.7(a) (voir critère de sortie) » ou renvoie simplement à P1.7(a).

### F7 — Clarté du risque slug Phase 4 (l.36)
- **Sévérité** : Major
- **Périmètre** : « Phase 4 », puce Risques (l.36)
- **Justification** : « l'échec d'épuisement des 99 tentatives garde le contrat actuel épinglé pendant la migration (erreur discriminée + code HTTP spécifiés post-migration dans le `service_test.go` pur) » juxtapose deux moments (pendant / post) sans marquer la frontière — un lecteur neuf ne sait pas quel comportement est garantit quand. Jargon comprimé « mécanique `TestAPISurface` » et caveat vague « plus commentaire de rappel à la maintenance » (quel commentaire, où ?).
- **Réécriture proposée** :
  > « Pendant la migration : conserver le comportement actuel d'échec (erreur discriminée + code HTTP actuels), déjà épinglés par la suite `tests/`. Après la migration : écrire la spécification cible dans le `service_test.go` pur. La liste `reservedSlugs` contient des racines de routes : le test contrat P0.3 échoue si une racine réelle manque, à compléter d'un commentaire pointant vers la liste. »

### F8 — Jugement subjectif « le plus complexe » (l.51)
- **Sévérité** : Nit
- **Périmètre** : titre « Phase 7 » (l.51)
- **Justification** : « (le plus complexe, en dernier) » est un jugement non mesurable ; l'information utile est l'ordre et la nature collage final.
- **Réécriture proposée** : « Phase 7 — Module `auth` : collage final (dépend des phases 3-6) ».

### F9 — Subjectifs « acceptable », « vrai » (l.76, l.140, l.143)
- **Sévérité** : Nit
- **Périmètre** : discipline commune (l.76 « la compilation verte est le vrai contrôle »), P1.7 (l.140 « acceptable en transitionnel » ; l.143 « épingle désormais le vrai câblage »)
- **Justification** : « vrai » et « acceptable » sont desQualifier subjectifs ; la règle ou le statut suffit.
- **Réécriture proposée** : l.76 « la compilation verte est le contrôle fait foi » → « la compilation verte fait foi » ; l.140 « toléré en transition, élimination requise en Phase 8 » (sans « acceptable ») ; l.143 « épingle le câblage de production ».

### F10 — « idem » ambigu (l.25)
- **Sévérité** : Nit
- **Périmètre** : Phase 3 Sources (l.25)
- **Justification** : « (logout user 200 "Successfully logged out", admin idem) » — `idem` ne précise pas si le message textuel, le code, ou les deux sont identiques ; un pin de contrat doit être exact.
- **Réécriture proposée** : « logout user et admin : 200 + `{"message": "Successfully logged out"}` » (à confronter au code réel avant application).

### F11 — Instruction vague Phase 8 (l.59)
- **Sévérité** : Nit
- **Périmètre** : « Phase 8 », premier point (l.59)
- **Justification** : « traiter la suppression non commitée `docs/test.md` » ne nomme pas la décision : supprimer le fichier local ? ne rien faire car non commité ?
- **Réécriture proposée** : « supprimer le fichier local non commité `docs/test.md` (s'il existe) ».

### F12 — Parenthèse redondante « (preuve de non-régression) » (l.22)
- **Sévérité** : Nit
- **Périmètre** : Phase 2 Validation (l.22)
- **Justification** : « suite `tests/` verte sans modification des assertions (preuve de non-régression) » — la parenthèse paraphrase la condition qui la précède ; « preuve » ajoute une promesse que la parenthèse ne renforce pas.
- **Réécriture proposée** : « suite `tests/` verte sans modification des assertions ».

### F13 — Déclaration de validité des file:line trop locale (l.76)
- **Sévérité** : Nit
- **Périmètre** : discipline commune des « Étapes détaillées » (l.76) ; refs `file:line` en Phases 2-7 (l.21 `controllers/user.go:19-33`, l.42 `models/user.go:122`, l.46 `models/admin.go:55-64`, l.55 `models/login.go:34-38`)
- **Justification** : l'avertissement « les numéros de ligne… décrivent l'état initial… deviennent obsolètes dès la première étape » est placé dans la section « Étapes détaillées — Phase 0 et Phase 1 uniquement » alors que les Phases 2-7 utilisent aussi des refs `file:line`.
- (**Éditorial : **RELOCATE** — remonter la phrase au niveau du titre « Detailed Implementation Plan » ou de la discipline commune couvrant l'ensemble des phases.)

### F14 — Inflation du gras et des parenthèses (transverse)
- **Sévérité** : Nit
- **Périmètre** : document entier (notamment l.26, l.28, l.164, l.167)
- **Justification** : gras sur des tiers de chaque puce, parenthèses imbriquées triples (p.ex. l.28 « (internal/app construit) », « (tolérance transitoire explicite ; définition complète…§4) »). L'emphasise perd sa fonction quand tout est en gras ; les parenthèses multi-clauses seraient plus lisibles en sous-puces ou en phrases.
- **Réécriture proposée** : réserver le gras aux invariants et aux carve-outs ; convertir les parenthèses >1 clause en sous-puces.

---

## Caveats — calibration

| Ligne | Formulation | Verdict | Action |
|---|---|---|---|
| l.26 | « rollback pré-Phase-3 borné aux consommations en grâce (re-login équivalent à une consommation normale) » | promesse d'équivalence UX non prouvée ici | **CALIBRATE** → « rollback pré-Phase-3 possible ; les sessions consommées en grâce sont perdues : l'utilisateur doit se re-connecter » |
| l.26 | « L'incident de session reste observable côté support (journalisé au marquage et à la récupération) » | « observable côté support » ne nomme aucune décision du support | **CALIBRATE** → « journalisé à la consommation et à la récupération (index `user_sessions:*`) — observabilité sweep » |
| l.26 | « purge défensive des reliques » | adjectif d'auto-justification + métaphore imprécise | **CALIBRATE** → « suppression des records consommés expirés que le TTL n'a pas encore évacués » |
| l.67 | « couverte par le test contrat P0.3, assertions `tests/` inchangées, smoke CI, `make e2e` Phases 1 et 7 » | backed par des filets nommés | KEEP |
| l.69 | Rollback (« branch par phase… revert jusqu'au dernier commit vert puis re-diagnostic ») | décisions concrètes de l'opérateur | KEEP |
| l.98 | « Ce test survit à tous les déplacements (les annotations suivent les handlers) » | mécanisme énoncé (lit le JSON commité) | KEEP |
| l.163 | « sans éditer la liste attendue » + exception « déplacement mécanique de helpers » | règle opérateur explicite, exception bornée | KEEP |
| l.167 | « 0 en sortie de Phase 7 » via `git ls-files` | critère mesurable | KEEP |

---

## Tableau récapitulatif

| # | Sévérité | Périmètre | Objet | Action |
|---|---|---|---|---|
| F1 | Major | l.3-4 | bandeaux méta redondants avec le design | RELOCATE + compression |
| F2 | Major | l.24-31 | mur de soft-consume, duplique design décision 7 | réécrire en sous-puces |
| F3 | Major | l.28/68/167 | fenêtre A2 ×3 | dédupliquer (1 occurrence complète) |
| F4 | Major | l.67 ↔ l.26 | carve-out redit | compression + renvoi |
| F5 | Major | l.13/18/21/134-136 | helpers + report httpx ×2-3 | canon unique (P1.6) |
| F6 | Major | l.60/131/140 | dette binding.Validator ×3 | canon unique (P1.7a) |
| F7 | Major | l.36 | phrase slug confuse + caveats vagues | réécriture |
| F8 | Nit | l.51 | « le plus complexe » | reformuler factuel |
| F9 | Nit | l.76/140/143 | « vrai », « acceptable » | reformuler |
| F10 | Nit | l.25 | « idem » ambigu | préciser contrat |
| F11 | Nit | l.59 | « traiter » vague | verbe décisionnel |
| F12 | Nit | l.22 | parenthèse paraphrase | DROP parenthèse |
| F13 | Nit | l.76 vs Phases 2-7 | disclaimer file:line mal scopé | RELOCATE |
| F14 | Nit | transverse | gras/parenthèses | règle d'emphase |

**Total : Blocker 0 · Major 7 · Nit 8.** Aucune modification de fond proposée.
