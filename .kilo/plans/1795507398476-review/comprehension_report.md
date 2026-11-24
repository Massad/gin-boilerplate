# comprehension_report.md — Stage 6 (Comprehension Test)

**Target:** `.kilo/plans/1795507398476-docker-pg-redis-dev.md` (full file, document read alone — no GoldfishContext per Stage 6 protocol)
**Reviewer:** egm-comprehension (fresh context)
**Semantic ranges:** FeatureRange 1..41 / DetailedRange 42..267

---

## What this proposes

The document proposes replacing the developer's unusable local database setup with a Docker Compose stack at the repository root that runs PostgreSQL and Redis in containers, published on the loopback interface only, on the non-standard host ports 15432 (PostgreSQL) and 16379 (Redis) so they avoid the neighbouring Docker stacks that already occupy the usual 5432/6379 ports. The two services use the official images `postgres:18-alpine` and `redis:8-alpine`, a superuser `postgres/postgres`, a default database `golang_gin_db`, `trust` authentication (said to be required by an existing connection-string parsing test and already in place in CI), a passwordless Redis, healthchecks that mirror CI, named volumes for persistence, and `restart: unless-stopped`. The stack is driven by four new Make targets (`postgres_up`, `redis_up`, `services_up`, `services_down`), and every Make target that needs a database (`run`, `test`, `e2e`, `migrate_up`, `migrate_down`) gains `services_up` as a prerequisite so the containers start automatically; the `.env` file and its `.env_rename_me` template are changed to point at the new ports, and the README, AGENTS.md, GitHub Actions workflow and memory-bank notes are updated so that the dev and CI versions stay "in parity".

## How the existing system works, per this document

The document describes the starting point as follows: the dev app can no longer start cleanly because neighbouring Docker stacks (honcho, teable, and others) have squatted the standard PostgreSQL/Redis ports and no native PostgreSQL or Redis is running any more inside the WSL Debian environment; the API already listens on port 9001 (via `.env`), the Vite dev server on 5173, and only 15432/16379 were verified free on both the WSL and the Windows side. The work is then spelled out as seven concrete edits: a root `docker-compose.yml` (full YAML given verbatim), four new Make targets plus the `services_up` prerequisite on the consuming targets, two changed lines in `.env`/`.env_rename_me`, documentation updates in the README, AGENTS.md and memory-bank, and a two-tag bump in the GitHub Actions services block (`postgres:15` → `postgres:18`, `redis:7` → `redis:8`). The document argues that no application code or test change is needed: the unit tests touch no port and no database, the `all`-tagged integration suite and the end-to-end cycle both read their host and port from `.env` and therefore follow the containers automatically (the e2e cycle recreating a disposable `artisan_e2e` database and flushing Redis database 1), the `trust` mode satisfies the connection-string password test that would otherwise fail with error 28P01, and the first-time bootstrap is just `make migrate_up DB_USER=postgres DB_PASS=postgres` applying goose migrations 000001-000003 (000002 creating the `app_user`/`app_admin` roles). An 11-step validation sequence, a risk list (daemon stopped, future port collision, destructive `down -v`, trust authentication) and an explicit out-of-scope list (application containerization, Redis TLS/ACL, strong dev passwords, changes to the recipes of the `test`/`e2e` targets) complete the plan.

## Findings

**1. 🟡 Major** — Task 3 and Task 4 (lines 140-141, 145-147): the README is told to keep "the native install valid (the app is driven by `.env`)", but `.env` is simultaneously hard-wired to the container ports 15432/16379 with "no other variable touched", so a developer who installs PostgreSQL/Redis natively ends up with an app that connects to nothing and no instruction to put the ports back. The two halves of the plan are mutually exclusive on the ports, and nothing tells the fresh reader which side wins or how to run the native path.

**2. 🟡 Major** — "Parité dev = CI" (lines 21, 29, 52, 69, 169, 242-244): the document claims dev and CI run "identical tags" and that CI is thereby "validated by construction", yet dev runs `postgres:18-alpine` / `redis:8-alpine` while CI is changed to `postgres:18` / `redis:8` — different variants (and, per the objective line, the images are even named without the `-alpine` suffix at all). The parity argument is the stated justification for moving to PG 18/Redis 8, so a reader checking the claim finds the tags are not identical and "proof" step 11 proves less than it says.

**3. 🟡 Major** — Task 3 and the compose comment (lines 40, 49, 135-141): the document never shows how the application turns `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` into an actual connection string, nor quotes the current `.env` lines it edits, so "ports only" and "`DB_HOST=127.0.0.1` unchanged" are claims about an unseen baseline. A fresh reader cannot confirm that `DB_PORT` is the variable the code actually reads (and that `REDIS_HOST` being a combined `host:port` is handled), which is exactly what task 3 depends on.

**4. 🟡 Major** — Contexte and the persistence decision (lines 5-16, 38): the new stack starts from empty named volumes, but the document never says what happens to the developer's existing native PostgreSQL database and Redis sessions from port 5432 — no mention of importing, migrating, or deliberately discarding that data. A developer with real local dev data has no answer to "where did my data go, and how do I bring it over".

**5. 🟢 Nit** — "miroir CI" / "identiques" (lines 36, 63, 170-171): the compose healthcheck is `pg_isready -U postgres -d golang_gin_db` (requires the database to exist) while the CI healthcheck string is never shown and CI's environment is declared unchanged — "identical" is unverified and plausibly slightly off.

**6. 🟢 Nit** — Tests section (lines 211-212): "wslrelay" is named without any definition, prerequisite check, or statement of whether it is installed or needed, yet automatic mirroring of the two ports to Windows is asserted as a fact.

**7. 🟢 Nit** — Design table, Images row (line 29): the stability and end-of-life claims ("stables vérifiées nov. 2026", "PG 19.0 sorti il y a 5 jours") are sourced to `endoflife.date` with no link and no date-stamped reference the reader can re-check.

**8. 🟢 Nit** — Contexte (lines 6, 12-13): "honcho" and "teable" are named without any indication of what they are or whether they matter to this project, and "Le patron local (chaque repli conventionnel s'est fait squatter)" is hard to parse on a first reading.

**9. 🟢 Nit** — Lines 34, 190, 194, 209: several technical terms used without expansion or pointer: "scram", "28P01", "RESP2", "drop/create FORCE", "skip" — each understandable to an experienced developer, but a fresh reader hitting all of them in one paragraph has to guess; no glossary.

**10. 🟢 Nit** — Validation step 2 and the risks list (lines 217-219, 248-249): the document requires Compose v2 for `--wait` without naming the minimum version, and says the remedy for a stopped daemon or Compose v1 is "documented" without saying in which file — the README and AGENTS.md edit lists (tasks 4 and 5) do not include that prerequisite.
