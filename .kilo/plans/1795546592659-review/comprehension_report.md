# Stage 6 — Comprehension Test

Target: `.kilo/plans/1795546592659-artisan-test-db.md` (FeatureRange 1-112 + DetailedRange 113-394 = full file).
Reviewer: `egm-comprehension` dispatched DIRECTLY, no document standard, no GoldfishContext (self-containment test — only the document was readable).

---

## What this proposes

Per the plan (lines 1-112), the Go test suites — the packages `tests` and `controllers` — will run exclusively on a disposable PostgreSQL database `artisan_test` that `make test` drops, recreates and migrates on every cycle, so the destructive global cleanup `DELETE FROM public.tenant` of `TestCleanUp` can never again wipe the dev database `golang_gin_db`. Pinning is done process-wide in each package's `TestMain` (`os.Setenv("DB_NAME", "artisan_test")`, mirrored on the already-shipped `REDIS_DB=2` Redis-index isolation), relying on the property that godotenv never overrides an already-set variable; if `artisan_test` is absent, the pool init fails loudly instead of silently falling back to the dev base. No new tool is written: the existing tool `cmd/e2edb` (the one that provisions the e2e base) is reused with prefixed environment variables (`E2E_DB_NAME=artisan_test REDIS_DB=2`), guarded by its existing allowlist and anti-DROP checks, with only its documentation generalized. Success is provable from the outside: a marker tenant (`dev-keep-marker`) inserted in the dev base must survive both `make test` and a direct `go test`, because any suite drift back to the dev base would delete it and fail the milestone. The work is phased in three phases — provisioning + pinning (T1-T7, delivered as one block because pinning without provisioning would break `make test`), proofs (marker, direct `go test`, e2e regression, cleanup), then documentation (T8) — each closing on a green milestone, with commit only on explicit request.

## How the existing system works, per this document

Today, `make test` (and, because `TestIntDB` loads `.env` itself via godotenv, any direct `go test` run outside make) executes the integration suites directly on the dev database `golang_gin_db`, and `TestCleanUp` in `tests/article_test.go` runs a global `DELETE FROM public.tenant`, so every test cycle erases dev tenants. The e2e cycle is already isolated: the Go tool `cmd/e2edb` recreates the disposable `artisan_e2e` database (allowlist `^[a-z][a-z0-9_]{2,62}$`, guards refusing `DB_NAME`/`postgres`/`template*` targets, quoted identifiers) and flushes the Redis index named by `REDIS_DB` — index 3 for e2e, refusing dev index 1 when the variable is unset — while each suite package's `TestMain` pins `REDIS_DB=2` for the Go suites. In CI, `DB_NAME: golang_gin_db` is set at job level, the job creates and migrates that base as superuser, runs a smoke test of the binary against it, and defines no `E2E_DB_ADMIN_*` variables. Two packages stay outside any database pinning by design: `middleware` (Redis only, no PostgreSQL) and `internal/db/pool_test.go` (read-only connections to whatever base is configured).

Both paraphrases could be written entirely from the document's own text; per the stage rules (no repository access, no baseline), the paraphrase is validated only against internal coherence, not against the actual codebase.

## Findings

**1. Nit — undefined decision label « garde D3 » (lines 24-25, 67, 392).**
The tag "D3" is an identifier from the earlier REDIS_DB plan's decision register; the document uses it (« garde D3 : refuse l'index 1 si absent », « garde Redis D3 », « garde D3 dans e2edb ») without ever saying what "D3" means. Its behavior is restated inline every time, so execution is unharmed, but a fresh reader hits a dead reference three times and may start looking for a guard literally named D3 in the code. Fix: drop the tag or expand it once — e.g. « garde D3 (décision n°3 du plan REDIS_DB) : refuse l'index 1 si REDIS_DB est absent ».

**2. Nit — prior-plan back-references assumed as read (line 69, line 94).**
Line 69 cites « l'analogue de `Options().DB == 2` du plan REDIS_DB » and line 94 « (rejet déjà acté pour le parseur dans le plan REDIS_DB) » — both presuppose the reader has the REDIS_DB plan in hand. In each case the operative rationale is restated inline (the `current_database()` assertion; "deux outils = deux sources de vérité"), so these parentheticals are decorative, but they are unresolvable for a naive reader. Fix: delete the parentheticals or reword them to self-standing statements (« même mécanisme de preuve que l'épinglage Redis »).

**3. Nit — ambiguity in line 24: « refuse l'index 1 si absent ».**
The antecedent of « absent » is unstated here; T1's English doc text (lines 120-121) disambiguates it as "when REDIS_DB is unset", but a reader meeting the French phrase first cannot tell whether it means "if REDIS_DB is absent" or "if index 1 is absent". Fix: « refuse l'index dev 1 quand `REDIS_DB` n'est pas posé ».

**4. Nit — « hors épinglage » for middleware can be misread (line 72 vs line 391).**
Line 72 puts `middleware` « hors épinglage — assumé, documenté », while line 391 states the REDIS_DB plan added a `TestMain REDIS_DB=2` to `middleware`. Under the definition at line 65 (« épinglage » = posing `DB_NAME`), the two are consistent, but a reader crossing both lines may briefly conclude middleware has no TestMain at all, and wonder whether T4 should touch it. Fix: on line 72, write « hors épinglage `DB_NAME` (son `TestMain REDIS_DB=2` existant reste inchangé) ».

**5. Nit — T3/T4 replace an existing TestMain whose full content is never shown (lines 160, 188).**
T3 says « Remplacer le `TestMain` existant … par : » and supplies complete new code, but the document never affirms that the existing TestMain contains only the `REDIS_DB=2` pin; a reader cannot conclude from the text alone that wholesale replacement loses nothing (line 391 only implies it, as a checklist item). Fix: one clause in T3, e.g. « l'actuel `TestMain` ne fait que poser `REDIS_DB=2` (plan REDIS_DB) — le remplacement est sans perte ».

**6. Nit — T2 env-passing style is unexplained (lines 143-147).**
The recipe sources `.env` on two separate lines and explicitly forwards `DB_HOST="$$DB_HOST" DB_PORT="$$DB_PORT" SSL="$$SSL"` to `migrate_up`, while `DB_NAME`/`DB_USER`/`DB_PASS` are passed as make arguments. The code is given verbatim so it executes, but a naive reader cannot tell whether the explicit forwardings are load-bearing or cosmetic, and what happens if one is dropped. Fix: one sentence stating why (e.g. « pattern miroir de la recette `e2e` : ces trois valeurs ne sont pas connues de la cible make sans passage explicite »).

**7. Nit — validation marker INSERT is not idempotent (lines 300-307).**
Phase 2 step 1 inserts `('dev-keep-marker','Dev Marker')` with no conflict handling. If the phase is replayed, or step 4's cleanup was skipped in a previous run, the INSERT fails on `tenant.slug` with no explanation in the document. Fix: `ON CONFLICT (slug) DO NOTHING`, or prepend a delete of any residual marker, or state « supprimer un éventuel marqueur résiduel avant l'INSERT ».

**8. Nit — undefined abbreviation « CUJ » (line 313).**
« (1 setup + 3 CUJ) » uses an unexpanded abbreviation (Critical User Journey). Purely informational — the milestone is just « `make e2e` → vert » — but it is unexplained jargon for a fresh reader. Fix: « 1 projet `setup` + 3 scénarios de bout en bout ».

**9. Nit — « masquage MSYS du pdftotext xpdf » never explained (lines 28, 156).**
The existing recipe's MSYS masking is cited as a fact of the recipe and required to stay unchanged (« Le `-p 1` existant et le masquage MSYS restent inchangés »), but the document never says what it guards against, so a reader porting the recipe to another environment does not know why removing it would break the suite. Fix: half a sentence, e.g. « le PATH Windows embarque un `pdftotext` xpdf incompatible que le masquage cache à la suite PDF ».

No Blocker and no Major findings: every phase, task, validation command and guard is executable from the document's own text, the definitions of `e2edb`, `artisan_test`, `REDIS_DB` indices, godotenv behavior and the fail-closed property are all given inline, and the two halves connect coherently (each decision maps to a task and to a milestone). The findings above are dead references, antecedent ambiguities and unexplained idioms that slow a fresh reader without stopping one.
