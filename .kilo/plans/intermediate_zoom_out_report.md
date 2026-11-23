# Goldfish Review Report — Stage 8 (Intermediate Zoom Out)

**Target:** `.kilo/plans/1795428465642-login-simple.md` (`1-235`)

## Reviewer Status

| Reviewer | Status |
| --- | --- |
| `egm-zoom-out` | completed |

## Prioritized Findings (Blockers -> Major -> Nits)

The Stage 8 reviewer re-examined the full document with the Stage 5-7 reports in context. It raised no new high-altitude findings; all findings below confirm prior stage findings on the same scopes.

### Unified-login timing side channel — Severity: Major (confirms R7-2; related to Stage 5 catalyst finding)

* **Reviewer Lens:** `egm-zoom-out`
* **Affected Scope:** `UnifiedLogin` semantics (lines 72-75), risks section (line 222)
* **Discrepancy Description:** The sequential dual-path login (« le chemin user prime ; l'admin n'est tenté que si le chemin user échoue (sauf tenant suspendu) ») avoids double bcrypt in the common case but introduces a timing side channel that leaks whether an email exists as a tenant user, contradicting « erreurs 406 génériques (pas d'énumération) ».
* **Rationale & Recommendation:** The simplicity is illusory because it fails a core security requirement, potentially necessitating future rework. Code search found no existing constant-time authentication pattern in the codebase to leverage, and the unified login as specified includes no constant-time handling. Add the mitigation (dummy-bcrypt on absent user, or parallel constant-time evaluation) or retract the enumeration-resistance claim.

### Risks & Mitigations treatment missing from the feature spec range — Severity: Major (confirms R7-1)

* **Reviewer Lens:** `egm-zoom-out`
* **Affected Scope:** Feature spec lines 1-154 vs. dedicated risks content at lines 219-225
* **Discrepancy Description:** The NFR section (lines 147-151) does not frame security measures as identified failure modes with specific mitigations; the dedicated risks content lives only in the detailed plan.
* **Rationale & Recommendation:** Splitting risk treatment across sections increases cognitive load and can lead to overlooked mitigations; the feature spec should carry the risk treatment alongside goals and architecture, per the standard.

### Contradictory Non-Goals labeling — Severity: Major (confirms R7-3)

* **Reviewer Lens:** `egm-zoom-out`
* **Affected Scope:** Non-Goals, line 35
* **Discrepancy Description:** « Vues secondaires fonctionnelles ... restent statiques » labels the views "fonctionnelles" while describing them as static.
* **Rationale & Recommendation:** Unambiguous scope language prevents misinterpretation during implementation; a developer could incorrectly implement functionality for these views, violating the non-goal and adding unnecessary work.

### Token lifetimes without rationale — Severity: Nit (confirms R7-4)

* **Reviewer Lens:** `egm-zoom-out`
* **Affected Scope:** Lines 20, 28
* **Discrepancy Description:** Lifetimes (15 minutes, 8 hours) are stated without threat-model or security-policy basis.
* **Rationale & Recommendation:** Document the rationale to prevent unjustified future changes.

### Redundant and informal language — Severity: Nit (confirms R7-5)

* **Reviewer Lens:** `egm-zoom-out`
* **Affected Scope:** Lines 12, 20, 24, 37, 69, 91
* **Discrepancy Description:** Redundant and informal phrasing increases cognitive load without technical value (e.g., line 12 « supprimée puis recréée à chaque cycle »).
* **Rationale & Recommendation:** Apply the copy edits listed in R7-5.

## Synthesizer Assessment

The intermediate zoom-out converged with the Stage 7 critic findings: no new Blockers, no new simplicity/YAGNI violations, no architectural dead-ends, and no premature-detail findings were raised at full-document altitude. The high-level architecture is judged to hold; the previously raised concerns remain the open items.
