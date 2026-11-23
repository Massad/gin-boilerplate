# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

## Outils de test (session 2026-11-23, plan stack-test-go)

- Quatre outils ajoutés et épinglés : **gofakeit/v7** (factories seedées par nom de test : `gofakeit.Seed(int64(crc32.ChecksumIEEE([]byte(t.Name()))))` en tête de chaque factory — la valeur 0 force un reseed aléatoire, d'où le garde 0→1), **testfixtures/v3** (fixtures YAML `tests/fixtures/` chargées via le pool admin, une fois par suite), **testcontainers-go** (Postgres éphémère opt-in — skip propre sans Docker : la lib n'expose pas de sonde daemon, le skip se fait par gestion d'erreur au démarrage), **mockery** CLI v2.53.8 dans le Makefile pattern `.cache/tools/` (`make generate_mocks` → mocks commités `internal/mocks/`, config `.mockery.yaml`, interfaces DBTX/Querier uniquement).
- goose en bibliothèque (v3.20.0, aligné CLI) exécute les migrations du conteneur éphémère (`internal/testutil/postgres.go`, Provider API).
- Pièges découverts à l'exécution : les rôles non-superuser ne peuvent ni `TRUNCATE` (privilège absent) ni `ALTER TABLE ... DISABLE/ENABLE TRIGGER ALL` (ownership requis) — le chargeur purge donc en `DELETE` (enfants d'abord, cascades actives) et un connecteur enveloppé intercepte les seuls batches de toggle de triggers de testfixtures ; contraintes FK restent actives, fichiers fixtures ordonnés par dépendance (admin_user, tenant, user, article).
- Course de données corrigée avec `-p 1` sur `make test` (Makefile + CI) : les packages tournent désormais en série — le purge de début de suite supprimait à l'occasion le tenant/admin que `controllers`/`main` (packages parallèles) venaient de seed, ce qui faisait échouer la suite ~1 run sur 4. La convention « suite séquentielle » est maintenant appliquée au niveau packages.
- La suite `all` a conservé ses garanties : ordre séquentiel de `tests/article_test.go` (variables de package, pas de testify/suite), tests RLS fail-closed intacts, `-race` en CI (gcc absent localement), skips Redis.
- Validé : `make test` vert (9 packages ; testcontainers SKIP propre), mocks idempotents (2× `make generate_mocks` → diff vide), `go mod verify`.

## État actuel (2026-11-22)

**Plan « Install Storybook in app/ » exécuté** (`.kilo/plans/1795339024940-install-storybook-front.md`). Points clés :

- Storybook **10.6.8** installé (`@storybook/react-vite`, addons docs/a11y/themes) — l'attente « ligne v11 » du plan était périmée : 11 n'existe qu'en beta, 10.6.8 est le major courant et ses peers acceptent Vite 8 / React 19.
- Harness tokens : `app/src/design-system/` (palettes couleurs/rayons/typo lues à l'exécution via `getComputedStyle` + pages MDX) — ancré sur le contrat jetons `index.css`, zéro rework pour le pont « Open Design ».
- Gaps préexistants corrigés : `pnpm lint` réparé (`app/eslint.config.js` flat config), `.storybook` ajouté au `tsconfig.json` (par le script init lui-même).
- Job CI `frontend` ajouté (Node 22 + pnpm, `--frozen-lockfile`, lint, build, build-storybook) — pipeline Go inchangé.
- Piège pnpm 11 : les scripts de build npm exigent l'approbation `allowBuilds` dans `app/pnpm-workspace.yaml` (esbuild approuvé) — sinon `pnpm install` échoue et tout `pnpm run` échoue via le contrôle deps pré-exécution.
- Validé : `pnpm lint` ✓, `pnpm build` ✓, `pnpm build-storybook` ✓, `storybook doctor` « project looks good ». Toute la doc repo mise à jour (README racine, `app/README.md`, AGENTS.md).
- Travail **non commité** (convention repo : commits quand l'utilisateur le décide).
- Note : une session locale concurrente a réécrit une partie des fichiers `design-system` (labels FR, hooks avec MutationObserver) pendant l'exécution ; l'état retenu est celui du disque, cohérent et validé suite au retrait des doublons.

## Récent (session 2026-11-21)

- Retrait intégral de l'accès public (principe « rien de public ») : lookup `GET /v1/public/:slug` supprimé (controller, `OneBySlug`, `GetTenantBySlug` — code sqlc régénéré), routes invoice démo supprimées (`invoice/` devient bibliothèque PDF sans routes), Swagger servi uniquement hors `ENV=PRODUCTION` ; `generate_docs` rejoué, AGENTS.md/README/memory bank mis à jour.

- Étape 12 (T18/E7) : retrait sqlx/lib/pq vérifié no-op, garde-fou CI « Check removed legacy drivers » (pattern sur imports quotés — le commentaire `lib/pq` de `internal/db/pool.go:70-72` est volontairement non matché), AGENTS.md + README.md mis à jour, docs Swagger non régénérées (aucune annotation touchée).
- Nettoyage : 40 `.kilo/*.sh` one-offs supprimés (+ secrets embarqués), 6 `*_report.md` de la phase design supprimés, `.cache/` (137 Mo) supprimé, binaire `gin-boilerplate` supprimé + pattern `.gitignore` corrigé (`/gin-boilerplate`).
- Revalidation complète : build, vet, unitaires purs (config, models), suite tag `all` sur 7 packages **y compris `./` (main_test.go, maillon manquant de l'étape 12)** — tout vert en `-race` (RLS prouvée en `app_user`/`app_admin`).
- Création de ce memory bank.

## Candidats pour la prochaine session (à arbitrer avec l'utilisateur, hors périmètre du plan actuel)

- Pont Open Design → réémission des variables de `app/src/index.css` (le contrat jetons est prêt, zéro rework attendu).
- Intégration Storybook × vitest (feature `test`), Chromatic optionnel, retrait `@varlock/vite-integration` (anomalie préexistante).
- Câbler `RequirePermission` (routes owner-only) + invitations/membres (colonne et map déjà posées).
- SPA PWA (repo séparé) — le fallback backend est déjà en place.
- Conséquences du nettoyage : `.cache/` sera retéléchargé au prochain `make install_*` ; tout le travail des plans reste non commité — à committer quand l'utilisateur le décide.

## Session 2026-11-22 (après-midi) - correction de l'environnement de test Windows

- `make test` échouait de 4 manières : DSN `invalid port` (valeurs de `.env` entre guillemets exportées littéralement dans le shell), Redis `ERR AUTH` (`.env` déclarait un `REDIS_PASSWORD` alors que le serveur local n'a pas de requirepass), données de test orphelines (panique d'un run précédent ayant sauté `TestCleanUp` → 406 au register puis panic), `pdftotext` absent.
- Corrections : `.env` et `.env_rename_me` déquotés (tous les loaders restent compatibles), `REDIS_PASSWORD` vidé (aligné sur CI `REDIS_PASSWORD: ""` et le Redis local), purge des lignes orphelines (`public.tenant` en cascade + `admin_user` de test + clés Redis `user_sessions:*`/`admin:*`), Poppler 26.09 installé via winget (`oschwartz10612.Poppler`, entrée User PATH — un *nouveau* terminal est nécessaire pour la voir, les shells déjà lancés gardent l'ancien PATH).
- Piège majeur : **Git for Windows embarque un xpdf pdftotext v4.00** (`C:\Program Files\Git\mingw64\bin\pdftotext.exe`, 1996-2017) qui domine le PATH dès que make exécute les recettes via son sh — son extraction `-layout` scramble les lignes et faisait échouer `TestGeneratePDFContent` de façon reproductible via make mais jamais via `go test` direct (deux binaires différents résolus). Prouvé en extrayant le même PDF avec les deux binaires.
- Cible `test` du Makefile corrigée en deux temps. Premier essai inerte : le PATH arrive du sh sous forme d'UNE seule ligne à deux-points (filtrer ligne à ligne exige `tr ':' '\n'`) et l'ancre `^(mingw64|usr)` rate les entrées préfixées d'un slash (`/mingw64/bin`, `/usr/bin`). Version finale : scrub fait **uniquement** si `case "$(uname -s)" in *MSYS*|*MINGW*)` (Linux garde `/usr/bin`, là où vit le vrai poppler), filtre `awk '!/^\/(mingw64|usr)\/bin/ && !/\/[Gg]it\/(mingw64|usr)\/bin/'` après `tr ':' '\n'`, puis `paste -sd ':' -`. La cible source aussi `.env` (`set -a`) pour supprimer les SKIP dépendants du shell appelant.
- Vérifié : `go vet` clean, `make test` exit 0, 8 packages ok, zéro SKIP, deux runs consécutifs stables. `-race` non joué localement (exige CGO/gcc ; chemin CI uniquement).
- Doc : README.md (section Testing simplifiée à `make test` + poppler Windows via winget + note « nouveau terminal » ; le bloc d'export manuel reste dans la section Migrations car `migrate_up` build `DB_DSN` à partir des variables make) et AGENTS.md (entrée `make test` mise à jour : auto-source `.env`, poppler requis, masquage xpdf). Tout non commité — commit à la décision de l'utilisateur.
