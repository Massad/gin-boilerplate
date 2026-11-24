# Stage 8 — Intermediate Zoom Out

Target: `.kilo/plans/1795546592659-artisan-test-db.md` (full file) with the Stage 5 zoom-out report in context. Reviewer: `egm-zoom-out` + standard `zoom-out-standard` (direct dispatch). Comprehension/Critic Stage 7 findings were NOT applied to the document (uninterrupted flow) — this pass audits the document as-is after knowing those passes ran.

---

# Second Zoom-Out Pass — plan `1795546592659-artisan-test-db.md`

Persona: egm-zoom-out (Skeptical Architect), Mode: REVIEW, second zoom-out pass. Standard: zoom-out-standard. All findings below cite plan line numbers; baseline grounding re-verified against `Makefile`, `cmd/e2edb/main.go`, `.github/workflows/ci.yml`, `tests/main_test.go`, `tests/article_test.go`, `controllers/main_test.go`, `controllers/auth_test.go`, `internal/db/pool_test.go`.

## Verdict

The document still holds as a whole. The architecture is not over-engineered: it reuses `cmd/e2edb` unchanged (the strongest reuse move available — l. 67), mirrors an already-validated repo pattern (`TestMain` pinning, l. 65), rejects the six obvious alternatives with real arguments (l. 89-98), and keeps the destructive actor on a disposable base — a structural fix, not a band-aid. Every factual anchor I re-checked is exact: `TestCleanUp`'s `DELETE FROM public.tenant` at tests/article_test.go:910 (plan l. 15's "~l. 908" is the enclosing function), `godotenv.Load` at :153, `ctx` at :156, `InitAdmin` at :158, the assertion order claim (assertion inserted before `InitRedis`/`LoadFixtures` at :159/:165) is consistent with plan l. 70 and T5; `auth_test.go` setup at :38-39 matches plan l. 213; the Makefile `test` recipe (:61-67) and `e2e` pattern (:153-158) match T2 exactly; ci.yml `DB_NAME: golang_gin_db` at :50 and the step at :64-74 match T7; e2edb's guard/allowlist/flush behavior (main.go:49, :108-136, :151-157) matches plan l. 25-26 and the Phase-1 expected outputs (l. 292-294); migration 000002:19-26 confirms the marker `INSERT (slug, name)` claim (all other columns defaulted). `internal/db/pool_test.go` confirmed connect-only, never writes — plan l. 339-341 is accurate.

**The Stage 5 verdict stands with one amendment**, detailed in Finding 1: the architecture, reuse, YAGNI, and altitude judgments all hold after the comprehension/critic passes, but the "No false compromise detected" conclusion of the first zoom-out does not survive — the retained unbounded sweep (l. 78-80) is a real false compromise that the first pass missed and the Critic pass correctly surfaced.

## Findings

**Finding 1 — A false compromise IS present: the global sweep's "harmlessness" is environmental, not structural**
- Severity: Major
- Affected scope: « Ce qui ne change pas » (l. 78-80), design table (l. 64)
- Simplicity rationale: The plan builds three layers of enforcement (process pinning l. 65/160-194, ownership assertions l. 70/196-219, fail-closed init l. 66) yet leaves the one destructive component — `DELETE FROM public.tenant` (l. 78-79: « il devient inoffensif en vivant sur `artisan_test` ») — unguarded. Its safety depends on every current and future package pinning correctly; T5/T6's assertions only protect the two suites they live in. Everything else in the plan is fail-closed; this is the single fail-open surface, and the document declares it harmless instead of closing it. This is a false compromise, not a YAGNI call.
- Fix direction: The minimal structural fix is the tripwire variant: one `current_database() == 'artisan_test'` assertion at the top of `TestCleanUp` (tests/article_test.go, already in the allowlist l. 360), converting the landmine into a loud failure. Scoping the DELETE to suite-owned rows (the pattern controllers/auth_test.go:48-50 already uses) is the deeper alternative but costs more design work; the tripwire is the proportionate change and keeps the sweep's stated benefit for ad-hoc direct runs (l. 79-80). This should be folded into the allowlisted T5-adjacent scope now, not deferred.

**Finding 2 — The objective overclaims exclusivity**
- Severity: Nit
- Affected scope: l. 41-44 vs. l. 72 and l. 339-341
- Simplicity rationale: « s'exécutent exclusivement sur une base jetable `artisan_test` » is falsified from day one by `internal/db/pool_test.go`, which connects to the configured base (`golang_gin_db` under make) — read-only, verified, and acknowledged at l. 72, but 250 lines away from the objective. A reader auditing the objective alone finds an undocumented-looking violation.
- Fix direction: Reword to the actual guarantee: suites that *write* run exclusively on `artisan_test`; the dev base is never written nor purged by any suite. One-line change, zero cost.

**Finding 3 — The concurrency analysis covers one pairing; the real surface is wider**
- Severity: Nit
- Affected scope: Risks (l. 331-333), design l. 68
- Simplicity rationale: The plan analyzes `make test` vs. `make e2e` only. Not covered: a direct `go test` in a second terminal (same index 2 — the provisioning FLUSHDB at l. 68 erases its live sessions mid-flight), two concurrent `make test` invocations (both DROP `artisan_test WITH (FORCE)`, disconnecting each other), and the shared Redis index 2 on which `TestCleanUp` (tests/article_test.go:916-928) deletes *all* `user_sessions:*`/`admin:*` keys, not just its own suite's. All low-probability, all documentation-class.
- Fix direction: One sentence in the Risks section or T8's AGENTS.md bullet: provisioning flushes index 2 and both destructive actors target `artisan_test` — do not run suites or `make test` concurrently.

**Finding 4 — The success criterion is a one-shot manual proof, not a durable guarantee**
- Severity: Nit
- Affected scope: Success criterion (l. 46-51), Validation Phase 2 (l. 300-317)
- Simplicity rationale: The dev marker proves pinning only on the day it is run; nothing persisted prevents a future regression to the dev base. The plan's own criterion language (« si une suite revenait sur la base dev … le jalon échouerait », l. 50-51) implies ongoing protection that no artifact delivers. Related: the Phase-2 marker `INSERT` is not rerun-safe (a failed run between milestones 1 and 4 leaves the row, and the rerun hits 23505).
- Fix direction: The Finding 1 tripwire is the durable form of this criterion and subsumes it; and make the seed idempotent (`DELETE ... WHERE slug='dev-keep-marker'` before the `INSERT`, or `ON CONFLICT DO UPDATE`) so milestone 1 is rerunnable. Both are changes to validation text only.

**Finding 5 — Verbatim code blocks duplicated across tasks (carried over from Part 1, still valid)**
- Severity: Nit
- Affected scope: T3 (l. 162-188) vs. T4 (l. 191-194); T5 (l. 202-206) vs. T6 (l. 210-219)
- Simplicity rationale: T4 and T6 repeat T3/T5 copy-for-copy; the plan-level equivalent of copy-paste code. No information added beyond a package name and path.
- Fix direction: One canonical definition + reference. The implementation agent loses nothing.

Part 1's other Nits (helper placement vs. `internal/testutil`, the `internal/db` CI-guard justification, `E2E_DB_NAME` semantic overload, CI validated only statically) remain valid as written and need no restatement here. Caveat on scope: T8's memory-bank and `.kilo/HANDOFF.md` line anchors (l. 266-283) were outside this pass's permitted grounding set and are unverified — they are documentation-only targets and carry no execution risk.

## Checked and dismissed (second pass)

- T2 recipe block (l. 141-148): the sub-make `migrate_up DB_NAME=artisan_test` invocation overrides both `-include .env` and `DB_NAME ?= golang_gin_db` exactly as the proven `e2e` line (Makefile:158) does; `REDIS_DB=2` correctly confined to the provisioning command (l. 150-152), leaving the pin to `TestMain`.
- T7 placement (after ci.yml:74) is correct: provisioning precedes Build/Vet and the `go test -tags=all` step; the smoke binary legitimately stays on `golang_gin_db` (its routes expect migrated dev state).
- e2edb needs zero logic change for dual use: `artisan_test` passes the allowlist (`^[a-z][a-z0-9_]{2,62}$`, main.go:49), the target≠`DB_NAME` guard, and the quoted identifier; the D3 refuse-dev-index guard behaves correctly with the explicit `REDIS_DB=2`.
- Expected-output milestones (l. 292-294) match e2edb's actual print statements (main.go:151, 156, 177).
- No new tooling, no new abstraction, no speculative interface anywhere in the plan — the altitude is appropriate for an execution-ready document; the verbatim code blocks in T3-T7 are instructions to an implementation agent, not weeds.

## Bottom line

No Blocker. One Major (Finding 1 — the retained unbounded sweep; the first zoom-out's "No false compromise detected" is hereby amended: this is the compromise) and four Nits; Nits 2-5 converge with findings already surfaced in the flow. The plan's provisioning, pinning, CI wiring, anchors, and reuse decisions all re-verify against the baseline. It remains executable as written; the Major is a small, already-allowlisted change (one assertion in `TestCleanUp`) that should be applied during T5's scope rather than deferred.
