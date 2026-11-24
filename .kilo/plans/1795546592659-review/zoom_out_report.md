# Stage 5 — Initial Zoom Out Review

Target: `.kilo/plans/1795546592659-artisan-test-db.md` (full file, 394 lines).
Reviewers: `egm-zoom-out` (standard: zoom-out-standard) + `egm-inventive-catalyst` (Mode: REVIEW). Dispatched as direct subagents in parallel. GoldfishContext: empty (user opted out; document contains no markdown links — nothing auto-added).

---

## Part 1 — egm-zoom-out (standard: zoom-out-standard)

# Zoom-Out Review — plan `1795546592659-artisan-test-db.md`

## Verdict

The plan passes the Zoom-Out Standard. It targets the root problem (destructive suites sharing the dev database), solves it with existing infrastructure (`cmd/e2edb` reused as-is, zero new tooling), mirrors an already-proven pattern in this repo (the REDIS_DB pin via `TestMain`), and its rejected-alternatives table (lines 89-98) is genuine, not decorative. Grounding checks against the repo confirmed the plan's factual claims: `TestCleanUp` runs `DELETE FROM public.tenant` at tests/article_test.go:907-911, `TestIntDB` loads `.env` itself at :153 and asserts `Options().DB == 2` at :160, `pgdb.InitAdmin` is at :158 (T5's "l. 158" and "ctx l. 156" are exact), the controllers setup is at controllers/auth_test.go:39 (plan's "~l. 38", close enough), the Makefile `test` recipe is at :61-67 with the `e2e` pattern at :153-158, ci.yml has `DB_NAME: golang_gin_db` at :50 and the "Configure disposable test services" step at :64-74, and e2edb's package doc/usage/variable paragraphs match the cited line ranges (cmd/e2edb/main.go:1-4, 20-23, 24-29). No Blocker found.

## Standard Checks

- **0. Context sufficiency** — Pass. The objective (lines 39-44), the failure mode (lines 35-37), and the success criterion (lines 46-51) are explicit, with file-and-line anchors verified against the repo. A fresh reader can act on this document.
- **1. Root problem focus** — Pass. Moving the destructive actor onto a disposable database is a structural fix, not a band-aid (e.g., the alternative of gutting `TestCleanUp`'s DELETE would only patch the `tests` package). Aligned with the repo's established philosophy: disposable DBs recreated per cycle, fail-closed pinning, one tool as single source of truth.
- **2. Simplicity & extensibility** — Pass. No speculative abstractions; reuse of e2edb is verified real (its allowlist `^[a-z][a-z0-9_]{2,62}$` at cmd/e2edb/main.go:49 accepts `artisan_test`, and its guards need no change). No dead-ends: the doc-only generalization (T1) matches the tool's already-generic contract.
- **3. Anti-weeds** — Mostly pass. Exact code blocks in an implementation plan destined for an implementation agent are appropriate altitude, not pseudocode-weeds. Two duplication wrinkles flagged below.
- **4. Leverage existing** — Pass. `migrate_up` sub-make invocation, `_env` sourcing, and `-mod=readonly` all follow the working `e2e` recipe verbatim. The rejection of a new `cmd/testdb` (line 94) and of raw `psql` (line 96) is correct per repo conventions.

## Findings

**Finding 1 — Verbatim code blocks duplicated across tasks add plan bulk, not information**
- Severity: Nit
- Affected scope: T3 (lines 162-188) vs T4 (191-194); T5 (lines 202-206) vs T6 (210-219)
- Rationale: T4 and T6 repeat, copy for copy, what T3 and T5 already specify; only the package name and file path differ. A duplicated 26-line block is the plan-level equivalent of copy-paste code. The same enforcement is achieved with one canonical definition plus a reference.
- Fix direction: Reduce T4 to "identical to T3 for package `controllers` (file tests/main_test.go:11-19 today)" and T6 to "same three-line assertion as T5, placed after `pgdb.InitAdmin` at controllers/auth_test.go:39". The implementation agent loses nothing.

**Finding 2 — The `current_database()` ownership assertion is duplicated in two packages with no stated reason against the existing helper home**
- Severity: Nit
- Affected scope: Design table line 70, T5/T6 (lines 196-219), Hors-scope line 347 (`internal/testutil/**` excluded)
- Rationale: The repo already routes shared test logic through `internal/testutil/` (fixtures, factories). Inline copies of the same assertion in `tests` and `controllers` are a small placement deviation for two call sites — defensible today (two users don't justify the abstraction), but the plan silently excludes `internal/testutil/**` instead of stating why. The "right placement" question is therefore unanswerable by a fresh reader.
- Fix direction: State in the Hors-scope line that the helper is intentionally deferred until a third pin package appears (YAGNI), or, alternatively, relocate the assertion to `internal/testutil` as a one-function addition if the allowlist can absorb it. Either way, make the trade-off explicit rather than implicit.

**Finding 3 — Success criterion wording vs. residual dev-database connection: gap in stated justification, not in logic**
- Severity: Nit
- Affected scope: Objective lines 41-44, design table row line 72, Risks lines 339-341
- Rationale: The objective ("la base dev n'est plus écrite ni purgée") is technically accurate — `internal/db/pool_test.go` only connects (pgx.Connect, internal/db/pool_test.go:25,43), no writes — and the exclusion is acknowledged. But the plan never states the actual reason `internal/db` cannot be pinned: the CI guard `git diff --exit-code internal/db` (ci.yml:80-83) makes any edit to that directory a build failure. A fresh reader must infer this; without it, the exclusion reads as "assumé" rather than as a forced constraint, which weakens the document's otherwise excellent traceability.
- Fix direction: Add one clause to the line-72 row or the Risks bullet: "pool_test excluded because any change to `internal/db/**` fails the CI generated-code check; its connections are read-only."

**Finding 4 — `E2E_DB_NAME` variable semantics become misleading after the dual use**
- Severity: Nit
- Affected scope: T2 (lines 141-148), design table line 67, Hors-scope lines 349-350
- Rationale: After this plan, `E2E_DB_NAME` is both (a) the Makefile variable `?=` to `artisan_e2e` (Makefile:138) and (b) the environment key under which the `test` recipe passes `artisan_test` while overriding the variable entirely. The rename was rejected as cosmetic churn (line 350) — a def call — but the consequence (a variable named "e2e" carrying a non-e2e value in the test recipe) is never acknowledged in the plan. This is exactly the kind of accepted inconsistency that should be written down so a future reader doesn't "fix" it in a divergent direction.
- Fix direction: One sentence in the rejected-alternatives row or T1 text: "the `E2E_DB_NAME` key name now overloads 'e2e'; the rename stays rejected, the doc-only generalization (T1) is the chosen boundary."

**Finding 5 — CI change validated only statically; first real CI run is not attached to any milestone**
- Severity: Nit
- Affected scope: T7 line 238-239, Risks lines 336-338
- Rationale: The plan adds two CI lines that only a push can exercise ("validation statique du YAML — limite assumée"). Correct and consistent with previous plans, but the validation milestones (lines 285-320) nowhere mention that a future CI run is the real acceptance test for T7; if T7 is wrong, the failure surfaces outside this plan's lifecycle.
- Fix direction: Add to Phase 2 or the Risks bullet: "the next real CI run is the binding proof of T7; a red 'Configure disposable test services' step maps to exactly these two lines."

## Items verified as NOT defects (checked, dismissed)

- **REDIS_DB=2 flush inside `make test` provisioning** (line 68): index 2 is exclusively the Go-suites index; deterministic and consistent with the TestMain pins. Symmetric with the e2e cycle's index 3.
- **Guard interplay**: e2edb's refuse-if-target==DB_NAME guard turns a misconfigured `.env` (`DB_NAME=artisan_test`, lines 328-330) into an explicit failure — fail-closed, coherent.
- **Per-cycle DROP/CREATE vs truncate vs create-if-absent** (lines 95, 97): deterministic-refresh choice is the minimal correct option, not over-engineering; the cost estimate (lines 334-335) is proportionate.
- **Sub-make `migrate_up DB_NAME=artisan_test`**: mirrors the proven `e2e` recipe line (Makefile:158); command-line variables override both `-include .env` and `DB_NAME ?=` defaults.
- **Two packages pinning the same DB under `-p 1`**: no new race introduced; both suites already shared one database.

No false compromise detected.

---

## Part 2 — egm-inventive-catalyst (Mode: REVIEW)

# EGM Review — Inventive Catalyst (Mode: REVIEW)

Document: `.kilo/plans/1795546592659-artisan-test-db.md` (lines 1-394). Baseline grounded against `Makefile`, `cmd/e2edb/main.go`, `.github/workflows/ci.yml`, `tests/main_test.go`, `tests/article_test.go`, `controllers/main_test.go`, `controllers/auth_test.go`.

**Step 0 — Contradiction existence check.** The core architecture is sound and not a false compromise: a disposable provisioned database with process-level pinning and fail-closed rejection mirrors the already-validated REDIS_DB pattern (plan l.65: « symétrique exact de `REDIS_DB=2` »). The rejected alternatives table (l.89-98) shows the author examined the obvious trade-off space (make-only export, new tool, create-if-missing, psql, truncate, parameterized name) and each rejection is argued, not assumed. No foundational false compromise is being accepted. The findings below are secondary, localized compromises.

---

## Finding 1 — The global sweep is preserved as a permanent landmine, not eliminated

1. **Severity**: Major
2. **Affected scope**: « Ce qui ne change pas » (plan l.78-80) and the design decision table (l.64: « deux acteurs destructeurs »)
3. **False compromise**: The plan declares the global `DELETE FROM public.tenant` in `TestCleanUp` (`tests/article_test.go:910`) harmless simply because it now runs on a disposable database: « le balayage global reste tel quel — il devient inoffensif en vivant sur `artisan_test` » (l.78-79). That "harmlessness" is not a property of the code — it is a property of the *environment being correct*. The plan itself builds three protection layers (TestMain pinning l.160-188, `current_database()` assertions l.196-219, fail-closed init l.66), yet the destroyer remains the one component whose safety depends on *every current and future* package remembering to pin. Add a fourth test package tomorrow that calls `pgdb.InitAdmin` without pinning, and the dev wipe returns silently — the assertions in T5/T6 only guard the two suites they live in. This is the repository's residual single point of catastrophic failure, accepted as "harmless".
4. **Proposed resolution**: **Inversion + Cheap Dispensables.** Once the database itself is the disposable segment, teardown-by-sweep is architecturally redundant: the single destructor should be e2edb's `DROP DATABASE ... WITH (FORCE)` (`cmd/e2edb/main.go:152`), which already guarantees a virgin state every provision. Concretely: scope `TestCleanUp` to suite-owned rows (delete by the suite's known emails/slugs, exactly the pattern `controllers/auth_test.go:48-50` already uses: targeted delete by `tenant.ID`) instead of an unbounded `DELETE FROM public.tenant`. The plan's own contrast (l.19-21: controllers suites are « ciblé, non destructeur global ») shows the scoped pattern is established in this repo. `tests/article_test.go` is already inside the modified-files allowlist (l.360), so this needs no allowlist expansion. An alternative, less invasive variant: leave the sweep but convert it from landmine to tripwire — assert inside `TestCleanUp` that `current_database() == 'artisan_test'` before executing the global `DELETE`, making any unpinned drift fail loudly instead of silently wiping dev.
5. **Secondary liabilities**: Removing or scoping the sweep means consecutive direct `go test` runs between two provisionings accumulate residue (the plan itself notes the sweep is « utile aux `go test` directs entre deux provisionnements », l.79-80) — HTTP flows seeding unique emails/slugs would collide with leftovers. The tripwire variant keeps this benefit intact while only adding one assertion; full removal additionally requires accepting ad-hoc-run residue as the caller's problem.

## Finding 2 — The Objective claims total exclusivity; the scope decision quietly waives it

1. **Severity**: Nit
2. **Affected scope**: « Objectifs » (plan l.41-44) versus « Périmètre non épinglé » (l.72) and Risques (l.339-341)
3. **False compromise**: The objective states the Go suites run « exclusivement sur une base jetable `artisan_test` » (l.41-44), yet line 72 declares `internal/db/pool_test.go` stays pinned to « la base configurée » — i.e., under `make test`, `pool_test.go` still opens connections to `golang_gin_db`, the dev database. Read-only and non-destructive, and explicitly assumed (« assumé, documenté »), but the objective as written is false from day one. A reader of the objective alone would audit for exclusivity and find an undocumented-looking violation; the reconciliation lives three sections away.
4. **Proposed resolution**: Reword the objective to what is actually guaranteed — "the *destructive* suites live exclusively on `artisan_test`; the dev base is never written nor purged by any suite" — or, more inventive but costlier, pin `DB_NAME` in `pool_test.go`'s TestMain too (it only needs *a* reachable database to exercise DSN parsing, not the dev one specifically). The latter would require amending the allowlist (l.366-368 forbids `internal/db/**`).
5. **Secondary liabilities**: The rewording is free. The extra pinning adds a third TestMain to maintain and breaks the stated principle of zero changes to `internal/db/**`; it also couples pool-parse tests to a database that `make test` may have just dropped and recreated mid-cycle (harmless with `-p 1`, but conceptually noisier).

## Finding 3 — Fail-closed is deliberate, but the failure is loud without being *actionable*

1. **Severity**: Nit
2. **Affected scope**: « Fail-closed » decision (plan l.66) and Risques (l.324-327)
3. **False compromise**: The plan treats « échec explicite à l'init des pools (`database "artisan_test" does not exist`) » as the safety mechanism, with the remedy confined to documentation (T8, l.250: « exige qu'un `make test` ait provisionné `artisan_test` une fois »). The developer who hits the failure gets a raw pgx error pointing at nothing. The trade-off "safety versus guided recovery" is accepted without examining whether both are achievable.
4. **Proposed resolution**: **Blessing in Disguise / Self-Service.** The `require.NoError(t, pgdb.InitTenant(...), "initializing the tenant pool")` call sites already carry custom failure messages (T5 targets `tests/article_test.go:157-158`; T6 targets `controllers/auth_test.go:38-39`). Extend those messages to name the remedy: e.g., "initializing the tenant pool — if the error says the database does not exist, run `make test` once to provision `artisan_test`". No logic change, no allowlist expansion (both files are already modified), error path stays fail-closed.
5. **Secondary liabilities**: None beyond slightly longer assertion strings; the guidance can drift if the provisioning command changes (mitigated by T8's documentation duty).

## Finding 4 — `FLUSHDB(2)` in the `make test` preamble is unguarded against the one concurrency the plan does not consider

1. **Severity**: Nit
2. **Affected scope**: Design decision « Recréation par cycle » (plan l.68) and Risques (l.331-333)
3. **False compromise**: The plan analyzes `make test` versus `make e2e` concurrency (l.331) — distinct databases, distinct indexes — but not `make test` versus an ad-hoc direct `go test` in a second terminal. Both target Redis index 2; the recipe's FLUSHDB (l.68: « le FLUSHDB de l'index 2 en préambule ») erases the live sessions of the concurrent direct run mid-flight, producing flaky failures far from their cause. The deterministic-start benefit is accepted; the collision cost is unnanounced.
4. **Proposed resolution**: Document it (one sentence in T8's `AGENTS.md` bullet: "a running `make test` flushes index 2 — do not run direct suites concurrently"), or as the inventive alternative, segment further: let direct runs use their own index via an override the recipe does not set — but note e2edb hard-couples flush to `REDIS_DB`, and the plan forbids logic changes to `cmd/e2edb` (l.134), so documentation is the proportionate fix.
5. **Secondary liabilities**: Documentation-only: risk of the note going stale; the flag-based alternative would breach the frozen tool scope and add a CLI surface.

## Finding 5 — Phase 2 marker validation is not rerun-safe

1. **Severity**: Nit
2. **Affected scope**: Validation Phase 2, milestone 1 (plan l.300-307)
3. **False compromise**: The marker proof inserts `('dev-keep-marker', 'Dev Marker')` into `golang_gin_db.tenant` with a bare `INSERT` (l.300-302). If a prior Phase 2 run failed between milestones 1 and 4 (the cleanup `DELETE ... WHERE slug='dev-keep-marker'` is milestone 4, l.315-317), the rerun hits the unique slug constraint with 23505 and blocks re-validation. The validation protocol assumes first-run success.
4. **Proposed resolution**: Make the seed idempotent — `DELETE FROM tenant WHERE slug='dev-keep-marker'` immediately before the `INSERT` in the same verification script (or `INSERT ... ON CONFLICT (slug) DO UPDATE SET name='Dev Marker'`), so milestone 1 is rerunnable after any partial failure without manual cleanup.
5. **Secondary liabilities**: None; a one-statement change to a manual validation script, no repo files touched.

**Summary**: one Major (the retained unbounded sweep — the only place where the plan's safety is environmental rather than structural), four Nits. No blocker: the provisioning path, pinning, CI wiring, and line-number anchors all check out against the baseline files (`Makefile:61-67`, `ci.yml:64-74`, `article_test.go:153-160/907-911`, `auth_test.go:38-39`, e2edb guards at `main.go:108-136`). The plan is executable as written; the Major is a design-completeness gap worth resolving in T2/T5's scope rather than deferring to a future plan.

---

*Both reviews returned by direct subagent dispatch (Goldfish isolation). Findings unapplied — consolidated at the end of the uninterrupted flow.*
