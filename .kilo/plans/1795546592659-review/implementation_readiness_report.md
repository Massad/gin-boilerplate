# Stage 9 — Implementation Readiness Test

Target: `.kilo/plans/1795546592659-artisan-test-db.md` — DetailedRange (lines 113-394).
Reviewers (direct dispatch, parallel): `egm-implementation-readiness` (standard implementation-readiness-standard), `egm-codebase-details` (standard implementation-readiness-standard), `egm-style` (no standard).

---

## Part 1 — egm-implementation-readiness (standard: implementation-readiness-standard)

# Implementation-Readiness Review — plan `1795546592659-artisan-test-db.md`, DetailedRange lines 113-394

**Verdict: conditionally ready.** T1-T7 are executable as written: the snippets compile against the current code, every anchor I checked matches the actual files, and the validation milestones are concrete and runnable. Two findings must be resolved before an implementation agent runs the plan: one scope-omission that makes the plan's stated objective false as written, and one internal contradiction in the allowlist that will false-fail the Phase 3 jalon.

## Findings

**1. Major — Root `main_test.go` (package `main`) writes to the dev database and is absent from every scope enumeration in the plan.**
- Affected scope: Risques (plan l. 339-341, which enumerates `internal/db/pool_test.go` as the only un-pinned Postgres-touching suite), Hors scope (l. 345-353), Allowlist interdits (l. 365-370), T8 AGENTS wording (l. 244-250).
- Finding: `main_test.go` (repo root, `//go:build all`) calls `testutil.NewAdminFixture` and `runCreateAdmin` then `pgdb.InitAdmin` twice (l. 40-54) and deletes its row in `t.Cleanup` (l. 58-60). Under `make test` (`-tags=all ./...`) it resolves `DB_NAME=golang_gin_db` from the sourced `.env` and **writes to the dev database**; under a bare `go test` it loads `.env` itself and does the same. The plan's objective (l. 43-44: "la base dev n'est plus écrite ni purgée par les suites — ni par make test, ni par go test direct") is therefore not met as stated, and the T8 documentation ("la base dev n'est plus purgée") would be subtly overbroad if written as specified. The dev-mark proof cannot detect it: the marker lives in `tenant`, while this suite only upserts/deletes its own `admin_user` row. The "Décisions de design" row enumerating the un-pinned remainder lives in the design half (l. 72) and also omits this file. An implementer who discovers the file mid-execution faces an undetermined design choice: pin package `main` too, or accept and document it — a violation of the zero-decision rule.
- Guidance: pick one and write it down. Cheapest: extend the un-pinned enumeration (Risques, l. 339-341 area) and the "Décisions" row with `main_test.go` (package `main`), justified as "upsert idempotent sur `admin_user` + suppression ciblée de sa propre ligne, aucune écriture sur `tenant` — même classe que les seeds ciblés de `controllers`", and soften the T8/README wording accordingly ("les destructeurs globaux et les suites métier vivent sur `artisan_test`" rather than "la base dev n'est plus écrite par les suites"). Alternative: add a `TestMain` to package `main` (then `main_test.go` must enter the allowlist's modified list). Either is fine; leaving it undocumented is not.

**2. Major — The allowlist contradicts itself on `controllers/main_test.go`.**
- Affected scope: Allowlist (plan l. 359 vs l. 369); Phase 3 jalon (l. 319-320: "git status + git diff relus face à l'allowlist").
- Finding: "Modifiés" (l. 359) lists `controllers/main_test.go` (T4), while "Interdits d'écriture" (l. 369) bans "`controllers/*.go` hors `auth_test.go`". `controllers/main_test.go` matches the forbidden clause. The Phase 3 jalon is literally a diff reviewed against this allowlist — as written, the plan's own expected diff violates it, and a strict implementer either stalls or must improvise an exemption (a decision the plan was supposed to pre-make).
- Guidance: reword the interdit to "`controllers/*.go` hors `auth_test.go` et `main_test.go`", or invert the convention to "`controllers/**` : aucun fichier hors `auth_test.go`/`main_test.go`". Same file, one-line fix.

**3. Nit — T6 anchor is off by one.**
- Affected scope: T6 (plan l. 213: "juste après `pgdb.InitAdmin` (l. ~38)").
- Finding: in `controllers/auth_test.go`, `InitAdmin` is at l. 39 (l. 38 is `InitTenant`). The `~` hedges it and the file is 146 lines, so no real ambiguity, but the exact number costs nothing.
- Guidance: cite l. 39.

**4. Nit — `TestCleanUp` anchor drifts in the context block.**
- Affected scope: Contexte (plan l. 15, "~l. 908") — outside the reviewed range but used by it.
- Finding: the function is at l. 907 and the `DELETE FROM public.tenant` at l. 910 of `tests/article_test.go`. Harmless given the `~`, noted for precision only.

## Verified-correct (no finding)

- **T1 (e2edb doc):** all three quoted anchors exist verbatim — package doc l. 1-4, Usage block l. 20-22, "pins the migration superuser (postgres/postgres) and REDIS_DB to 3" inside the l. 24-29 paragraph. Doc-only change; the e2edb unit test asserts behavior, not comments, so the milestone replay is sound.
- **T2 (Makefile):** insertion point before `@echo -e "🔍 Running tests..."` (l. 62) is correct; the snippet is a token-for-token clone of the proven `e2e` recipe pattern (l. 153-158: sourcing, `E2E_DB_ADMIN_*` prefix, un-tabbed recipe continuation, sub-`make` with `DB_HOST/DB_PORT/SSL` passthrough). Command-line `DB_NAME=artisan_test DB_USER=postgres DB_PASS=postgres` correctly override the `-include .env` variables; the `E2E_DB_NAME=artisan_test` command prefix does not collide with the Makefile's `E2E_DB_NAME ?= artisan_e2e` variable. The `REDIS_DB=2` prefix stays scoped to the e2edb call, so the suites' index keeps coming from `TestMain` — consistent with the plan's stated intent.
- **T3/T4 (TestMain):** both current files are exactly the REDIS_DB-only variant the plan replaces; `os.Setenv` after the shipped env wins, and `godotenv.Load("../.env")` (tests/article_test.go l. 153) does not override already-set variables — the pinning mechanism is sound. `middleware/main_test.go` indeed has no PostgreSQL use in that package (verified by grep), so leaving it unchanged is correct.
- **T5/T6 (assertions):** `pgdb.AdminStore()` returns `Store` with exported `Pool *pgxpool.Pool` (internal/db/pool.go l. 15-19), so `Pool.QueryRow` exists. In `tests/article_test.go`, `InitAdmin` is at l. 158 exactly, `ctx` at l. 156, `InitRedis`/`LoadFixtures` at l. 159/165 — insertion "after l. 158, before any write" is exact. In `controllers/auth_test.go`, `context`, `require`, and `pgdb` are all already imported.
- **T7 (CI):** the step is at ci.yml l. 64-74 with `make migrate_up DB_USER=postgres DB_PASS=postgres` at exactly l. 74; the job env has `DB_NAME: golang_gin_db` (l. 50) and **no** `E2E_DB_ADMIN_*` (verified — the plan's justification for the explicit prefix is correct); keep-smoke-on-`golang_gin_db` is real. 10-space indentation matches.
- **Validation milestones are executable:** `go test ./cmd/e2edb ./db` without `-tags=all` genuinely runs both packages (neither test file carries a build tag); the expected e2edb outputs "Dropping the database artisan_test" / "Creating the database artisan_test" / "Flushing the Redis index 2" match the actual `fmt.Printf` strings (cmd/e2edb/main.go l. 151/156/177); the marker columns match migration 000002 l. 19-26 exactly (only `slug`/`name` have no defaults); roles are genuinely idempotent (DO blocks, l. 7-16) so re-migrating `artisan_test` works; the allowlist regex passes for `artisan_test`; `pool_test.go` really is read-only (one control connect + `SELECT current_database()`), so the plan's "assumé, non destructif" classification of it is honest.
- **T8 anchors:** README Testing isolation paragraph exists (l. 161-167); `memory-bank/techContext.md` is 41 lines with l. 41 last ("fin de fichier" claim accurate); `systemPatterns.md` Tests bullet at l. 41; `progress.md` last Historique entry is 23 (so "24" is correct); `activeContext.md` has the « Historique résumé » section; `.kilo/HANDOFF.md` l. 27 contains the quoted « base `golang_gin_db` (goose version 3) » text. AGENTS.md targets (make-test bullet l. 8, CI bullet l. 18, `cmd/e2edb/` structure bullet l. 33, « Tests Redis » conventions bullet l. 70) all exist.
- **Environment notes, phasing, fail-closed stance, alternatives, and out-of-scope sections** are coherent and consistent with the verified baseline (REDIS_DB plan artifacts present in `tests`/`controllers`/`middleware`/`e2edb`/Makefile/CI as claimed).

Bottom line: fix finding 1 (decide and document the fate of root `main_test.go`) and finding 2 (allowlist exemption for `controllers/main_test.go`), and the plan is executable first-try by a fresh agent without further design input.

---

## Part 2 — egm-codebase-details (standard: implementation-readiness-standard)

# Verification Report — Plan `.kilo/plans/1795546592659-artisan-test-db.md`, DetailedRange (lines 113-394)

Audited against the real codebase at `\\wsl.localhost\Debian\home\sylvaindeb\projects\go-artisan-services` (read-only). The design half (lines 1-112) was not audited per instructions, except where the range itself restates structural facts.

## Findings

### 1. Allowlist contradicts itself for `controllers/main_test.go`
- **Severity**: Major
- **Affected scope**: Plan lines 358-370 (§ Allowlist).
- **Discrepancy**: `controllers/main_test.go` is listed under **Modifiés** (line 359), but the forbidden list says `controllers/*.go` hors `auth_test.go` (line 369). Read literally, `controllers/main_test.go` is a `*.go` file that is not `auth_test.go`, so the forbidden rule outlaws a file the Modifiés list requires touching. A mechanical phase-3 allowlist enforcement would block the legitimate T4 edit.
- **Fix direction**: Reword the forbidden entry to `controllers/*.go` hors `auth_test.go` et `main_test.go` (or exempt the two named files explicitly), matching the qualified exception already used for `cmd/e2edb/main.go` ("commentaires de doc uniquement").

### 2. T6 anchor line reference drifted by one line
- **Severity**: Nit
- **Affected scope**: Plan lines 212-213 (T6) vs `controllers/auth_test.go`.
- **Discrepancy**: The plan says the assertion goes "juste après `pgdb.InitAdmin` (l. ~38)". In the file, `pgdb.InitTenant` is at `controllers/auth_test.go:38` and `pgdb.InitAdmin` is at `controllers/auth_test.go:39`. The prose instruction ("just after InitAdmin, before any seed") is unambiguous regardless, so this is a line-reference drift only.
- **Fix direction**: Update to "(l. 39)".

### 3. T2 target-header anchor slightly imprecise
- **Severity**: Nit
- **Affected scope**: Plan line 153 (T2, "En-tête du target (l. 55-60)") vs `Makefile`.
- **Discrepancy**: The comment header of the `test` target actually spans `Makefile:52-60` (starts with `## RUN TESTS` at line 52; the bullets run 53-60). Lines 55-60 cover only the last two bullets. Cosmetic: the instruction ("add one heading line") remains executable.
- **Fix direction**: Cite `Makefile:52-60`.

No Blocker found. No undocumented edge case in `cmd/e2edb` contradicts the plan; no non-existent reference found among items the plan claims exist today.

## Anchors verified as CORRECT

**T1 vs `cmd/e2edb/main.go`** — exact:
- Package doc to be edited: `cmd/e2edb/main.go:1-4` — current text "prepares the disposable environment of the end-to-end suite: … refuses the dev index 1 when REDIS_DB is unset" matches the plan's quoted source text (plan l. 117-123).
- Usage block: line 22 (`//	E2E_DB_NAME=artisan_e2e REDIS_DB=3 go run ./cmd/e2edb`) — the "second example line" insertion point is correct.
- Variables paragraph: lines 24-29 contain verbatim "pins the migration superuser (postgres/postgres) and REDIS_DB to 3" and the "E2E_DB_ADMIN_* when it is set, from DB_USER/DB_PASS otherwise" fallback — both as quoted in the plan.
- All other logic untouched per T1; guards confirmed: allowlist regex line 49 (verbatim `^[a-z][a-z0-9_]{2,62}$`; `artisan_test` matches), target==DB_NAME guard (l. 112-114), protected cluster names (l. 65-69), Redis D3 guard refusing unset REDIS_DB → index 1 (l. 132-134), identifier quoted via `pgx.Identifier.Sanitize` (l. 149).

**T2 vs `Makefile`**:
- Insert-before anchor exact: `@echo -e "🔍 Running tests..."` is `Makefile:62` (plan l. 138 says l. 62 — confirmed).
- Recipe facts (context l. 27-28) confirmed: `test: services_up` (l. 61), `.env` sourcing (l. 63), MSYS pdftotext masquerade (l. 64-66), `go test -mod=readonly -count=1 -p 1 -v -tags=all ./...` (l. 67).
- e2e pattern matches the plan's template exactly: `Makefile:153-158` (echo l. 153; `@if [ -f .env ]…` + `E2E_DB_ADMIN_USER=postgres E2E_DB_ADMIN_PASS=postgres E2E_DB_NAME=$(E2E_DB_NAME) REDIS_DB=3 go run -mod=readonly ./cmd/e2edb` l. 154-155; recursive `migrate_up DB_NAME=… DB_USER=postgres DB_PASS=postgres DB_HOST="$$DB_HOST" DB_PORT="$$DB_PORT" SSL="$$SSL"` l. 157-158). The plan's T2 snippet is a faithful literal-hardcoded mirror.

**T3/T4 vs current pinned files**:
- `tests/main_test.go` (20 lines): `//go:build all`, package `tests`, `TestMain` pinning only `REDIS_DB=2` (`os.Setenv("REDIS_DB", "2")` at l. 15) — the plan's full replacement (T3 code block, plan l. 162-188) is a coherent superset of the current file.
- `controllers/main_test.go` (20 lines): byte-for-byte the same shape, package `controllers` — T4 replacement consistent.
- `middleware/main_test.go` exists and is left out — consistent with T4's parenthetical.

**T5 vs `tests/article_test.go`**:
- `TestIntDB` confirmed: godotenv at l. 153, `ctx := context.Background()` at l. 156, `pgdb.InitTenant` l. 157, `pgdb.InitAdmin` l. 158, `db.InitRedis()` l. 159, existing Redis index assertion l. 160, `LoadFixtures` l. 165 — insertion point "after InitAdmin, before InitRedis/LoadFixtures" is exactly between l. 158 and l. 159; no write precedes it.
- Imports: l. 22 `pgdb "github.com/Massad/gin-boilerplate/internal/db"`, l. 27 `require`, l. 8 `context` — the planned snippet compiles against current imports.
- `TestCleanUp` global sweep confirmed: `DELETE FROM public.tenant` at l. 910 (function header l. 907) — within the plan's cited range.
- Pool usability: `internal/db/pool.go:15-19` — `Store` embeds `*Queries` and exposes exported `Pool *pgxpool.Pool`; `pgxpool.Pool.QueryRow` is the standard API; the same `pgdb.AdminStore().Pool` pattern is already used in the suite (`tests/article_test.go:910`, `controllers/auth_test.go:49,69`).

**T6 vs `controllers/auth_test.go`**:
- Setup block of `TestRefreshOnlyConsumesMatchingOwner` spans l. 27-39; `pgdb.InitAdmin(context.Background(), os.Getenv)` at l. 39; the first write (`testutil.NewTenantFixture` / `CreateTenant`) at l. 44-46. Inserting right after l. 39 precedes every write, as the plan requires. `context`, `require`, `pgdb` all imported (l. 7, 16, 21).

**T7 vs `.github/workflows/ci.yml`** — all anchor claims confirmed:
- `DB_NAME: golang_gin_db` at job env line 50.
- Step "Configure disposable test services" at lines 64-74; `make migrate_up DB_USER=postgres DB_PASS=postgres` at line 74 — exact insertion anchor.
- The job env (lines 42-55) defines `DB_ADMIN_USER`/`DB_ADMIN_PASS` but NO `E2E_DB_ADMIN_*` — the plan's justification (e2edb would fall back to non-superuser `DB_USER`/`DB_PASS`) is confirmed by `cmd/e2edb/main.go:187-192`.
- Redis grep guard exists at lines 89-91 ("Check no hardcoded Redis index", `! git grep -n "InitRedis(1)"`).
- `-p 1` present in the Test step (line 103). Smoke runs with job env (so `golang_gin_db`) as the plan states. `DB_HOST`/`DB_PORT`/`REDIS_HOST`/`SSL` all present in job env, so the added two lines' env needs are met.

**Validation milestones**:
- e2edb expected print strings exist verbatim as printf templates: `"Dropping the database %s"` (`cmd/e2edb/main.go:151`), `"Creating the database %s"` (l. 156), `"Flushing the Redis index %d"` (l. 177) — with `E2E_DB_NAME=artisan_test REDIS_DB=2` the outputs are exactly the three strings awaited at plan l. 292-293.
- Marker INSERT claim: `db/migrations/000002_multi_tenant.sql:19-26` — `tenant(id, slug, name, status, created_at, updated_at)`; only `slug`/`name` are NOT NULL without defaults; `tenant` has no RLS (platform table), so a superuser INSERT with just `(slug, name)` succeeds; the cited line range is exact.
- `go test ./cmd/e2edb ./db` packages exist and are service-free: `cmd/e2edb/main_test.go` (pure anti-DROP guard tests) and `db/db_test.go` (pure: `InitRedis` only constructs a client without connecting — `db/db.go:37-48` — so no Redis service needed).
- Fail-closed claim: `internal/db/pool.go:22-40` — `Connect` pings and returns the error; absent `artisan_test` surfaces `database "artisan_test" does not exist` explicitly; no silent fallback.
- Pinning mechanism sound under make: `Makefile:13-14` (`-include .env` + `export`) exports `DB_NAME=golang_gin_db`, but `os.Setenv` in `TestMain` overrides the inherited value and `godotenv.Load` (tests/article_test.go:153) never overrides an already-set variable — the plan's épinglage reasoning (plan l. 65) is factually correct.

**Allowlist completeness (files claimed to exist)**:
- `tests/main_test.go`, `controllers/main_test.go`, `tests/login_test.go` (forbidden — exists), `internal/db/pool_test.go` (forbidden — exists), `cmd/e2edb/main.go`, `Makefile`, `.github/workflows/ci.yml`, `AGENTS.md`, `README.md`, `memory-bank/{techContext,systemPatterns,activeContext,progress}.md`, `.kilo/HANDOFF.md` — all exist. No file named Modifiés is missing.

**T8 doc anchors** — all present where cited:
- `AGENTS.md:8` (`make test` bullet), `:33` (`cmd/e2edb/` Structure bullet), `:70` ("Tests Redis" conventions bullet), `:18` (CI bullet).
- `README.md`: "### Testing" at line 135; the isolation paragraph from the REDIS_DB plan spans lines 161-167 (Redis index distribution sentence) — the T8 target is real and extendable.
- `memory-bank/techContext.md:41` is the last line and contains the Redis index distribution clause — extendable as planned.
- `memory-bank/systemPatterns.md`: "## Tests" section at lines 39-42; line 41 is the first bullet — the plan's target wording matches.
- `memory-bank/progress.md`: "Historique des étapes" — last numbered entry is 23 at line 51 (REDIS_DB plan) — adding entry 24 is correct and consistent.
- `memory-bank/activeContext.md`: "## Historique résumé" at line 64 — the dated-line insertion point exists.
- `.kilo/HANDOFF.md:27` contains verbatim "base `golang_gin_db` (goose version 3)" inside the stack-dev bullet — the replacement target is exactly as described.

**Environment notes sanity**: `docker-compose.yml` service name `postgres` referenced by the Phase 2 `docker compose exec -T postgres psql` commands is valid (Makefile `postgres_up` targets the same service). Go path note and PowerShell-5.1/WSL bridging notes match the known environment constraints.

## Summary

The implementation range is executable with zero pending design decisions: every file anchor, quote, line reference (within ±1-2 lines), print string, migration column claim, and pool API claim checks out against the current tree. One Major finding: the allowlist's forbidden rule for `controllers/*.go` must be reworded to exempt `main_test.go`, or the T4 edit conflicts with its own write boundary. Two Nits on line-reference precision (T6 "l. ~38" vs actual l. 39; T2 header "l. 55-60" vs actual l. 52-60).

---

## Part 3 — egm-style (no standard)

# Style Review — `1795546592659-artisan-test-db.md`, lines 113-394 (implementation half T1-T8, Validation, Risques, Hors scope, Allowlist, Notes d'environnement)

Overall: the range is dry, imperative, and executor-oriented. Task instructions carry full code blocks, exact anchors, and verbatim expected outputs; the allowlist and interdiction lists are explicit. No Blocker findings. The issues below are precision leaks, unexpanded abbreviations, and recurring editorial inconsistencies.

## Findings

**1. Major — Environment script location is implied, never specified (l. 372-378).**
The note says "écrire le script dans un fichier puis `wsl.exe ... < /mnt/kilo/<chemin> | bash`" and states the mount maps `C:\Users\schat\.kilo` → `/mnt/kilo`, but never says where the script file must be created. A fresh executor will plausibly write it to the workspace or a temp dir, where `/mnt/kilo/<chemin>` cannot reach it. The instruction promises a working procedure but under-specifies the one precondition it depends on. Fix: state the constraint explicitly — the script must live under `C:\Users\schat\.kilo\...`, e.g. `C:\Users\schat\.kilo\tmp\run.sh` ↔ `/mnt/kilo/tmp/run.sh`, and give one worked example command.

**2. Nit — Unexpanded abbreviations (l. 156, l. 312).**
"masquage MSYS" (l. 156) and "1 setup + 3 CUJ" (l. 312) use abbreviations never expanded anywhere in the reviewed range. MSYS is skimmable from l. 28 (out of range) but the executor's range gives no gloss; CUJ (Critical User Journey) appears bare. Fix: one-word glosses on first use — "masquage MSYS (pdftotext de Git for Windows)" and "3 CUJ (scénarios utilisateur critiques)".

**3. Nit — Approximate line anchor mixed with exact anchors (l. 213).**
"juste après `pgdb.InitAdmin` (l. ~38)" is the only tilde-approximated anchor in a range where every other reference is exact. Either the line is known or it is not; the tilde signals to the executor that the anchor might have drifted without saying how to confirm it. Fix: make it exact, or replace the tilde with a disambiguating anchor ("l. 38, le seul appel `InitAdmin` du fichier" — note: it actually is l. 39).

**4. Nit — Two ellipsis conventions in replacement quotes (T1, l. 117 vs l. 132).**
The first replacement pair truncates with the single-glyph "…" (l. 117), the second with ASCII "..." (l. 132). In a document whose purpose is verbatim-editable quoted strings, mixed conventions invite inconsistent transcription. Fix: pick one form and apply it to both.

**5. Nit — Caps emphasis used as a recurring rhetorical device (l. 134, l. 138, l. 199, l. 224, l. 310, l. 382, l. 389).**
AUCUNE, AVANT, APRÈS, EST, TOUJOURS, PAS — used at least seven times in the range. When everything is shouted, the truly critical guards (e.g. `< /dev/null` on l. 382) lose rank. Fix: keep caps only for irreversible/destructive guards; downgrade the rest to bold or plain.

**6. Nit — SQL statement split mid-string (l. 300-302).**
`INSERT INTO tenant (slug, name) VALUES ('dev-keep-marker','Dev Marker')` wraps inside the string literal across two prose lines. For a command the executor is meant to replay verbatim through a `docker compose exec` pipeline, a line break in the quoted SQL is a transcription hazard. Fix: keep the SQL on one physical line and let only the annotation wrap.

**7. Nit — Before/after comparison for the marker count is implicit (l. 303-307).**
"capturer `SELECT count(*) FROM tenant` ... le marqueur → toujours présent et le count inchangé" implies a second capture after `make test` but never instructs it. The comparison — the actual evidence — is left to inference. Fix: state the pairing explicitly: "capturer le count AVANT, relancer la même requête APRÈS `make test` ; les deux valeurs doivent être égales".

**8. Nit — Anglicisms and personifications (l. 324-326, l. 332, l. 349, l. 374, l. 307/316).**
"voulu fail-closed" (l. 325), "sessions stale" (l. 332), "churn cosmétique" (l. 349), "Le dépôt vit sur `\\wsl.localhost\...`" (l. 374), "le balayage ... vit ailleurs" (l. 307). A fresh reader parses "le dépôt vit sur" as odd. Fix: neutral verbs — "s'exécute sur", "cible", "réside sur"; "sessions obsolètes" for stale; "coût cosmétique" for churn.

**9. Nit — Terse code-speak without decode (l. 290; l. 347).**
"`./db` = pur, sans services" and "`internal/testutil/**` (conteneur éphémère propre)" compress meaning into fragments. "Propre" here means "son propre conteneur", not "clean"; "pur" means unit tests without PostgreSQL/Redis. Fix: "`./db` : tests unitaires purs, sans PostgreSQL ni Redis" and "(conteneur éphémère auto-géré, hors provisioning)".

**10. Nit — Vague validation criterion in T7 (l. 237-239).**
"validation statique du YAML (indentation, place de l'étape), symétrie avec les recettes make" — "symétrie avec les recettes make" names no concrete check. Fix: make it a named check: "le bloc e2edb de l'étape CI doit être identique à la ligne e2edb de la recette `test` (env identique)".

**11. Nit — Under-quoted command in T8 (l. 244-246).**
The AGENTS.md bullet is told to describe provisioning "avant `go test -p 1`" — the actual recipe is `go test -p 1 -v -tags=all` (l. 28 of the design half). Fix: quote the full flag set or say "la commande de test existante (`-p 1 -v -tags=all`)".

**12. Nit — Ambiguous possessive phrase (l. 349-350).**
"le doc T1 documente l'usage double" — the intent is "la tâche T1 documente (via le commentaire de paquet) l'usage double". Fix: "le commentaire de paquet de T1 documente l'usage double".

**13. Nit — Arrow symbol overloaded (T1 l. 117-133; Validation l. 288-317; l. 327).**
"→" means "replace with" in T1's edit pairs and "produces / must yield" in the Validation milestones. Fix: use "→" only for results; use "remplacer X par Y" prose in T1.

**14. Nit — Informal heading phrasing (l. 390).**
"**Entrée à rejouer avant de commencer :**" — elliptical. Fix: "Prérequis à vérifier avant de commencer :" with the same five items.

## What holds up

- Imperative precision in the tasks: every insertion gives an anchor line, a verbatim block, and explicit ordering constraints.
- Claim calibration is good where it matters: the Validation milestones quote the exact expected e2edb outputs, the marker test names the failure mode that would betray drift, and the CI limitation is stated as a static-validation-only check rather than overpromised.
- Editorial consistency: French quotes « » used uniformly, em-dashes consistent, code fences correctly typed (`make`, `go`, `yaml`), bold reserved for labels and keys.
- The environment notes are the strongest part of the range (stdin-eating `docker compose exec -T`, Go PATH, git-noise files pinned as do-not-touch) except for finding 1.
