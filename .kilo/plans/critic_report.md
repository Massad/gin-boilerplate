# Goldfish Review Report — Stage 7 (Critic Review)

**Target:** `.kilo/plans/1795428465642-login-simple.md` (`1-154`)

## Reviewer Status

| Reviewer | Status |
| --- | --- |
| `egm-design` | completed |
| `egm-structural` | completed |
| `egm-style` | completed |

## Prioritized Findings (Blockers -> Major -> Nits)

### R7-1 — No "Risks & Mitigations" treatment in the feature spec range — Severity: Major

* **Reviewer Lens:** `egm-design`
* **Affected Scope:** Feature spec lines 1-154, in particular `## Non-Functional Requirements` (lines 147-151); dedicated risks content sits at lines 219-225, outside the audited range
* **Discrepancy Description:** The feature-spec-standard requires a Risks & Mitigations treatment that identifies potential failure modes, security implications, scalability limits, and how they will be mitigated. The active scope contains no such section: the NFR block lists security measures (« erreurs 406 génériques (pas d'énumération) », « identité du tenant uniquement dans le JWT signé », « mot de passe jamais sérialisé (`json:"-"`) ») but does not frame them as identified failure modes with mitigations, and addresses neither scalability limits nor failure modes. The only dedicated risks content (« Risques & garde-fous » : garde anti-`DROP DATABASE`, sémantique de repli, refresh unique, RLS) appears in the detailed plan at lines 219-225, outside the spec.
* **Rationale & Recommendation:** Add a "Risks & Mitigations" section within the feature spec (lines 1-154) enumerating concrete failure modes (e.g., unified-login fallback timing, Redis unavailability, token theft via `localStorage`/XSS), scalability limits, and specific mitigations for each. Synthesizer note: `egm-structural` assessed the standard's required areas as present at header level (it maps "Risks & Mitigations" to the NFR section); the design finding addresses the substance of the risk treatment, which structural review does not evaluate.

### R7-2 — Unified-login timing side channel contradicts the enumeration-resistance claim — Severity: Major

* **Reviewer Lens:** `egm-design`
* **Affected Scope:** Plan lines 25 (claim), 63 (CUJ 3), 72-75 (`UnifiedLogin` semantics), 97-101 (flow diagram)
* **Discrepancy Description:** The spec asserts account-enumeration resistance: « erreurs 406 génériques (pas d'énumération de comptes) » (line 25), and CUJ 3 requires « aucune indication du côté (user/admin) tenté » (line 63). But the fallback semantics — « le chemin user prime ; l'admin n'est tenté que si le chemin user échoue (sauf tenant suspendu) » (lines 73-75) — run bcrypt only when the email exists in the user table. An absent user email short-circuits past bcrypt and produces a measurably faster response than the existing-user-with-wrong-password case, leaking whether an email is registered as a tenant user. No constant-time handling or dummy-bcrypt mitigation is described.
* **Rationale & Recommendation:** Either make the fallback constant-time (always run bcrypt against a dummy hash when the user is absent, or always evaluate both paths before responding) or retract the enumeration-resistance claim and document the residual timing risk. Related but distinct from the Stage 5 inventive-catalyst finding on the same mechanism (dual-account lockout vs. enumeration timing).

### R7-3 — Contradictory labeling in Non-Goals: « Vues secondaires fonctionnelles » — Severity: Major

* **Reviewer Lens:** `egm-style`
* **Affected Scope:** `## Non-Goals`, second bullet (line 35)
* **Discrepancy Description:** « Vues secondaires fonctionnelles : « Mot de passe oublié », « Demander un accès », « Créer une organisation » restent statiques (bannière « fonctionnalité à venir » à la soumission) » — the label « fonctionnelles » contradicts the description « restent statiques ». The noun phrase is intended as a non-goal ("making the secondary views functional is out of scope"), but as written it can be read as calling the views functional.
* **Rationale & Recommendation:** Reword the non-goal to remove the contradiction, e.g. « Rendre fonctionnelles les vues secondaires : ... restent statiques ... », so the excluded scope is unambiguous.

### R7-4 — Fixed token lifetimes asserted without rationale — Severity: Nit

* **Reviewer Lens:** `egm-design`
* **Affected Scope:** Lines 20 (15-minute tenant access token), 28 (8-hour admin token)
* **Discrepancy Description:** « Le token access tenant expire en 15 minutes » and « admin (8 h, sans refresh) » are security-relevant constants stated without rationale, threat-model basis, or reference to a security policy.
* **Rationale & Recommendation:** Document the rationale for each lifetime or tie them to a stated security policy so reviewers can confirm they are appropriate.

### R7-5 — Wordiness, redundancy, and informal phrasing across the spec — Severity: Nit

* **Reviewer Lens:** `egm-style`
* **Affected Scope:** Lines 12, 20, 24, 37, 69, 91 (line references normalized by the synthesizer)
* **Discrepancy Description & Rationale:** Six editorial nits, none changing meaning:
  - Line 12: « supprimée puis recréée à chaque cycle » — redundant; recreation implies prior deletion.
  - Line 20: « la session casse au bout d'un quart d'heure » — the verb « casse » is informal for technical writing, and the clause repeats the already-stated 15-minute expiry.
  - Line 24: « Intégrer fidèlement la maquette complète » — the adverb « fidèlement » adds subjective emphasis unnecessary for a goal; « complète » is redundant given the detailed breakdown that follows (2 volets, 4 cartes, responsive).
  - Line 37: « E2E dans la CI (local uniquement), CI e2e. » — confusing and redundant; suggest « E2E local uniquement, pas de CI e2e. »
  - Line 69: « Table plateforme sans RLS — aucune politique à ajouter » — redundant; a table without RLS inherently requires no RLS policies.
  - Line 91: « liens et « Retour à la connexion » comme la maquette » — filler; fidelity to the mockup is already specified elsewhere in the same line and section.
* **Rationale & Recommendation:** Apply the copy edits.

## Zero-Finding Lenses

* `egm-structural` reported: "No findings." — the feature spec is structurally compliant with the standard: heading hierarchy is sequential (single H1 at line 1, H2 sections, H3 only under CUJs and Technical Plan), formatting is well-formed (consistent `*` bullets, one valid ordered sub-list at lines 109-114, one properly closed code fence at lines 96-105, no tables, no Markdown links), and the standard's required content areas (Problem Statement, Goals & Non-Goals, Proposed Architecture, Alternatives Considered, Risks & Mitigations via NFR) are present within the scope.
