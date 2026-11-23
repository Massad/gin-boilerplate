# Goldfish Review Report — Stage 5 (Initial Zoom Out)

**Target:** `.kilo/plans/1795428465642-login-simple.md` (`1-235`)

## Reviewer Status

| Reviewer | Status |
| --- | --- |
| `egm-zoom-out` | completed |
| `egm-inventive-catalyst` | completed |

## Prioritized Findings (Blockers -> Major -> Nits)

### Dual-path login fallback accepts an unexamined reliability compromise — Severity: Major (advisory)

* **Reviewer Lens:** `egm-inventive-catalyst`
* **Affected Scope:** `models/login.go` (`UnifiedLogin`) — plan lines 72-75 (semantics), line 222 (risk guardrail)
* **Discrepancy Description:** The plan states « Un même email présent dans les deux tables ne peut pas sonder les deux mots de passe. » The unified login tries the user path first and falls back to the admin path only when the user path fails for reasons other than tenant suspension. When the same email exists in both the user table and `admin_user` with different passwords, a superadmin typing their credentials at `POST /v1/login` cannot authenticate: the user lookup succeeds, bcrypt fails, and the admin path is never attempted. The plan presents this restriction as an unavoidable security property without examining alternatives.
* **Rationale & Recommendation:** Inventive resolution via the "Extraction / Taking Out" pattern: evaluate both authentication paths independently (in parallel, with constant-time discipline to prevent timing-based account enumeration), then return success for the matching account type, preferring the user when both match. **Secondary liabilities if adopted:** every login performs two bcrypt/lookup operations instead of a conditional one; a constant-time comparison requirement; more complex error handling to merge dual-path results. Advisory only — the sequential design is a deliberate, documented trade-off; the author may instead keep it and state an explicit policy (e.g., platform admin emails must never overlap tenant user emails), which the plan currently does not do.

## Zero-Finding Lenses

* `egm-zoom-out` audited the full file against `zoom-out-standard.md` (context sufficiency, root-problem focus, simplicity, anti-weeds, reuse, premature details) and reported: "No findings."
