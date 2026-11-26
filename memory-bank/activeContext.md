# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

## Session 2026-11-26 — gate G-1 (certification plan-phases) PASS + lot L0 (filet de sécurité) PASS

**Cycle G-1** (egm-review + egm-resolve, session dédiée, personas chargés du repo kilo) : 3 personas livrés (`style_findings.md` F1-F14 déjà présent, `implementation_readiness_findings.md` F1-F7, `codebase_details_findings.md` F1-F9) ; consolidation dédupliquée `1795597264251-phases-review/consolidated_report.md` (Blocker 1 · Major 12 · Nit 12) ; passe egm-resolve `goldfish_review_report.md` **Status: resolved** (auto accept : 22 PROPOSE FIX appliqués au plan-phases, 0 CONTEST, 0 ESCALATE, 1 DECLINE intégré) — plan-phases **marqué certifié** 2026-11-26, split-review tâches 5-6 `[Fait]`. Résiduel lots l.46 corrigé (formulation dureable pré-flight L0, validé utilisateur). Restes ouverts : harmonisation optionnelle lots l.47 (« 9 git mv »), drift design l.73 (`docs/test.md`) réservé au drift-report de clôture L8.

**Lot L0** (branche `refactor/modular-phase-0` fourchue de `modular_archi` @ `07d81c4`) : P0.1 smoke `ci.yml` corrigé (commit utilisateur `30ae675` ; re-add périmé de `/v1/invoice*` issu d'une sauvegarde d'éditeur observée puis écarté) ; P0.2 baseline figée — **découverte environnement : la batterie Go/make s'exécute dans WSL** (`\\wsl.localhost` refuse les locks du toolchain Go côté Windows), `make test` vert ×3 (11/11 paquets `ok`, `artisan_test` recréée, index Redis 2), régénérations idempotentes (blocs docs/mocks identiques à HEAD) ; **correctif outillage `959f50e`** : `.mockery.yaml` requiert `version: 2` et le Makefile renommé `MOCKERY_CLI_VERSION`/`MOCKERY_CLI_BIN` (l'`export` global du Makefile exposait `MOCKERY_VERSION`, que mockery v2.53.8 lit comme override de la clé config `version` → `'version' cannot parse value as 'bool'` — jamais détecté : la CI n'exécute pas `generate_mocks`) ; P0.3 `tests/api_surface_test.go` — `TestAPISurface` (égalité bidirectionnelle exacte des 14 routes `/v1`, vert **sans éditer la liste attendue**, commit `67c9e77`) ; gate **PASS** : `.kilo/plans/1795597264251-impl-log/L0-gate.md` (CI distante à valider au push — action opérateur).

**Lot L1** (branche `refactor/modular-phase-1`, commit code `3dc7869`) : P1.1-P1.5 — `internal/platform/{config,redisdb,db,middleware,mocks,validate}` extraits (packages/exports inchangés, 11+7 sites d'imports) ; P1.7 — `internal/app.NewRouter(Deps{TemplateGlob, Env})` recopie verbatim du wiring+routes `/v1`+Swagger+SPA, `cmd/api/{main,main_test}.go` (fin du paquet racine), `SetupRouter` supprimé (38 appels → `testRouter()`, la suite épingle le câblage de production) ; équivalence routeur vérifiée par test transitoire (PASS, supprimé au commit — preuve au gate) ; P1.8 — Makefile/ci.yml/playwright sur `./cmd/api`, swag `-g cmd/api/main.go`, `git diff --exit-code internal/platform/db`, AGENTS.md chemins MAJ. **Dérive outillage corrigée** : `make generate_mocks` était un no-op silencieux (la clé de schéma `version: 2` se décode dans le bool du flag `--version` avec viper weak-typed → sémver+exit avant génération ; l'ajout L0 de la clé retiré, renommage make `MOCKERY_CLI_*` conservé) — génération DBTX/Querier réelle, idempotence ×3 vérifiée. Batterie : `make test` 11/11 `ok` (incl. `cmd/api`, `platform/{config,db,redisdb}`), docs régénérés blobs = HEAD, `make e2e` **4 passed** (webServer `./cmd/api`). Gate **PASS** `1795597264251-impl-log/L1-gate.md`. Prochaine étape : **L2** (contrôleurs → `internal/modules/article`, pilote du pattern module) — fourche `refactor/modular-phase-2` après push/merge du lot.

## Session 2026-11-23 — plan « vue de connexion simple » (`.kilo/plans/1795428465642-login-simple.md`) : backend + frontend + e2e verts, vérification finale close (passe de correction du drift incluse)

Implémentation complète : items 1-30 faits et vérifiés, item 31 = ce fichier. Arbre de travail **partiellement commité** (`3ac3aae` détrack `app/dist`, `57992de` job CI frontend + harnais e2e + memory bank + plan + drift report) ; les **corrections de drift** restent en working tree **non commité** (aucun commit demandé pour cette passe).

### Backend — VERT (sessions précédentes)

- Migration goose `000003_admin_name.sql` : `admin_user.name varchar NOT NULL DEFAULT 'Superadmin'` (table plateforme, sans RLS) ; `db/queries/admins.sql` aligné (`GetAdminByEmail` renvoie `name`, `CreateAdmin` upsert idempotent qui remet le nom à jour) ; code sqlc régénéré (`internal/db` commité).
- `models/login.go` — `UnifiedLogin` : tentative utilisateur tenant puis repli superadmin, réponse **discriminée par `kind`** (`tenant_user` → user + tenant + tokens access/refresh ; `superadmin` → admin id/email/name + token access). Email inconnu → bcrypt contre un hash factice (coût constant, pas d'énumération temporelle) ; tenant suspendu → 403 propagé ; politique assumée : **un email superadmin ne recouvre jamais un email utilisateur tenant** (utilisateur trouvé + mauvais mot de passe → 406 direct, sans tentative admin).
- `controllers/auth.go` — `AuthController.Login` (200 / 403 / 406, annotations Swagger) ; route `POST /v1/login` ; CLI `-create-admin` enrichi du flag `-name` (défaut « Superadmin », upsert idempotent).
- Tests : `tests/login_test.go` (suite tag `all`) — user OK + Redis, admin OK + `name`, 406 mauvais mot de passe, 406 email inconnu (assert du bcrypt factice), 403 tenant suspendu, 406 validation, endpoints legacy toujours 200 ; `make test` complet vert, `go vet` vert.

### Frontend — VERT (session 2026-11-23)

- `app/src/auth/session.ts` : clé unique `atrium.session`, écriture localStorage (« Rester connecté ») ou sessionStorage, lecture avec repli, purge.
- `app/src/lib/api.ts` : wrapper `/v1` ; interception 401 centralisée — tenant : **une seule** promesse de refresh partagée + **un seul** rejeu de la requête ; échec du refresh → purge + `/login` ; admin (401, pas de refresh) → purge + `/login`.
- `app/src/auth/AuthContext.tsx` (session au boot, login/logout, navigation par `kind`), `app/src/router.tsx` (gardes `beforeLoad` : session absente → `/login`, kind incompatible → page correspondante), `app/src/pages/LoginPage.tsx` (4 vues `login | create | request | forgot`, SSO masqué, vues secondaires inertes avec bannière « à venir », aperçu de slug), `AdminHomePage` / `TenantHomePage` (« Admin » / « Tenant » + « Bienvenue {name} »), `App.tsx` (composition du design : `QueryClientProvider` → `RouterContextProvider` → `AuthProvider` → `RouterProvider`), `AuthLayout` + story, styles `.shell/.brand-*` dans `atrium.css`, config vitest (jsdom + `@testing-library/jest-dom`).
- 22 tests React verts (`session.test.ts`, `api.test.ts`, `LoginPage.test.tsx`) ; `pnpm lint`, `pnpm build`, `pnpm test`, `pnpm build-storybook` verts.
- Cause-racine des échecs initiaux (documentée) : le premier rendu d'un `RouterProvider` est **asynchrone** → `waitFor` obligatoire dans les asserts ; `createRouter({ routeTree: undefined })` ne panic pas (routeur sans routes rendu vide).

### E2E — VERT (porte de phase 3 franchie cette session)

- Harnais relu contre le contrat §E2E du plan, corrigé, **puis exécuté** : `cmd/e2edb` (drop/create de `artisan_e2e` en superuser épinglé `E2E_DB_ADMIN_USER/PASS` + `FLUSHDB` Redis db 1 ; garde durcie : refus si nom vide, `DB_NAME`, `postgres`, `template*`), cible `make e2e`, `app/playwright.config.ts` (webServer `go run .`, PORT 9100, `DB_NAME=artisan_e2e`, origine unique `app/dist`, projet `setup`), `app/e2e/setup.ts` (seed de l'owner tenant via `POST /v1/user/register`) + `app/e2e/login.spec.ts` (CUJ 1/2/3) (à la date du 2026-11-23 ; l'isolation des index 1/2/3 est en vigueur depuis le plan REDIS_DB).
- **3 runs consécutifs de `make e2e` verts (4/4 tests)** ; chromium Playwright 1.65.1 installé sous WSL (une fois : `pnpm -C app exec playwright install chromium`). Détails du contrat retenus : emails `admin@e2e.local` (superadmin CLI) et `owner@e2e.local` (owner tenant seedé), assertions d'URL (`/admin`, `/`, `/login`).
- Correctifs harnais : include vitest limité à `src/**` (sinon vitest collectait les specs Playwright → `pnpm test` en échec) ; `app/tsconfig.json` inclut `e2e` + `playwright.config.ts` (le harnais est désormais type-checké par `pnpm build`) ; `.gitignore` : `/e2edb` (+ `.exe`).
- Effet de bord attendu du cycle : le `FLUSHDB` purge la base Redis 1 (sessions de l'app de dev) (à la date du 2026-11-23 ; l'isolation des index 1/2/3 est en vigueur depuis le plan REDIS_DB — le `FLUSHDB` ne purge plus que l'index 3, les sessions dev subsistent en 1).

### Correctif hors périmètre (réparé cette session, non planifié)

- `internal/testutil/postgres.go` + `postgres_test.go` : la colonne goose est `version_id` (et non `version`) sous goose v3.20.0, et le conteneur n'est considéré prêt que sur le log « database system ready to accept connections » (2 occurrences) — la seule écoute du port 5432 laissait passer un `FATAL 57P03`.

### CI + documentation (items 29-30)

- `.github/workflows/ci.yml` : job `frontend` — étape `pnpm test` insérée **entre** `pnpm build` et le smoke test `pnpm build-storybook` (YAML validé).
- `AGENTS.md` : `POST /v1/login` (réponses discriminées, bcrypt factice, politique non-recouvrement), migration 000003, flag `-name`, `make e2e` (prérequis `playwright install`), tests React (`app/src/**`, e2e exclu de l'include vitest), bump `SQLC_VERSION` v1.30.0 (flaky : relancer la cible en cas de panic), `cmd/e2edb`.

### Passe drift 2026-11-23 (3 divergences du drift report corrigées)

- **Item 18 — composition `App.tsx`** : le fichier suit désormais la composition du design — `QueryClientProvider` → `RouterContextProvider` (`@tanstack/react-router`) → `AuthProvider` → `RouterProvider` ; l'intermédiaire technique est requis car `AuthContext` appelle `useNavigate()` **au-dessus** du `RouterProvider` ; la route racine de `app/src/router.tsx` ne porte plus que l'`Outlet`. Non-régression verte : `pnpm test` 22/22, `pnpm lint`, `pnpm build`.
- **Items 6/10 — branche superadmin sans refresh** : `models/auth.go` — `Token.RefreshToken` taggé `json:"refresh_token,omitempty"` (la clé disparaît du payload quand elle est vide) ; `controllers/auth.go` — la branche superadmin renvoie `models.Token{AccessToken}` (filaire inchangé) + annotations décrivant l'union discriminée par `kind` ; `tests/login_test.go` — helper `tokenPayload` + asserts (superadmin → `refresh_token` **absent**, tenant → les deux, endpoints legacy verrouillés) ; `docs/` régénérés (`make generate_docs`, idempotent) — le schéma `models.Token` généré conserve la propriété `refresh_token` (swag n'honore pas `omitempty`), l'absence côté superadmin est documentée dans la description d'endpoint.
- **Item 31 — memory bank resynchronisée** : `activeContext.md` (bloc de vérification finale passé en « clos », composition `App.tsx` corrigée) + `progress.md` (point de vérification `[x]`, entrée d'historique 21, « limites connues » à jour) + ligne de statut du plan (`.kilo/plans/1795428465642-login-simple.md`, section « Implementation Status »).

## Vérification finale du plan — CLOS (fait et vérifié le 2026-11-23)

- `gofmt -l .` → **vide** (dérive `tests/login_test.go` repliée) ; `git diff --exit-code internal/db` → **vide dès la 1re passe** (sqlc v1.30.0, garde CI satisfaite) ; `make generate_docs` → **idempotent** (doc régénérée à l'identique) ; `go vet -mod=readonly ./...` → vert.
- `make test` → **VERT** (exit 0, 66 PASS / 0 FAIL / 0 SKIP, 10 packages `ok`, PostgreSQL + Redis actifs) ; frontend : `pnpm lint` exit 0, `pnpm test` 22/22, `pnpm build` exit 0.
- Décision `app/dist` **actée** : 10 fichiers retirés de l'index (commit `3ac3aae`) + `dist` dans `app/.gitignore` → plus de diff fantôme CRLF/LF, et le smoke test CI Go retrouve le fallback démo (« Gin Boilerplate ») → **redevient vert** ; les fichiers restent sur disque (dev, `make e2e`).
- Point à solder au prochain commit : `e2edb.exe` est **encore suivi** (un fichier déjà suivi n'est pas ignoré par `.gitignore`) → `git rm --cached e2edb.exe`.
- Limite assumée : la CI réelle reste **non exécutable localement** (aucun push demandé) — jobs validés statiquement seulement.
- Les **corrections de drift** (`app/src/App.tsx`, `app/src/router.tsx`, `models/auth.go`, `models/login.go`, `controllers/auth.go`, `tests/login_test.go`, `docs/`) restent en **working tree non commité**.

## Leçons durables (poste Windows + WSL)

- Le Go **Windows** ne peut pas builder sur le partage `\\wsl.localhost\...` (`RLock: Fonction incorrecte`) → **toutes** les commandes (go / make / git / pnpm) se jouent via `wsl -d Debian -- bash -lc 'cd /home/sylvaindeb/projects/go-artisan-services && …'` ; le Go WSL n'est pas dans le PATH par défaut (`export PATH=/usr/local/go/bin:$PATH`).
- `pnpm` **uniquement sous WSL** : `app/node_modules` y est installé (liens symboliques Linux) ; pnpm côté Windows tente de purger le dossier (`ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`).
- `pdftotext` (poppler-utils) installé sous WSL → les tests PDF de `./invoice` peuvent y tourner.
- WSL : PostgreSQL (127.0.0.1:15432) et Redis (127.0.0.1:16379) joignables — stack Docker `artisan-dev` (`make services_up`) ; port 9100 libre (serveur e2e).
- PowerShell 5.1 : pas de heredoc, pas de `&&` ; passer par `wsl -d Debian -- bash -lc` pour les enchaînements.
- Piège d'outillage : la capture de sortie console perd les blocs contenant des caractères non-ASCII → écrire les sorties dans un fichier (`> /tmp/x.txt`) puis le relire.

## Historique résumé (détail dans `progress.md`)

- 2026-11-21 : plan multi-tenant RLS terminé (schéma/RLS, pools, auth Redis, articles, backoffice admin, CI, Swagger).
- 2026-11-22 : Storybook 10.6.8 + harness tokens dans `app/` ; correction de l'environnement de test Windows (`.env` déquoté, `REDIS_PASSWORD` vidé, poppler winget, cible `test` qui masque le xpdf de Git for Windows).
- 2026-11-23 (matin) : stack de test Go (mockery, gofakeit, testfixtures, testcontainers, goose v3.20.0) + helpers `internal/testutil/`.
- 2026-11-23 (après-midi) : plan login simple — phase backend (items 1-10) puis phase frontend (items 11-23).
- 2026-11-23 (soir) : phase e2e exécutée et verte, CI + AGENTS.md + memory bank (items 29-31) ; puis passe de correction du drift (3 divergences) et vérification finale close (gofmt, sqlc, docs, `make test` verts ; `app/dist` détracké et commité).
- 2026-11-24 (soir) : isolation des index Redis (plan REDIS_DB — dev 1, suites Go 2, cycle e2e 3) + base jetable PostgreSQL `artisan_test` pour les suites Go (plan PG, `DB_NAME` épinglé par `TestMain`, provisioning dans `make test`) ; `make test` et `make e2e` verts, marqueurs dev intacts.
