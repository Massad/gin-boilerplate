# Goldfish Annotated Review Report

Status: resolved (2026-11-21 — auto accept : 19 PROPOSE FIX appliqués ; ESCALATE M6 tranché par l'auteur : Sentry conservé en v1 ; ESCALATE M5 tranché par l'auteur : servage depuis `app/dist` confirmé, fallback démo conservé si dist absent ; CONTEST M8, M10-titres, N7 documentés ; DECLINE enregistrés)

**Target:** `.kilo/plans/1795265473956-frontend-stack-design-draft.md` (full file, lignes 1-318)
**Source findings:** `1795265473956-consolidated_review_report.md` (27 tâches consolidées, 49 constats bruts, 5 stages)
**Catégories:** PROPOSE FIX = 19 · CONTEST = 3 · ESCALATE = 2 · DECLINE = 3 (2 sous-parts regroupées : M10 = contest + fix, N10 = 2 fixes + decline)

## Prioritized Findings & AI Suggestions

### B1 — Couche auth cookies backend déjà livrée ; document présenté comme « à faire » [`egm-zoom-out`, `egm-implementation-readiness`, `egm-codebase-details`] - Severity: Blocker
* **Affected Scope:** `(#L13)`, `(#L40)`, `(#L102)`, `(#L108-L117)`, `(#L133-L137)`, `(#L236)`, `(#L240-L267)`, `(#L304)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Vérifié dans le dépôt : cookies (`models/auth.go:19-74,365-394`), refresh par cookie (`controllers/auth.go:31-35,102-105`), login/logout tenant+admin (`controllers/user.go:68-74,135`, `controllers/admin.go:55-56,142`), servage SPA + `/assets` (`main.go:76-214`), tests d'assertions cookies (`tests/article_test.go:144,364,789,834`, `controllers/auth_test.go:96-181`, `middleware/auth_test.go:214,243`). Constantes réelles `RefreshCookieName`/`AdminCookieName`, `Secure` via `ENV=PRODUCTION` **ou** `SSL=TRUE`.
* **Rationale:** Le plan décrit un état « avant » qui n'existe plus : un exécutant réappliquerait du travail fait (régressions) ; la règle zéro-décision du standard exige que le plan reflète le réel.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Recalage du document sur l'existant : marquer « LIVRÉ » + vérification uniquement ; corriger au passage constantes, condition `Secure` et références de lignes erronées.
* ***Proposed Edits:***
    * Location: L13 (fin) — Change: «…sur l'API `/v1` existante ; la couche backend visée par le pivot cookie est **déjà livrée** (couche auth cookies — vérification uniquement, voir plus bas) ; aucune autre modification backend n'est requise en v1. » (remplace «…existante, sans modification du backend en v1.»)
    * Location: L13 (milieu) — Change: «en chunk chargé à la demande (lazy)» à la place de «en chunk lazy» (joint à M11).
    * Location: L40 — Change: «Toute modification de l'API métier en v1 : la **couche auth cookies** (`login`/`refresh`/`logout`, `AdminAuth`, tests) est livrée et testée ; les autres gaps restent documentés pour v1.5 (…)»
    * Location: L102 — Change: «La SPA consomme les routes `/v1` existantes (articles, public). Les contrats auth ont été modifiés par la couche cookies **livrée au 2026-11-21** (login/refresh/logout, admin login/logout, `AdminAuth` par cookie — voir « Couche auth cookies backend — livrée »). Les manques backend restent documentés en v1.5 (`GET /user/me`, câblage RBAC, invitations).» (supprime « cookie refresh » des manques v1.5)
    * Location: L108 (titre) — Change: «### Couche auth cookies backend — livrée (constat d'existant, vérification uniquement)»
    * Location: L110 — Change: «(LIVRÉ) `POST /v1/user/login` : pose le cookie refresh (`httpOnly`, `Secure` si `ENV=PRODUCTION` ou `SSL=TRUE`, `SameSite=Strict`, `Path=/v1/token`) ; réponse `{message, user, tenant, token: {access_token}}` — `refresh_token` absent du JSON.»
    * Location: L111 — Change: «(LIVRÉ) …réponse = `{access_token}` + cookie ré-émis (nouveau refresh).»
    * Location: L112, L113 — Change: préfixe «(LIVRÉ)» ; L113 ajoute «admin login répond `{message}`» + «(fallback header Authorization)».
    * Location: L116 — Change: «(LIVRÉ) Tests backend : … assertions … déjà en place.»
    * Location: L133 (titre) — Change: «### Servage — en place dans `main.go` (lignes 76-214)»
    * Location: L135, L136 — Change: préfixe «(LIVRÉ)» ; L135 «…servent `./app/dist/index.html` s'il existe (`os.Stat`), avec `Cache-Control: no-cache`, sinon fallback template démo.»
    * Location: L236 — Change: «Ordre d'exécution recommandé : la couche auth backend est **livrée** (items 1-11 : vérification uniquement), puis le scaffolding `/app` (items 12-18), les features front (items 19-32), enfin l'intégration CI et les docs (items 33-38).»
    * Location: L240 — Change: «#### Backend — couche auth cookies (LIVRÉE — vérification uniquement)»
    * Location: L242-244 (item 1) — Change: «(LIVRÉ) … en place : `RefreshCookieName`/`AdminCookieName` (`models/auth.go:21,25`), `sessionCookie` (httpOnly, `SameSite=Strict`, `Path` scopé, `Secure` via `cookieSecure()` = `ENV=PRODUCTION` **ou** `SSL=TRUE`, `models/auth.go:36-51`), `SetRefreshCookie`/`ClearRefreshCookie`/`SetAdminCookie`/`ClearAdminCookie` (`models/auth.go:55-74`), `ExtractAdminTokenMetadata` lit le cookie (`models/auth.go:365-394`).»
    * Location: L246-247 (item 2) — Change: «(LIVRÉ) … `Login` pose le cookie refresh (`user.go:68`) et répond `{message, user, tenant, token: {access_token}}` via `models.LoginToken` (`user.go:69-74`) ; `Logout` retire le cookie (`user.go:135`).» (corrige «struct `TokenDetails`»)
    * Location: L249-250 (item 3) — Change: «(LIVRÉ) … `Refresh` lit le refresh **dans le cookie** (`auth.go:31-35`), … répond `{access_token}` (`auth.go:102-105`) et ré-émet le cookie (`auth.go:104»).»
    * Location: L252-253 (item 4) — Change: «(LIVRÉ) … `Login` pose le cookie admin et répond `{message}` (`admin.go:55-56`) ; `Logout` retire le cookie (`admin.go:142`).»
    * Location: L255-256 (item 5) — Change: «(LIVRÉ) … `AdminAuth` lit le token depuis le cookie (`models/auth.go:365-394`)…»
    * Location: L261-262 (item 7) — Change: «(LIVRÉ) … en place (`main.go:76-214`) : `serveSPAIndex` (`os.Stat`, fallback template démo), servage `/assets` (`max-age=31536000, immutable`), `r.GET("/")`, `NoRoute` (JSON 404 `/v1`/non-GET), `LoadHTMLGlob` et `r.Static("/public")` conservés.»
    * Location: L264-267 (items 8-11) — Change: «(LIVRÉ)» avec références de tests réelles (`article_test.go:144,364,789,834`, `auth_test.go:96,140,148-181`, `middleware/auth_test.go:214,243`) ; item 11 : «(`docs/` déjà à jour — annotations `AccessTokenResponse`/`MessageResponse`/`UserLoginResponse` en place ; `make generate_docs` + `git diff` uniquement si annotations modifiées)».
    * Location: L304 — Change: préfixe «(état livré — vérification)» sur la ligne Backend du Verification Plan.

### B2 — Item 6 invalide : `./forms/auth.go` inexistant ; struct `Token` encore utilisé [`egm-implementation-readiness`, `egm-codebase-details`] - Severity: Blocker
* **Affected Scope:** `(#L258-L259)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** `forms/` = validator.go, errors.go, admin.go, user.go, article.go (pas de `auth.go`). Le struct `Token` (tag `refresh_token`) vit dans `models/auth.go:114-118` et reste retourné par `UserModel.Login` (`models/user.go:67`) : « n'a plus d'usage » est faux.
* **Rationale:** Cible de suppression inexistante ; exécuter l'item échouerait, supprimer `models.Token` casserait le login.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Remplacer l'item par un constat factuel sans action.
* ***Proposed Edit:*** Location: L258-259 — Change: «6. ~~`./forms/auth.go`~~ (aucune action) — **cible corrigée** : `forms/` ne contient pas de `auth.go` ; le struct `Token` (refresh en body) vit dans `models/auth.go:114-118` et reste retourné par `UserModel.Login` (`models/user.go:67`) : **aucune suppression**.»

### M1 — Contradiction « sans modification backend / aucune modification API » vs contrats auth altérés [`egm-inventive-catalyst`] - Severity: Major
* **Affected Scope:** `(#L13)`, `(#L40)`, `(#L102)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Le document affirme « sans modification du backend en v1 » et « Aucune modification de l'API /v1 en v1 » tout en modifiant les contrats login/refresh/logout/admin (L109-114).
* **Rationale:** Changements de contrat casseurs présentés comme inexistants ; un lecteur peut croire l'API intangible.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Mêmes lignes que B1 : les modifications M1 (L13, L40, L102) résolvent la contradiction en nommant les endpoints modifiés et leurs contrats livrés.
* ***Proposed Edit:*** Location: L13, L40, L102 — Change: fusionné avec les éditions B1 correspondantes.

### M2 — Contrat API erroné et contradictoire pour `types.ts` [`egm-codebase-details`, `egm-style`] - Severity: Major
* **Affected Scope:** `(#L280)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Le plan prétend `{id int64, message}` / listes `articles` ; réalité (`controllers/article.go`) : create `{message, id}` avec id **string** UUID, listes clé `results`, détail clé `data`, update/delete `{message}`. La phrase « contrat post-modification » contredit « formes à figer d'après les contrôleurs ».
* **Rationale:** `types.ts` serait construit sur des formes fausses et non opposables.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Figer le contrat réel (livré) dans le plan comme référence de `types.ts`.
* ***Proposed Edit:*** Location: L280 — Change: «types alignés sur le **contrat backend livré** (figé ici d'après les contrôleurs actuels) : login `{message, user, tenant, token: {access_token}}`, register `{user, tenant}`, refresh `{access_token}`, articles create `{message, id}` (id **string** UUID), listes `{results}` / détail `{data}`, update/delete `{message}`, admin login/logout `{message}` ; tenants list/statut : figer depuis `controllers/admin.go».»

### M3 — CSRF : modes de défaillance de SameSite non documentés [`egm-inventive-catalyst`] - Severity: Major
* **Affected Scope:** `(#L114)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** `SameSite` est un comportement navigateur (dégradation possible en Lax) ; un XSS ne vole pas le token mais peut émettre des requêtes avec le cookie attaché. Le document affirme « couvert » sans état du résiduel.
* **Rationale:** Un lecteur de ce doc d'authentification a besoin du résiduel pour juger la posture ; la décision (pas de jeton dédié) reste celle de l'auteur.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Documenter le résiduel assumé sans changer la décision (Alternative 13).
* ***Proposed Edit:*** Location: L114 — Change: «CSRF : `SameSite=Strict` + scoping `Path` (le cookie refresh n'est envoyé qu'à `/v1/token/refresh`, le cookie admin qu'aux routes `/v1/admin`) ; aucun jeton CSRF requis ; le middleware CORS existant reste en same-origin. Résiduel connu et assumé en v1 : dégradation navigateur possible en Lax, et un XSS peut émettre des requêtes avec le cookie attaché (sans vol du token) — fenêtre bornée par expiration (7 j / 8 h) ; Alternative 13 écarte le jeton double-submit ; une vérification d'`Origin` sur les mutateurs reste une option v1.5.»

### M4 — Rejet du 100 % httpOnly : complexité client déplacée non assumée [`egm-inventive-catalyst`] - Severity: Major
* **Affected Scope:** `(#L199-L202)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Alternative 11 rejetée « portée backend disproportionnée » sans état du contrecoût : l'hybride crée la boucle 401→refresh (single-flight), le retry ×1, la purge de session (CUJ 2/5).
* **Rationale:** La justification doit présenter le trade-off total, pas seulement le côté backend.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Une phrase assumant le déplacement de complexité clôt le gap sans rouvrir la décision.
* ***Proposed Edit:*** Location: L202 (fin de la rationale) — Change: ajouter « Trade-off assumé explicitement : l'hybride déplace la complexité vers le client (boucle 401→refresh single-flight, retry ×1, purge de session, CUJ 2 et 5) — l'empreinte backend minimale prime en v1. »

### M5 — Fallback silencieux vers le template démo en production [`egm-inventive-catalyst`] - Severity: Major
* **Affected Scope:** `(#L135)`, `(#L230)`, `(#L262)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Si `app/dist` est absent en production, le produit sert silencieusement les gabarits de démo. Le fallback est délibéré (smoke test CI sans Node), mais sa posture en PRODUCTION n'est pas tranchée.
* **Rationale:** Décision d'auteur : garder le fallback (simplicité opérationnelle) ou fail-closed en prod (404) — impacte du code livré et le smoke test CI.
* ***AI Suggestion:*** ESCALATE
* ***Justification:* ** Résolution hors portée du texte : demande une décision (et un éventuel changement backend/CI). Question à l'auteur : que fait `main.go` en production quand `app/dist` manque — fallback démo conservé, ou 404 ?
Il faut ajouter app/dist

### M6 — Sentry SaaS imposé en v1 (YAGNI) [`egm-zoom-out`] - Severity: Major
* **Affected Scope:** `(#L46)`, `(#L104-L106)`, `(#L287)`, `(#L297)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Dépendance tierce + secret CI + étape build pour un MVP solo ; bénéfice non prouvé en v1 (Error Boundaries + logs suffiraient).
* **Rationale:** Décision enregistrée (« SaaS type Sentry », L46) mais re-questionnée par la critique YAGNI : garder ou décaler.
* ***AI Suggestion:*** ESCALATE
* ***Justification:*** Le texte peut supprimer Sentry, mais la décision appartient à l'auteur (elle avait été tranchée gate-par-gate). Question : Sentry reste en v1 ou passe en v1.5 (Error Boundaries + logs seuls en v1) ?
On peut décaler sentry pour le moment

### M7 — e2e sur backend réel sans isolation de données [`egm-inventive-catalyst`] - Severity: Major
* **Affected Scope:** `(#L143)`, `(#L291-L292)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Aucune isolation par run ni garantie de nettoyage ; un échec pollue PostgreSQL/Redis et déstabilise les runs suivants.
* **Rationale:** Les e2e créent tenans/articles/tenants — sans purge, la base de test se dégrade.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Spécifier l'isolation par run (tenant dédié + purge teardown).
* ***Proposed Edits:***
    * Location: L143 (CI, après « cache des navigateurs ») — Change: ajouter « ; e2e isolés par run : tenant dédié au run (suffixe unique) purgé en teardown (succès comme échec) ».
    * Location: L291 (item 32) — Change: ajouter en fin « ; isolation : tenant dédié par run (suffixe unique) + purge en teardown ».

### M8 — Chute d'altitude du plan d'implémentation (détails code/champs) [`egm-zoom-out`] - Severity: Major
* **Affected Scope:** `(#L234-L300)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** « Détails d'implémentation (champs, chemins exacts) au lieu d'interfaces/contrats » ; la liste fichiers/champs devrait sortir du design.
* **Rationale:** Selon le reviewer, ces détails contreviennent au contrôle d'altitude.
* ***AI Suggestion:*** CONTEST
* ***Justification:*** Preuve lisible : le format `egm-draft` prescrit précisément cette granularité. `skills/egm-draft/SKILL.md` (Step 5, Draft Implementation) : « Draft a detailed implementation plan… add enough detail so that with the plan alone you can implement it without having to make any impactful decisions », « Enumerate Files: List EVERY file that will be changed, created, or deleted », « NO CODE ». Et `skills/egm-draft/references/design_doc.md` : « ### Target Files to Create/Modify … Do not use vague language or TBDs », placée **après** les sections de haut niveau. La granularité critiquée est le format de l'artefact, validé en downstream par les mêmes templates. La part valide du constat (formes de contrat figées, L280) est traitée par M2.

### M9 — Risques du servage statique + alternative proxy/CDN absente [`egm-design`] - Severity: Major
* **Affected Scope:** `(#L226-L232)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Aucune posture de déploiement documentée (binaire unique, pas de proxy/CDN dans le dépôt) ; alternative NGINX/CDN non pesée.
* **Rationale:** Un lecteur a besoin du modèle de déploiement et de la condition de réévaluation ; benchmarks et architecture multi-réplicas hors périmètre v1 (solo, un seul binaire).
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Un bullet NFR enregistre l'hypothèse de déploiement et la condition de réévaluation — le nécessaire pour un lecteur, sans sur-spécifier.
* ***Proposed Edit:*** Location: après L230 (NFR) — Change: insérer « * **Déploiement** : binaire unique servant l'API + les fichiers statiques ; pas de proxy inverse ni CDN dans l'infra visée en v1 (aucune config de déploiement dans le dépôt). Charge statique assumée pour le trafic visé ; alternative (proxy/CDN dédié pour décharger le Go) réévaluée si la charge ou la réplication l'exige. » (l'ajout d'une Alternative NGINX/CDN complète est jugé disproportionné : l'hypothèse tient en un bullet)

### M10 — Langue des titres ; « Open Questions » mal labellisée [`egm-style`] - Severity: Major
* **Affected Scope:** `(#L42)`, titres L49/78/147/226, « Reason for Rejection » L152-222 (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Titres et libellés en anglais dans un document français ; la section « Open Questions » ne pose aucune question (décisions déjà résolues).
* **Rationale:** Titres anglais : incohérents avec le corpus. Open Questions : trompeur.
* ***AI Suggestion:*** CONTEST (titres) + PROPOSE FIX (Open Questions)
* ***Justification:*** (a) Titres : preuve lisible — `skills/egm-draft/references/design_doc.md` prescrit les intitulés exacts « ## Overview », « ## The Problem », « ## Goals », « ## Open Questions », « ## Critical User Journeys (CUJs) », « ## Alternatives Considered », « ## Non-Functional Requirements », « ## Detailed Implementation Plan » et la rubrique « **Reason for Rejection** » ; renommer casserait l'audit structurel EGM. (b) En revanche l'intitulé interne trompeur se corrige sans toucher au heading.
* ***Proposed Edit:*** Location: L44 — Change: «Résolues lors du draft :» → «Décisions résolues pendant la rédaction (conservées pour mémoire ; à retirer à la finalisation) :»

### M11 — Jargon non défini à la première occurrence [`egm-style`] - Severity: Major
* **Affected Scope:** `(#L13)`, `(#L59)`, `(#L111)`, `(#L250)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** « chunk lazy », « single-flight », « cookie rotated », « repose le cookie rotated » sans définition ni équivalent français.
* **Rationale:** Termes pivot du design token ; définis une fois, ils levèrent l'ambiguïté pour un lecteur front ou non-front.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Définition ciblée aux termes pivot seulement (pas de sweep global) ; fusionne avec les éditions B1 sur les mêmes lignes (L13, L111, L250).
* ***Proposed Edits:***
    * Location: L13 — Change: «en chunk chargé à la demande (lazy)» (fusionné dans B1).
    * Location: L59 — Change: «(rotation atomique single-flight, nouvel access en mémoire, cookie rotated)» → «(rotation atomique single-flight : N requêtes 401 concurrentes déclenchent un seul refresh ; nouvel access en mémoire, cookie ré-émis)».
    * Location: L111, L250 — Change: «cookie rotated»/«repose le cookie rotated» → «ré-émet le cookie (nouveau refresh)» (fusionné dans B1).

### M12 — Décision httpOnly répétée à l'identique (Goals L26 vs L47/L228) [`egm-style`] - Severity: Major
* **Affected Scope:** `(#L26)`, `(#L47)`, `(#L228)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Le même énoncé décisionnel apparaît presque mot pour mot dans Goals, Open Questions et NFR (CUJ 4 en récit est acceptable) — risque d'incohérence à la mise à jour.
* **Rationale:** Source de vérité unique pour la décision, ailleurs un renvoi.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Garder la décision complète en L47 (résolu) et la norme en L228 (NFR) ; Goals renvoie. CUJ 4 conservé (récit de scénario).
* ***Proposed Edit:*** Location: L26 — Change: remplacer la puce par « * Aucun token de durée longue n'est accessible au JS (détail de la décision : L47 ; norme : NFR L228). La session survit au reload sans re-login (refresh silencieux via cookie). »

### M13 — Allégations non calibrées [`egm-style`] - Severity: Major
* **Affected Scope:** `(#L206)`, `(#L297)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** « la plus grande fenêtre XSS restante » (L206, Bloqueur en Stage 7) : superlatif non mesuré ; « garantit le smoke test » (L297) : certitude non démontrée.
* **Rationale:** Un doc d'auth ne doit porter que des affirmations vérifiables.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Remplacer par des faits mesurables.
* ***Proposed Edits:***
    * Location: L206 — Change: «…lisible par le JS — la fenêtre d'exposition XSS la plus longue du design (access tenant borné à 15 min en mémoire).»
    * Location: L297 — Change: «le fallback template permet d'exécuter le smoke test Go sans build Node».

### M14 — README/memory-bank/docs désalignés avec l'état livré [`egm-codebase-details`] - Severity: Major
* **Affected Scope:** `(#L267)`, `(#L298-L300)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** La section Authentication du README (L220-236) décrit encore tokens JSON + refresh en body + « stored in Redis » ; item 37 ne prévoit que d'ajouter une section front ; item 11 ne dit pas si `docs/` est déjà sync.
* **Rationale:** L'outil d'intégration (items 36-38) doit couvrir les fichiers réellement faux.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Étendre item 37 (README Authentication) ; item 11 qualifié ; item 36 déjà correct.
* ***Proposed Edits:***
    * Location: L267 (item 11) — Change: inclus dans B1 (état `docs/` précisé).
    * Location: L299 (item 37) — Change: ajouter « ; **corriger aussi la section Authentication (README L220-236)** qui décrit encore des tokens en JSON (refresh en body, `stored in Redis`) ».

### M15 — Verification Plan : outillage front absent, étape sécurité ambiguë [`egm-implementation-readiness`, `egm-style`] - Severity: Major
* **Affected Scope:** `(#L271)`, `(#L305)`, `(#L307)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Les vérifications front référencent `/app`, pnpm, vitest, playwright qui n'existent pas encore — pas d'étape 0 ni de versions fixées. L307 « cookie visible httpOnly » est contradictoire (présence ≠ attribut).
* **Rationale:** Le plan de vérification doit être exécutable sans ambiguïté.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Étape 0 explicite + versions fixées + critères observables pour la vérification sécurité.
* ***Proposed Edits:***
    * Location: L271 (item 12, fin) — Change: ajouter « ; versions fixées : Node LTS + `packageManager` pnpm (Corepack) dans `package.json` ».
    * Location: L305 — Change: préfixer « **Front — étape 0 (préalable)** : items 12-18 réalisés (scaffolding), versions fixées. Puis : » devant la liste des commandes.
    * Location: L307 — Change: «…Application > Cookies : cookie `refresh` présent et `httpOnly` (`Secure` en prod, `Path=/v1/token`), cookie `admin` présent et `httpOnly` (`Path=/v1/admin`) ; aucune clé de token dans `localStorage` ; suspension d'un tenant via la console admin → CUJ 3 vérifié dans l'onglet tenant.»

### N1 — Claims JWT décodés utilisés pour le contexte UI [`egm-inventive-catalyst`] - Severity: Nit
* **Affected Scope:** `(#L74)`, `(#L126)`, `(#L281)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Tension entre « client agit sur des claims non vérifiés » et « le serveur reste seul juge ».
* **Rationale:** Usage purement affichage, non autorisation ; borné par validation serveur.
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Le document le dit déjà (L126 : « le serveur reste seul juge de validité ») ; le renforcement proposé (tirer le contexte après refresh) changerait le bootstrap pour un gain marginal. Constat enregistré pour la mise en œuvre.

### N2 — Stack front en versions récentes sans justification [`egm-zoom-out`] - Severity: Nit
* **Affected Scope:** `(#L84-L88)`, `(#L271-L272)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** React 19, Vite, RR v7, TanStack v5, Tailwind v4 récents sans filet.
* **Rationale:** Choix auteur (stack alignée sur ses compétences, L13).
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Décision d'auteur enregistrée ; les lockfiles (item 12) verrouillent exactement les versions — le filet demandé existe dans le plan.

### N3 — Références de ligne backend approximatives [`egm-zoom-out`] - Severity: Nit
* **Affected Scope:** `(#L53)`, `(#L64)`, `(#L133)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** `user.go:100` (réponse register), `auth.go:88-91`, `user.go:57-59`, `main.go 173-195` — approximatifs.
* **Rationale:** La traçabilité en souffre ; les références symboliques sont robustes.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Passer aux références symboliques ; L133 corrigé sous B1.
* ***Proposed Edits:***
    * Location: L53 — Change: «(`controllers/user.go:100`)» → «(handler `Register`, `controllers/user.go`)».
    * Location: L64 — Change: «(401, `controllers/auth.go:88-91`)» → «(401 — tenant suspendu rejeté dans `Refresh`, `controllers/auth.go`)» ; «(`controllers/user.go:57-59`)» → «(`Login`, `controllers/user.go`)».

### N4 — Alternative « Sentry auto-hébergé » annoncée mais absente [`egm-inventive-catalyst`] - Severity: Nit
* **Affected Scope:** `(#L106)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** L106 promet une documentation dans Alternatives Considered — aucune des 15 alternatives ne la traite.
* **Rationale:** Allégation documentaire fausse.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Corriger la phrase (les options sont écartées en une ligne, sans entrée Alternative complète — disproportion).
* ***Proposed Edit:*** Location: L106 (fin) — Change: «Les options self-hosted et différé sont documentées dans Alternatives Considered.» → «Les options self-hosted (infrastructure à opérer) et différée (métier sans observabilité au lancement) ont été écartées pour v1.»

### N5 — Risque de l'access admin 8 h non quantifié [`egm-design`] - Severity: Nit
* **Affected Scope:** `(#L113-L114)`, `(#L226-L227)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Fenêtre 8 h acceptée sans quantification ni détection d'anomalie.
* **Rationale:** Posture httpOnly + SameSite retenue par le pivot ; 401 → re-login documenté (CUJ 4).
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Acceptation du risque implicite dans le pivot httpOnly (L202, L207) ; la détection d'anomalie relève du backend v1.5/v2. Constat enregistré.

### N6 — Résilience de la rotation Redis non attestée [`egm-design`] - Severity: Nit
* **Affected Scope:** `(#L110-L111)`, `(#L226-L227)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Comportement en cas d'échec de refresh/Redis non spécifié côté doc.
* **Rationale:** CUJ 2 définit déjà le comportement front : refresh échoué → purge + redirection login.
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Le chemin d'échec existe (CUJ 2, L59) ; enregistrer que tout échec Redis → 401 → login est suffisant pour v1 solo.

### N7 — Intitulé « Non-Functional Requirements » hors vocabulaire du standard [`egm-structural`] - Severity: Nit
* **Affected Scope:** `(#L226)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Le contenu (sécurité, cache, rollout) correspond à « Risks & Mitigations ».
* **Rationale:** Alignement terminologique avec le standard feature-spec.
* ***AI Suggestion:*** CONTEST
* ***Justification:*** Preuve lisible : `skills/egm-draft/references/design_doc.md` L104 prescrit exactement « ## Non-Functional Requirements » comme section du template EGM ; le renommage casserait le template et l'audit structurel downstream. Le contenu risque/mitigation y est bien traité.

### N8 — Em dash / encodage (mojibake possible) [`egm-structural`] - Severity: Nit
* **Affected Scope:** `(#L1)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Dash ém — selon l'encodage.
* **Rationale:** Fichier UTF-8 ; le template EGM lui-même en emploie.
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Environnement UTF-8 confirmé ; pas d'action.

### N9 — Mise en gras décorative [`egm-style`] - Severity: Nit
* **Affected Scope:** `(#L19-125)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** Gras appliqué à des termes ordinaires.
* **Rationale:** Préférence rédactionnelle.
* ***AI Suggestion:*** DECLINE
* ***Justification:*** Le gras signale les pivots du design (httpOnly, mémoire, contrats) ; dé-grasser tout est du churn sans lecture fonctionnelle.

### N10 — Remplissage, fautes et codes non documentés (regroupé) [`egm-style`, `egm-implementation-readiness`] - Severity: Nit
* **Affected Scope:** `(#L7)`, `(#L236)`, `(#L243)`, `(#L244)`, `(#L287)`, `(#L288)`, `(#L304)`, `(#L306)` (.kilo/plans/1795265473956-frontend-stack-design-draft.md)
* **Discrepancy Description:** « ce harnes » (typo, L7) ; 406 non documenté (L288) ; « non-invasive » et question PII tenant_id dans Sentry (L287) ; métaphores « verts » (L304/306) ; phrase tautologique L236 ; champ Type/Scope maigre (L243/244) ; statuts par item absents.
* **Rationale:** Mix : fautes et codes = corrigeables ; le reste = préférences ou couvert ailleurs.
* ***AI Suggestion:*** PROPOSE FIX (2 edits) + DECLINE (rest)
* ***Justification:*** « harnes » = typo objective ; 406 = information fonctionnelle. Les autres : L236/L243/L244/statuts résolus par les éditions B1 ; « verts »/« non-invasive »/PII = enregistrés (la question PII `tenant_id`-dans-Sentry reste ouverte pour la mise en œuvre).
* ***Proposed Edits:***
    * Location: L7 — Change: «En mode plan de ce harnes,» → «En mode plan de ce harness,».
    * Location: L288 — Change: «mapping 401/403/406/5xx» → «mapping 401/403/406 (violation d'unicité côté backend)/5xx».
    * DECLINE (enregistré) : L287 « non-invasive » + question PII ; L304/306 métaphores CI ; L236 tautologie (supprimée de fait par l'édition B1) ; L243-244 Type/Scope maigre (prescrit par le template ; précisé dans l'édition B1 de l'item 1) ; statuts par item (apportés par les marqueurs « (LIVRÉ) » de B1).

---

## Résumé des décisions (Step 2)

| Catégorie | Nombre | Items |
| --- | --- | --- |
| PROPOSE FIX | 19 | B1, B2, M1, M2, M3, M4, M7, M9, M10(Open Questions), M11, M12, M13, M14, M15, N3, N4, N10(×2) |
| CONTEST | 3 | M8, M10(titres), N7 — preuves : `skills/egm-draft/SKILL.md` (Step 5) + `skills/egm-draft/references/design_doc.md` (L9-104, L139-145, L161) |
| ESCALATE | 2 | M5 (fallback démo en prod), M6 (Sentry v1 ou v1.5) |
| DECLINE | 3 (+parts) | N1, N2, N5, N6, N8, N9, N10(reste) |
