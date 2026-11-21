# Installation des dépendances front uniquement dans `/app` — périmètre verrouillé

* **Date :** 2026-11-21
* **Statut :** Prêt pour exécution
* **Supersede (pour l'exécution) :** `1795265473956-frontend-stack-design-draft.md` et `1795265473956-goldfish_review_report.md` — conservés comme historique de design, **non exécutables**. Le présent plan est la seule référence d'action du chantier.

## Objectif unique

Ramener `/app` à l'état « dépendances de la stack front installées, rien d'autre » :

```
app/package.json · app/pnpm-lock.yaml · app/pnpm-workspace.yaml · app/.gitignore (+ node_modules/ installé, non commité)
```

## Constat de l'état actuel (vérifié dans le dépôt)

* `/app` contient DÉJÀ les deps installées : `package.json` (pnpm@11.20.0 épinglé via `packageManager`, engines node ≥ 22.12), `pnpm-lock.yaml` (~107 Ko), `pnpm-workspace.yaml`, `.gitignore` (`node_modules`), `node_modules/`. Les deps couvrent la stack du design draft : React 19, react-router, @tanstack/react-query, react-hook-form + @hookform/resolvers, zod, tailwindcss + @tailwindcss/vite, base shadcn (clsx, tailwind-merge, class-variance-authority, @radix-ui/react-slot), @sentry/react + web-vitals ; devDeps : vite, @vitejs/plugin-react, typescript, vitest, @testing-library/react, msw, @playwright/test, eslint (flat) + typescript-eslint + react-hooks, prettier, jsdom, @types/*.
* Hors périmètre laissé par le chantier stoppé : 6 fichiers de config (`vite.config.ts`, `tsconfig.json`, `eslint.config.js`, `.prettierrc`, `playwright.config.ts`, `index.html`), `src/` (~50 fichiers), `dist/` (build).
* Arbre de travail backend : couche auth cookies NON COMMITEE (modifications `controllers/`, `models/`, `middleware/`, `main.go`, `tests/`, suppression `forms/auth.go`, nouveau `models/auth_test.go`) — état « LIVRÉ » selon le design draft. **Hors périmètre : ne ni modifier, ni réverter, ni commiter.**

## Décisions verrouillées (auteur, 2026-11-21)

1. Le hors-périmètre de `/app` est **supprimé**, avec sauvegarde préalable.
2. `@sentry/react` + `web-vitals` **restent** dans les deps (décision enregistrée « Sentry conservé en v1 ») : `package.json` / `pnpm-lock.yaml` inchangés sur ce point.
3. Fin de chantier : un **commit isolé** des 4 fichiers deps uniquement.

## Tâches (ordre strict)

1. **Sauvegarde** : copier `app/` hors `node_modules` vers `C:\Users\schat\AppData\Local\Temp\kilo\app-backup-20261121\` (ex. `tar -a -cf "C:\Users\schat\AppData\Local\Temp\kilo\app-backup-20261121.zip" --exclude=app/node_modules app`). **Vérifier que la sauvegarde contient `src/`, `dist/` et les 6 configs AVANT toute suppression.**
2. **Suppression du hors-périmètre** : `app/src/`, `app/dist/`, `app/vite.config.ts`, `app/tsconfig.json`, `app/eslint.config.js`, `app/.prettierrc`, `app/playwright.config.ts`, `app/index.html`. Rien d'autre.
3. **Vérification de l'installation** : dans `/app`, `pnpm install --frozen-lockfile` → doit sortir 0 tel quel (aucune modification de deps décidée). Si désynchronisation lockfile/manifest détectée : ne PAS bricoler les versions ; exécuter `pnpm install` pour resynchroniser et le signaler dans le rapport.
4. **Contrôle de l'état final** : `ls app` = exactement 5 entrées : `node_modules/`, `package.json`, `pnpm-lock.yaml`, `pnpm-workspace.yaml`, `.gitignore`. Aucun autre fichier créé dans `/app`.
5. **Commit isolé** : `git add app/package.json app/pnpm-lock.yaml app/pnpm-workspace.yaml app/.gitignore` ; vérifier que l'index ne contient QUE ces 4 chemins (`git diff --cached --stat`) ; commit : `chore(app): install frontend stack dependencies (pnpm lockfile)`. Ne jamais `git add -A` ni `git add .` : les modifications backend non commitées et `.kilo/plans/` doivent rester hors index.
6. **Sanity backend** : `go build -mod=readonly .` (aucun fichier Go touché par ce chantier ; en cas d'échec préexistant hors `/app` : le signaler sans corriger).

## Vérification (critères d'acceptation)

* `pnpm install --frozen-lockfile` exit 0 dans `/app`.
* `ls app` = 5 entrées exactement (cf. tâche 4).
* Sauvegarde complète et lisible (`src/`, `dist/`, 6 configs) avant suppression.
* `git log -1 --stat` : un seul commit, exactement les 4 fichiers deps.
* `git status` après commit : modifications backend (M/…, `D forms/auth.go`, `?? models/auth_test.go`, `?? .kilo/plans/`) INTACTES et toujours non commitées ; `app/` n'apparaît plus en non-suivi (`node_modules` ignoré via `app/.gitignore`).
* `go build -mod=readonly .` OK.

## Périmètre verrouillé — INTERDIT (tout le reste)

* Créer tout fichier dans `/app` au-delà des 4 fichiers deps (pas de configs, pas de `src/`, pas de `components.json`).
* Modifier `package.json` ou le lockfile (deps, versions, scripts inchangés — Sentry conservé).
* Toute action backend : édition Go, revert des modifications non commitées, migrations, `docs/`, requêtes sqlc.
* `Makefile`, `.github/workflows/ci.yml`, `README.md`, `AGENTS.md`, `memory-bank/` : intouchés.
* Exécuter `shadcn init/add`, `playwright install` (binaires navigateurs), `vite build`, `tsc`, lint ou tests — aucun outillage requis à ce stade (pas de configs = comportement attendu).
* Tout commit autre que le commit isolé deps (jamais de commit mélangeant le backend).

## Notes / risques

* La suppression (tâche 2) est **irréversible** (fichiers non suivis par git) → sauvegarde de la tâche 1 obligatoire et vérifiée avant.
* `dist/` absent → le serveur Gin sert le template démo via le fallback `NoRoute` (comportement LIVRÉ et voulu, cf. M5 du rapport Goldfish).
* `pnpm-workspace.yaml` est conservé : config d'installation pnpm (`minimumReleaseAgeExclude`), elle fait partie de la couche deps.
* Le statut git backend non commité (couche cookies) sera tranché séparément par l'auteur (commit dédié ou abandon) — jamais pendant ce chantier.
* Chantiers ultérieurs (scaffolding, features, CI, e2e, docs) : à planifier un par un à partir du design draft conservé en historique ; ce plan ne les autorise pas.
