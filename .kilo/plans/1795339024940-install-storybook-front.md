# Install Storybook in `app/` — design-system-ready groundwork

* **Author:** misterlys (proposed — from GitHub remote; confirm)
* **Date:** 2026-11-22
* **Status:** FINAL — all decisions confirmed by user: token-palette witness, generated docs pages, pre-existing gap fixes, minimal frontend CI smoke job, repo documentation updates.

## Overview

Install Storybook (current major, v11 line; Vite 8 and React 19 supported) in `app/` as development groundwork, before the first component wave. Story conventions are established now so every future shadcn/Base UI component ships documented from day one. The design system currently lives in an external Figma-like tool ("Open Design"); its later integration must require zero rework of what is installed today — so the install is anchored on the **CSS token contract** (`app/src/index.css` oklch variables + `.dark` class), not on any throwaway component.

## Analysis (Goldfish synthesis, 7 lenses, deduplicated)

Four parallel fresh-context reviews (Contrarian/YAGNI+Cost, DX+Migration, SRE+Debt, Inventive Catalyst). Three of four rejected the initial hypothesis as-is. Key accepted findings:

1. **Witness choice matters (High).** A shadcn Button story proves the chain but is throwaway: it will be deleted when the design system's own components arrive, and it duplicates the shadcn registry docs. The inventive lens proposed a better witness: **token-palette stories** (color swatches, radius, typography) rendered at runtime from `index.css` CSS variables. They prove the full chain (SB + Tailwind 4 + dark toggle) with zero components, and they survive the design system swap because tokens are the stable contract.
2. **No empty "Design System" docs page (Medium).** A placeholder page with no content becomes dead code. Instead: docs pages built from the same CSS variables (colors/radius/typography). When the Open Design bridge later re-emits variables in the same format, docs regenerate — no structural rework.
3. **Theme toggle must use the app's real mechanism (Medium).** `app/src/index.css` uses class-based dark mode (`@custom-variant dark (&:is(.dark *))`). `@storybook/addon-themes` `withThemeByClassName` toggles exactly the `.dark` class — validated against current SB docs.
4. **Typecheck and lint gaps (High, pre-existing for lint).** `app/tsconfig.json` includes only `["src", "vite.config.ts"]` — `.storybook/` config would escape `tsc --noEmit` (used by `pnpm build`). And `app/` has **no eslint config at all** while `pnpm lint` exists (pre-existing gap; `eslint .` currently errors). Both must be fixed for stories to be typechecked/linted like the rest of the code.
5. **CI covers Go only (High for SRE).** `.github/workflows/ci.yml` has no frontend job; `app/` is invisible to CI. Without at least a `storybook build` smoke, broken config/stories merge silently.
6. **YAGNI objection documented (not adopted).** The contrarian lens recommends deferring Storybook until real components exist (zero components today; SB pulls a large dep tree; lighter alternatives exist — Ladle, react-catalog). Mitigations that keep install-now defensible: the token harness gives Storybook immediate content, conventions are cheaper to set before the component wave, and everything installed stays valid after the design system integration. The user's explicit decision to install now stands; this objection is recorded, not dismissed.
7. **Version/policy notes (Low).** `minimumReleaseAge` is not set in the repo (only `minimumReleaseAgeExclude` in `pnpm-workspace.yaml`), so fresh SB releases resolve normally. Pin exact versions via the committed lockfile; `storybook doctor` documented for triage. `@varlock/vite-integration` is pinned in dependencies but referenced nowhere in `vite.config.ts` — pre-existing anomaly, flagged for separate cleanup (out of scope here).

## Goals vs. Non-Goals

**In-Scope Goals**
- Storybook installed in `app/` (framework `@storybook/react-vite`, builder auto-detected from `vite.config.ts`), features: `docs`, `a11y` (no `test` feature, no `ai`).
- Preview wired to the app's real stack: `import "../src/index.css"` + `withThemeByClassName` (light `''` / dark `'dark'`, default light).
- Stories convention: colocated `src/**/*.stories.tsx` (CSF), autodocs via docs addon; strict-TS compatible (`_`-prefix unused args).
- Token harness: token-palette stories + "Design System" docs pages reading CSS variables at runtime from `index.css` (no extra parser, no extra dep, auto light/dark).
- Init-generated example stories removed (no dead code).
- Gaps fixed so the new files are covered: `.storybook` added to `tsconfig.json` include; minimal eslint flat config added in `app/` (fixes pre-existing broken `pnpm lint`).
- Minimal frontend CI job (`frontend`): pnpm 11 + Node 22, `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm build`, `pnpm storybook build` — isolated from the Go pipeline.
- `app/README.md`: scripts (`storybook dev`/`storybook build`), `storybook doctor`, frozen-lockfile policy, troubleshooting.
- Repo docs updated (root `README.md` frontend section, `AGENTS.md` commands/structure, `memory-bank/` session update) — in French, matching existing documentation.

**Out-of-Scope Non-Goals**
- Open Design → code bridge (token export, design-tool sync): future task; this install only guarantees zero-rework landing.
- Interaction tests / vitest-Storybook integration (`test` feature, `@storybook/addon-vitest`): deferred phase.
- Chromatic or visual-regression hosting.
- Any product component creation (including shadcn Button) — pending Open Decision 1.
- Removal of `@varlock/vite-integration`.

## Alternatives Weighed

- **Defer Storybook until components exist (contrarian lens):** rejected by user decision; recorded as objection 6.
- **Lighter alternative (Ladle / react-catalog):** rejected — ecosystem momentum, docs/MDX, a11y and themes addons, and future vitest integration align with the design-system ambition.
- **Witness = shadcn Button + story:** rejected by user (throwaway, duplicates registry docs, couples early stories to base-vega component styling).
- **Empty "Design System" docs page:** rejected — dead placeholder; replaced by token-generated docs pages.
- **JS theme state instead of `.dark` class:** rejected — would test a mechanism the app does not use.

## Proposed Implementation (ordered)

1. In `app/`: `pnpm dlx storybook@latest init --builder vite --features docs a11y` (non-interactive). Expect the v11 line (min Vite 6.3; Vite 8 supported). Let init add `storybook`/`build-storybook` scripts and `.storybook/` scaffolding.
2. `.storybook/main.ts` (CSF 4, `defineMain`): stories globs `../src/**/*.mdx`, `../src/**/*.stories.@(ts|tsx)`; addons: docs, a11y, themes. Remove any init features not selected.
3. `.storybook/preview.tsx`: `import "../src/index.css"`; `addonThemes.withThemeByClassName({ themes: { light: '', dark: 'dark' }, defaultTheme: 'light' })`.
4. Delete init-generated example stories/stubs.
5. `tsconfig.json`: `include: ["src", "vite.config.ts", ".storybook"]`. Verify `pnpm build` (tsc --noEmit) passes with strict flags on config files.
6. Add `eslint.config.js` (flat): `@eslint/js` + `typescript-eslint` + `eslint-plugin-react-hooks` + `eslint-config-prettier`; confirm `pnpm lint` now succeeds on the whole `app/` tree including `.storybook/` and stories.
7. Token harness: `src/design-system/` — palette stories (colors, radius, typography) reading `getComputedStyle(document.documentElement)` variables at runtime; MDX docs pages "Introduction" + "Design Tokens" rendering from the same module. Both auto-reflect light/dark via the toolbar.
8. `app/README.md`: scripts, local dev URL (http://localhost:6006 via `pnpm storybook`; Vite app stays on its own port 5173), static output `storybook-static/` via `pnpm storybook build`, `storybook doctor`, frozen-lockfile installs, dark-mode toolbar note.
9. Repo documentation (French, matching existing docs): root `README.md` — frontend section (Storybook scripts, local URL, story conventions); `AGENTS.md` — Commands (add `pnpm storybook` / `pnpm storybook build` / `pnpm lint` in `app/`) and Structure (describe `app/.storybook/`, `src/design-system/`, colocated stories convention); `memory-bank/` — update `activeContext.md`, `techContext.md`, `progress.md` at session end (repo convention: memory bank updated each session).
10. CI: add a `frontend` job to `.github/workflows/ci.yml` — `runs-on: ubuntu-latest`; `pnpm/action-setup` (pnpm 11) + `actions/setup-node` (Node 22, pnpm cache on `app/pnpm-lock.yaml`); `working-directory: app`; steps: `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm build`, `pnpm storybook build`. The Go jobs stay untouched (no interaction with `make test`, RLS checks, legacy-driver guards).
11. Validation (see below). Commit `package.json` + lockfile together.

## Decisions (resolved or open)

1. **RESOLVED — Witness: token palette (user-confirmed).** No shadcn Button is created; the palette stories + token docs pages are the sole validation witness.
2. **RESOLVED — Docs pages generated from CSS variables (user-confirmed).** No empty placeholder page.
3. **RESOLVED — Pre-existing gaps fixed in this task (user-confirmed):** eslint flat config added in `app/` (repairs the currently broken `pnpm lint`), `.storybook` added to `tsconfig.json` include.
4. **RESOLVED — CI: minimal frontend smoke job (user-confirmed).** A dedicated `frontend` job is added to `.github/workflows/ci.yml`, independent of the Go pipeline.

## Risks & Mitigations

- **Token swap at Open Design integration:** stories/docs reference CSS variables only (never hardcoded hex/classes tied to base-vega names beyond the variables themselves) → future swap = edit `index.css` variables in one place.
- **Dep-tree growth (~hundreds of transitive packages):** dev-only, lockfile-pinned; accepted cost, documented.
- **Story drift:** no interaction tests yet; mitigated by the docs-page contract and (if adopted) the CI build smoke. Deeper enforcement deferred with the `test` feature.
- **Strict TS on story patterns:** use `_`-prefixed args; autodocs typing comes from CSF types.

## Validation Plan

- `pnpm lint` and `pnpm build` pass in `app/` (now covering `.storybook/` and stories).
- `pnpm storybook build` completes without errors (smoke).
- Manual: `pnpm storybook` → toolbar dark/light toggles flip `.dark` on the story container (palette visibly switches), docs pages render, a11y panel runs on the palette stories.
- `git status` review: no `.env`/secrets touched; lockfile committed.
- CI: `frontend` job green on push (frozen install, lint, build, storybook build); Go jobs unaffected.

## Key Decisions & Staging for Next Steps

- Decisions carried forward: colocated CSF stories; `.dark`-class theme toggle; token-contract anchoring; English/plain artifact language.
- Future phases (separate tasks): Open Design bridge (token export pipeline), Storybook vitest integration (`--features test` equivalent), optional Chromatic, `@varlock` cleanup.
