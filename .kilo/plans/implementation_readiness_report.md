# Goldfish Review Report — Stage 9 (Implementation Readiness Test)

**Target:** `.kilo/plans/1795428465642-login-simple.md` (`155-235`)

## Reviewer Status

| Reviewer | Status |
| --- | --- |
| `egm-implementation-readiness` | completed (28 raw findings) |
| `egm-codebase-details` | completed (no findings) |
| `egm-style` | completed (6 Nit findings) |

## Synthesis Note

The implementation-readiness reviewer applied the zero-decision rule (implementation-readiness-standard.md) at its strictest: 27 of its 28 raw findings are Blocker-severity, most repeating one pattern — the plan does not inline the current content of the files it modifies, so the implementer must read them. The synthesizer deduplicated the 28 raw findings into the grouped items below, corrected factual errors against the document text, and recorded one standard requirement no reviewer flagged (R9-5). The reviewer's item numbering drifts by one for plan items 26-31 (it labels `cmd/e2edb` as item 26; the plan numbers it 27), and plan item 26 (`app/e2e/login.spec.ts`, line 199) received no raw finding — its scenarios are specified by CUJ references.

## Prioritized Findings (Blockers -> Major -> Nits)

### R9-1 — UI fidelity work depends on a maquette that is not reachable from the repository — Severity: Blocker

* **Reviewer Lens:** `egm-implementation-readiness` (raw findings on plan items 15, 19, 20)
* **Affected Scope:** Plan items 15 (`LoginPage.tsx`, line 188), 19 (`AuthLayout.tsx` + stories, line 192), 20 (`atrium.css`, line 193); maquette reference at line 6
* **Discrepancy Description:** The plan requires fidelity to the maquette (« fidèle à `login.html` », styles « transposés de la maquette », « volets marque/formulaire + barre mobile ») but the maquette lives outside the repository (`C:\Users\schat\AppData\Roaming\Open Design\...login.html`, line 6) and is included neither in the plan nor in the working tree. An implementer who cannot open that file cannot execute items 15/19/20 correctly on the first attempt; the component structure and CSS to transpose exist only in that file.
* **Rationale & Recommendation:** Commit the maquette (or an export of it: the four views' structure and the `.shell/.brand-*/.mobile-brand/.form-card` styles) into the repository or attach it to the plan; alternatively state explicitly how the implementer receives it.

### R9-2 — api.ts error model and 401-refresh integration point left to the implementer — Severity: Blocker

* **Reviewer Lens:** `egm-implementation-readiness` (raw findings on plan items 11, 13)
* **Affected Scope:** Plan item 11 (`app/src/lib/api.ts`, line 184), item 13 (`AuthContext.tsx`, line 186); behavior pinned at plan lines 90, 104, 224
* **Discrepancy Description:** The plan pins the refresh behavior (tenant: one attempt via `/v1/token/refresh`, tokens replaced, replay once; admin: purge → `/login`) but not the mechanism: the wrapper's exported API, its error contract (throw vs. typed result, error shape), where the 401 is intercepted, and how concurrent 401s serialize the single refresh attempt. These are design choices under the standard's zero-decision rule, and CUJ 2's silent-refresh journey depends on them.
* **Rationale & Recommendation:** Add a short interface spec, e.g.: « `api.ts` exporte `request(path, opts)` ; non-2xx lève `ApiError{status, body}` ; l'interception 401 vit dans `api.ts`, attend un refresh unique partagé via `/v1/token/refresh`, remplace les tokens via `session.ts`, rejoue une fois, sinon purge + `/login` » — or equivalent.

### R9-3 — Local interface decisions left open in new frontend files — Severity: Major

* **Reviewer Lens:** `egm-implementation-readiness` (raw findings on plan items 12, 14, 16, 17, 22, 24, and the zod scope of item 15)
* **Affected Scope:** Plan items 12 (`session.ts`, line 185), 14 (`router.tsx`, line 187), 15 (zod schema scope, line 188), 16/17 (home pages, lines 190-191), 22 (`app/src/test/setup.ts`, line 195), 24 (`app/package.json`, line 197)
* **Discrepancy Description:** Genuinely open but local choices: the exported function signatures of `session.ts`; how the home pages consume the session (context hook vs. direct read); the zod validation scope — « validation zod (email, requis) » does not state whether the password is also enforced client-side; the exact content of the vitest setup file; and item 24's wording — « `test` devient `vitest run` (le setup existant `--passWithNoTests` reste tolérant avant ajout des tests) » — is ambiguous about the final script content (`vitest run` with zero tests fails without `--passWithNoTests`).
* **Rationale & Recommendation:** One-line interface notes per module; state the zod schema scope and the final `test` script verbatim. Synthesis note: raw severity was Blocker; synthesized to Major because these choices are local and low-risk, unlike R9-1/R9-2.

### R9-4 — "Plan does not inline current file contents" class — Severity: Nit (rebutted as design gaps)

* **Reviewer Lens:** `egm-implementation-readiness` (16 raw findings: plan items 2, 4, 5, 6, 7, 8, 9, 18, 21, 23, 25, 27, 28, 29, 30, 31)
* **Affected Scope:** All Modify/Regen items and convention-referencing items across lines 161-207
* **Discrepancy Description (as reported):** For Modify/Regen items the plan does not reproduce the current content/signatures of the touched files, so the implementer "must inspect" them.
* **Synthesis & Rebuttal:** The standard's zero-decision rule prohibits design choices — choosing schemas, API paths, dependencies, or error-handling strategies — not reading existing repository files. The plan states the precise change per file (satisfying the standard's "Actionable Steps: precise logic changes, data transformations, and API updates for each file"), and `egm-codebase-details` verified in this same stage that every current-state claim matches the actual code, so the implementer's reading will match the plan's assumptions. Three raw claims are rebutted with document evidence: the down-migration SQL IS specified (line 69: « down : `ALTER TABLE admin_user DROP COLUMN name` » — raw finding on item 1 is factually wrong and dropped); the e2edb safety guards ARE specified (lines 110, 203, 221: refusal when the target name is empty or equal to `DB_NAME`); and the integration-test conventions ARE documented in-repo (AGENTS.md "Tests" section, `tests/article_test.go`, with the plan enumerating the required cases at line 178). Residual value: for a fully self-contained handoff, current signatures could be inlined — recorded as Nit.

### R9-5 — Phasing & Checkpoints requirement of the standard unmet — Severity: Major (synthesizer observation)

* **Reviewer Lens:** synthesizer (no reviewer raised it; recorded for the consolidated report)
* **Affected Scope:** Whole detailed-plan structure (lines 157-235)
* **Discrepancy Description:** The implementation-readiness standard requires partitioning the work into logical, incremental phases, each leaving the codebase in a stable, buildable, testable state. The plan is a flat 31-item list with implicit ordering (backend → frontend → tools/CI) and an end-state Verification Plan (lines 209-217); no intermediate checkpoints are defined.
* **Rationale & Recommendation:** Add explicit phases with gates, e.g.: Phase 1 backend (items 1-10) gated by `go build`, `go vet`, `make test` green; Phase 2 frontend (items 11-23) gated by `pnpm lint`, `pnpm build`, `pnpm test`; Phase 3 e2e harness + docs (items 24-31) gated by `make e2e`.

### R9-6 — Editorial nits in the detailed plan — Severity: Nit

* **Reviewer Lens:** `egm-style`
* **Affected Scope & Style Rationale (6 findings, line references verified):**
  - Line 170 (item 5, `models/login.go`): « Aucune requête SQL hors `db/queries` (aucune requête nouvelle nécessaire) » — synonym redundancy; the parenthetical repeats the clause.
  - Line 184 (item 11, `api.ts`): « base de tous les appels » — mild filler for a fetch wrapper.
  - Line 185 (item 12, `session.ts`): « pure, testable » — subjective quality judgment about unwritten code, and « pure » is in tension with the described localStorage/sessionStorage side effects.
  - Line 188 (item 15, `LoginPage.tsx`): « fidèle à `login.html` » — subjective, unverifiable from the plan; state factually what is reproduced (the four views and their switching).
  - Line 221 (Risques & garde-fous): « la base applicative ne peut pas être supprimée par erreur » — absolute, uncalibrated claim restating the guard; CALIBRATE to the mechanism (refus si vide ou égal à `DB_NAME`).
  - Lines 229-231 (Future Improvements): « **The idea** : » — identical English phrase opening all three bullets in a French document; use French and drop the repetition.
* **Rationale & Recommendation:** Apply the copy edits; none change meaning.

## Zero-Finding Lenses

* `egm-codebase-details` (very thorough audit against the actual codebase): "No findings. After auditing the detailed implementation plan (lines 155-235) against the codebase, all current-state claims in the document were verified accurate, and no contradictions or non-existent references were found. The plan's descriptions of existing files match their actual current content." This covers the existing routes (`/v1/user/login`, `/v1/admin/login`, `/v1/token/refresh`), the `admin_user` schema, `db/queries/admins.sql`, `models/admin.go`, the `-create-admin` CLI and `main_test.go`, `tests/article_test.go`, the design-system components, the TanStack Router and test dependencies in `app/package.json`, `App.tsx`, `vite.config.ts`, the Makefile targets, the CI frontend job, `AGENTS.md`, and the memory-bank files.
* `egm-style` reported no Blocker or Major findings in the detailed range: "The active scope is objective, concise, and well structured."

## Post-Review Re-Audit Addendum (2026-11-23)

The user provided the maquette project path (the `login.html` referenced at plan line 6) as approved review context. Per the one-re-audit allowance, a fresh `egm-codebase-details` reviewer (Mode: REVIEW, artifact = `login.html`, 1,158 lines) re-audited the maquette-dependent findings:

* **R9-1 is RESOLVED:** the maquette is reachable and every maquette-dependent claim verifies against the file with line evidence — no contradictions, no non-existent references. Residual recommendation: copy `login.html` into the repository so the reference does not depend on the machine-local AppData path.
* **Two new Major findings:** login-card navigation elements omitted from item 15 — the « Mot de passe oublié ? » link beside « Rester connecté » (maquette lines 729-735) and the footer « Demander un accès »/« Créer une organisation » entry links (lines 751-754). Six grouped Nits: `.brand-mark`/`.form-pane`/`.form-head` sub-components, spinner + `aria-busy`, `.field-help` texts, view-switch title/focus behaviors, terms checkbox, unused tenant-selector CSS. Recorded in the consolidated report as C-12, C-13, C-14.
* **Author decision (2026-11-23):** the login-card links (« Mot de passe oublié ? », footer « Demander un accès »/« Créer une organisation ») and « Rester connecté » are confirmed displayed with frontend-only behavior (view switching, storage choice — no backend). Plan spec (line 91) and item 15 updated; C-12 and C-13 resolved in the consolidated report.
* **Author decision (2026-11-23, second):** C-2 (R9-2) resolved — the api.ts-centric throwing design was adopted (`request(path, opts)`, `ApiError{status, body}`, tenant 401 interception with a single shared refresh promise, replay once, purge + `/login` on failure). Plan lines 88, 90, 104, 184, 186, 196, and 224 updated.
* All other Stage 9 findings stand unchanged.
