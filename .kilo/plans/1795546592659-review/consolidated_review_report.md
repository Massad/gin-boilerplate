# Consolidated Review Report — plan `1795546592659-artisan-test-db.md`

Flow: Uninterrupted Full Flow (Stages 5→6→7→8→9). Preset: consolidated, all rubrics.
Ranges: FeatureRange 1-112 (design) / DetailedRange 113-394 (implementation). GoldfishContext: empty (user opted out; no markdown links in the document). Reviewer dispatch: direct subagents (Goldfish isolation); 10 review runs, 0 failed.

Stage reports (same folder):
- `zoom_out_report.md` (egm-zoom-out + egm-inventive-catalyst)
- `comprehension_report.md` (egm-comprehension — 9 Nits, no Blocker/Major; both paraphrases hold)
- `critic_report.md` (egm-design + egm-structural + egm-style, FeatureRange)
- `intermediate_zoom_out_report.md` (egm-zoom-out, 2nd pass — first pass's "No false compromise detected" amended)
- `implementation_readiness_report.md` (egm-implementation-readiness + egm-codebase-details + egm-style, DetailedRange)

## Headline verdict

**No Blocker. The plan is conditionally ready:** every anchor, snippet, guard claim, print string, migration column and pool API claim verified correct by the codebase-details pass (extensive verified-correct lists in the stage reports — the strongest grounding of the whole flow). The blockers-to-green are a handful of Majors: one architectural hole (the retained global sweep is fail-open), one uncovered suite (root `main_test.go` writes the dev base), one self-contradictory allowlist line, two overclaimed guarantees (provisioning self-sufficiency, concurrency safety), one execution-method precondition (script location), and recurring dangling cross-references. All have docs-only or single-line fixes; none requires reopening the design.

**Counts:** 68 raw findings → 0 🔴 Blocker · 8 🟡 Major · 13 🟢 Nit consolidated items.

## 🔴 Blockers

None.

## 🟡 Major consolidated items

**M-1 — The retained global sweep is a fail-open surface (false compromise).**
`TestCleanUp`'s `DELETE FROM public.tenant` is declared harmless merely because the environment is right; a future unpinned package would silently wipe dev again (the only fail-open spot in an otherwise fail-closed plan). *Fix (recommended, proportionate)*: tripwire — add the same `current_database() == 'artisan_test'` assertion at the top of `TestCleanUp` (tests/article_test.go, already in the allowlist; one line, keeps the sweep's benefit for ad-hoc direct runs). *Deeper alternative*: scope the DELETE to suite-owned rows (pattern already used by controllers) — costs more, leaves residue between provisionings. *Fold into the T5 task scope now.*
Sources: Z2-1, Z3-1 (the Stage-8 pass explicitly amends the Stage-5 "No false compromise" verdict).

**M-2 — Root `main_test.go` (package `main`) writes the dev base and is absent from every scope enumeration.**
Under `make test` (and bare `go test` + CI) it upserts/deletes its own `admin_user` row on `golang_gin_db` — read-only nowhere, but never global. The objective (« la base dev n'est plus écrite ni purgée par les suites ») is false as written; T8 wording would inherit the overbroad claim; the dev-marker proof cannot catch it (marker lives in `tenant`). *Fix*: pick one and write it down — (a) document the exception (extend the un-pinned enumeration at l. 72 + Risks with `main_test.go`, classed as targeted-row cleanup like the controllers pattern; soften objective + T8 wording to "aucune suite ne purge ni n'écrit la base dev par une opération globale — seules exceptions documentées : opérations ciblées sur leurs propres lignes") — no allowlist change; or (b) pin package `main` via `TestMain` (requires adding the file to the allowlist's Modifiés). Zero-decision rule demands the plan pre-make this choice.
Sources: R-1, Z2-2, Z3-2.

**M-3 — The allowlist contradicts itself on `controllers/main_test.go`.**
`Modifiés` includes it; `Interdits` bans "`controllers/*.go` hors `auth_test.go`" — which matches it. The Phase 3 jalon (diff reviewed against the allowlist) false-fails on the plan's own expected diff. *Fix*: reword the interdit to « `controllers/*.go` hors `auth_test.go` et `main_test.go` » (mirrors the qualified e2edb carve-out).
Sources: R-2, CD-1.

**M-4 — The provisioning contract claims more than it delivers: `make test` still needs a migrated dev base.**
`internal/db/pool_test.go` connects (read-only) to the configured base (`golang_gin_db`) under `make test`/CI; the skip condition never fires there. On a fresh clone — or after the documented `docker compose down -v`, which even erases roles — `make test` provisions `artisan_test` then goes **red in `internal/db`** because `golang_gin_db` doesn't exist. Nothing in the plan states this ordering prerequisite. *Fix*: one Risks entry + the T8 make-test doc bullets state it (« `make test` reste dépendant d'une première `make migrate_up` sur la base dev, à cause d'`internal/db/pool_test.go` — lecture seule mais connexion réelle ») — docs-only, allowlist-compliant.
Sources: Crit-1 (design critic), grounded at internal/db/pool_test.go:17-33.

**M-5 — Concurrency safety is asserted for one pairing; the real surface is wider.**
Covered: `make test` vs `make e2e` (DB/Redis level genuinely safe — verified: DO-block role creation is exception-safe). Not covered: `make test`/`make e2e` both transitively run `install_goose` writing the same GOBIN binary (installation race); the provisioning `FLUSHDB(2)` erases the live sessions of a concurrent direct `go test`; two concurrent `make test` DROP each other's `artisan_test WITH (FORCE)`; `TestCleanUp` deletes ALL `user_sessions:*`/`admin:*` on index 2, not just its own. All documentation-class. *Fix*: narrow the Risks bullet (cluster-level safe, command level not serialized) + one T8 doc line (operational rule: never run `make test`/`make e2e`/direct suites concurrently in the same checkout).
Sources: Crit-2, Z2-4, Z3-3.

**M-6 — The environment-notes procedure depends on an unstated precondition: script location.**
`tr -d '\r' < /mnt/kilo/<chemin> | bash` works only if the script lives under `C:\Users\schat\.kilo\...` (mounted `/mnt/kilo`); the note never says so, and a fresh executor will plausibly write it in the workspace where the mount can't reach it. *Fix*: one sentence + worked example (`C:\Users\schat\.kilo\tmp\run.sh` ↔ `/mnt/kilo/tmp/run.sh`).
Sources: Y-1 (stage-9 style, rated Major for good cause).

**M-7 — "plan REDIS_DB" is a dangling cross-reference, four times, never named.**
Lines 10, 69, 93/94 lean on the sibling plan (facts, proof mechanism, rejection rationale) without its path; three reviewers flagged it independently (one rated the style instance Major). *Fix*: name the file once at first mention — `.kilo/plans/1795535384311-redis-db-isolation.md` — and shorten later mentions.
Sources: C2-2, C3-4, K6-1, K6-2.

**M-8 — Load-bearing trade-offs marked with the bare word "assumé" (and two related one-liners).**
Line 68 (FLUSHDB 2 as "deterministic — assumé") and line 72 ("hors épinglage — assumé, documenté") assert ownership without the trade-off; separately the plan never acknowledges the `E2E_DB_NAME`-now-carries-non-e2e overload, and never binds the (static-only) CI change to its real acceptance test. *Fix*: three cheap one-liners — line 68: "l'index 2 ne porte que les données des suites, jamais les sessions dev (1)"; line 72: name the residual risk (no PostgreSQL writes in middleware/pool_test) and state the YAGNI deferral (Z1-2); plus "le nom `E2E_DB_NAME` surcharge 'e2e' — le renommage reste rejeté, la doc-only T1 est la frontière choisie" (alt., Z1-4) and "le prochain run CI réel est la preuve engageante de T7 ; une étape rouge y renvoie à ces deux lignes" (Z1-5).
Sources: C3-3, Z1-4, Z1-5, Z1-2.

## 🟢 Nit consolidated items

**N-1 — Section-level duplication:** Non-objectifs (l. 53-58) vs « Ce qui ne change pas » (l. 74-87) state the same facts twice (style Major carried in as a consolidation Nit — keep full list in Non-objectifs, cross-reference the behavioral implications). Sources: C3-2.
**N-2 — Helper-placement rationale:** state why the `current_database()` helper is not in `internal/testutil` (YAGNI until a third pin package) and why `internal/db` can't be pinned (CI `git diff --exit-code internal/db` guard) — both currently inferable only. Sources: Z1-2, Z1-3.
**N-3 — Anchor precision pack:** T6 « l. ~38 » → l. 39; `TestCleanUp` ~l. 908 → l. 907 (DELETE at 910); T2 « En-tête du target (l. 55-60) » → l. 52-60 (« ## RUN TESTS » at 52); unify file/line formats (four styles in the design half). Sources: C2-3, C3-7, R-3, R-4, CD-2, CD-3, Y-3.
**N-4 — Caps discipline:** ALL-CAPS mid-sentence ×4 in the design half (`AUSSI`, `AUCUNE`, `c'EST`, `AVANT`) and ×7 in the implementation half; keep caps only for irreversible guards, neutralize the rest (bold convention already exists). Sources: C3-1, Y-5.
**N-5 — Unexpanded jargon:** « CUJ » → « scénarios de bout en bout »; « masquage MSYS » → gloss « (pdftotext de Git for Windows) »; « le smoke » → « le smoke test du binaire »; « pur, sans services » → « tests unitaires purs, sans PostgreSQL ni Redis »; « conteneur éphémère propre » → « conteneur éphémère auto-géré, hors provisioning ». Sources: K6-8, K6-9, C3-12, Y-2, Y-9.
**N-6 — Marker proof hardening:** make the seed rerun-safe (`DELETE FROM tenant WHERE slug='dev-keep-marker'` before the INSERT, or `ON CONFLICT DO UPDATE`); keep the SQL on one physical line (currently split mid-string — transcription hazard); state the before/after count pairing explicitly. Sources: Z2-5, K6-7, Z3-4, Y-6, Y-7.
**N-7 — Actionable fail-closed messages:** extend the pool-init `require.NoError` messages (both call sites are already modified by T5/T6) to name the remedy (« si l'erreur dit que la base n'existe pas : lancer `make test` une fois »). Sources: Z2-3.
**N-8 — Comprehension leftovers:** disambiguate l. 24 « refuse l'index 1 si absent » → « quand `REDIS_DB` n'est pas posé »; line 72 → « hors épinglage `DB_NAME` (son `TestMain REDIS_DB=2` existant reste inchangé) »; T3 — add "l'actuel TestMain ne fait que poser REDIS_DB=2, le remplacement est sans perte"; T2 — one sentence explaining the three explicit `DB_HOST/DB_PORT/SSL` forwardings (mirror pattern of the e2e recipe). Sources: K6-3, K6-4, K6-5, K6-6.
**N-9 — Style pack:** neutralize personification (« vivent/vit sur », « tape », « le dépôt vit sur », « sessions stale » → « s'exécutent sur », « sessions obsolètes »); CI color slang consistency (« verts/rouge » → keep consistently or use « réussis/échoue »); flatten overloaded parentheticals (l. 67, l. 47-49); split the run-on status sentence (l. 3-6); unbold mid-prose emphasis (l. 67); « symétrique » state-once-then-point (l. 65/68/95); expand terse asides (« dommage historique » → « perte déjà subie, non restaurée par ce plan »; « fail-closed, mais inutilement » → « échouerait en base absente — fail-closed, sans bénéfice »); name the migration-replay precedent (`artisan_e2e`) for « à l'identique » (l. 84-86); « le doc T1 » → « le commentaire de paquet de T1 » (l. 349-350); T1 arrow overload (use prose "remplacer X par Y"); T7 static validation → named check (« le bloc e2edb de l'étape CI doit être identique à la ligne e2edb de la recette `test` »); T8 « avant `go test -p 1` » → « la commande de test existante (`-p 1 -v -tags=all`) »; « Entrée à rejouer » → « Prérequis à vérifier ». Sources: C3-5, C3-6, C3-8, C3-9, C3-10, C3-11, C3-13, C3-14, Y-8, Y-10, Y-11, Y-12, Y-13, Y-14.
**N-10 — Task duplication:** T4 duplicates T3's 26-line block, T6 duplicates T5's — replace with one canonical definition + reference (package name and anchor differ only). Sources: Z1-1, Z3-5.
**N-11 — Optional durable guard (defer/accept):** mirror the Redis CI grep with a guard forbidding the literal `golang_gin_db` in non-comment Go files under `tests/`/`controllers/` — converts the one-shot marker proof into a retained defense. Not required for readiness; decide on adoption. Sources: C1-4, Z3-4 (durable part).
**N-12 — Record the measured provisioning cost:** the "≈ quelques secondes" estimate is unbacked and paid on every `make test`/CI run; measure once in Phase 1 or 2 and record the figure. Sources: C1-3.
**N-13 — Credential note:** one sentence in T8 noting the pinned superuser `postgres/postgres` is dev/CI-only (same silence already exists in the `e2e` recipe). Sources: C1-5.

## Traceability matrix (68 raw findings → consolidated items)

| Raw finding | Stage source | Consolidated item |
|---|---|---|
| Z1-1 (dup code blocks) | 5 zoom-out | N-10 |
| Z1-2 (testutil placement) | 5 zoom-out | M-8, N-2 |
| Z1-3 (internal/db CI-guard reason) | 5 zoom-out | M-4, N-2 |
| Z1-4 (E2E_DB_NAME overload) | 5 zoom-out | M-8 |
| Z1-5 (CI static validation binding) | 5 zoom-out | M-8 |
| Z1 — "No false compromise detected" | 5 zoom-out | amended by Z3 (Stage 8) |
| Z2-1 (sweep landmine, Major) | 5 catalyst | M-1 |
| Z2-2 (exclusivity overclaim) | 5 catalyst | M-2 |
| Z2-3 (fail-closed not actionable) | 5 catalyst | N-7 |
| Z2-4 (FLUSHDB(2) concurrency) | 5 catalyst | M-5 |
| Z2-5 (marker not rerun-safe) | 5 catalyst | N-6 |
| K6-1 (garde D3 undefined) | 6 comprehension | N-8 (fold: name plan + D3 once — see M-7 too) |
| K6-2 (back-references) | 6 comprehension | M-7 |
| K6-3 (« index 1 si absent ») | 6 comprehension | N-8 |
| K6-4 (middleware hors épinglage) | 6 comprehension | N-8 |
| K6-5 (TestMain replacement content) | 6 comprehension | N-8 |
| K6-6 (T2 env-passing why) | 6 comprehension | N-8 |
| K6-7 (marker INSERT idempotency) | 6 comprehension | N-6 |
| K6-8 (CUJ) | 6 comprehension | N-5 |
| K6-9 (MSYS masking) | 6 comprehension | N-5 |
| C1-1 (make test not self-provisioning, Major) | 7 design | M-4 |
| C1-2 (concurrency overclaim, Major) | 7 design | M-5 |
| C1-3 (cost unbacked) | 7 design | N-12 |
| C1-4 (marker one-shot / durable guard) | 7 design | N-11 |
| C1-5 (credentials note) | 7 design | N-13 |
| C1-6 (presentational gap, no action) | 7 design | accepted as family format |
| C2-1 (Risks absent from design half, Major) | 7 structural | M-7 sibling — see note below* |
| C2-2 (unresolvable REDIS_DB ref) | 7 structural | M-7 |
| C2-3 (TestCleanUp anchor) | 7 structural | N-3 |
| C3-1 (caps shouting) | 7 style | N-4 |
| C3-2 (section duplication) | 7 style | N-1 |
| C3-3 ("assumé" uncalibrated) | 7 style | M-8 |
| C3-4 (dangling REDIS_DB) | 7 style | M-7 |
| C3-5 (color jargon) | 7 style | N-9 |
| C3-6 (personification) | 7 style | N-9 |
| C3-7 (anchor formats) | 7 style | N-3 |
| C3-8 (parentheticals) | 7 style | N-9 |
| C3-9 (run-on status) | 7 style | N-9 |
| C3-10 (mid-prose bold) | 7 style | N-9 |
| C3-11 ("à l'identique" evidence) | 7 style | N-9 |
| C3-12 ("le smoke") | 7 style | N-5 |
| C3-13 ("symétrique" fatigue) | 7 style | N-9 |
| C3-14 (terse asides) | 7 style | N-9 |
| Z3-1 (false compromise confirmed, Major) | 8 zoom-out-2 | M-1 |
| Z3-2 (exclusivity overclaim) | 8 zoom-out-2 | M-2 |
| Z3-3 (concurrency surface wider) | 8 zoom-out-2 | M-5 |
| Z3-4 (criterion one-shot + rerun-safe) | 8 zoom-out-2 | N-6, N-11 |
| Z3-5 (dup code blocks carried) | 8 zoom-out-2 | N-10 |
| R-1 (root main_test.go unwritten scope, Major) | 9 readiness | M-2 |
| R-2 (allowlist contradiction, Major) | 9 readiness | M-3 |
| R-3 (T6 anchor) | 9 readiness | N-3 |
| R-4 (TestCleanUp anchor) | 9 readiness | N-3 |
| CD-1 (allowlist contradiction, Major) | 9 codebase-details | M-3 |
| CD-2 (T6 anchor) | 9 codebase-details | N-3 |
| CD-3 (T2 header anchor) | 9 codebase-details | N-3 |
| Y-1 (script location Major) | 9 style | M-6 |
| Y-2 (MSYS/CUJ) | 9 style | N-5 |
| Y-3 (tilde anchor) | 9 style | N-3 |
| Y-4 (ellipsis conventions) | 9 style | N-10's scope (quote hygiene) — folded into N-1/N-10 |
| Y-5 (caps overuse) | 9 style | N-4 |
| Y-6 (SQL split mid-string) | 9 style | N-6 |
| Y-7 (count pairing implicit) | 9 style | N-6 |
| Y-8 (anglicisms/personification) | 9 style | N-9 |
| Y-9 ("pur"/"propre") | 9 style | N-5 |
| Y-10 (T7 static check vague) | 9 style | N-9 |
| Y-11 (T8 under-quoted command) | 9 style | N-9 |
| Y-12 ("le doc T1") | 9 style | N-9 |
| Y-13 (arrow overload) | 9 style | N-9 |
| Y-14 ("Entrée à rejouer") | 9 style | N-9 |

\* C2-1 note (Risks section absent from the design half, rated Major by the structural reviewer): the § Risques exists (l. 322-343, substantively strong) but sits in the implementation tail — the document follows the established plan-family layout of this repo (the REDIS_DB plan uses the same layout and passed the same protocol). Cheapest conforming fix if wanted: add a « Risques → § Risques / cas d'échec » pointer at the end of the design half. The consolidated report keeps it linked to the M-group rather than granting a standalone Major: the design-half reader does reach an explicit pointer at l. 46 (« Risques et limites → § Risques / cas d'échec » appears in the family's Objectifs section in the REDIS_DB plan; in THIS plan that pointer line is absent — add it, one line, close on M-8's calibration pass).

## Suggested resolution order

1. **M-3** (one-line allowlist fix — unblocks the Phase 3 jalon).
2. **M-1 + M-2** (design decisions to pre-make: sweep tripwire in T5 scope; main_test.go classification + objective rewording).
3. **M-4 + M-5 + M-8 + M-6** (docs-only wording fixes, one editing pass over Risques/T8/Notes).
4. **M-7** (one path citation at l. 10).
5. 🟢 N-1..N-13 (mostly single-line text fixes; N-6/N-3 touch executable validation text — worth doing before execution).
