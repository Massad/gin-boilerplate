# Stage 8 — Intermediate Zoom Out (design)

> Persona : egm-zoom-out (subagent isolé, standard `zoom-out-standard`). Document post-résolution Stage 7 : 167 lignes.
> GoldfishContext : plan-phases + artefacts d'audit (`zoom_out_report.md` Stage 5, `goldfish_review_report.md` resolved).
> Ankrage : numérotation COURANTE du design (167 lignes).

## Verdict

**The intermediate state holds at elevation level.** The 25 fixes applied after Stage 7 did not destabilize the architecture: §4's dependency rule is now factually accurate on its first read (consumer-side definition stated explicitly for both orchestration ports and technical services, l.151), the A2 window now carries its own closure mechanism verbatim (l.151: `git ls-files` = 0), the new §5 groups risks at the right altitude without duplicating the plan's operational content (rollback correctly delegated, l.159), and §6 is a clean inventory with an honest cross-reference to the Non-objectifs (l.163) instead of the contradicting « Questions ouvertes » section. Decision 7 is long, but its depth is justified altitude, not weeds: it is the migration's only security-critical behavioral carve-out, and Stage 7's Blockers correctly demanded the policy be fixed in the design rather than deferred to execution. No pseudocode, no premature optimization, no scope creep beyond the already-acté carve-out. However, the flagship decision 7 makes a **continuity claim that is contradicted by the actual baseline mechanism in the codebase**, which weakens the migration's rollback guarantee — the one genuinely new elevation-level concern this round produced. One false compromise exists only residually, in items consciously deferred without being recorded.

**No false compromise detected** in the newly added content: the soft-consume mechanism, the 5-minute fixed grace with threat analysis, the replay policy (401 delivered / single undelivered recovery / fail-closed crash), and the A2 closure criterion are all priced decisions with named costs, not strawman rejections.

## Findings

### 🟡 Major — Decision 7's "no-change" continuity claim contradicts the baseline: the session record is a Redis STRING, not a hash, and the SHA-256 element is unmotivated

- **Affected scope**: Decision 7 (l.39); propagated identical wording to §3.10 (l.126) and plan-phases Phase 3.
- **Quote (l.39)**: « le record de session reste le hash Redis `user_sessions:*` (**aucune migration de schéma**) ; le script Lua remplace le DEL par un marquage atomique : champs `consumed_at` + `delivered` écrits dans le hash »
- **Why**: Verified in the codebase: the current mechanism stores each session half as a **plain STRING** under the token's UUID key — `db.GetRedis().Set(td.RefreshUUID, value, ...)` (models/auth.go:123), value format « userID:tenantID:role » — and `user_sessions:{userID}` is a **SET index** of UUIDs, not the record (models/auth.go:127, comment l.111-112). The Lua script does a `GET` + string-prefix match, not a hash read (models/auth.go:262-268). Decision 7 therefore asserts « reste » over what is in fact a **format conversion** (string → hash fields `consumed_at`/`delivered`) and conflates the index key with the record — a fresh reader who opens `models/auth.go` cannot reconcile the design with the code it claims to preserve atomically and byte-exactly. Two cascading elevation-level consequences: (a) **rollback coherence** — §5 delegates rollback to the plan's « revert au dernier commit vert » (l.159), but a revert to the pre-Phase-3 binary against Redis records converted to hash format silently breaks every in-flight refresh (string-prefix Lua vs. hash records) → forced re-login for all users; the migration's core guarantee (« revert étape par étape », l.43) does not survive Redis state divergence, and the design never acknowledges the deployment/rollout window; (b) the « hash sha256 du refresh token stocké (jamais le token en clair) » element has **no explained functional role** — the token's UUID key already binds the record to the token, and today the token is never stored in clear anyway (only its UUID key) — so this reads as directive-translated detail with no mechanism need, exactly the kind of speculative element the zoom-out standard rejects. Not a Blocker: the target mechanism is clearly constructible, ownership logic survives unchanged, and the intent is unambiguous — but this coherence break must be fixed before Stage 9, otherwise every implementer of Phase 3 discovers it at execution time.
- **Simplicity rationale**: False continuity in a design whose entire value proposition is « migration = nothing changes except the acted carve-out ». Naming the format change (and its rollout/rollback consequence) costs one sentence; discovering it in code review costs a phase.

### 🟢 Nit — The in-process sweep adds lifecycle complexity for zero functional delta: TTL already evicts everything the sweep would remove

- **Affected scope**: Decision 7 (l.39); §5 (l.157).
- **Quote (l.39)**: « **Cleanup** : sweep in-process du module `auth` (ticker ~5 min) supprimant les records consommés dont la grâce est expirée — TTL Redis = éviction primaire, sweep = filet + observabilité »
- **Why**: At consume time the TTL is reduced to the grace — so Redis itself evicts every consumed record exactly at grace expiry. The sweep can never remove anything TTL would not, by construction. « filet » protects against nothing TTL doesn't (unless one assumes TTL corruption, unstated), and « observabilité » is provided by the logging already mandated. An in-process background ticker introduces goroutine lifecycle (startup, shutdown, testability of the race invariant) into a migration whose non-objective is adding behavior. This was review-mandated (Stage 7 R1) and consciously priced, so it passes — but at elevation level it is the round's clearest YAGNI residue and the design should either name the sweep's actual failure mode it nets against, or downgrade it.

### 🟢 Nit — Ambiguous referent of the `recovered` flag inside the carve-out

- **Affected scope**: Decision 7, replay policy (ii) (l.39).
- **Quote (l.39)**: « pendant la grâce → **une seule** re-émission de récupération (même forme 200 que le happy path, flag `recovered` interdit une seconde), sinon `401` »
- **Why**: Nothing states whether `recovered` is an internal hash field (state machine of the record) or a JSON response field. If it is a response field, it contradicts « même forme 200 que le happy path » within the same parenthetical — a self-contradiction in the migration's single contract carve-out, the precise spot where byte-exactness matters most. If it is the record field that forbids a second recovery, one word (« champ du hash ») fixes it.

### 🟢 Nit — §5 scalability bullet contains a literal overstatement contradicted three lines above

- **Affected scope**: §5 (l.157).
- **Quote (l.157)**: « la migration ne modifie ni les deux pools PostgreSQL (~`app_admin` BYPASSRLS, ~`app_user`), ni les TTL Redis, ni le profil de croissance multi-tenant »
- **Why**: Decision 7 (l.39) explicitly changes TTL usage on the consume path. The same bullet then qualifies itself, so intent is clear, but a naive reader hits an intra-paragraph contradiction on exactly the word « TTL ». Minor wording fix: « ni les TTL Redis *des sessions actives* ».

### 🟢 Nit — Three Stage 5/7-era deferrals remain unrecorded in §6's reports list

- **Affected scope**: §6, « Reports restants » (l.167); Préambule (l.24); §3.9 (l.125).
- **Quote (l.167)**: « Reports restants — casse des clés de réponse (`"Message"`/`"message"`), CORS configurable, refresh httpOnly, câblage RBAC — : détaillés dans « Non-objectifs » »
- **Why**: The original zoom-out (Persona 2, Finding 3) proposed a one-line config extraction of the hardcoded CORS origin — consciously declined (Stage 5 gate), but the declination is recorded nowhere in this document; Persona 2 Finding 4 asked for a post-Phase-8 backlog item normalizing the raw `err.Error()` payloads (`controllers/user.go:94`) — this item is **absent from the §6 list** (only four reports are named), so it has silently dropped out of the audit trail; Persona 1's Nit on thin `model.go`/`dto.go` files was never actioned nor declined (l.45). Add the error-payload normalization to the reports list (or record its declination) so the deferral ledger closes.

## Explicit checks (Stage 5 cross-examination)

- **Stage 5 findings status**: the two Majors fully resolved — dependency-rule contradiction fixed by the rewritten l.151 (verified against decisions 4 and 6 and the plan); buried SPA debts surfaced (l.125, reserve (c), anchor P1.7). Cross-domain transaction coupling resolved by decision 6. Refresh-token loss = decision-7 carve-out, decided. Four Nits partially open, folded into the last finding above. Nothing from Stage 5 silently lost except the error-payload backlog item.
- **Coherence of new content**: decision 7 ↔ l.18 ↔ §3.10 ↔ §5 ↔ §6 mutually consistent (same policy, same 5-min figure, same test list, « aucun reliquat ouvert restant » accurate); §4 ↔ decision 4 ↔ decision 6 terminology aligned. No anchor regressions detected in the 158→167 rewrite.
- **Altitude**: decision 7's field-level detail defensible as policy-level; no pseudocode; no premature optimization beyond the sweep nit and the unmotivated SHA-256 element.
- **Résumé** : 0 Blocker, 1 Major, 4 Nits.

## Re-audit de certification (pass 2, 2026-11-25)

> egm-zoom-out isolé (subagent), lecture ciblée + vérification code (`models/auth.go:119-127/247/262-276`), 168 lignes.

**5/5 findings fixés ; 0 nouveau Blocker, 0 nouveau Major, 0 nouveau Nit d'élévation.** Le document est certifié au niveau zoom-out ; la passe de résolution n'a pas déstabilisé l'architecture (decision 7 désormais exacte face à la base de code, rollback borné écrit, ledger §6 complet). Deux notes mot-niveau sous seuil (shorthand sweep, liste de tests §5) — alignées après l'audit sans nouvelle re-audit (réglage de précision, sous seuil de certification).

