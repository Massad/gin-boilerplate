# Goldfish Review Report

**Target:** `.kilo/plans/1795265473956-frontend-stack-design-draft.md` (DetailedRange, lignes 234-318)
**Stage:** 9 — Implementation Readiness Test
**Standards:** `implementation-readiness-standard.md` (egm-implementation-readiness, egm-codebase-details) ; aucun standard (egm-style)
**GoldfishContext:** `1795265473956-concepts-brief-frontend-stack.md`
**Vérification:** chaque constat vérifié contre les fichiers réels du dépôt `go-artisan-services`.

## Reviewer Status

| Reviewer | Status |
| --- | --- |
| `egm-implementation-readiness` | completed |
| `egm-codebase-details` | completed |
| `egm-style` | completed |

## Verdict Implementation Readiness

Le plan **ne passe pas** le test d'aptitude à l'implémentation (règle zéro-décision). État de référence vérifié du dépôt : la couche auth cookie backend décrite aux items 1-11 est **déjà entièrement implémentée et testée** (`models/auth.go:19-74` constantes + cookies, `models/auth.go:365-394` `AdminAuth` par cookie, `controllers/user.go:68-74,135`, `controllers/auth.go:31-35,102-105`, `controllers/admin.go:55-56,142`, `main.go:76-214` servage SPA, tests `controllers/auth_test.go:96,140,148-181`, `middleware/auth_test.go:214,243`, `tests/article_test.go:144,364,789,834`). Le travail réellement restant : frontend `/app` (items 12-32) et intégration monorepo (items 33-38). Or le plan mélange « déjà fait » et « à faire » et contient des cibles invalides et des contrats erronés.

## Priorized Findings (Blockers -> Major -> Nits)

### Items 1-11 (couche auth cookies backend) déjà implémentés et testés — plan obsolète sur ce périmètre - Severity: 🔴 Blocker
* **Reviewer Lens:** `egm-implementation-readiness` + `egm-codebase-details`
* **Affected Scope:** lignes 240-267 (items 1-11), 236 (ordre d'exécution), 304-305 (vérification backend)
* **Discrepancy Description:** Le plan présente comme « modifications à apporter » : constantes/helpers cookies, login posant le cookie et répondant `{access_token}` sans `refresh_token`, refresh lu dans le cookie avec rotation, admin login/logout cookies + réponse `{message}`, `AdminAuth` par cookie, servage SPA + `/assets` dans `main.go`, mise à jour des 3 fichiers de tests, régénération swagger. Tout est déjà en place (références exactes ci-dessus, y compris assertions cookies dans les tests et annotations swagger `AccessTokenResponse`/`MessageResponse`/`UserLoginResponse` déjà présentes). Un exécutant réappliquerait des changements existants (régressions, conflits). Le « Type/Scope » de l'item 1 cite des constantes `refresh_cookie`/`admin_cookie` inexistantes (réalité : `RefreshCookieName`/`AdminCookieName`, `models/auth.go:21,25`) et un `Secure` conditionné à `ENV=PRODUCTION` seul (réalité : `ENV=PRODUCTION` **ou** `SSL=TRUE`, `models/auth.go:36-38`). L'item 7 cite « lignes 173-195 » de `main.go` (réalité : 76-214).
* **Rationale & Recommendation:** Marquer les items 1-11 « déjà implémenté — vérification uniquement » (`go build`/`go vet`/`make test`) et retirer du périmètre de modification ; recentrer le plan sur items 12-38 ; corriger constantes, condition `Secure` et références de lignes.

### Item 6 invalide : cible `./forms/auth.go` inexistante ; struct `Token` encore utilisé - Severity: 🔴 Blocker
* **Reviewer Lens:** `egm-implementation-readiness` + `egm-codebase-details`
* **Affected Scope:** lignes 258-259
* **Discrepancy Description:** `forms/auth.go` n'existe pas (`forms/` = validator.go, errors.go, admin.go, user.go, article.go). Le struct `Token` avec `refresh_token` vit dans `models/auth.go:114-118` — pas dans `forms/` — et est encore retourné/utilisé par `UserModel.Login` (`models/user.go:67`). « N'a plus d'usage » est faux ; la cible de suppression est fausse.
* **Rationale & Recommendation:** Supprimer l'item 6, ou le recadrer sur `models/auth.go` après avoir prouvé l'absence d'appelant (ce qui n'est pas le cas aujourd'hui).

### Contrat articles erroné et contradictoire pour `types.ts` - Severity: 🟡 Major
* **Reviewer Lens:** `egm-codebase-details` + `egm-style`
* **Affected Scope:** ligne 280 (item 21)
* **Discrepancy Description:** Le plan prétend des formes `{id int64, message}` et listes `articles` « à figer d'après les contrôleurs ». Réalité (`controllers/article.go`) : `Create:49` renvoie `{"message", "id": uuid.String()}` — id **string** (UUID), pas int64 ; `All:72` clé `"results"` ; `One:103` clé `"data"` ; Update/Delete `{"message"}`. Le plan annonce en outre un alignement « sur le contrat post-modification backend » puis le déclare non figé dans la même phrase.
* **Rationale & Recommendation:** Figer dans le plan les formes réelles (id string UUID, clés `results`/`data`/`message`) comme contrat de référence de `types.ts`, avec un gate du Verification Plan avant le build front.

### Vérification front inapplicable en l'état : outillage et répertoires absents - Severity: 🟡 Major
* **Reviewer Lens:** `egm-implementation-readiness`
* **Affected Scope:** lignes 305-306
* **Discrepancy Description:** `pnpm --dir app …`, `tsc --noEmit`, `make e2e_spa` référencent `/app`, pnpm, vitest, playwright et des cibles Makefile qui n'existent pas encore ; aucune étape d'initialisation (scaffolding, installation deps, création de cibles) ni version Node fixée n'est prévue avant ces vérifications.
* **Rationale & Recommendation:** Ajouter une étape 0 d'initialisation frontend (items 12-18) préalable au plan de vérification front, avec version Node/pnpm fixée.

### `README.md` (section Authentication) contredit le code livré - Severity: 🟡 Major
* **Reviewer Lens:** `egm-codebase-details`
* **Affected Scope:** ligne 299 (item 37) ; `README.md:220-236`
* **Discrepancy Description:** Le prévoit seulement d'ajouter une section front au README, mais la section Authentication existante décrit l'ancien comportement (« login returns access_token and refresh_token », refresh par token en body, « both tokens stored in Redis ») — faux depuis le pivot cookie.
* **Rationale & Recommendation:** Étendre l'item 37 : mettre à jour la section Authentication du README (cookies httpOnly, refresh via cookie) en plus d'ajouter la section front.

### Items d'intégration : couverture de mise à jour incomplète (memory-bank, AGENTS.md, docs) - Severity: 🟡 Major
* **Reviewer Lens:** `egm-codebase-details`
* **Affected Scope:** lignes 267 (item 11), 298-300 (items 36-38)
* **Discrepancy Description:** `memory-bank/projectbrief.md:17` et `activeContext.md:19` listent encore « SPA PWA (repo séparé) » et le cookie httpOnly hors périmètre ; `AGENTS.md:19` ne connaît pas les commandes front ni `/app` ; `docs/` swagger semble déjà synchronisé (annotations présentes) sans que l'item 11 précise s'il reste quelque chose à régénérer.
* **Rationale & Recommendation:** Étendre les items 36-38 aux fichiers exacts à corriger (projectbrief, activeContext, AGENTS.md, section Authentication du README) et qualifier l'état réel de `docs/` (à jour ou à régénérer).

### Étape de vérification sécurité ambiguë (« cookie visible httpOnly ») - Severity: 🟡 Major
* **Reviewer Lens:** `egm-style`
* **Affected Scope:** lignes 291 (item 32), 307
* **Discrepancy Description:** Formulation contradictoire : un cookie httpOnly n'est pas lisible du JS ; sa présence dans DevTools ne prouve pas l'attribut. Ni onglet, ni attributs (`httpOnly`, `Secure`, `Path`), ni contrôle `localStorage` distinctement spécifiés.
* **Rationale & Recommendation:** Réécrire l'étape : « Application > Cookies : cookies `refresh` (`httpOnly`, `Secure` en prod, `Path=/v1/token`) et `admin` (`httpOnly`, `Path=/v1/admin`) présents ; aucune clé `token*` dans `localStorage` ».

### Allégation non calibrée (« garantit le smoke test ») - Severity: 🟡 Major
* **Reviewer Lens:** `egm-style`
* **Affected Scope:** ligne 297 (item 35)
* **Discrepancy Description:** « le fallback template garantit le smoke test sans Node » : certitude non démontrée, non vérifiable depuis le document.
* **Rationale & Recommendation:** Reformuler (« permet l'exécution du smoke test Go sans build Node ») et déplacer la vérification dans le Verification Plan avec condition explicite (`os.Stat` sur `app/dist`).

### Nits (regroupés)

* **Ligne 244** — champ « Type / Scope : session/tokens Go » vide d'information (`egm-style`).
* **Ligne 236** — « Chaque phase se termine par ses vérifications » tautologique avec la section qui suit (`egm-style`).
* **Ligne 293-300** — aucun statut par item (déjà fait / à faire) pour distinguer le périmètre restant (`egm-implementation-readiness`).
* **Lignes 267** — préciser l'état de synchronisation de `docs/` (`egm-codebase-details`).
* **Lignes 279, 290** — « single-flight » jargon non défini : « N appels 401 concurrents → un seul refresh » (`egm-style`).
* **Ligne 250+** — franglais non naturalisé (« repose le cookie rotated », « mort-vivant », « proxy serve-side », « shell SPA », « toaster », « mirroir », « RHF »… (`egm-style`).
* **Ligne 287** — « non-invasive » subjectif ; question PII (`tenant_id` dans Sentry) non traitée (`egm-style`).
* **Ligne 288** — code 406 sans définition ni origine (violation d'unicité backend → 406/409) (`egm-style`).
* **Lignes 304, 306** — métaphores CI (« verts », « de bout en bout ») à remplacer par des critères précis (`egm-style`).
