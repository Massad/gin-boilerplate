# Stage 7 — Critic Review

Target: `.kilo/plans/1795546592659-artisan-test-db.md` — FeatureRange (lines 1-112).
Reviewers (direct dispatch, parallel): `egm-design` (standard feature-spec-standard), `egm-structural` (standard feature-spec-standard), `egm-style` (no standard). Findings unapplied at this point (uninterrupted flow — consolidation at the end).

---

## Part 1 — egm-design (standard: feature-spec-standard)

# Design-Critic Review — plan `1795546592659-artisan-test-db.md` (FeatureRange: lines 1-112)

**Mode**: REVIEW. **Scope audited**: design half (lines 1-112); task half skimmed only for coherence of references. Baseline grounding read: AGENTS.md, Makefile, `cmd/e2edb/main.go`, `.github/workflows/ci.yml`, `tests/main_test.go`, `tests/article_test.go` (TestIntDB + TestCleanUp), `controllers/main_test.go`, `controllers/auth_test.go` (setup block), `docker-compose.yml`, `internal/db/pool.go`, `internal/db/pool_test.go`, `db/migrations/000002_multi_tenant.sql`, `middleware/main_test.go`.

## Overall verdict

The core mechanism is sound and verifiable: pin `DB_NAME=artisan_test` at process level in each suite's `TestMain` (exact symmetric of the already-validated `REDIS_DB=2` pin), provision via the existing `e2edb` tool guarded by its own guardrails, and prove membership with `current_database()` assertions. I could not construct a plausible path where the suites still write the dev base after this plan. Every line anchor I checked is accurate (Makefile `test:` target at l. 61, echo at l. 62; ci.yml `DB_NAME: golang_gin_db` at l. 50, step at l. 64-74; `TestCleanUp` sweeping `public.tenant` at article_test l. 907-910; `InitAdmin` at l. 158 with `ctx` at l. 156; auth_test pool init at l. 38-39). The alternatives section is concrete and technical (6 entries, no appeal-to-maintainer) — it satisfies the Alternatives Rule. The marker criterion (`dev-keep-marker` slug planted in the very table the global sweep empties) is genuinely falsifiable. That said, two claims in the Risks section are broader than the evidence, and one operational prerequisite is silently inherited rather than stated.

## Findings

### 1. Major — `make test` is presented as self-provisioning, but it is not: the dev-base prerequisite is inherited silently

**Affected scope**: plan l. 41-44 (objective), l. 66 (fail-closed), l. 72 (decision "Périmètre non épinglé"), l. 322-327 (Risks), l. 242-258 (T8 docs); grounded in `internal/db/pool_test.go` (l. 17-33), Makefile l. 13, `.github/workflows/ci.yml` l. 70.

**Finding**: The objective claims the Go suites run "exclusivement" on `artisan_test` and the Risks section says "tout `make test` provisionne" (l. 327). But `internal/db/pool_test.go` runs under the same `go test -tags=all ./...` invoked by `make test`, and it connects to the **configured** database (`golang_gin_db`) — its skip condition (l. 18) is always false under `make test`/CI because `DB_HOST`/`DB_NAME` are set by `.env`/job env. It is indeed read-only (`SELECT current_database()` only), so the "dev never written" goal holds, but the *self-sufficiency* implied by l. 327 is false: on a fresh clone, or after the documented destructive reset (`docker compose down -v`, which even erases roles), `make test` provisions `artisan_test` and migrates it, then fails red in `internal/db` because `golang_gin_db` doesn't exist. Nothing in the design half forces `make migrate_up` on the dev base first, and T8's documented prerequisite covers only the reverse case (`go test` direct needs `artisan_test`). The pre-existing-ness of the failure (suites previously ran on dev) does not excuse the plan: it changes the provisioning contract and its own success narrative ("make test provisionne tout ce dont les suites ont besoin") is wrong as stated.

**Recommendation**: Add an explicit Risks entry: "`make test` reste dépendant de `golang_gin_db` migré (première `make migrate_up`) à cause de `internal/db/pool_test.go` — lecture seule mais connexion réelle". Extend the T8 `make test` doc bullet (AGENTS + README) with the ordering prerequisite. Note the write-forbidden scope (`internal/db/**` in the Allowlist) makes docs the only fix channel — which is fine, but then the docs must actually say it.

### 2. Major — the concurrency assertion ("`make test` et `make e2e` concurrents : sûr") is broader than its evidence

**Affected scope**: plan l. 331-333 ("rôles idempotents, index Redis distincts — sûr"); grounded in `cmd/e2edb/main.go` l. 140-161, `db/migrations/000002_multi_tenant.sql` l. 7-15, Makefile l. 114-130.

**Finding**: The database-level claim *does* hold under inspection — I checked the migration: role creation uses `EXCEPTION WHEN duplicate_object THEN NULL` (l. 7-15), which is genuinely exception-safe, not check-then-act, so two concurrent migrates on distinct cluster databases cannot race each other into a duplicate-role failure; e2edb drops distinct databases with `WITH (FORCE)` and flushes distinct Redis indexes. But "sûr" for the two *recipes* is a stronger claim than what has been proven: `make test` (new T2) and `make e2e` (l. 152-158) both transitively run `migrate_up → install_goose`, which `go install`s to the **same** shared GOBIN path (`.cache/tools/goose/v3.20.0/goose`). Two concurrent recipe executions race on that single binary write — transient rename/ENOENT failures are realistic, particularly through the Git-for-Windows `sh` route documented in the Makefile header. The design asserts cluster-level safety and silently inherits command-level unsafety.

**Recommendation**: Either (a) narrow the claim to "$DB$/$Redis$ level safe; the goose install step is not serialized", (b) state the operational rule "don't run `make test` and `make e2e` concurrently in the same checkout", or (c) note that a first `migrate_up` warms the cache and subsequent races are benign — but pick one explicitly. As written, the Risks section overclaims.

### 3. Nit — the cost-per-cycle estimate is unbacked

**Affected scope**: plan l. 334-335.

**Finding**: "DROP/CREATE + migrations sur base vide ≈ quelques secondes" is an estimate with no measurement, and it matters — this cost is paid on *every* `make test` and every CI run (the new CI lines add a second full migrate). Three migrations on three columns of DDL is very likely fast, but the persona rule is: don't accept assertions without proof.

**Recommendation**: Measure once during Phase 2 (or Phase 1 milestone 3, which already runs it) and record the number in the plan or memory-bank; the CI timeout (15 min) has enough headroom, but say so with a figure, not a vibe.

### 4. Nit — the dev-marker proof is a one-shot, not a retained regression defense

**Affected scope**: plan l. 46-51 (success criterion), Phase 2.1 (l. 300-307).

**Finding**: The marker test is excellent as a falsifiable acceptance proof, but after Phase 2's cleanup step (`DELETE ... WHERE slug='dev-keep-marker'`, l. 315-317) nothing retains the protection: future drift (a third suite added without pinning, someone re-pinning `DB_NAME`) is caught only by the two `current_database()` assertions. That is a real but pointwise defense.

**Recommendation**: Cheap permanent guard to consider (docs or CI, not code): forbid the literal `golang_gin_db` in non-comment Go files under `tests/` and `controllers/`, mirroring the existing CI "hardcoded Redis index" grep guard (ci.yml l. 89-91). Not required for this plan to pass, but the asymmetry between the Redis guard (automated) and the DB pin (assertion-only) invites the same guard.

### 5. Nit — hardcoded dev superuser credentials in the new recipe line

**Affected scope**: plan l. 144 (T2), l. 231 (T7); Makefile l. 155 for the precedent.

**Finding**: `E2E_DB_ADMIN_USER=postgres E2E_DB_ADMIN_PASS=postgres` repeated in `make test` copies the `e2e` recipe pattern exactly; consistent, acceptable for dev/CI containers. But the design half never says this is dev/CI-only and must never be a production credential — the same silence exists in the `e2e` recipe, so this is a repo-wide pattern issue, not introduced here.

**Recommendation**: One sentence in the T8 doc bullet ("superuser dev/CI postgres/postgres, comme `make e2e`") suffices; no code change.

### 6. Nit — minor presentational gap against the standard

**Affected scope**: plan l. 60-72 vs. standard "Proposed Architecture" requirement.

**Finding**: The Alternatives Rule and the Risks & Limits requirement are satisfied. The "Proposed Architecture" is conveyed through the decision table rather than a component/data-flow description; combined with the pinned literals decision (l. 69) this is adequate for an EGM implementation plan, but a fresh reader gets the full picture only after the task list. Acceptable given the format; no action strictly required.

## Coherence check (design half vs. task list)

Coherent throughout: T1-T7 map one-to-one onto the decisions table rows; the Phase 1 must-land-as-a-block rationale (l. 105-107: pinning without provisioning breaks `make test`) is correct; the expected e2edb outputs in Phase 1 milestone 3 ("Dropping the database artisan_test" / "Creating..." / "Flushing the Redis index 2") match the actual `Printf` strings in `cmd/e2edb/main.go` l. 151/156/177; the CI insertion's rationale (`E2E_DB_ADMIN_*` absent from the job env — verified, ci.yml l. 42-55) is accurate; Phase 2's regression re-run of `make e2e` correctly acknowledges e2edb being shared by both cycles. The godotenv non-override assumption at l. 65 matches godotenv.Load semantics and the T3 snippet ordering (TestMain sets `DB_NAME` before any test loads `../.env`).

Findings 1 and 2 are the ones worth resolving before the plan enters execution; both have docs-only or claim-narrowing fixes that fit the existing Allowlist.

---

## Part 2 — egm-structural (standard: feature-spec-standard)

# Structural Goldfish — REVIEW report

Document: `.kilo/plans/1795546592659-artisan-test-db.md`, reviewed range lines 1-112 (design half). Line 113+ is out of scope; cross-references from the range into 113+ were only checked for resolvability, not audited for content.

## Structure map (extracted from range 1-112)

- H1: title (l.1) — `# Base jetable artisan_test …`
- Status block: bold paragraph (l.3-6), unlabeled
- H2 sections, in order: `Contexte` (l.8), `Objectifs et hors-objectifs` (l.39), `Décisions de design (fixées)` (l.60), `Ce qui ne change pas` (l.74), `Alternatives rejetées` (l.89), `Phasage` (l.100)
- Two tables in range: decision table (l.62-72), alternatives table (l.91-98). Both are well-formed Markdown: single header row, separator row, consistent pipe counts, no broken cells, no stray pipes inside cells (long cells at l.64, l.67 use em-dashes and code spans, not pipes).
- No H3 in range; no skipped heading levels; all H2s under the single H1. Hierarchy is clean.

## Findings

**1. Major — Mandatory "Risks & Mitigations" section absent from the design half**
- Affected scope: entire range 1-112; the section exists only at l.322 (`## Risques / cas d'échec`), across the Tasks and Validation sections.
- Description: the Feature-Spec Standard lists Risks & Mitigations as mandatory. As delimited, the design half ends at `Phasage` with no failure-mode analysis at all — a fresh reader of the design half sees fixed decisions (fail-closed, DROP WITH FORCE, CI assumptions) with zero stated risk of each. The risks do exist in the document (l.322-343, and they are substantively good — they even mirror decision-table rows), but they are placed in the implementation tail, far from the design decisions they qualify. This is an ordering/coverage deviation from the standard's required section set for the design half.
- Fix direction: either carry a condensed risk summary (or an explicit "see § Risques" pointer) at the end of the design half, or relocate/reduplicate the risk section so the design half is self-contained. Alternatively, the orchestrator's range definition should include l.322-343 as design content — but as the document stands, the design half alone does not satisfy the standard.

**2. Nit — Recurring unresolvable internal reference "plan REDIS_DB"**
- Affected scope: l.10 (fact foundation), l.69 (`analogue de Options().DB == 2 du plan REDIS_DB`), l.93, l.94 (`rejet déjà acté pour le parseur dans le plan REDIS_DB`).
- Description: the reviewed range leans on a sibling plan for grounding and even for a prior rejection rationale, four times, without ever naming the file. A fresh reader cannot open "plan REDIS_DB". Note the standard's Alternatives Rule: rejection rationale "rejet déjà acté dans le plan REDIS_DB" (l.94) is the one rationale in the alternatives table that delegates its justification to an unlinked document (its other half, "deux sources de vérité", is self-contained — so the rule is not violated, just weakened).
- Fix direction: cite the plan's filename (e.g., `.kilo/plans/<…>-redis-db.md`) once at first mention (l.10).

**3. Nit — Stale-by-one line anchor for `TestCleanUp`**
- Affected scope: l.15 — `TestCleanUp (tests/article_test.go, ~l. 908)`.
- Description: baseline `tests/article_test.go` has the doc comment at l.901 and `func TestCleanUp` at l.907; the plan says ~l.908. The `~` hedge covers it, but all other anchors in the range are exact, breaking the internal convention of precision.
- Fix direction: correct to l.907 (or l.901-907).

## Verified consistent (no action)

- **Heading hierarchy**: no level skips, no out-of-order sections, no misspelled headers. `Status` paragraph is unlabeled but matches the plan family's convention.
- **Line anchors within the range, all checked against repo baseline and exact**: `TestIntDB`/`godotenv.Load` l.153 ✓, CI `DB_NAME: golang_gin_db` l.50 ✓, CI step l.64-74 ✓ (name and migrate line both correct), Makefile recipe 61-67 ✓ (target at 61, `go test … -p 1 -v -tags=all` at 67), e2edb allowlist `^[a-z][a-z0-9_]{2,62}$` at `cmd/e2edb/main.go:49` ✓ (`artisan_test` passes: lowercase start, 12 chars), D3 guard refusing unset-index dev flush (main.go:132-133) ✓, `DROP DATABASE … WITH (FORCE)` (main.go:152) ✓, `E2E_DB_ADMIN_*` superuser override with `DB_USER`/`DB_PASS` fallback (main.go:185-189) ✓ (supports the l.32-33 CI claim, and the job env block at ci.yml:42-55 indeed defines no `E2E_DB_ADMIN_*`).
- **Cross-references resolve**: `§ Phasage` (l.4 → l.100) ✓; `§ Validation, jalons 1-4` (l.108 → l.285, which indeed holds exactly 4 Phase-2 milestones) ✓; T-numbers used at l.4/104/110 resolve to definitions at l.113+; `TestMain` ×2 assumption (l.65) is grounded — baseline `tests/main_test.go` and `controllers/main_test.go` both exist.
- **Internal consistency of claims**: "ni par `make test`, ni par `go test` direct" (l.44) matches the alternative rationale (l.93); "le smoke du binaire reste sur `golang_gin_db`" (l.31) matches "le smoke et le binaire restent sur `golang_gin_db`" (l.71); marker slug `dev-keep-marker` (l.47) is used coherently in the success criterion; Phase 1 item list (l.105-106) maps 1:1 onto T1-T7; "Périmètre non épinglé" row (l.72) gives the decision table symmetric coverage of what is left out.
- **Alternatives rule**: six rejected alternatives with concrete, technical rationales (l.91-98) — the mandatory evaluation of alternatives is satisfied; no rejection is attributed to a maintainer without grounds. One rationale (l.94) partially leans on the unlinked REDIS_DB plan (finding 2).
- **Standard coverage in range**: Problem statement (Contexte + Symptôme/Bénéfice l.35-37), Goals & Non-Goals (l.39-58), Proposed architecture (decision table l.60-72 + non-change list l.74-87) — present and well-ordered. Only Risks & Mitigations is missing from the half (finding 1).

## Verdict

Structurally sound range with one real gap (risks not in the design half, Major) and two Nit-level reference hygiene issues. No blockers, no formatting or hierarchy violations.

---

## Part 3 — egm-style (no standard)

# Style Review — FeatureRange lines 1-112 (`1795546592659-artisan-test-db.md`)

## Major findings

**1. Blocker-none; tone: ALL-CAPS emphasis mid-sentence (Major)**
- Scope: lines 18 (`AUSSI`), 33 (`AUCUNE`), 55 (`c'EST`), 70 (`AVANT`).
- Rationale: shouting in a status register. The document already has a bolding convention for emphasis (lines 14, 36, 64, 67); four capitalized fragments introduce a second, incompatible mechanism and read as informal. Rewrite in neutral prose: "tape également la base dev", "le job ne définit aucune variable", "c'est précisément la base dev", "avant toute écriture". One mechanism (bold) for one purpose.

**2. Duplicated scope content: Non-objectifs vs "Ce qui ne change pas" (Major)**
- Scope: lines 53-58 vs lines 74-87.
- Rationale: the same four facts (e2edb doc-only, `.env` unchanged, `make e2e` untouched, migrations/queries untouched) are stated twice nearly verbatim 20 lines apart, and a third time partially in the decisions table (line 67) and line 78. Redundancy at section level; the two sections will drift independently as the plan evolves. Fix: keep the full list in one section (Non-objectifs is the better home for scope, "Ce qui ne change pas" for behavioral implications) and cross-reference from the other.

**3. "assumé" as an unexplained editorial hand-wave (Major — uncalibrated caveat)**
- Scope: line 68 ("…donne un état Redis de départ déterministe — assumé") and line 72 ("restent hors épinglage — assumé, documenté").
- Rationale: "assumé" asserts ownership without giving the reader the trade-off, which is what actually matters: what breaks if this is wrong, and who decides. Action: **CALIBRATE** — replace with the concrete rationale in one clause (line 68: FLUSHDB index 2 only ever holds suite-run data, never dev sessions on index 1 — say so); **RELOCATE** for line 72 — name the section where the out-of-pinning decision is documented, and state the residual risk (middleware and pool_test have no PostgreSQL writes to lose).

**4. Dangling cross-reference "le plan REDIS_DB" (Major)**
- Scope: lines 10, 69, 93, 94.
- Rationale: the text leans on a sibling document four times ("est implémenté et validé ce jour", "analogue de `Options().DB == 2`", "rejet déjà acté pour le parseur") but never names its file. A fresh reader (this persona included) cannot verify any of the borrowed claims. Fix: cite the file path once at first mention (line 10) and reuse a short form thereafter.

## Nit findings

**5. Test-status color jargon (Nit)**
- Scope: lines 12 ("verts"), 51 ("vert à tort"), 106-107 ("rouge").
- Rationale: CI-status slang ("green/red") used in three places; unambiguous to CI practitioners, imprecise to everyone else, and mixed with straight verbs elsewhere. Fix: "réussis"/"échoue" once each, or keep the slang but use it consistently — do not alternate between registers.

**6. Informal personification of suites (Nit)**
- Scope: lines 21 ("vivent sur la même base"), 37 ("vit sur une base jetable"), 80 ("en vivant sur"), and colloquial "tape" (lines 18, 93).
- Rationale: tests do not "live" or "hit" databases in formal engineering prose. Fix: "s'exécutent sur", "cible", "écrit dans".

**7. Four incompatible file/line reference formats (Nit)**
- Scope: line 14 (`tests/article_test.go, ~l. 908`), line 17 (`godotenv.Load, l. 153`), line 27 (`Makefile:61-67`), line 29 (`ci.yml l. 50`).
- Rationale: also mixed precision — `~l. 908` is approximate while every other reference is exact. Fix: one format (`file:line`) throughout; resolve the approximation at line 14.

**8. Overloaded parenthetical packing (Nit)**
- Scope: line 67 (three parentheticals plus two semicolons in one sentence), lines 46-51 (parenthetical containing its own colon list: "(choisi à dessein dans la table … : `tenant`, pas une table épargnée)").
- Rationale: nesting depth forces re-reading; the parenthetical at 47-49 is doing sub-list work. Fix: break line 67's rationales into sub-bullets; flatten the line 47-49 aside into its own sentence.

**9. Run-on status sentence (Nit)**
- Scope: lines 3-6.
- Rationale: one sentence with an em-dash break and two semicolons, plus the bureaucratese "du ressort d'un agent d'implémentation". Fix: split into two sentences; "sera implémenté par un agent dédié" or simply "hors de portée de ce document".

**10. Mid-prose bold emphasis (Nit)**
- Scope: line 67 ("seule la **documentation** de l'outil se généralise").
- Rationale: bold is reserved elsewhere for leads and table keys; mid-sentence bold here competes with the caps instances (finding 1). Fix: unbold; the sentence stresses itself fine.

**11. "à l'identique" claimed without evidence pointer (Nit — CALIBRATE)**
- Scope: lines 84-86 ("les migrations se rejouent à l'identique sur `artisan_test` (rôles idempotents, DO blocks de 000002)").
- Rationale: this is a claim about tested behavior; the evidence exists but is not named — `artisan_e2e` already replays the same migrations every e2e cycle. Fix: cite that precedent in the sentence; otherwise the claim outruns its stated basis.

**12. "le smoke" clipped jargon (Nit)**
- Scope: lines 31, 71.
- Rationale: "le smoke du binaire" / "le smoke" assume the reader expands "smoke test" the same way. Fix: "le smoke test du binaire" (once full, then "le smoke test").

**13. "symétrique" fatigue (Nit)**
- Scope: lines 65 ("symétrique exact"), 68 ("symétrique d'`artisan_e2e`"), 95 ("symétrique d'`artisan_e2e`"), with "analogue" at line 69.
- Rationale: the mirror-pattern argument is stated four times; by the third occurrence it needs no restatement. Fix: state it once with full force (line 65), compress the repeats to a pointer.

**14. Terse asides that under-explain (Nit)**
- Scope: lines 57-58 ("dommage historique"), 98 ("fail-closed, mais inutilement").
- Rationale: both compress a full thought into three words; "dommage historique" could be misread as "historic damage worth noting" rather than "permanently lost, outside this plan's scope". Fix: "perte déjà subie, non restaurée par ce plan" / "échouerait en base absente — fail-closed, sans bénéfice".

## Format checks passed (no finding)

- French quoting is consistent: « » for quoted names/instructions (lines 30, 94, 96), straight quotes only inside code spans for verbatim output (line 66). Correct.
- Spaced em-dashes (—) are used consistently as sentence-internal breaks (lines 3, 64, 68, 72, 93).
- The Critère de succès (lines 46-51) and the Décisions table are models of calibration: concrete, measurable, and every claim has a named mechanism. Keep them as is.
