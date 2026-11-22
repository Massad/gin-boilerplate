# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

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
