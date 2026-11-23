# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

## Session 2026-11-23 — plan « vue de connexion simple » (`.kilo/plans/1795428465642-login-simple.md`) : backend + frontend + e2e verts, vérification finale en cours

Implémentation complète : items 1-30 faits et vérifiés, item 31 = ce fichier. Arbre de travail **non commité** (aucun commit demandé pour l'instant).

### Backend — VERT (sessions précédentes)

- Migration goose `000003_admin_name.sql` : `admin_user.name varchar NOT NULL DEFAULT 'Superadmin'` (table plateforme, sans RLS) ; `db/queries/admins.sql` aligné (`GetAdminByEmail` renvoie `name`, `CreateAdmin` upsert idempotent qui remet le nom à jour) ; code sqlc régénéré (`internal/db` commité).
- `models/login.go` — `UnifiedLogin` : tentative utilisateur tenant puis repli superadmin, réponse **discriminée par `kind`** (`tenant_user` → user + tenant + tokens access/refresh ; `superadmin` → admin id/email/name + token access). Email inconnu → bcrypt contre un hash factice (coût constant, pas d'énumération temporelle) ; tenant suspendu → 403 propagé ; politique assumée : **un email superadmin ne recouvre jamais un email utilisateur tenant** (utilisateur trouvé + mauvais mot de passe → 406 direct, sans tentative admin).
- `controllers/auth.go` — `AuthController.Login` (200 / 403 / 406, annotations Swagger) ; route `POST /v1/login` ; CLI `-create-admin` enrichi du flag `-name` (défaut « Superadmin », upsert idempotent).
- Tests : `tests/login_test.go` (suite tag `all`) — user OK + Redis, admin OK + `name`, 406 mauvais mot de passe, 406 email inconnu (assert du bcrypt factice), 403 tenant suspendu, 406 validation, endpoints legacy toujours 200 ; `make test` complet vert, `go vet` vert.

### Frontend — VERT (session 2026-11-23)

- `app/src/auth/session.ts` : clé unique `atrium.session`, écriture localStorage (« Rester connecté ») ou sessionStorage, lecture avec repli, purge.
- `app/src/lib/api.ts` : wrapper `/v1` ; interception 401 centralisée — tenant : **une seule** promesse de refresh partagée + **un seul** rejeu de la requête ; échec du refresh → purge + `/login` ; admin (401, pas de refresh) → purge + `/login`.
- `app/src/auth/AuthContext.tsx` (session au boot, login/logout, navigation par `kind`), `app/src/router.tsx` (gardes `beforeLoad` : session absente → `/login`, kind incompatible → page correspondante), `app/src/pages/LoginPage.tsx` (4 vues `login | create | request | forgot`, SSO masqué, vues secondaires inertes avec bannière « à venir », aperçu de slug), `AdminHomePage` / `TenantHomePage` (« Admin » / « Tenant » + « Bienvenue {name} »), `App.tsx` (QueryClientProvider + AuthProvider + RouterProvider), `AuthLayout` + story, styles `.shell/.brand-*` dans `atrium.css`, config vitest (jsdom + `@testing-library/jest-dom`).
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

## Reste à faire (en cours — vérification finale du plan)

- `make test` complet (PostgreSQL + Redis actifs) ; `make generate_sqlc` → diff **vide** sur `internal/db` (garde CI `git diff --exit-code internal/db`) ; `make generate_docs` (Swagger à jour après les annotations).
- Points en suspens :
  - (a) `app/dist/` est **suivi par git** et chaque `pnpm build` le réécrit en **LF** alors que la version committée est en **CRLF** → diff fantôme sur `app/dist/index.html` (restauré via `git checkout` à chaque fois) ; à trancher : normaliser LF dans l'arbre ou ignorer `app/dist`.
  - (b) dérive `gofmt` **préexistante** sur `tests/login_test.go` (2 lignes d'alignement de tags de struct, introduites par le commit `7a68626`) → à replier au passage.
  - (c) la CI réelle n'est pas exécutable localement (aucun push demandé) : seuls les jobs sont validés statiquement.
- Aucun commit demandé : tout le travail du plan reste en **working tree non commité**.

## Leçons durables (poste Windows + WSL)

- Le Go **Windows** ne peut pas builder sur le partage `\\wsl.localhost\...` (`RLock: Fonction incorrecte`) → **toutes** les commandes (go / make / git / pnpm) se jouent via `wsl -d Debian -- bash -lc 'cd /home/sylvaindeb/projects/go-artisan-services && …'` ; le Go WSL n'est pas dans le PATH par défaut (`export PATH=/usr/local/go/bin:$PATH`).
- `pnpm` **uniquement sous WSL** : `app/node_modules` y est installé (liens symboliques Linux) ; pnpm côté Windows tente de purger le dossier (`ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`).
- `pdftotext` (poppler-utils) installé sous WSL → les tests PDF de `./invoice` peuvent y tourner.
- WSL : PostgreSQL (127.0.0.1:5432) et Redis (127.0.0.1:6379) joignables, port 9100 libre (serveur e2e).
- PowerShell 5.1 : pas de heredoc, pas de `&&` ; passer par `wsl -d Debian -- bash -lc` pour les enchaînements.
- Piège d'outillage : la capture de sortie console perd les blocs contenant des caractères non-ASCII → écrire les sorties dans un fichier (`> /tmp/x.txt`) puis le relire.

## Historique résumé (détail dans `progress.md`)

- 2026-11-21 : plan multi-tenant RLS terminé (schéma/RLS, pools, auth Redis, articles, backoffice admin, CI, Swagger).
- 2026-11-22 : Storybook 10.6.8 + harness tokens dans `app/` ; correction de l'environnement de test Windows (`.env` déquoté, `REDIS_PASSWORD` vidé, poppler winget, cible `test` qui masque le xpdf de Git for Windows).
- 2026-11-23 (matin) : stack de test Go (mockery, gofakeit, testfixtures, testcontainers, goose v3.20.0) + helpers `internal/testutil/`.
- 2026-11-23 (après-midi) : plan login simple — phase backend (items 1-10) puis phase frontend (items 11-23).
- 2026-11-23 (soir) : phase e2e exécutée et verte, CI + AGENTS.md + memory bank (items 29-31) ; reste la vérification finale puis le commit, non demandé.
