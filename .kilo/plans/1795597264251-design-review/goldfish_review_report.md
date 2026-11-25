# Goldfish Annotated Review Report

**Target:** `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-design.md` (full file, 158 lines) — résolution des findings du Stage 7 (`critic_report.md`), round 1
**Artefacts liés:** `user_directives_2026-11-25.md` (directives architecturales actées au gate Stage 5)

## Prioritized Findings & AI Suggestions

### [R1] décision 7 — stockage soft-consume contradictoire [`egm-design`, id design-1-1] — Severity: Blocker
*   **Affected Scope:** l.37 (décision 7) ; rép. plan-phases Phase 3
*   **Discrepancy Description:** `consumed_at` nullable sur « l'enregistrement de session » implique une colonne durable (PostgreSQL) alors que les sessions vivent en Redis (`user_sessions:*`, Lua) ; aucune migration goose planifiée ; TTL de grâce en plage (5-10 min) au lieu d'un choix.
*   **Rationale:** le seul changement comportemental de la migration est inconstructible tel qu'écrit ; Redis n'a pas d'« annulable ».
*   ***AI Suggestion:*** PROPOSE FIX (fusion avec design-1-4 : cleanup job)
*   ***Justification:*** critique exacte — Elephant : le mécanisme cible est Redis-hash, pas SQL ; la plage TTL et le vocabulaire SQL viennent de la directive utilisateur, traduites littéralement.
*   ***Proposed Edit:***
    *   Location: décision 7 (list item), réécriture complète
    *   Change: spécifier le mécanisme Redis : le record de session reste le hash `user_sessions:*` — soft-consume = opérations du script Lua remplacées par marquage atomique (`consumed_at` + `delivered` champs du hash, hash sha256 du refresh token stocké) + TTL réduit à la grâce **fixée à 5 min** (min(grâce, TTL restant)) au lieu de DEL immédiat ; **aucune migration de schéma** (aucun enregistrement déplacé vers PostgreSQL) ; atomicité Lua inchangée (une seule opération atomique par consommation) ; ownership inchangé (hash comparé avant consommation). Cleanup job : balayage en-process du module `auth` (ticker, cadence ~5 min) supprimant les records consommés dont la grâce est expirée (TTL Redis = éviction primaire, sweep = filet + observabilité) ; invariant de course : le sweep ne touche jamais un record dont la grâce est active et jamais un record non consommé.

### [R2] décision 7 — politique de rejeu différée à l'exécution [`egm-design` design-1-2 + `egm-structural` structural-1-4 + style N8] — Severity: Blocker
*   **Affected Scope:** l.37, l.154 (et puce P3 du plan)
*   **Discrepancy Description:** « choix figé en exécution » pour un choix sécuritaire critique : re-émission pendant la grâce laisserait un token volé rejoué obtenir une paire fraîche après rotation légitime (indistinguable d'une réessai honnête) ; 401 préserve le signal mais casse la reprise réseau.
*   **Rationale:** la norme exige que les implications de sécurité soient tranchées dans le design, pas déléguées.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** critique exacte — Elephant : le discrimineur honnête existe (le champ `delivered` du marquage) et permet une politique sûre ET récupérable.
*   ***Proposed Edit:*** (intégré à la réécriture R1)
    *   Change: politique tranchée dans le design : (i) rejeu d'un jeton **déjà délivré** (`delivered=1`) → `401`, même forme de réponse que le chemin invalide actuel (byte-exact sur ce chemin) ; (ii) rejeu d'un jeton marqué consommé **non délivré** (`delivered=0` : l'application a échoué avant émission de la nouvelle paire) pendant la grâce → **une seule re-émission de récupération** (même forme 200 que le happy path, flag `recovered` interdit une seconde), sinon `401` ; (iii) crash serveur entre consommation et lecture → échec fermé (`delivered` inconnu = non récupérable, 401). Analyse de menace écrite : un token volé rejoué pendant la grâce n'obtient une paire que si la première tentative légitime n'a rien délivré ; après délivrance légitime, rejeu → 401 (grâce fixe 5 min pour borner l'exposition). Mises à jour en cascade : §6 (reliquat « figé en exécution » → tranché), puce P3 du plan-phases (politique copiée), critère « sous tests dédiés » conservé (rejeu/grâce/recovery/fails-closed).

### [R3] objectif l.18 non qualifié [`egm-design`, id design-1-3] — Severity: Major
*   **Affected Scope:** Objectifs, l.18
*   **Discrepancy Description:** « Contrat HTTP inchangé au byte près pendant toute la migration » contradit le carve-out décision 7 non mentionné ici.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.18. Change: « Contrat HTTP inchangé au byte près pendant toute la migration (sauf le carve-out acté de la décision 7 — chemin « rejeu pendant grâce » du soft-consume) ; … »

### [R4] cleanup job non spécifié [`egm-design`, id design-1-4] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX — résolu par la réécriture R1 (owner in-process module auth, cadence ~5 min, invariant de course écrit). Pas d'édition séparée.

### [R5] justification « make e2e Phases 1 et 7 » non démontrée [`egm-design`, id design-1-5] — Severity: Major
*   **Affected Scope:** §3 item 9, l.123
*   **Discrepancy Description:** assertion (« les autres phases ne modifient que des branches de câblage ») contredite par la Phase 3 (rebranchement de tout le middleware chain) ; l'argument réel (suite tests/ = fillet contrat de bout en bout via app.NewRouter) n'est pas énoncé.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.123, phrase « `make e2e` est rejoué… ». Change: expliciter l'argument : « `make e2e` est rejoué aux Phases 1 et 7 — les deux seules phases qui changent l'entrypoint (`main.go` → `cmd/api`) et l'assemblage du routeur, donc la couche que seul le cycle Playwright observe (processus réel + SPA servis) ; toute régression de câblage des autres phases (middleware rebranchés en Phase 2-3 inclus) se manifeste comme une violation du contrat HTTP, couverte de bout en bout par la suite `tests/` qui pilote `app.NewRouter` (vrais handlers, middlewares, réponses) et par le smoke CI à chaque branche. »

### [R6] section « Risks & Mitigations » absente en en-tête [`egm-structural`, id structural-1-1] — Severity: Blocker
*   **Affected Scope:** structure du document (l.149→l.151)
*   **Discrepancy Description:** la norme exige une rubrique risques identifiable par en-tête (failure modes, security, scalability, mitigations) ; le contenu existe mais dispersé sous en-têtes trompeurs ; aucune analyse scalability.
*   ***AI Suggestion:*** PROPOSE FIX (fusion part avec design-1-7)
*   ***Proposed Edit:*** Location: nouvelle section `## 5. Risques et atténuations` insérée avant (ex-)Questions ouvertes. Change (design-level, sans dupliquer l'opérationnel du frère) :
    - Carve-out de contrat (décision 7) : le seul chemin déviant, sa politique tranchée (R2) et son invariant d'ownership ; tests dédiés exigés.
    - Risque de la fenêtre A2 : imports croisés transitoires — atténué par « platform ne jamais modules », décompte `git ls-files` = 0 en Phase 7 (critère de sortie).
    - Scalabilité/statu quo : la migration ne modifie ni les pools (2), ni les TTL Redis, ni la croissance multi-tenant ; soft-consume borne les records consommés par TTL de grâce (5 min) + sweep ; le Unit of Work réutilise la mécanique Tx existante (`Store.InTx` + `WithTenant`) sans nouvelle contention.
    - Détection de régression : test contrat P0.3, suite `tests/` inchangée, smoke CI par branche (couverte §3.9/§3.10).
    - Rollback/exécution : reporté explicitement à la rubrique « Risques transverses et rollback » du plan-phases (lien).

### [R7] « Questions ouvertes » contredit le contenu [`egm-structural` structural-1-2 + `egm-style` style-1-3 (M3)] — Severity: Major
*   **Affected Scope:** l.151-158
*   **Discrepancy Description:** 5/6 items tranchés/différés ; l.156=l.24 et l.158=l.23 verbatim (Non-objectifs) ; l.154 contient une entrée close.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: section recréée en `## 6. Décisions et reports` (après la nouvelle section 5) : une intro (« arbitrages tranchés et reports hors migration ; l'inventaire complet des non-objectifs reste en Préambule »), puis : (1) montées de versions → follow-up post-Phase 8 ; (2) soft-consume : tranché 2026-11-25, politique R2 appliquée — **aucun reliquat ouvert restant** ; (3) une seule puce « les reports restants (casse des clés de réponse, CORS configurable, refresh httpOnly, câblage RBAC) sont listés dans Non-objectifs ». Les 4 puces dupliquées (l.155-l.158) sont supprimées.

### [R8] prose cassée à l'ancre A2 [`egm-structural` structural-1-3 + `egm-style` N7] — Severity: Major
*   **Affected Scope:** l.149 (parenthèse A2)
*   **Discrepancy Description:** fragment télégraphique « Impact architectural ici) : » + virgule de fusion dans « Contexte : …migraient de façon décalée, une interdiction… » ; article superflu « le 2026-11-25 ».
*   ***AI Suggestion:*** PROPOSE FIX (artefact d'édition du resolve Stage 6 — corrigé en réécrivant)
*   ***Proposed Edit:*** Location: l.149. Change: « **Fenêtre transitoire Phases 3-6 — décision A2** (« A2 » est l'identifiant de l'arbitrage Blocker porté par `goldfish_review_report.md`, racine de `.kilo/plans/` — review antérieure du plan original, tranché le 2026-11-25 en gate interactif, Option 1 « Tolérance transitoire explicite ») : une interdiction immédiate des imports croisés aurait imposé de déplacer chaque paquet et chacun de ses importeurs dans un seul mouvement ; les répertoires producteurs et consommateurs migrent de façon décalée. Impact : les imports racine↔modules sont tolérés **dans les deux sens**, uniquement pour le code en cours de migration ; la fenêtre est refermée en Phase 7 (suppression des répertoires racine) et l'interdiction vérifiée par le critère de sortie (décompte `git ls-files` = 0). » — fin de parenthèse déplacée pour une phrase à verbe.

### [R9] provenance « (validation utilisateur) » sans ancre [`egm-structural`, id structural-1-5] — Severity: Major
*   **Affected Scope:** titre de section l.29
*   **Discrepancy Description:** sept décisions revendiquent une validation utilisateur sans lien vérifiable (ni session, ni artefact) ; dates sans référence d'artefact (décisions 6/7/4 étendue).
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.29 (titre de section). Change: titre → « ### Décisions structurantes actées (validation utilisateur) » suivi d'une ligne d'ancrage : « > Provenance : validations recueillies interactivement pendant la rédaction (cycles egm-brainstorm du plan original) ; arbitrages A2/B10 des reviews antérieures : `goldfish_review_report.md` (racine de `.kilo/plans/`, Status: resolved) ; directives architecturales du 2026-11-25 : `1795597264251-design-review/user_directives_2026-11-25.md`. » Les dates 2026-11-25 inline (décisions 4/6/7) restent — cohérentes avec l'ancre.

### [R10] métadonnées EGM redondantes (l.4) [`egm-style`, id style-1-1 (M1)] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.4. Change: « > Review EGM — Elephant-Goldfish Model, protocole de review multi-agent. Cycle `feature_spec` d'egm-review : 5 zoom-out initial, 6 test de compréhension, 7 review critique, 8 zoom-out intermédiaire ; rapports et reports dans `1795597264251-design-review/`. Les ancrages de lignes des rapports antérieurs (`zoom_out_report.md`, `goldfish_review_report.md`, racine de `.kilo/plans/`) suivent la numérotation de l'original. » (répétition « Step `feature_spec` (Stages 5-8) » supprimée)

### [R11] « platform » vs « plateforme » [`egm-style`, id style-1-2 (M2)] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.44. Change: « le platform n'importe jamais les modules et ne doit pas dupliquer leur contrat de contexte » → « la plateforme n'importe jamais les modules et ne doit pas dupliquer leur contrat de contexte ». Canon fixé : « plateforme » en prose, `platform` réservé aux références de package (`(platform)` de la même liste accepté comme référence d'emplacement).

### [R12] point 10 dédoublé [`egm-style`, id style-1-4 (M4) + N9] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.124 réécrit en un passage sans répétition : couverture énumérée une fois (suite, matrice, propriété refresh, 7 tests middleware, CLI, garde e2edb, unitaires purs), puis trois statuts (« non couverte / épinglée par les tests / couverte par propriété d'ownership ») sans re-démonstration ; « cf. fin de point » supprimé ; terminologie byte-près canon conservée ; pointer soft-consume conservé.

### [R13] synonymes de préservation en rotation (§3.9) [`egm-style`, id style-1-5 (M5) + N6] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX (partie : réserves)
*   ***Proposed Edit:*** Location: l.123. Change: « **Trois réserves préservées telles quelles et à traiter après** » → « **Trois réserves, à traiter après la migration** » ; « (c) … est recopiée telle quelle » → « est recopiée à l'identique » ; « préservé tel quel (copie verbatim) » (titre du point) → « (recopié à l'identique) » ; « précisément » supprimé (dans la réécriture R5). Canon : « recopié à l'identique » pour l'action de migration, « inchangé au byte près » pour le contrat.

### [R14] cinq libellés pour byte-exactness [`egm-style`, id style-1-6 (M6)] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX (partie : l.37, l.124)
*   ***Proposed Edit:*** Location: l.37 + l.124. Change: « restent byte-exact » → « restent inchangés au byte près » (2 occurrences l.124) ; l.68 (plan) déjà « byte-exact » → inchangé (frère cohérent non requis ici) ; canon « inchangé au byte près ».

### [R15] N1 : « transvers » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.8. Change: « socle technique transvers » → « socle technique transversal ».

### [R16] N2 : accord « repositories » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.36. Change: « chaque module expose ses repository » → « chaque module expose ses repositories ».

### [R17] N3 : « que cartographiés » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.115. Change: « mêmes singletons (…) et vars de package `*Model` que cartographiés en §1 » → « identiques à ceux cartographiés en §1 ».

### [R18] N5 : « contrat HTTP doit rester vert » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.41. Change: « le contrat HTTP doit rester vert à chaque étape » → « le contrat HTTP doit rester inchangé à chaque étape ».

### [R19] N12 : « multiplie » / « n'auront qu'à » non calibrés [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.12 + l.111. Change: « chaque évolution (membres/invitations, billing) multiplie le câblage manuel et le risque de régression d'isolement » → « …ajoute par domaine du câblage manuel et un risque supplémentaire de régression d'isolement » ; « n'auront qu'à câbler `RequirePermission(...)` » → « se borneront à câbler `RequirePermission(...)` ».

### [R20] N13 : « en central » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.119. Change: « non implémentées en central » → « non centralisées ».

### [R21] N14 : « Interdit à la main » [`egm-style`, Nit] — PROPOSE FIX
*   Location: l.62. Change: « Interdit à la main (sauf pool.go) » → « Édition manuelle interdite (sauf pool.go) ».

### [R22] renvois plan-phases sans lien markdown [`egm-structural`, id structural-1-7 (partie)] — Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (les 2 renvois plan-phases uniquement)
*   ***Proposed Edit:*** Location: déc. 7 (l.37) + §3.9 (c) (l.123). Change: chemins backticks → liens markdown `[…-plan-phases.md](1795597264251-architecture-modulaire-par-domaine-plan-phases.md)`. Les autres références (l.66, l.121 : plan distinct cité une fois) restent en backticks — DECLINE sur ce résidu.

### [R23] lignes 117 et 260 non réclamées [`egm-structural`, id structural-1-8] — Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.3 (design) + l.3 (plan-phases). Change: append « ; 117 et 260 de l'original : lignes vides (aucun contenu perdu) ».

### [R24] titre §3 sans les zones de vigilance [`egm-structural`, id structural-1-9] — Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.113. Change: « ## 3. Couplages et points de friction » → « ## 3. Couplages, points de friction et zones de vigilance ».

### [R25] anatomie modèle du préambule [`egm-structural`, id structural-1-10] — Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: l.8. Change: « (handler / service / repository / dto / tests) » → « (handler / service / repository / dto / model / tests) ».

## DECLINE (record only)

- **[design-1-6] Renumbering vs ancrages historiques (l.3-4)** — accepté tel quel, cross-ref explicite déjà en en-tête.
- **[design-1-7] Risques délégués au frère (l.113-124)** — le manquement réel (risque décision 7 sans analyse) est résolu par R1/R2/R6 ; le reste délégation saine du pair.
- **[structural-1-6] Numérotation d'en-têtes asymétrique** — renumbering « Préambule » invasif pour des ancres externes (`§4 du design` référencé par le frère) ; cosmétique.
- **[structural-1-7, résidu] backticks pour les plans distincts cités une fois** — coût > bénéfice.
- **[style N4] wiring/câblage mélangés** — deux registres établis (anglicisme technique vs français) ; churn massif pour un gain minime.
- **[style N10]** — partiellement résolu par R7 (l.156 supprimée) ; résidu acceptable.
- **[style N11] répétition ×3 du constat smoke/wiring (l.12, l.67, l.120)** — l.67 = cellule d'inventaire (contexte légitime) ; l.12 porte la définition, §3.6 le constat.
- **[style N15] divers ajustements mineurs non listés item par item** — couverts en partie par les réécritures R5/R12/R13.
- **[style N6 « précisément »]** — appliqué via R5.
- **[style N8 « le retient »]** — résolu par R2 (politique tranchée).

## Status

Status: resolved (2026-11-25)
Résolution : **auto accept après interactive breakdown — 25/25 items acceptés** (Items 1-25 : R1-R25), 0 skip, 0 discuss, 0 contest, 0 escalate. 10 clusters DECLINE (record only) : design-1-6, design-1-7, structural-1-6, structural-1-7 (résidu backticks plans distincts), style N4, N10, N11, N15 — aucun ouvert restant. Éditions appliquées en un batch : design `1795597264251-architecture-modulaire-par-domaine-design.md` (158 → 167 lignes, sections 5 « Risques et atténuations » et 6 « Décisions et reports » ajoutées, décision 7 entièrement spécifiée : records `consumed_at`/`delivered`, grâce 5 min fixée, politique de rejeu tranchée, sweep in-process, invariant de course, analyse de menace) + plan-phases (puce P3 copie exacte de la politique, carve-out Risques transverses mis au canon, note lignes 117/260 en en-tête). Structure revérifiée post-application. Handoff → Stage 8 (Intermediate Zoom Out).

---

## Pass 2 — Stage 8 findings (Intermediate Zoom Out), 2026-11-25

**Target:** `1795597264251-architecture-modulaire-par-domaine-design.md` (post-pass-1, 167 lignes) — findings de `intermediate_zoom_out_report.md`.

### [R2-1] Claim de continuité décision 7 faux [`egm-zoom-out` Major] — Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX (preuves vérifiées à l'Elephant : `models/auth.go:123/127/262-268` — records STRING sous clé UUID, `user_sessions:*` = index SET, Lua `GET`+préfixe+`DEL`).
*   ***Proposed Edit:*** décision 7 réécrite : record = chaîne STRING existante adressée par l'UUID du refresh JWT (token jamais stocké, en clair ni haché — l'intention « hash du token » de la directive est satisfaite par l'adressage jti) ; `user_sessions:{userID}` = index de révocation, pas le record ; état encodé dans la valeur (actif inchangé → consommé `consumed:<delivered|recovered>:<value>`, TTL grâce 5 min) ; sha256 abandonné (sans rôle mécanique, doublonnait l'adressage) ; note Rollout/rollback ajoutée : records actifs au format inchangé → rollback pré-Phase-3 borné aux consommations en grâce (re-login équivalent à une consommation normale). Cascade : §3.10, puce P3 du plan-phases.

### [R2-2] Sweep = résidu YAGNI (TTL suffit) [`egm-zoom-out` Nit] — PROPOSE FIX
*   ***Proposed Edit:*** requalification : éviction TTL primaire et suffisante ; sweep = observabilité (compte+log) + purge défensive des reliques, jamais la dépendance de purge ; coût lifecycle assumé (start/stop lié au service, testable). Cascades : §5, décision 7, puce P3.

### [R2-3] Référent du flag `recovered` ambigu [`egm-zoom-out` Nit] — PROPOSE FIX
*   ***Proposed Edit:*** décision 7 (ii) + puce P3 : « état `recovered` encodé dans la valeur du record, jamais dans la réponse HTTP » — lève l'ambiguïté et exclut la contradiction avec « même forme 200 ».

### [R2-4] §5 « ni les TTL Redis » surexprimé [`egm-zoom-out` Nit] — PROPOSE FIX
*   ***Proposed Edit:*** §5 : « ni les TTL Redis des sessions actives » + « éviction par TTL primaire ».

### [R2-5] §6 : item backlog `err.Error()` manquant + déclinaisons non consignées [`egm-zoom-out` Nit] — PROPOSE FIX
*   ***Proposed Edit:*** §6 : bullet « Reports et dettes hors migration » complété (normalisation payloads d'erreur via `httpx` + erreurs typées, post-Phase 8 — Stage 5 Persona 2 F4) + bullet « Déclinaisons consignées » (extraction CORS immédiate déclinée au gate Stage 5 ; gabarits minces coût assumé).

**Status (pass 2): resolved (2026-11-25 — résolution directe au gate Stage 8, 5/5 fixés, aucun CONTEST/ESCALATE)**
