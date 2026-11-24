# Goldfish Annotated Review Report

**Status:** resolved (auto-accept, 2026-11-24 — 21 corrections appliquées au plan cible)

**Target:** `.kilo/plans/1795546592659-artisan-test-db.md` (`full file`)
**Findings source:** `consolidated_review_report.md` (68 raw findings → 0 Blocker · 8 Major · 13 Nit)
**Verification note:** every Major claim re-checked against the codebase in this session — `tests/article_test.go:907-910` (global sweep), root `main_test.go:22-60` (targeted admin_user upsert/delete), `controllers/auth_test.go:39` (InitAdmin anchor), `internal/db/pool_test.go:17-33` (real connect, skip never fires under make), `Makefile:61-67/114-130` (test recipe, install_goose shared GOBIN), `cmd/e2edb/main.go:1-29` (doc anchors), `ci.yml:50/64-74` (job env + step anchors).

Suggestion legend: **CONTEST** (disproves with reviewer-checkable evidence) · **ESCALATE** (none this round) · **PROPOSE FIX** (valid, precise edit defined) · **DECLINE** (nit not worth fixing). Ordering: Majors first, then Nits.

---

## 🟡 Majors

### [M-1] [`Z2-1 + Z3-1`] - Severity: Major
* **Affected Scope:** [Lines 78-80](.kilo/plans/1795546592659-artisan-test-db.md#L78-L80) (« Ce qui ne change pas ») + T5 ([Lines 196-208](.kilo/plans/1795546592659-artisan-test-db.md#L196-L208))
* **Discrepancy Description:** The retained global sweep (`DELETE FROM public.tenant`, `TestCleanUp`, tests/article_test.go:910) is declared harmless merely because the environment is right; a future un-pinned package would silently wipe dev again — the single fail-open spot in an otherwise fail-closed plan.
* **Rationale:** Stage 8 explicitly amends Stage 5's "No false compromise". Verified: the sweep run unguarded at tests/article_test.go:907-910.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Tripwire = one line inside `TestCleanUp`, a file already in `Modifiés`; keeps the sweep's benefit for ad-hoc direct runs between provisionings. The deeper alternative (suite-owned-row scoping) is disproportionate.
* ***Proposed Edit:***
  * Location: `### T5` (after the assertion snippet, l. 206-208).
  * Change: add bullet: « - **Garde du balayage (`TestCleanUp`, l. 907)** : insérer en tête de `TestCleanUp` la même assertion `current_database() == 'artisan_test'` (une ligne, AVANT le DELETE de l. 910 → **avant**) ; le fichier est déjà dans l'allowlist : le balayage devient fail-closed contre un futur paquet non épinglé. »
  * Location: l. 78-80 (bullet TestCleanUp).
  * Change: replace « **`TestCleanUp`** : le balayage global reste tel quel — il devient inoffensif en vivant sur artisan_test (et reste utile aux `go test` directs entre deux provisionnements). » with « **`TestCleanUp`** : le balayage global reste tel quel, précédé de la garde `current_database() == 'artisan_test'` (T5) — inoffensif sur la base jetable, fail-closed si un futur paquet non épinglé atterrissait dessus ; son bénéfice pour les `go test` directs entre deux provisionnements est conservé. »

### [M-2] [`R-1 + Z2-2 + Z3-2`] - Severity: Major
* **Affected Scope:** [Lines 41-44](.kilo/plans/1795546592659-artisan-test-db.md#L41-L44) (Objectif) + l. 72 (Décisions) + [Lines 347-348](.kilo/plans/1795546592659-artisan-test-db.md#L347-L348) (Hors scope) + § Risques + T8 ([Lines 247-248](.kilo/plans/1795546592659-artisan-test-db.md#L247-L248))
* **Discrepancy Description:** Root `main_test.go` (package `main`) writes the dev base (upsert/deletes its own `admin_user` row on `golang_gin_db` under `make test`/CI/`go test -tags=all`) and is absent from every scope enumeration; the objective (« la base dev n'est plus écrite ni purgée par les suites ») is false as written, and T8 wording would inherit the overbroad claim.
* **Rationale:** Verified: root `main_test.go` performs `runCreateAdmin` upserts + targeted `DELETE FROM public.admin_user WHERE email=$1` (main_test.go:58-60) — targeted-row cleanup, same pattern as controllers; the dev-marker proof cannot catch it (marker lives in `tenant`).
* ***AI Suggestion:*** PROPOSE FIX (option a — document the exception; zero-decision rule resolved: pinning package `main` is NOT chosen, no allowlist change)
* ***Justification:*** Option (a) is docs-only, matches the file's actual behavior (verified targeted cleanup), and avoids widening an already large allowlist. Option (b) would add a 4th pinned `TestMain` for one test that never writes globally — churn without risk reduction.
* ***Proposed Edit:***
  * Location: l. 43-44 (Objectif).
  * Change: replace « la base dev `golang_gin_db` n'est plus écrite ni purgée par les suites — ni par `make test`, ni par `go test` direct. » with « aucune suite n'opère de balayage global ni de purge sur la base dev `golang_gin_db` — ni par `make test`, ni par `go test` direct ; exception documentée : root `main_test.go` continue d'écrire/cleanup sa propre ligne `admin_user` (opération ciblée, non globale). »
  * Location: l. 72 (row « Périmètre non épinglé »).
  * Change: extend enumeration with « et root `main_test.go` (upsert ciblé de sa propre ligne `admin_user`) » — see combined M-8/N-8 rewrite below.
  * Location: l. 347-348 (Hors scope) — add: « - root `main_test.go` (paquet `main`, tag `all`) : n'écrit que sa propre ligne `admin_user` sur la base configurée — exception ciblée documentée, non épinglée. »
  * Location: § Risques — add bullet: « - **root `main_test.go`** : sous `make test`/CI, upsert + cleanup de sa propre ligne `admin_user` sur `golang_gin_db` (main_test.go:58-60) — ciblé, non destructeur global ; même classe que les cleanups controllers ; non épinglé, exception documentée. »
  * Location: l. 247-248 (T8, bullet `make test`).
  * Change: replace « + « la base dev n'est plus purgée ; les destructeurs des suites vivent sur `artisan_test` » » with « + « aucune opération globale ne touche plus la base dev ; les destructeurs des suites s'exécutent sur `artisan_test` ; seule écrite résiduelle : la ligne `admin_user` de root `main_test.go` (ciblée, cf. Risques) » »

### [M-3] [`R-2 + CD-1`] - Severity: Major
* **Affected Scope:** [Lines 365-370](.kilo/plans/1795546592659-artisan-test-db.md#L365-L370) (Allowlist, Interdits)
* **Discrepancy Description:** `Modifiés` includes `controllers/main_test.go`; `Interdits` bans « `controllers/*.go` hors `auth_test.go` » — which matches it. The Phase 3 jalon (diff re-read against the allowlist) false-fails on the plan's own expected diff.
* **Rationale:** Verified: l. 359 vs l. 369. The qualified carve-out pattern already exists for e2edb (doc-only).
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** One word removes the self-contradiction; the Phase 3 jalon then compares against a consistent contract.
* ***Proposed Edit:***
  * Location: l. 369.
  * Change: replace « `tests/login_test.go`, `controllers/*.go` hors `auth_test.go`, et la » with « `tests/login_test.go`, `controllers/*.go` hors `auth_test.go` et `main_test.go`, et la »

### [M-4] [`C1-1 (Crit-1)`] - Severity: Major
* **Affected Scope:** [Lines 339-341](.kilo/plans/1795546592659-artisan-test-db.md#L339-L341) (§ Risques, bullet `internal/db/pool_test.go`) + T8 ([Lines 248-250](.kilo/plans/1795546592659-artisan-test-db.md#L248-L250))
* **Discrepancy Description:** The provisioning contract claims more than it delivers: `make test` still needs a migrated dev base. `internal/db/pool_test.go` connects (read-only) to the configured base (`golang_gin_db`) under `make test`/CI; the skip condition never fires there. On a fresh clone — or after `docker compose down -v` — `make test` provisions `artisan_test` then goes red in `internal/db` because `golang_gin_db` doesn't exist. Nothing states this ordering prerequisite.
* **Rationale:** Verified: pool_test.go:17-33 skips only when `DB_HOST`/`DB_NAME` are unset — never under `make test` (`.env` sourced by the recipe); the control connection at l. 25 is a real connect.
* ***AI Suggestion:*** PROPOSE FIX (docs-only, allowlist-compliant)
* ***Justification:*** One Risks sentence + one T8 doc line close the gap without any code change; claiming self-sufficiency would overpromise.
* ***Proposed Edit:***
  * Location: l. 339-341 (row/bullet pool_test.go) — extend.
  * Change: after « lecture seule non destructive ; assumé, documenté » append « ; **prérequis implicite de `make test`** : la base dev configurée doit exister et être migrée (`internal/db/pool_test.go` s'y connecte pour de vrai — le skip ne se déclenche que hors `.env` ; sur un clone frais ou après `docker compose down -v`, faire d'abord `make migrate_up`). »
  * Location: T8, bullet `make test` (l. 248-250) — add: « + « `make test` reste dépendant d'une première `make migrate_up` sur la base dev, à cause d'`internal/db/pool_test.go` (lecture seule, mais connexion réelle à la base configurée) » ».

### [M-5] [`C1-2 (Crit-2) + Z2-4 + Z3-3`] - Severity: Major
* **Affected Scope:** [Lines 331-333](.kilo/plans/1795546592659-artisan-test-db.md#L331-L333) (§ Risques, bullet concurrence) + T8
* **Discrepancy Description:** Concurrency safety is asserted for one pairing (`make test` vs `make e2e` — DB/Redis level genuinely safe, verified: DO-block role creation is exception-safe). Real surface wider: both transitively run `install_goose` writing the same GOBIN binary (Makefile:114-130, `.cache/tools/goose/`); the provisioning `FLUSHDB(2)` erases the live sessions of a concurrent direct `go test`; two concurrent `make test` DROP each other's `artisan_test WITH (FORCE)`; `TestCleanUp` deletes ALL `user_sessions:*`/`admin:*` on index 2, not just its own (tests/article_test.go:916-928).
* **Rationale:** All documentation-class; verified Makefile GOBIN pinning and the Redis key sweep scope.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Narrow the claim (cluster-level safe, command level not serialized) + one operational rule line: docs-only, honest.
* ***Proposed Edit:***
  * Location: l. 331-333 (bullet « make test et make e2e concurrents »).
  * Change: replace « sûr ; le DROP WITH FORCE déconnecte au passage les sessions stale de la base ciblée » with « **sûr au niveau cluster** (bases distinctes, rôles idempotents, index Redis distincts, DO blocks exception-safes) — **mais pas sérialisé au niveau commandes** : `make test`/`make e2e`/`go test` direct lancés en même temps dans le même checkout écriraient les mêmes fichiers (install_goose → même GOBIN) ou se purgeraient mutuellement (FLUSHDB(2) d'un provisioning vs sessions d'un `go test` en cours ; deux `make test` concurrents se DROP l'un l'autre ; `TestCleanUp` vide TOUTES les clés `user_sessions:*`/`admin:*` de l'index 2). Règle opérationnelle : un seul de ces cycles à la fois par checkout. »
  * Location: T8, bullet conventions de test (l. 254-258) — append one line: « ; jamais `make test`, `make e2e` ou `go test` direct en parallèle dans le même checkout (Risques). »

### [M-6] [`Y-1`] - Severity: Major
* **Affected Scope:** [Lines 372-378](.kilo/plans/1795546592659-artisan-test-db.md#L372-L378) (Notes d'environnement)
* **Discrepancy Description:** The environment-notes procedure depends on an unstated precondition: script location. `tr -d '\r' < /mnt/kilo/<chemin> | bash` works only if the script lives under `C:\Users\schat\.kilo\...` (mounted `/mnt/kilo`); a fresh executor will plausibly write it in the workspace where the mount can't reach it.
* **Rationale:** The note announces the mount in parentheses but never binds the `<chemin>` to it.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** One sentence + a worked example removes the only non-deduced step of the execution window.
* ***Proposed Edit:***
  * Location: l. 376-378.
  * Change: replace « écrire le script dans un fichier puis `wsl.exe -d Debian -e bash -c "tr -d '\r' < /mnt/kilo/<chemin> | bash"` (le partage `C:\Users\schat\.kilo` est monté `/mnt/kilo` dans WSL). » with « écrire le **dans `C:\Users\schat\.kilo\...` exclusivement** (seul partage monté `/mnt/kilo` dans WSL — un script écrit dans le workspace est hors de portée du mount), puis `wsl.exe -d Debian -e bash -c "tr -d '\r' < /mnt/kilo/<chemin> | bash"` ; exemple : `C:\Users\schat\.kilo\tmp\run.sh` ↔ `/mnt/kilo/tmp/run.sh`. » (drop the stray « le » after « écrire »)

### [M-7] [`C2-2 + C3-4 + K6-1 + K6-2`] - Severity: Major
* **Affected Scope:** [Line 10](.kilo/plans/1795546592659-artisan-test-db.md#L10) (first mention) + back-references at l. 65/69/93-94/261-262/386/390-391
* **Discrepancy Description:** « plan REDIS_DB » leans on the sibling plan (facts, proof mechanism, rejection rationale) four times without its path; three reviewers flagged it independently.
* **Rationale:** Verified: the sibling is `.kilo/plans/1795535384311-redis-db-isolation.md` (present in-tree, reviewed with the same protocol).
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Naming the file once at first mention makes every back-reference resolvable; later mentions can stay terse.
* ***Proposed Edit:***
  * Location: l. 10.
  * Change: replace « le plan REDIS_DB — isolation des index Redis 1/2/3 — est implémenté » with « le plan REDIS_DB (`.kilo/plans/1795535384311-redis-db-isolation.md`) — isolation des index Redis 1/2/3 — est implémenté ».

### [M-8] [`C3-3 + Z1-2 + Z1-4 + Z1-5`] - Severity: Major
* **Affected Scope:** [Line 68](.kilo/plans/1795546592659-artisan-test-db.md#L68) + [Line 72](.kilo/plans/1795546592659-artisan-test-db.md#L72) (« assumé ») + l. 349-350 (E2E_DB_NAME overload) + T7 (l. 237-239, CI binding) + C2-1 (Risks pointer, design half)
* **Discrepancy Description:** Load-bearing trade-offs marked with the bare word « assumé » assert ownership without the trade-off; the `E2E_DB_NAME` non-e2e overload is never acknowledged; the static-only CI change is never bound to its real acceptance test. C2-1 companion: a « Risques » pointer is missing from the design half.
* **Rationale:** All three one-liners are cheap and docs-only; the consolidated report closes C2-1 on this same calibration pass.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Calibration is the difference between an owned trade-off and a shrug; the fixes cost four lines total.
* ***Proposed Edit:***
  * Location: l. 68 (row « Recréation par cycle », tail).
  * Change: replace « donne un état Redis de départ déterministe — assumé » with « donne un état Redis de départ déterministe — l'index 2 ne porte que les données des suites, jamais les sessions dev (index 1) ».
  * Location: l. 72 (row « Périmètre non épinglé » ; combined with M-2/N-2/N-8 rewrites into one final row).
  * Change: replace the whole row with « | Périmètre non épinglé | `middleware` (Redis seulement : son `TestMain REDIS_DB=2` existant reste inchangé, aucun PostgreSQL dans le paquet), `internal/db/pool_test.go` (lecture seule non destructive ; non épinglable : garde CI `git diff --exit-code internal/db`), root `main_test.go` (upsert ciblé de sa propre ligne `admin_user`) restent hors épinglage — chaque cas nommé dans § Risques ; l'épinglage de `internal/testutil` (helper `current_database()` commun) est refusé par YAGNI jusqu'à un troisième paquet à épingler | ».
  * Location: l. 349-350 (Hors scope, renommage bullet).
  * Change: replace « `cmd/e2edb` (churn cosmétique ; le doc T1 documente l'usage double). » with « `cmd/e2edb` (churn cosmétique ; le commentaire de paquet de T1 documente l'usage double). Le nom `E2E_DB_NAME` surcharge « e2e » depuis les suites Go — renommage rejeté ici, la doc-only T1 est la frontière choisie. »
  * Location: l. 237-239 (T7, validation statique).
  * Change: replace « La CI n'est pas exécutable localement : validation statique du YAML (indentation, place de l'étape), symétrie avec les recettes make. » with « La CI n'est pas exécutable localement : validation statique nommée — le bloc e2edb de l'étape CI doit être identique à la ligne e2edb de la recette `test` (indentation et place de l'étape comprises). Le prochain run CI réel est la preuve engageante de T7 ; une étape rouge y renverra à ces deux lignes. »
  * Location: end of § « Objectifs et hors-objectifs » (after l. 58) — closes C2-1.
  * Change: add pointer line: « *Risques et limites → § Risques / cas d'échec.* »

---

## 🟢 Nits

### [N-1] [`C3-2`] - Severity: Nit
* **Affected Scope:** [Lines 53-58](.kilo/plans/1795546592659-artisan-test-db.md#L53-L58) vs [86-87](.kilo/plans/1795546592659-artisan-test-db.md#L74-L87) (« Ce qui ne change pas »)
* **Description:** Section-level duplication: Non-objectifs and « Ce qui ne change pas » state the same facts twice (e2edb doc-only; e2e inchangé).
* **Rationale:** Keep the full list in Non-objectifs; cross-reference in « Ce qui ne change pas ».
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Cheap dedupe; the TestCleanUp bullet changes anyway (M-1).
* ***Proposed Edit:***
  * Location: l. 76-77 and l. 83.
  * Change: replace the two bullets « **`cmd/e2edb` logique** : ... » and « **`make e2e`** : recette inchangée. » with one bullet: « **`cmd/e2edb` et `make e2e`** : identiques au plan Non-objectifs — zéro changement de logique ou de recette. »

### [N-2] [`Z1-2 + Z1-3`] - Severity: Nit
* **Affected Scope:** T5 (l. 196-208) + [Line 72](.kilo/plans/1795546592659-artisan-test-db.md#L72)
* **Description:** Helper-placement rationale unspoken: why the `current_database()` helper is not in `internal/testutil` (YAGNI until a third pin package) and why `internal/db` can't be pinned (CI guard).
* **Rationale:** Folded into the M-8 rewrite of l. 72 (both rationales) — only the T5 half remains here.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** One clause makes the placement decision explicit where the assertion snippet lives.
* ***Proposed Edit:***
  * Location: end of T5 (after the M-1 tripwire bullet).
  * Change: add: « - (Le helper d'assertion n'atterrit pas dans `internal/testutil` : un seul paquet à épingler en plus de `tests` — YAGNI jusqu'à un troisième ; `internal/db` est de toute façon intouchable, garde CI `git diff --exit-code internal/db`.) »

### [N-3] [`C2-3 + C3-7 + R-3 + R-4 + CD-2 + CD-3 + Y-3`] - Severity: Nit
* **Affected Scope:** l. 15, l. 153, l. 213
* **Description:** Anchor precision pack: `TestCleanUp` « ~l. 908 » → l. 907 (DELETE à 910) ; T6 « l. ~38 » → l. 39 ; T2 « l. 55-60 » → l. 52-60 (« ## RUN TESTS » à 52) ; unify line formats.
* **Rationale:** Verified against the files: article_test.go:907/910, auth_test.go:39, Makefile:52.
* ***AI Suggestion:*** PROPOSE FIX (executable-validation anchors, worth doing before execution)
* ***Justification:*** Anchors are used as diff/verification inputs; wrong ones misdirect the executor.
* ***Proposed Edit:***
  * l. 15: « (tests/article_test.go, ~l. 908) » → « (tests/article_test.go, l. 907 ; DELETE à l. 910) »
  * l. 153: « - En-tête du target (l. 55-60) : » → « - En-tête du target (l. 52-60, « ## RUN TESTS » à l. 52) : »
  * l. 213: « juste après `pgdb.InitAdmin` (l. ~38) : » → « juste après `pgdb.InitAdmin` (l. 39) : »

### [N-4] [`C3-1 + Y-5`] - Severity: Nit
* **Affected Scope:** design half ×4 (l. 18, 32, 55, 70) + implementation half ×7 (l. 134, 138, 151, 198, 224, 388×2)
* **Description:** ALL-CAPS mid-sentence shouting ×11; keep caps only for irreversible guards, neutralize the rest (bold convention already exists).
* **Rationale:** None of these is an irreversible guard; none triggers a destructive action by itself.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Bold-lowercase is the existing emphasis convention; caps stay reserved for genuinely irreversible guards (none here).
* ***Proposed Edit:*** (merge with l. 18 rewrite of N-9 « tape »)
  * l. 18: « tape AUSSI la base dev » → « écrit **aussi** la base dev »
  * l. 32: « ne définit AUCUNE variable » → « ne définit **aucune** variable »
  * l. 55: « : c'EST la base dev » → « : c'est la base dev »
  * l. 70: « AVANT toute écriture » → « **avant** toute écriture »
  * l. 134: « AUCUNE autre retouche » → « **aucune** autre retouche »
  * l. 138: « Insérer AVANT la ligne » → « Insérer avant la ligne »
  * l. 151: « n'est PAS exporté » → « n'est pas exporté »
  * l. 198: « et AVANT `db.InitRedis()` » → « et **avant** `db.InitRedis()` »
  * l. 224: « ajouter APRÈS la ligne » → « ajouter après la ligne »
  * l. 388: « à ne PAS toucher, à ne PAS commiter » → « à ne pas toucher, à ne pas commiter »

### [N-5] [`K6-8 + K6-9 + C3-12 + Y-2 + Y-9`] - Severity: Nit
* **Affected Scope:** l. 31, 28, 156, 290, 312, 348
* **Description:** Unexpanded jargon: « CUJ », « masquage MSYS », « le smoke », « pur, sans services », « conteneur éphémère propre ».
* **Rationale:** Fresh-reader comprehension; zero-cost glosses.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** All five glosses remove a lookup (or a guess) at first read.
* ***Proposed Edit:***
  * l. 31: « le smoke du binaire reste sur » → « le smoke test du binaire reste sur »
  * l. 28: « masquage MSYS du pdftotext xpdf » → « masquage MSYS (pdftotext embarqué de Git for Windows) du pdftotext »
  * l. 290: « db = pur, sans services » → « db = tests unitaires purs, sans PostgreSQL ni Redis »
  * l. 312: « (1 setup + 3 CUJ ; » → « (1 setup + 3 scénarios de bout en bout ; »
  * l. 348: « (conteneur éphémère propre). » → « (conteneur éphémère auto-géré, hors provisioning). »

### [N-6] [`Z2-5 + K6-7 + Z3-4 + Y-6 + Y-7`] - Severity: Nit
* **Affected Scope:** [Lines 300-307](.kilo/plans/1795546592659-artisan-test-db.md#L300-L307) (Phase 2 jalon 1)
* **Description:** Marker proof hardening: seed not rerun-safe (INSERT without prior delete), SQL split mid-string across lines (transcription hazard), before/after count pairing implicit.
* **Rationale:** Verified: `tenant.slug` is unique — a leftover marker breaks the re-run of the proof.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Rerun-safe seed + single-line SQL + explicit pairing make the jalon idempotent and copy-paste-safe.
* ***Proposed Edit:***
  * Location: jalon 1 (l. 300-307).
  * Change: replace the seed sentence with: « 1. **Marqueur dev** : sur `golang_gin_db`, d'abord `DELETE FROM tenant WHERE slug='dev-keep-marker';` puis, sur une seule ligne : `INSERT INTO tenant (slug, name) VALUES ('dev-keep-marker','Dev Marker');` (colonnes vérifiées : db/migrations/000002_multi_tenant.sql:19-26 — tout le reste porte des défauts) ; capturer le count **avant** (`SELECT count(*) FROM tenant`) ; lancer `make test` → vert ; capturer le count **après** — le marqueur **toujours présent** et count avant == count après : preuve que le balayage global de `TestCleanUp` s'exécute ailleurs. »

### [N-7] [`Z2-3`] - Severity: Nit
* **Affected Scope:** T5 (existing InitNoError messages, article_test.go:157-158) + T6
* **Description:** The fail-closed pool-init errors are explicit but not actionable.
* **Rationale:** Both call sites are already touched by T5/T6 — zero extra surface.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Names the remedy at the exact point of failure; converts a red suite into a one-line recovery.
* ***Proposed Edit:***
  * Location: T5, before the assertion insertion paragraph.
  * Change: add: « - **Messages actionnables (N-7)** : étendre les `require.NoError` des deux `Init*` (l. 157-158) : « ...initializing the tenant pool — if the error says the database does not exist, run `make test` once to provision artisan_test; never fall back to the dev database ». Même message T6 (l. 38-39). »

### [N-8] [`K6-3 + K6-4 + K6-5 + K6-6`] - Severity: Nit
* **Affected Scope:** l. 24, l. 72, T3 (l. 160), T2 (l. 147)
* **Description:** Comprehension leftovers: « refuse l'index 1 si absent » ambiguous; middleware un-pinned rationale; TestMain replacement content; why T2 explicitly forwards `DB_HOST/DB_PORT/SSL`.
* **Rationale:** « si absent » reads as « if the index is absent »; the middleware TestMain question and the T2 forwardings need one clause each.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** One phrase/one sentence each; l. 72 is already rewritten wholesale by M-8 (folded there).
* ***Proposed Edit:***
  * l. 24: « (garde D3 : refuse l'index 1 si absent) » → « (garde D3 : refuse l'index 1 quand `REDIS_DB` n'est pas posé) »
  * T3 (l. 160): after « Remplacer le `TestMain` existant (tag `all`, package `tests`) par : » insert « (l'actuel pose uniquement `REDIS_DB=2` ; le remplacement garde les deux épinglages, sans perte) »

### [N-9] [`C3-5..C3-14 + Y-8..Y-14`] - Severity: Nit
* **Affected Scope:** multiple (listed per line)
* **Description:** Style pack: personification (« tape », « vivent/vit sur », « sessions stale »), color slang (already consistent — keep « vert(s)/rouge » as-is), overloaded parentheticals (l. 67, l. 47-49), run-on status sentence (l. 3-6), mid-prose bold (l. 67), « symétrique » fatigue (l. 65/68/95), terse asides (l. 57-58, l. 98), « à l'identique » evidence (l. 84-86), « le doc T1 » (l. 349, folded by M-8), T1 arrow overload, T7 named check (folded by M-8), T8 quoted command, « Entrée à rejouer » (l. 390).
* **Rationale:** Wording-class; each rewrite is local.
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** All are mechanical, meaning-preserving rewrites with verified anchors.
* ***Proposed Edit:***
  * l. 3-6: split the status sentence after « exécution » : « **Statut** : document prêt pour exécution. L'étape suivante est l'implémentation T1-T8 ... »
  * l. 47-49: split the e2edb bullet into two sub-sentences: « recrée déjà n'importe quelle base autorisée via `E2E_DB_NAME`, allowlist `^[a-z][a-z0-9_]{2,62}$` (artisan_test passe). Il purge l'index Redis nommé par `REDIS_DB` — garde D3 : refuse l'index 1 quand `REDIS_DB` n'est pas posé ; gardes anti-DROP : cible == `DB_NAME`, `postgres`, `template*` refusés, identifiant quoté par pgx. »
  * l. 67 (row Réutilisation, with M-8/mid-prose-bold): flatten parentheticals + unbold « **documentation** » (final text from the M-8 l. 72-consistent family edit; also drop « (préfixe limité à cette commande) » into prose).
  * l. 20: « mais vivent sur la même base » → « mais s'exécutent sur la même base »
  * l. 37: « vit sur une base jetable » → « s'exécute sur une base jetable »
  * l. 57-58: « (dommage historique) » → « (perte déjà subie, non restaurée par ce plan) »
  * l. 68: « symétrique d'`artisan_e2e` » → « comme `artisan_e2e` »
  * l. 93: 「 et tape la base dev」 → « et écrit la base dev »
  * l. 84-86: « les migrations se rejouent à l'identique sur `artisan_test` » → « les migrations se rejouent à l'identique sur `artisan_test` (précédent éprouvé : `artisan_e2e` du cycle e2e) »
  * l. 95: « symétrique d'`artisan_e2e` » → « identique au traitement d'`artisan_e2e` »
  * l. 98: « (fail-closed, mais inutilement) » → « (échouerait en base absente — fail-closed, sans bénéfice) »
  * T1 (l. 117-123): replace the arrow chain with prose: « remplacer la phrase d'en-tête « prepares the disposable environment of the end-to-end suite » par : ... »
  * l. 246: « avant `go test -p 1` » → « avant la commande de test existante (`-p 1 -v -tags=all`) »
  * l. 390: « **Entrée à rejouer avant de commencer :** » → « **Prérequis à vérifier avant de commencer :** »
  * l. 248/255: « vivent sur `artisan_test` » → « s'exécutent sur `artisan_test` » (T8 bullets, two occurrences)
  * Colors: keep « vert(s)/rouge » consistently — already uniform across l. 12/106/290-295/386/394: no change (recorded as accepted convention).

### [N-10] [`Z1-1 + Z3-5`] - Severity: Nit
* **Affected Scope:** T4 (l. 191-194) + T6 (l. 210-219)
* **Description:** Claimed task duplication: T4 duplicates T3's 26-line block; T6 duplicates T5's snippet.
* **Rationale:** Half-disproved: T4 already contains one sentence (« Identique à T3, package `controllers`. », l. 191-194) — no 26-line block in the current file. The T6 half holds: l. 215-219 re-lists T5's five-line snippet.
* ***AI Suggestion:*** PARTIAL CONTEST + PROPOSE FIX (T6 half)
* ***Justification:*** CONTEST evidence the reviewer could read: l. 191-194 of the target file. The T6 duplication is real and worth removing — a single canonical definition keeps T5's snippet (its N-7 message extension applies once).
* ***Proposed Edit:***
  * Location: T6 (l. 213-219).
  * Change: replace the snippet block with: « Insérer la même assertion que T5 (même trois lignes, `context.Background()` inchangé), juste après `pgdb.InitAdmin` (l. 39). »

### [N-11] [`C1-4 + Z3-4 (durable part)`] - Severity: Nit
* **Affected Scope:** § Alternatives rejetées (l. 91-98)
* **Description:** Optional durable guard: mirror the Redis CI grep with a guard forbidding the literal `golang_gin_db` in non-comment Go files under `tests/`/`controllers/` — converts the one-shot marker proof into a retained defense. Not required for readiness; adoption is a choice.
* **Rationale:** Adopting would widen T7 (ci.yml) beyond its « 2 lignes » scope and add allowlist pressure; the doc is docs-only.
* ***AI Suggestion:*** PROPOSE FIX (record the deferral decision; adoption itself left as a follow-up the user can request)
* ***Justification:*** The zero-decision rule cuts both ways: the plan must record that the durable guard was considered and consciously deferred, not passed over silently.
* ***Proposed Edit:***
  * Location: § Alternatives rejetées — new row.
  * Change: add: « | Garde CI anti-littéral (`golang_gin_db` banni des fichiers Go non-commentés de `tests/`/`controllers/`, miroir du grep Redis) | rendrait durable la preuve one-shot du marqueur ; hors contract de ce plan docs-only (élargirait T7 au-delà des 2 lignes) — déferré, adoption possible en suivi | »

### [N-12] [`C1-3`] - Severity: Nit
* **Affected Scope:** [Lines 334-335](.kilo/plans/1795546592659-artisan-test-db.md#L334-L335) + Phase 1 jalon 3
* **Description:** The measured provisioning cost is unbacked (« ≈ quelques secondes ») and is paid on every `make test`/CI run.
* **Rationale:** Measure once at Phase 1's jalon 3 (the provisioning output is already visible there).
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Converts an estimate into a recorded figure at zero extra execution cost.
* ***Proposed Edit:***
  * Location: l. 334-335.
  * Change: replace « ≈ quelques secondes — acceptable pour le déterminisme obtenu » with « figure mesurée une fois au jalon 3 de la Phase 1 (chronométrer le bloc e2edb + migrate_up de la recette) et remplacée ici en fin d'exécution — à confronter au déterminisme obtenu ».

### [N-13] [`C1-5`] - Severity: Nit
* **Affected Scope:** T8 (AGENTS.md bullet, l. 243-250)
* **Description:** The pinned superuser `postgres/postgres` (T2 recipe + CI) is dev/CI-only but never stated — same silence as the `e2e` recipe.
* **Rationale:** One sentence in T8; the e2e recipe keeps its existing silence (parity).
* ***AI Suggestion:*** PROPOSE FIX
* ***Justification:*** Documents the credential scope at the same spot where the pins are defined in doc-form.
* ***Proposed Edit:***
  * Location: T8, bullet `make test` (l. 244-246).
  * Change: append to the provisioning clause: « (superuser `postgres/postgres` épinglé : dev/CI uniquement — même convention que la recette `e2e` ; prod via `ALTER ROLE`) ».

---

## Consolidated decision state

- **PROPOSE FIX:** 20 items (M-1..M-8 incl. folded C2-1; N-1..N-9, N-11..N-13; N-10 fix-half).
- **CONTEST:** 1 partial (N-10, T4 half — evidence: target l. 191-194 already a reference).
- **ESCALATE:** none.
- **DECLINE:** none outright (colors sub-part of N-9 recorded as already-consistent: keep).

Editing passes once accepted (suggested order): ① M-3 (one-line allowlist) → ② M-1 + M-2 (T5 scope + objective/allowlist/risks) → ③ M-4 + M-5 + M-6 + M-8 + M-7 (Risques/T8/Notes calibration pass) → ④ N-3 + N-6 (executable anchors/proof text) → ⑤ remaining N items (text polish).

---

## Résolution (auto-accept — 2026-11-24)

- **Appliquées :** 20 items PROPOSE FIX intégraux + la moitié valide de N-10 (référence T6 en place du snippet dupliqué) = **21 corrections** toutes éditées dans le plan cible.
- **Contestées, non appliquées :** 1 — N-10 moitié T4 (preuve : le fichier cible l. 213-216 contient déjà « Identique à T3, package `controllers` »).
- **Escaladées :** 0. **Refusées (DECLINE) :** 0 (sous-partie « couleurs » de N-9 : déjà cohérente, conservée telle quel).
- **Post-édition :** balayage de vérification propre (aucun caps restant, aucune personnification restante, ancres/chemins posés) ; « limite assumée » du bullet CI conservée volontairement — liée désormais à sa preuve engageante (T7).
- **Item reporté hors plan (décision enregistrée) :** N-11 — la garde CI anti-littéral est déferrée via la nouvelle ligne d'« Alternatives rejetées » ; son adoption reste possible en suivi.
