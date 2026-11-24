# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

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

- Harnais relu contre le contrat §E2E du plan, corrigé, **puis exécuté** : `cmd/e2edb` (drop/create de `artisan_e2e` en superuser épinglé `E2E_DB_ADMIN_USER/PASS` + `FLUSHDB` Redis db 1 ; garde durcie : refus si nom vide, `DB_NAME`, `postgres`, `template*`), cible `make e2e`, `app/playwright.config.ts` (webServer `go run .`, PORT 9100, `DB_NAME=artisan_e2e`, origine unique `app/dist`, projet `setup`), `app/e2e/setup.ts` (seed de l'owner tenant via `POST /v1/user/register`) + `app/e2e/login.spec.ts` (CUJ 1/2/3).
- **3 runs consécutifs de `make e2e` verts (4/4 tests)** ; chromium Playwright 1.65.1 installé sous WSL (une fois : `pnpm -C app exec playwright install chromium`). Détails du contrat retenus : emails `admin@e2e.local` (superadmin CLI) et `owner@e2e.local` (owner tenant seedé), assertions d'URL (`/admin`, `/`, `/login`).
- Correctifs harnais : include vitest limité à `src/**` (sinon vitest collectait les specs Playwright → `pnpm test` en échec) ; `app/tsconfig.json` inclut `e2e` + `playwright.config.ts` (le harnais est désormais type-checké par `pnpm build`) ; `.gitignore` : `/e2edb` (+ `.exe`).
- Effet de bord attendu du cycle : le `FLUSHDB` purge la base Redis 1 (sessions de l'app de dev).

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
