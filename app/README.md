# Frontend — `app/`

Workspace frontend (Vite + React 19 + TypeScript strict + Tailwind CSS 4).
Versions épinglées via `packageManager` (pnpm 11) et `engines` (Node ≥ 22.12) ;
corepack fournit le pnpm exact (`corepack pnpm …` si `pnpm` n'est pas sur le
PATH, ou `corepack enable pnpm` pour un shim permanent — pas d'installation
globale).

## Commandes (depuis `app/`)

| Commande | Rôle |
| --- | --- |
| `pnpm dev` | serveur Vite de l'app — http://localhost:5173 |
| `pnpm storybook` | Storybook — http://localhost:6006 |
| `pnpm build-storybook` | build statique Storybook → `storybook-static/` |
| `pnpm build` | `tsc --noEmit` puis build Vite → `dist/` |
| `pnpm lint` | ESLint (flat config ; `.storybook/` et stories incluses) |
| `pnpm test` | Vitest |
| `node_modules/.bin/storybook doctor` | diagnostic Storybook (triage versions/addons) |

## Storybook

- Configuration dans `.storybook/` : `main.ts` (CSF via `defineMain`,
  framework `@storybook/react-vite`, addons **docs / a11y / themes**) et
  `preview.tsx` (import de `../src/index.css`,
  `withThemeByClassName` light `''` / dark `'dark'`, défaut light).
- Le sélecteur de thème de la barre d'outils bascule exactement la classe
  `.dark` — le mécanisme réel de l'app (`@custom-variant dark`), rien à
  synchroniser.
- Stories colocalisées : `src/**/*.stories.tsx` (CSF) ; pages MDX à côté.
- `src/design-system/` : palettes (couleurs, rayons, typographie) lues à
  l'exécution depuis les variables CSS de `src/index.css` — le **contrat de
  jetons** partagé avec l'outil « Open Design ». Réémettre les variables dans
  le même format suffira à mettre à jour pages et stories (zéro rework).

## Dépendances et install

- Toujours passer par `pnpm install` (lockfile commité) ; en CI :
  `pnpm install --frozen-lockfile`.
- Les scripts d'installation des dépendances sont soumis à approbation
  (pnpm 11) : la liste `allowBuilds` vit dans `pnpm-workspace.yaml`
  (`esbuild` est approuvé). Après l'ajout d'une dépendance qui exécute un
  script, l'ajouter ici — sinon `pnpm install` échoue avec
  `ERR_PNPM_IGNORED_BUILDS` et tout `pnpm run` avec le contrôle de deps
  pré-exécution.

## Dépannage

- `pnpm: command not found` → corepack (voir en-tête) ; ne pas installer
  pnpm globalement.
- Crash de `pnpm storybook` avec
  `spawn /mnt/c/.../powershell.exe ENOENT` → WSL sans monture `C:` (`wsl.conf`
  `[automount] enabled = false`) : Storybook n'y ouvrira jamais le navigateur
  lui-même. Le script passe déjà `--no-open` — ouvrir
  http://localhost:6006 manuellement.
- Erreur d'addons ou de versions → `storybook doctor`.
- `ERR_PNPM_IGNORED_BUILDS` → section "Dépendances et install" (approbation
  `allowBuilds`).
