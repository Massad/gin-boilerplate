# Consolidated Review Report

**Target:** `.kilo/plans/1795428465642-login-simple.md` — « Vue de connexion simple (sans SSO) » (235 lines)
**Mode:** Consolidated report with all review rubrics — uninterrupted full flow (Stages 5-9, no checkpoints)
**Semantic split (at review time):** `## Detailed Implementation Plan` at line 155 — FeatureRange `1-154`, DetailedRange `155-235`. After same-day author remediation the file has 249 lines with the boundary at line 163; line references in the findings refer to the reviewed 235-line version.
**GoldfishContext:** none (user instruction). The document contains no Markdown links; the maquette path outside the working directory and the inline reference to `1795428465642-stack-test-go.md` were not included.
**Date:** 2026-11-23

## Verdict

The plan survives the review structurally, factually, and at high altitude. Structure fully complies with the feature-spec standard; a naive reader paraphrased it accurately with zero comprehension findings; the initial zoom-out found no over-engineering, YAGNI violations, or dead ends; and the strongest result — a very thorough codebase audit verified every current-state claim in the document as true against the actual code (zero contradictions, zero non-existent references). The actionable work before implementation: pin the `api.ts` error model and 401-refresh mechanics, add the login-card navigation links to item 15, resolve the enumeration-timing contradiction in the unified-login design together with the overlapping-email policy, bring the risk treatment into the feature spec range, fix the Non-Goals wording, and add explicit phase checkpoints. The implementation-readiness reviewer's 28 raw findings reduce, after deduplication and evidence-based correction, to 2 Blockers and 2 Majors plus one rebutted class; everything else is editorial. Post-review addendum (same day): the user provided the maquette; the Stage 9 Blocker C-1 is resolved — a targeted re-audit verified every maquette-dependent claim — and two new Majors (C-12, C-13) surfaced from the maquette scan. All Blockers and Majors are now resolved by same-day author remediation (C-1 via the provided maquette context; C-2, C-3, C-12, C-13 by explicit decisions; C-4..C-7 and nits C-8..C-10 by approved batch edits). The plan is implementation-ready; the only remaining open item is the optional maquette-detail checklist (C-14).

## Reviewer Coverage (11 reviewer instances, all completed)

| Stage | Reviewer | Result |
| --- | --- | --- |
| 5 — Initial Zoom Out | `egm-zoom-out` | No findings |
| 5 — Initial Zoom Out | `egm-inventive-catalyst` | 1 Major (advisory) |
| 6 — Comprehension Test | `egm-comprehension` | Paraphrase only, no findings |
| 7 — Critic Review | `egm-design` | 2 Major, 1 Nit |
| 7 — Critic Review | `egm-structural` | No findings |
| 7 — Critic Review | `egm-style` | 1 Major, 6 Nit (spec range) |
| 8 — Intermediate Zoom Out | `egm-zoom-out` | 3 Major, 2 Nit — all confirmations of Stage 7, nothing new |
| 9 — Implementation Readiness | `egm-implementation-readiness` | 28 raw findings → 2 Blocker, 2 Major, 1 rebutted class (see Corrections) |
| 9 — Implementation Readiness | `egm-codebase-details` | No findings — all current-state claims verified true |
| 9 — Implementation Readiness | `egm-style` | 6 Nit (detailed range), no Blocker/Major |
| Post-review re-audit | `egm-codebase-details` | Maquette provided as context: all 7 fidelity claims verified, C-1 resolved; 2 new Major, 6 new Nit |

## Consolidated Action Items

### 🔴 Blockers

**C-1 — Maquette not reachable from the repository.**
Items 15/19/20 (plan lines 188, 192-193) require fidelity to `login.html`, which lives outside the working tree (line 6) and is neither committed nor excerpted. Without it, `AuthLayout`, `LoginPage`, and the `atrium.css` styles cannot be implemented as specified.
*Action:* commit the maquette or an export of it (four views' structure + `.shell/.brand-*/.mobile-brand/.form-card` styles), or state in the plan how the implementer receives it.
*Status: **RESOLVED** (post-review, 2026-11-23 — see Addendum).* The user supplied the maquette project path; a fresh `egm-codebase-details` re-audit verified all seven maquette-dependent claims against `login.html` with line evidence — no contradictions, no non-existent references. Residual recommendation: copy `login.html` into the repository (e.g. `docs/maquettes/`) so the reference no longer depends on the machine-local AppData path.

**C-2 — `api.ts` error model and 401-refresh integration point unspecified.**
The plan pins the refresh behavior (lines 90, 104, 224) but not the wrapper's exported API, its error contract, where 401 is intercepted, or how concurrent 401s serialize the single refresh attempt. CUJ 2 depends on this.
*Action:* add a short interface spec to the plan (exported request function, error type, interception location, single shared refresh promise, replay-once semantics, failure → purge + `/login`).
*Status: **RESOLVED by author decision** (2026-11-23).* The api.ts-centric throwing design was adopted: `request(path, opts)` returns parsed JSON, non-2xx throws `ApiError{status, body}`; api.ts intercepts tenant 401s with a single shared refresh promise (`/v1/token/refresh`, tokens replaced via `session.ts`, replay once; failure or admin → purge + `/login`). Plan lines 88, 90, 104, 184 (item 11), 186 (item 13), 196 (item 23), and 224 updated.

### 🟡 Major

**C-3 — Unified-login fallback: enumeration timing vs. the "pas d'énumération" claim, and the overlapping-email lockout.**
The sequential semantics « le chemin user prime ; l'admin n'est tenté que si le chemin user échoue » (lines 72-75, 222) create: (a) a timing side channel — an absent user email skips bcrypt and responds measurably faster than an existing-user-with-wrong-password case, leaking tenant-user registration status and contradicting lines 25 and 63; (b) a lockout — an email present in both tables with different passwords can never authenticate the admin path (Stage 5 catalyst, advisory). Raised by `egm-design` (Stage 7), confirmed by `egm-zoom-out` (Stage 8), first flagged by `egm-inventive-catalyst` (Stage 5).
*Action:* either mitigate (dummy bcrypt on user miss, or parallel constant-time evaluation of both paths — declared cost: doubled bcrypt/lookups, constant-time requirement, more complex error merging) or retract the enumeration-resistance claim and document the residual timing risk; separately, either document an explicit no-overlap policy for admin emails or accept the lockout explicitly.
*Status: **RESOLVED by author decision** (2026-11-23).* Dummy bcrypt on user miss adopted — unknown-email and wrong-password attempts now cost the same, so the « pas d'énumération » claim holds for body, status, and time — and the overlapping-email lockout is accepted with an explicit no-overlap policy (a superadmin email never matches a tenant user email). Plan lines 25, 63, 73-75, 149, 178 (item 9), and 222 updated. This also closes the Stage 5 inventive-catalyst advisory.

**C-4 — No "Risks & Mitigations" treatment inside the feature spec range.**
The feature-spec standard requires failure modes, security implications, scalability limits, and mitigations within the spec; the spec's NFR block (lines 147-151) lists security measures without framing them as failure modes, and the only dedicated risks content sits at lines 219-225 in the detailed plan. (`egm-structural` verified section presence at header level; the gap is substantive, not structural.)
*Action:* add a Risks & Mitigations section to lines 1-154 (candidate entries: fallback timing, Redis unavailability, token theft via `localStorage`/XSS, scalability limits).
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* A `## Risks & Mitigations` section was added to the spec (post-remediation lines 153-159): temporal enumeration (dummy bcrypt), Redis unavailability, token theft via browser storage, email-overlap policy, scalability limits.

**C-5 — Contradictory Non-Goals label: « Vues secondaires fonctionnelles ».**
Line 35 labels the views "fonctionnelles" while stating they « restent statiques »; as written it can be read as calling the views functional. Raised by `egm-style`, confirmed by `egm-zoom-out`.
*Action:* reword, e.g. « Rendre fonctionnelles les vues secondaires : … ».
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* The Non-Goals bullet was reworded: « Fonctionnalité backend des vues secondaires : … restent affichées et commutables côté client, mais inertes à la soumission … » (post-remediation line 36).

**C-6 — Local interface decisions left open in new frontend files.**
`session.ts` exported API (line 185), home-pages session consumption (lines 190-191), zod schema scope — « (email, requis) » does not say whether password is client-enforced (line 188), vitest setup content (line 195), and the final `test` script wording in item 24 (line 197: `vitest run` vs. `--passWithNoTests` is ambiguous).
*Action:* one-line interface notes per module; state the zod scope and the final `test` script verbatim. (Synthesized down from raw Blocker severity: local, low-risk choices.)
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* Interface notes added: `session.ts` exports `readSession()` / `writeSession(session, { remember })` / `purgeSession()` (item 12); `AuthContext` exposes `useAuth()` (item 13); router guards read `readSession()` (item 14); zod scope = email + password required (spec line 91, item 15); home pages consume `useAuth()` (items 16-17); vitest setup specified (item 22); final scripts `"test": "vitest run"` / `"e2e": "playwright test"` with `--passWithNoTests` removed (item 24).

**C-7 — No explicit phases or checkpoints (synthesizer observation, not a reviewer finding).**
The implementation-readiness standard requires incremental phases that each leave the tree buildable and testable; the plan is a flat 31-item list with only an end-state Verification Plan.
*Action:* add phase gates, e.g. Phase 1 backend (items 1-10): `go build`, `go vet`, `make test` green → Phase 2 frontend (items 11-23): `pnpm lint`, `pnpm build`, `pnpm test` green → Phase 3 e2e + docs (items 24-31): `make e2e` green.
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* A `### Phases & points de contrôle` section was added to the detailed plan (post-remediation lines 217-221): Phase backend (items 1-10) gated by `make generate_sqlc`, `make generate_docs`, `go build`, `go vet`, `make test`; Phase frontend (items 11-23) gated by `pnpm lint`, `pnpm build`, `pnpm test`, `pnpm build-storybook`; Phase e2e + docs (items 24-31) gated by `make e2e`.

**C-12 — Login card's « Mot de passe oublié ? » link undocumented.** (post-review re-audit)
Maquette lines 729-735: the login card's `.row-between` row holds the « Rester connecté » checkbox and a « Mot de passe oublié ? » link (`data-od-id="forgot-link"`, line 734) that switches to the forgot view. Plan item 15 lists the four views but not this trigger link or its placement.
*Action:* add to item 15: link « Mot de passe oublié ? » beside « Rester connecté », switches to view `forgot`.
*Status: **RESOLVED by author decision** (2026-11-23).* The link is displayed as in the maquette and switches to the `forgot` view — frontend-only, no backend. « Rester connecté » is confirmed displayed with its planned frontend-only storage role (localStorage vs sessionStorage — no backend involvement). Plan spec (line 91) and item 15 updated.

**C-13 — Login card footer entry links undocumented.** (post-review re-audit)
Maquette lines 751-754: the login card's `.form-foot` holds « Demander un accès » (`data-od-id="signup-link"`) and « Créer une organisation » (`data-od-id="tenant-help-link"`) — the entry points to two of the four views. Plan item 15 does not document how users reach the `request` and `create` views from the login screen.
*Action:* add to item 15: footer links switch to views `request` and `create`.
*Status: **RESOLVED by author decision** (2026-11-23).* Both footer links are displayed as in the maquette and switch to the `request` / `create` views — frontend-only, no backend. Plan spec (line 91) and item 15 updated.

### 🟢 Nit

**C-8 — Token lifetimes without rationale.** Lines 20, 28: 15 minutes / 8 hours stated without threat-model or policy basis. Document the rationale.
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* Rationales documented: the 15-minute access token bounds the exploitation window of a stolen token (line 20); the 8-hour admin session is a workday-length platform session with admin refresh out of scope (line 28).
**C-9 — Editorial nits in the spec range (6).** Lines 12 (« supprimée puis recréée »), 20 (« casse » + repeated expiry), 24 (« fidèlement … complète »), 37 (« E2E dans la CI (local uniquement), CI e2e »), 69 (« aucune politique à ajouter »), 91 (« comme la maquette » filler).
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* All six copy edits applied.
**C-10 — Editorial nits in the detailed range (6).** Lines 170 (redundant parenthetical), 184 (« base de tous les appels »), 185 (« pure, testable » — subjective and in tension with storage side effects), 188 (« fidèle à `login.html` » — state factually what is reproduced), 221 (CALIBRATE: absolute « ne peut pas être supprimée par erreur » → describe the guard mechanism), 229-231 (« **The idea** : » ×3 — English repetition in a French document).
*Status: **RESOLVED** (author-approved batch, 2026-11-23).* All six copy edits applied, including the CALIBRATE on the `DROP DATABASE` guardrail and the « **Idée** : » relabeling.
**C-11 — "Plan does not inline current file contents" (16 raw readiness findings, rebutted as a class).** Reading existing repository files is execution, not a design choice; the plan's per-file changes are precise, and `egm-codebase-details` verified all current-state descriptions match the actual code. Residual value only: inline current signatures for a fully self-contained handoff.
**C-14 — Maquette details the plan omits (grouped, post-review re-audit).** Sub-components of documented panes (`.brand-mark`, `.form-pane`, `.form-head` — maquette lines 112, 185, 195); submit-button spinner + `aria-busy` (lines 436-467, 738-741); `.field-help` texts beyond the slug preview (line 867); view-switch document-title change and first-input focus (lines 977-997); terms checkbox in the create view (lines 807-810 — inert per the plan's static-views non-goal). The maquette also ships unused tenant-selector CSS (lines 210-320, no rendered HTML — ignore). *Action:* optional fidelity checklist additions; `login.html` remains the reference.

## Traceability Matrix

| Consolidated | Priority | Source stages | Reviewer(s) | Raw references | Plan locations |
| --- | --- | --- | --- | --- | --- |
| C-1 | ✅ resolved (re-audit) | 9 + re-audit | egm-implementation-readiness; egm-codebase-details | raw #13/#17/#18 (maquette aspect); re-audit: 7/7 claims verified | items 15/19/20; line 6 |
| C-2 | ✅ resolved (author decision) | 9 | egm-implementation-readiness | raw #9/#11; adopted design: api.ts-centric throwing | items 11/13; lines 88-90, 104, 184, 186, 196, 224 |
| C-3 | ✅ resolved (author decision) | 5, 7, 8 | egm-inventive-catalyst; egm-design; egm-zoom-out | Stage 5 catalyst finding; R7-2; Stage 8 confirmation; adopted: dummy bcrypt + accepted lockout | lines 25, 63, 73-75, 149, 178, 222 |
| C-4 | ✅ resolved (author decision) | 7, 8 | egm-design; egm-zoom-out | R7-1; Stage 8 confirmation; Risks & Mitigations section added | spec, post-remediation lines 153-159 |
| C-5 | ✅ resolved (author decision) | 7, 8 | egm-style; egm-zoom-out | R7-3; Stage 8 confirmation; reworded | post-remediation line 36 |
| C-6 | ✅ resolved (author decision) | 9 | egm-implementation-readiness | raw #10/#12/#13(zod)/#14/#15/#20/#22; interface notes applied | items 12-17, 22, 24 |
| C-7 | ✅ resolved (author decision) | 9 | synthesizer observation | R9-5; phase gates added | detailed plan, Phases section |
| C-8 | ✅ resolved (author decision) | 7, 8 | egm-design; egm-zoom-out | R7-4; Stage 8 confirmation; rationales applied | lines 20, 28 |
| C-9 | ✅ resolved (author decision) | 7 | egm-style | R7-5 (6 nits); copy edits applied | lines 12, 20, 24, 37, 69, 91 |
| C-10 | ✅ resolved (author decision) | 9 | egm-style | R9-6 (6 nits); copy edits applied | items 5, 11, 12, 15; guardrails; Future Improvements |
| C-11 | 🟢 | 9 | egm-implementation-readiness | R9-4 (16 raw findings) | multiple Modify/Regen items |
| C-12 | ✅ resolved (author decision) | re-audit | egm-codebase-details | re-audit finding 1 | item 15; maquette lines 729-735 |
| C-13 | ✅ resolved (author decision) | re-audit | egm-codebase-details | re-audit finding 2 | item 15; maquette lines 751-754 |
| C-14 | 🟢 | re-audit | egm-codebase-details | re-audit findings 3-9 | items 15/19/20 |

## Corrections & Rebuttals (synthesizer, with document evidence)

1. **Down-migration SQL:** raw readiness finding "item 1 does not provide the down migration SQL" is factually wrong — line 69 specifies « down : `ALTER TABLE admin_user DROP COLUMN name` ». Dropped.
2. **e2edb safety guards:** raw finding "omits safety guards" is factually wrong — the guard (refusal when the target name is empty or equal to `DB_NAME`) is specified at lines 110, 203, and 221. Dropped.
3. **Test conventions:** raw finding "does not detail those conventions" overreaches — the conventions are documented in-repo (AGENTS.md "Tests" section; `tests/article_test.go`) and the plan enumerates the required test cases (line 178, line 196).
4. **Severity synthesis:** R9-3 (C-6) downgraded from raw Blocker to Major (local, low-risk choices); R9-4 (C-11) downgraded from raw Blocker to Nit (reading existing files is not a design choice; accuracy verified by `egm-codebase-details`). Rationale recorded in the stage report.
5. **Reviewer numbering drift:** the readiness reviewer labels plan items 27-31 as 26-30; this report uses the plan's own numbering. Plan item 26 (`app/e2e/login.spec.ts`) received no raw finding.
6. **Stage 5 zoom-out terseness:** the one-line "No findings" result was accepted as a valid zero-findings report; the Stage 8 zoom-out re-ran the same lens on the full file with prior reports in context and raised no new altitude findings, serving as a cross-check.

## Post-Review Addendum (2026-11-23): Maquette Context Provided

The user supplied the maquette project path (`...\dc409479-487d-4fdb-be25-8351ec3976e6`) — containing `login.html` (1,158 lines), the exact file referenced at plan line 6 — as approved review context. Per the one-re-audit allowance, a fresh `egm-codebase-details` reviewer re-audited the maquette-dependent findings with the file in context:

- **C-1 resolved.** All seven maquette-dependent claims verified with line evidence: the 4 view cards (`login-card` line 687, `createOrgCard` 758, `requestAccessCard` 824, `forgotCard` 890) and their `showView()` switching (lines 977-997); the two-panel `.shell` grid (line 80) with `.brand-pane`/`.form-pane`; the responsive collapse at 680px with `.mobile-brand`; the SSO block and « ou continuer avec » divider (lines 744-748) to hide; the error banner text as an exact prefix of maquette line 702; the « Rester connecté » checkbox (lines 730-733); the slug preview (lines 785, 1073-1085); and all six documented class names (`.shell` 80, `.brand-pane` 90, `.brand-hero` 135, `.brand-quote` 153, `.mobile-brand` 545, `.form-card` 191). The password toggle also exists in the maquette (`.pwd-toggle`, lines 371-385) — the plan sources it from the existing `PasswordInput` component, so no defect.
- **Two new Majors (C-12, C-13):** the login card's « Mot de passe oublié ? » link and the footer « Demander un accès »/« Créer une organisation » entry links — material navigation elements item 15 omits.
- **Six grouped Nits (C-14):** minor maquette details the plan does not enumerate (sub-components, spinner, help texts, title/focus behaviors, terms checkbox, unused tenant-selector CSS).
- **Author remediation (same day, second pass):** the user confirmed that « Mot de passe oublié ? », « Rester connecté » and the footer « Demander un accès »/« Créer une organisation » links must all be displayed with frontend-only behavior — view switching and the storage choice, no backend work (consistent with the existing non-goals: static secondary views, no register wiring). Plan spec (line 91) and item 15 updated accordingly; C-12 and C-13 resolved.
- **Open items after the author's first remediation:** 1 Blocker (C-2), 5 Majors (C-3, C-4, C-5, C-6, C-7), Nit groups C-8..C-11 and C-14.
- **Author remediation (third pass, same day):** the user adopted the api.ts-centric throwing design for C-2 — `request(path, opts)` returns parsed JSON, non-2xx throws `ApiError{status, body}`; api.ts intercepts tenant 401s with a single shared refresh promise (`/v1/token/refresh`, tokens replaced via `session.ts`, replay once; failure or admin → purge + `/login`). Plan lines 88, 90, 104, 184 (item 11), 186 (item 13), 196 (item 23), and 224 updated. C-2 resolved.
- **Open items now:** 5 Majors (C-3, C-4, C-5, C-6, C-7), Nit groups C-8..C-11 and C-14.
- **Author remediation (fourth pass, same day):** the user adopted the dummy-bcrypt resolution for C-3 — on an unknown user email, `UnifiedLogin` runs bcrypt against a fixed dummy hash before falling through to the admin path (constant cost, no temporal enumeration); the sequential semantics are kept, and the overlapping-email lockout is accepted with an explicit no-overlap policy (superadmin emails never match tenant user emails). Plan lines 25, 63, 73-75, 149, 178 (item 9), and 222 updated. C-3 and the Stage 5 catalyst advisory resolved.
- **Open items now:** 4 Majors (C-4, C-5, C-6, C-7), Nit groups C-8..C-11 and C-14.
- **Author remediation (fifth pass, same day):** the user approved the full batch — C-4 (new `## Risks & Mitigations` spec section), C-5 (Non-Goals reword), C-6 (interface notes), C-7 (phase gates), C-8 (token-lifetime rationales), C-9/C-10 (12 editorial copy edits). **All Blockers and Majors are resolved; the plan is implementation-ready.** The remediated plan has 249 lines with the `## Detailed Implementation Plan` boundary at line 163 (FeatureRange 1-162, DetailedRange 163-249). Line references in the findings above refer to the reviewed 235-line version. Remaining open: the optional maquette-detail checklist (C-14, advisory only) and the informational C-11 note.

## Stage Reports

- `.kilo/plans/zoom_out_report.md` (Stage 5)
- `.kilo/plans/comprehension_report.md` (Stage 6)
- `.kilo/plans/critic_report.md` (Stage 7)
- `.kilo/plans/intermediate_zoom_out_report.md` (Stage 8)
- `.kilo/plans/implementation_readiness_report.md` (Stage 9)
