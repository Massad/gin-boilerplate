# Stage 5 — Initial Zoom Out Review (design)

> Review de certification du document `1795597264251-architecture-modulaire-par-domaine-design.md` (156 lignes).
> Ancrages = numérotation courante du fichier design (pas l'original 268 lignes).
> GoldfishContext : `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`.
> Deux personas Goldfish isolés (subagents à contexte frais), consolidés par l'orchestrateur.

---

## Persona 1 — egm-zoom-out (standard : zoom-out-standard)

Mode: REVIEW. Evaluated against the Zoom-Out Standard (context sufficiency, root problem focus, simplicity/YAGNI, anti-weeds, leverage existing). Anchors are to the design document's actual current lines. Baseline context: the sibling execution plan `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`.

### Verdict

The design passes the prerequisite check: the objective and its motivation are explicit (lignes 10-18: problem = global wiring, no DI, business logic in handlers, duplicated route setup; cost of not migrating stated). The root problem — non-injectable global wiring — is addressed structurally (composition root `internal/app`, module layout, consumer-defined ports), not with a band-aid on one handler. Alternatives are genuinely considered and rejected with named costs (lignes 37-42), including the correct YAGNI rejection of full hexagonal architecture. Speculative work is explicitly deferred (version bumps, RBAC wiring, CORS config, httpOnly refresh, API-key case homogenization — lignes 149-156). `internal/platform/httpx` (l.135) is deferred to Phase 2 until the first real caller — good anti-YAGNI posture. I verified in the codebase that no existing `ParseUUIDParam`/`MapPgxError`/`AbortMessage`-style helpers exist, so httpx is not reinventing anything, and the design's claim that `getUserID`/`getTenantID`/`getRole` helpers exist at `controllers/user.go:19-33` is accurate. The pre-existing CI drift claim (l.119) is factual: `.github/workflows/ci.yml:139,157,166` still probes `/v1/invoice` routes.

Findings below.

### Findings

#### 🟡 Major — The single stated dependency rule is contradicted by the design's own cross-module wiring
- **Affected scope**: §4 Layout cible, dependency rule; tension with « Décisions structurantes » item 4.
- **Quote (l.147)**: « modules → modules uniquement via **interfaces définies côté consommateur** (ports) »
- **Why**: Decision 4 (l.34) has « `user`/`admin` consomment le `TokenService` d'auth » and the plan defines `TenantProvisioner` inside module `tenant` — both are provider-defined interfaces/services exported by the providing module, not consumer-defined ports. Only the `auth` ports (`UserAuthenticator`/`AdminAuthenticator`/`UserReader`, l.34) and `ArticleRepository` follow the consumer-defined rule. The universal rule as written is therefore false on its first read, and a fresh reader cannot tell which cross-module contracts are ports defined by consumers versus interfaces exported by providers. This is an elevation-level boundary rule that must be stated accurately (e.g., « modules → modules via interfaces, définies côté consommateur pour les ports d'orchestration, côté fournisseur pour les services techniques comme `TokenService` »), or the rule will be enforced inconsistently during review of each phase.
- **Simplicity rationale**: Not a complexity finding — a coherence finding. An imprecise architectural rule invites future cycles or ad-hoc exceptions.

#### 🟡 Major — A real architectural compromise is buried in prose without being flagged as a decision
- **Affected scope**: §3 item 9, SPA contract.
- **Quote (l.121)**: « le plan recopie `main.go:127-212` dans `internal/app` (fallback SPA `NoRoute` : 404 JSON sur `/v1*`/non-GET uniquement, `serveSPAIndex` no-cache, `/assets` `immutable`, `/public` statique, Swagger hors prod) »
- **Why**: Copying 85 lines of router wiring verbatim into a shared composition root (imported both by `cmd/api` and by tests) preserves the contract but converts `main.go` into a dead reference during the window — the plan runs an equivalence test (P1.7) precisely because two wirings now coexist. The plan documents two debts there: the global `binding.Validator` mutation inside `NewRouter` and the Δ surface with the old router. The design document itself, which is the elevation-level artifact, never surfaces these debts; a reader of the design alone cannot act on them. The design should state them (or reference the plan's debt list explicitly) so the exit criteria in Phase 7/8 include their elimination.
- **Simplicity rationale**: This is an honest, defensible transitional compromise — not a false one — but compromises at this altitude must be visible at this altitude too.

#### 🟢 Nit — "Uniformity" justifies thin boilerplate files
- **Affected scope**: Alternatives considered, item 3.
- **Quote (l.41)**: « Coût assumé : gabarit parfois mince (`dto.go`/`model.go`), accepté pour l'uniformité. »
- **Why**: Accepting near-empty files solely for template uniformity is a mild YAGNI violation: files whose existence is motivated by the template rather than by content. The compromise is consciously priced, so this is not a blocker. Recommend one sentence allowing a module to omit `model.go`/`dto.go` when its content would be a pure re-alias — the layout rule should serve the migration, not perform it.

#### 🟢 Nit — Transitive-import tolerance window is safe but stated as one rule in two voices
- **Affected scope**: §4 dependency-rule paragraph.
- **Quote (l.147)**: « **Fenêtre transitoire Phases 3-6 (décision A2)** : les imports racine↔modules sont tolérés **dans les deux sens**, uniquement pour le code en cours de migration »
- **Why**: The design correctly states « `platform` n'importe jamais `modules` (invariant absolu, toute la migration) » in the same paragraph, and the plan backs the window's closure with a measurable exit criterion (`git ls-files` count = 0, Phase 7). No action needed beyond ensuring the design repeats the closure criterion verbatim, since the design's own paragraph names the window without its enforcement mechanism.

#### 🟢 Nit — Claims requiring off-document context to verify
- **Affected scope**: §3 item 7, §2 extension-point paragraph.
- **Quote (l.109)**: « les futures invitations/membres (listées en backlog `memory-bank/progress.md:56`) »
- **Why**: The design leans on external references (`memory-bank/progress.md:56`, `.kilo/plans/1795284300000-remove-public-access.md`, prior review reports in `.kilo/plans/`, l.4) for load-bearing claims (backlog existence, the history of the removed invoice routes). Each reference is legitimate, but the design should state the essential fact inline (as it correctly does for the CI drift at l.119) rather than delegating its justification to a file a fresh reviewer may not read.

### Explicit checks

- **No false compromise detected** in the rejected-alternatives list: full hexagonal (l.41), version upgrades during migration (l.22, l.151), RBAC wiring (l.23, l.156), CORS configurability (l.24, l.154), httpOnly refresh (l.25, l.155) are all rejected with real reasons and honest costs, not strawmen. The incremental 9-phase approach for a codebase this size (l.39, cost acknowledged: « 9 phases à exécuter ») is proportionate given the byte-exact HTTP contract and RLS fail-closed invariants it protects — the design runs `make e2e` only at Phases 1 and 7 (l.66 of the plan) which shows the ceremony is capped, not maximal.
- **No premature optimization detected**: no caching, partitioning, async queues, or speculative abstractions anywhere.
- **No altitude drops detected**: the document's density of file/line anchors is cartography of existing code (baseline evidence), not implementation detail — no pseudocode, no helper signatures beyond interface contracts.
- **Reuse verified by code search**: no existing equivalents to the proposed `httpx` helpers or module-level repositories exist in the codebase; the proposal reuses Gin, sqlc, pgx pools, mockery, and the existing test suite as the contract net rather than building parallel verification — pattern-consistent with AGENTS.md.
- **Architectural dead-ends**: none. Platform-never-imports-modules plus consumer-side ports preserve future expansion (members/invitations need only wire `RequirePermission` in `internal/app`, l.109); the module boundary does not preclude the future billing domain (l.64, l.89).

---

## Persona 2 — egm-inventive-catalyst (Mode: REVIEW)

Target read: `.kilo/plans/1795597264251-architecture-modulaire-par-domaine-design.md` (156 lines, all anchors are that file's own numbering). Contradiction check (REVIEW step 0) applied first: several compromises in the document are genuinely examined and balanced, not false — the dependency-version freeze (lines 22, 151), the incremental-vs-big-bang choice (line 39), and the rejection of full hexagonal architecture (line 41) each carry an explicit, weighed cost. No transform manufactured for those. Four genuine false compromises / unexamined liabilities remain. None reach Blocker.

### Finding 1 — 🟡 Major

**Affected scope:** Section « Questions ouvertes » (line 152, refresh flow)
**Anchored passage:** "si la relecture utilisateur échoue après `ConsumeRefreshAuth`, le jeton est détruit sans contrepartie (session perdue) — comportement actuel, épinglé et inchangé pendant la migration ; backlog : rendre l'incident observable/réversible (événement de session côté support) plutôt qu'un simple 401." — line 152

**False compromise:** The document freezes a user-hostile failure mode as an immutable law: a transient read failure (database hiccup, pool exhaustion) after token consumption permanently destroys a valid, non-revoked session. Pinning this behavior during migration is correct; the compromise is that the named backlog reduces the resolution to a support-desk observability event — an administrative workaround — without examining whether the mechanical contradiction itself (atomic consume vs. read-failure window) can be eliminated. That is a bottleneck presented as unavoidable.

**Proposed resolution:** *Beforehand Cushioning* (catalog #4): two candidate transforms for the backlog, both preserving the existing ownership semantics. (a) Soft-consume: the Lua script marks the token consumed with a short TTL instead of deleting it; a failed post-consumption read reverses the mark within the window, and the grace TTL expires for the normal path. (b) Reorder: read user first, then consume via the ownership-checked script — closes the loss window but opens a different one (the token survives slightly longer). Either way, the backlog should name the chosen pattern and its tests now, not defer to an unspecified "support event".

**Secondary liabilities:** (a) A grace window weakens the immediate-revocation property of `ConsumeRefreshAuth` and adds TTL management plus new race tests; replay inside the grace window must be explicitly bounded or blocked (second-consume check). (b) Reordering breaks the atomicity currently proven by `TestRefreshOnlyConsumesMatchingOwner` if the ownership check is not replicated on the earlier path; both options add failure-path logic the current single-step script avoids.

### Finding 2 — 🟡 Major

**Affected scope:** Section 3 « Couplages et points de friction », item 4 (line 116) — cross-domain transaction design; materialized in the plan's `TenantProvisioner.CreateInTx(ctx, q *pgdb.Queries, ...)` port
**Anchored passage:** "Les futurs repos `tenant` et `user` doivent partager le même `Store.InTx` — le service d'orchestration (user) pilote la Tx, les deux repos exposent des opérations « in-tx » sur `*pgdb.Queries`." — line 116

**False compromise:** The document accepts, without examining an alternative, that every cross-domain write couples two modules through a platform concrete type: the `tenant` port signature must expose `*pgdb.Queries`, and the `user` module owns commit semantics for a write inside another domain's repository. This is stated as a requirement ("doivent") rather than a weighed choice. The very domain extension the migration is meant to enable (members/invitations, line 109) will replicate this pattern, so the coupling compounds with each new multi-table write.

**Proposed resolution:** *Intermediary* (catalog #8) and/or *Inversion* (catalog #5). Intermediary: a platform-level unit-of-work (`pgdb` exposes `WithinTx(func(q *pgdb.Queries) error)`), with module ports limited to in-TX operations so that `*pgdb.Queries` never appears in a cross-module port signature — the raw type stays inside `platform/db` and the orchestration wiring. Inversion: a `RegisterTenantWithOwner(provision func(...))`-style provisioner owns the loop body and calls back into `user`-supplied logic, so `user` never touches tenant tables through a hand-piloted transaction.

**Secondary liabilities:** A unit-of-work intermediary adds one orchestration hop and narrows the escape hatch for genuinely mixed operations beyond the two current cases; the callback/inversion version hurts stack-trace readability and complicates error typing across the callback boundary (errors from two modules now flow through one closure). Both add interface surface that the byte-for-byte contract constraint makes awkward to retrofit later — which is why the unexamined point matters now.

### Finding 3 — 🟢 Nit

**Affected scope:** Section 3, item 9, reserve (a) (line 121) — CORS
**Anchored passage:** "(a) `Access-Control-Allow-Origin` est codé en dur `http://localhost` (`middleware/cors.go:10`) — sans effet en single-origin, mais à rendre configurable (`internal/platform/config`) dès qu'une origine croisée existera" — line 121

**False compromise:** A hardcoded platform constant is defended by a purely reactive trigger ("as soon as a cross-origin becomes real"). The migration itself moves this middleware into `internal/platform/config`'s neighborhood in Phase 1 — extracting the string as one config read costs a single line, yet the choice defers it indefinitely, so the "hardcoded platform value" anti-pattern the migration exists to remove is knowingly retained in its own target directory.

**Proposed resolution:** Plain extraction under *Dynamics* (catalog #6): read the origin from `internal/platform/config` (env-driven, defaulting to the current literal), removing the hardcode without changing single-origin behavior.

**Secondary liabilities:** One more env variable to document and validate; a misconfigured origin only surfaces when cross-origin becomes real, so a config-validation note (not a migration change) is needed.

### Finding 4 — 🟢 Nit

**Affected scope:** Section 3 « Couplages et points de friction », item 5 (line 117) — error mapping
**Anchored passage:** "`controllers/user.go:94` renvoie `err.Error()` brut au client (messages models — pas de fuite SQL, mais format non maîtrisé)." — line 117

**False compromise:** The migration plans `httpx.MapPgxError` (Phase 2) for status codes but leaves the *payload text* problem unaddressed: an uncontrolled internal error string is being frozen as an unchanged byte-for-byte contract, and no post-migration backlog item exists to normalize it — while other follow-ups (case homogenization, httpOnly, CORS) are each listed in the open-questions section. An architecture decision (typed domain errors vs. passthrough) exists and is simply absent.

**Proposed resolution:** Add a post-Phase-8 backlog item: *Extraction* (catalog #2) — route all client-facing error payloads through the `httpx` responders with typed domain errors instead of handler-level `err.Error()` passthrough, in one coordinated change.

**Secondary liabilities:** The change is client-visible to the SPA, so it requires a coordinated frontend review of consumed error messages; typed domain errors add an error-hierarchy layer to maintain per module, and any normalization must respect the constant-cost login error policy (406 generic path must not start leaking distinctions).

### Summary

Four findings: 2 Major (refresh token destroyed on transient read failure with a support-workaround backlog instead of a designed fix — line 152; cross-domain transaction coupling user→tenant via `*pgdb.Queries` stated as a law with no intermediary/inversion alternative weighed — line 116), 2 Nit (CORS hardcode kept by a reactive trigger in the very directory the migration builds to remove hardcodes — line 121; uncontrolled `err.Error()` payloads frozen as contract with no backlog item — line 117). The remaining architectural choices audited (version freezes, incremental phasing, hexagonal rejection, RBAC extension point, transitory import window with closure criterion) are examined compromises with declared costs and require no inventive transform.
