# Stage 7 — Critic Review (design) — rapport consolidé

> Round 1. 3 personas Goldfish isolés (subagents contexte frais), dispatch direct.
> Standard : `feature-spec-standard` (egm-design, egm-structural) ; egm-style = persona seul (aucun standard désigné).
> TargetFile : `1795597264251-architecture-modulaire-par-domaine-design.md` (158 lignes, FeatureRange = 1..158).
> GoldfishContext : `…-plan-phases.md`. Aucun reviewer en échec.
>
> Index : 2 Blockers, 7 Majors, ~20 Nits. Doublons flaggés (≈) dans la table de consolidation.

## Consolidation (FindingsList)

| ID | Reviewer | Sévérité | Scope | Objet |
|---|---|---|---|---|
| design-1-1 | egm-design | 🔴 Blocker | l.37 (décision 7) | Mécanisme de stockage soft-consume contradictoire/inimplémentable (`consumed_at` nullable vs sessions Redis ; aucune migration goose planifiée ; TTL de grâce en plage non tranchée) |
| design-1-2 | egm-design | 🔴 Blocker | l.37, l.154 | Politique de rejeu (« 401 invalid_grant » vs re-émission) différée à l'exécution sans analyse de risque (vol de token pendant la grâce) — ≈ structural-1-4 |
| design-1-3 | egm-design | 🟡 Major | l.18 | Objectif « byte près » non qualifié vs carve-out décision 7 |
| design-1-4 | egm-design | 🟡 Major | l.37 | Cleanup job non spécifié (owner, cadence, invariant de course avec la grâce) |
| design-1-5 | egm-design | 🟡 Major | l.123 | Justification « make e2e Phases 1 et 7 » non démontrée (Phase 3 rebranche tout le middleware chain) |
| design-1-6 | egm-design | 🟢 Nit | l.3-4 | Renumérotation vs ancrages historiques (acceptable, constat) |
| design-1-7 | egm-design | 🟢 Nit | l.113-124 | Contenu risques inline + délégation au frère — le réel manquement = analyse de risque décision 7 (couvert par design-1-1/2) |
| structural-1-1 | egm-structural | 🔴 Blocker | document | Section « Risks & Mitigations » absente (non vérifiable à l'en-tête) ; contenu risques dispersé sous en-têtes trompeurs ; aucune analyse scalability |
| structural-1-2 | egm-structural | 🟡 Major | l.151-158 | Titre « Questions ouvertes » contredit le contenu (5/6 items tranchés/différés) — ≈ style-M3 |
| structural-1-3 | egm-structural | 🟡 Major | l.149 | Prose cassée dans la parenthèse A2 : fragment « Impact architectural ici) : » + virgule de fusion |
| structural-1-4 | egm-structural | 🟡 Major | l.37, l.154 | Choix de sécurité reporté à l'exécution — ≈ design-1-2 |
| structural-1-5 | egm-structural | 🟡 Major | l.29+ | Provenance « (validation utilisateur) » sans ancre vérifiable (aucun lien artefact/session) |
| structural-1-6 | egm-structural | 🟢 Nit | l.6, l.151 | Numérotation d'en-têtes asymétrique (Préambule/Questions non numérotés vs 1-4) |
| structural-1-7 | egm-structural | 🟢 Nit | l.3, l.37, l.121, l.123 | Style de renvoi croisé incohérent (lien markdown vs backticked paths) |
| structural-1-8 | egm-structural | 🟢 Nit | l.3 | Lignes 117 et 260 de l'original non réclamées par le mapping de scission |
| structural-1-9 | egm-structural | 🟢 Nit | l.113 | §3 « Couplages » héberge des items hors couplage (7, 9, 10) |
| structural-1-10 | egm-structural | 🟢 Nit | l.8 vs l.140 | Anatomie module du préambule omet `model.go` |
| style-1-1 (M1) | egm-style | 🟡 Major | l.4 | Métadonnées EGM redondantes (stages listés deux fois dans la même phrase) |
| style-1-2 (M2) | egm-style | 🟡 Major | l.33, l.44 | « platform » vs « plateforme » (2 genres, 1 concept) |
| style-1-3 (M3) | egm-style | 🟡 Major | l.151-158 | Questions ouvertes : items tranchées + duplication verbatim des Non-objectifs (l.156=l.24, l.158=l.23) |
| style-1-4 (M4) | egm-style | 🟡 Major | l.124 | Point 10 : même analyse deux fois (~1/3 du texte réductible) |
| style-1-5 (M5) | egm-style | 🟡 Major | l.123 | Synonymes de préservation en rotation (tel quel/verbatim/copie/à l'identique) |
| style-1-6 (M6) | egm-style | 🟡 Major | l.18, l.37, l.78, l.123, l.124 | 5 libellés pour la contrainte byte-exactness (canon recommandé : « inchangé au byte près » + « recopié à l'identique ») |
| style-1-7 (N1) | egm-style | 🟢 Nit | l.8 | « transvers » → « transversal » |
| style-1-8 (N2) | egm-style | 🟢 Nit | l.36 | « ses repository » → « ses repositories » |
| style-1-9 (N3) | egm-style | 🟢 Nit | l.115 | « mêmes singletons … que cartographiés » → « identiques à ceux cartographiés » |
| style-1-10 (N4) | egm-style | 🟢 Nit | l.12, l.42, l.115+ | « wiring » vs « câblage » mélangés |
| style-1-11 (N5) | egm-style | 🟢 Nit | l.41 | « contrat HTTP doit rester vert » (catégorie : vert = CI, pas contrat) |
| style-1-12 (N6) | egm-style | 🟢 Nit | l.123 | « précisément » (emphase inutile) |
| style-1-13 (N7) | egm-style | 🟢 Nit | l.149 | Fragment A2 + virgule de fusion + « le 2026-11-25 » (article superflu) — ≈ structural-1-3 |
| style-1-14 (N8) | egm-style | 🟢 Nit | l.37 | « le retient » ambigu (→ « si la politique retenue est la re-émission ») |
| style-1-15 (N9) | egm-style | 🟢 Nit | l.124 | « cf. fin de point » auto-référence inutile |
| style-1-16 (N10) | egm-style | 🟢 Nit | l.123, l.156 | « sans effet en single-origin » vs « inopérante » (dérive de synonyme) |
| style-1-17 (N11) | egm-style | 🟢 Nit | l.12, l.67, l.120 | Même constat « wiring épinglé par smoke CI » ×3 (connexions « et » vs « + ») |
| style-1-18 (N12) | egm-style | 🟢 Nit | l.12, l.111 | « multiplie » (spéculatif non calibré) ; « n'auront qu'à » (minimisation non calibrée) |
| style-1-19 (N13) | egm-style | 🟢 Nit | l.119 | « en central » (registre oral → « centralisé ») |
| style-1-20 (N14) | egm-style | 🟢 Nit | l.62 | « Interdit à la main » (ellipse → « édition manuelle interdite ») |
| style-1-21 (N15) | egm-style | 🟢 Nit | (var.) | Divers ajustements de formulation listés dans le rapport du reviewer (section Mineurs N1-N15 ci-dessous) |

> Doublons pour résolution : [design-1-2 ≡ structural-1-4], [structural-1-2 ≡ style-1-3], [structural-1-3 ≡ style-1-13]. Un même texte résout chaque cluster.

## Rapports individuels

### egm-design (standard : feature-spec-standard)

Audited against `feature-spec-standard.md`, from the architecture/risk angle. The document is dense, well-anchored, and mostly passes the standard's structural requirements (Problem Statement §"Problème", Goals & Non-goals, Proposed Architecture §4, Alternatives §"Alternatives considérées"). But it carries one fundamentally underspecified design — the only behavioral carve-out of the whole migration — and that is where the audit concentrates.

#### Findings

**1. 🔴 Blocker — Line 37 (and echo at line 154)**
> « `consumed_at` nullable sur l'enregistrement de session + hash du refresh token (jamais le token en clair) + TTL de grâce 5-10 min »

The storage mechanism of the decision is spec'd to the point of being contradictory and unimplementable by a fresh engineer. Sessions today live in **Redis** (`user_sessions:*`, managed by the Lua script, plan Phase 3 explicitly lists « repository.go (sessions Redis ...) »). Redis has no « nullable column » concept — `consumed_at` on an « enregistrement de session » implies a durable record, i.e. a schema change, yet **no goose migration is planned anywhere** (plan-phases lists zero schema changes; the design forbids nothing here but plans nothing either). If the session record moves to PostgreSQL, which pool touches it (RLS is tenant-scoped, sessions are platform-level), when does it migrate, and what happens to the Lua atomicity? If it stays in Redis, « consumed_at nullable » is meaningless vocabulary. The design also leaves the grace TTL as a range (« 5-10 min ») rather than a decision. This is the only behavioral change in the entire migration, and its core mechanism cannot be built from the document as written.

**2. 🔴 Blocker — Lines 37, 154**
> « rejeu pendant la grâce → `401 invalid_grant` (ou re-émission si la politique de sécurité le retient — choix figé en exécution, sous tests dédiés) »

The standard requires security implications be identified and mitigated; this document defers its single security-sensitive decision to execution with no threat analysis. The two options are not equivalent security policies: « re-émission » means a stolen refresh token replayed inside the grace window yields fresh tokens even after the legitimate owner rotated — it converts the soft-consume mechanism's replay signal into noise, since attacker replay and honest client retry become indistinguishable. « 401 » preserves the signal but hard-breaks network-retry recovery. A design doc that ships the mechanism (grace window) before the policy that makes the mechanism safe is designing backwards. This choice must be decided in the design, with the stolen-token-in-grace-window failure mode written down, or the carve-out must be cut from the migration.

**3. 🟡 Major — Line 18 vs lines 37, 154**
> « Contrat HTTP inchangé au byte près pendant toute la migration »

The stated objective is contradicted by decision 7's carve-out (« seul le chemin “rejeu pendant grâce” peut différer du comportement actuel »). The carve-out is acknowledged in plan-phases, but the design's own Objectives section still asserts byte-exactness without qualification. A naive reader joins the project, sees line 18, assumes no behavioral change during migration, and does not audit Phase 3. Amend the objective (« inchangé, sauf carve-out acté décision 7 ») or the guarantee is a lie by document structure.

**4. 🟡 Major — Line 37**
> « job de cleanup des enregistrements périmés »

A new operational component with no design: no owner (in-process ticker? cron? CLI?), no schedule, no failure mode. If the cleanup job dies, the session records (wherever they live per finding 1) grow unboundedly; if it races the grace window, it can delete a record a replay is entitled to consult. The standard's « operational complexities ... and how they will be mitigated » is unmet for this component. Specify owner, cadence, and the cleanup↔grace race invariant.

**5. 🟡 Major — Line 123**
> « `make e2e` est rejoué aux Phases 1 et 7 précisément parce que ce sont les deux seules phases qui changent l'entrypoint (`main.go` → `cmd/api`) et la structure de `router.go` »

Unbacked assertion. Phase 3 rewires every protected route constructor (`TokenAuth(svc)`, `AdminAuth(svc)`) and deletes `middleware/auth.go` — that is a structural change to `internal/app` routing, not a « branche de câblage ». Phase 2 introduces the first module into `internal/app`. The justification for not running e2e in the phases that rewrite the middleware chain rests on a claim the plan itself contradicts. Either widen the e2e gate to Phases 2-3-7, or justify why the suite `tests/` (which exercises `app.NewRouter` end to end) is sufficient for exactly that class of change — with that argument, not by assertion.

**6. 🟢 Nit — Lines 4, 3-4**
> « Les ancrages de lignes des rapports antérieurs ... suivent la numérotation de l'original »

Correctly self-aware, but the document deliberately renumbers against prior review reports' anchors. Acceptable given the cross-reference is explicit at the top; noted only because stale anchors in the historical reports can no longer be reconciled against this file.

**7. 🟢 Nit — Lines 113-124 vs the standard's "Risks & Mitigations"**
The design carries its risk content inline (§3 « points de friction », invariants §1, carve-outs) and delegates the consolidated « Risques transverses et rollback » to plan-phases. The pair satisfies the standard; the design standalone does not, and finding 2 shows the delegation is not always safe — the one risk-bearing change invented *in this document* (decision 7) is precisely the one whose risk analysis is missing. If the design introduces a carve-out, the design carries its risks.

#### What passes, in aggregate

Alternatives evaluation (lines 39-44): four real alternatives, each rejected for concrete trade-offs including costs of the *chosen* path — this satisfies the Alternatives Rule cleanly, no maintainer attribution without links. Problem statement line 12 quantifies the cost of inaction. Non-goals (lines 22-27) are explicit and repeated consistently with the open-questions section. Dependency rules (line 149) and the transient-window decision A2 come with both rationale and a measurable exit criterion (`git ls-files` = 0). Mapping tables (§1, §2) use verifiable file/line anchors. Invariants (lines 75-80) are preserved verbatim against current behavior. The security posture of the *current* system (constant-cost bcrypt, RLS fail-closed, admin/tenant pool split) is documented and protected, not renegotiated by the migration.

**Bottom line:** decisions 1-6 are executable as written. Decision 7 is not: it blocks the only behavioral change through the entire spec with a contradictory storage story and a deferred security policy. Resolve findings 1-2 in the design before any Phase 3 work; the rest are correctable in text.

### egm-structural (standard : feature-spec-standard)

#### Mapping structurel (extrait)

```
H1 (l.1)
  └─ H2 Préambule (l.6)           [non numéroté]
       ├─ H3 Problème (l.10) / Objectifs (l.14) / Non-objectifs (l.20)
       ├─ H3 Décisions structurantes actées (l.29)   [1-7]
       └─ H3 Alternatives considérées (rejetées) (l.39)  [1-4]
  └─ H2 1. Cartographie de l'existant (l.48) → H3 (l.50, l.73)
  └─ H2 2. Inventaire des domaines métier (l.82) → H3 (l.93)
  └─ H2 3. Couplages et points de friction (l.113)
  └─ H2 4. Layout cible (l.126)
  └─ H2 Questions ouvertes (l.151)  [non numéroté]
```

#### Constats

**🔴 Blocker — Section obligatoire « Risks & Mitigations » absente du document (aucun en-tête de risque sur l'ensemble de la plage)**
- Périmètre : trou structurel entre l.149 (fin de §4) et l.151 (`## Questions ouvertes`) ; fragments de contenu disséminés aux l.37, l.121, l.123, l.124.
- La norme « Risks & Limits » : « The specification must explicitly call out its own constraints. It must identify potential failure modes, security implications, scalability limits, and how they will be mitigated », et « Risks & Mitigations » figure dans la liste de sortie requise. Aucune section de risque à l'en-tête : la dérive CI (l.121) est enfouie dans « Couplages et points de friction », le carve-out de rejeu (l.37) au dixième sous-point d'une décision, les réserves SPA (a)(b)(c) (l.123) dans le point 9 du §3, et les risques d'exécution externalisés dans « Risques transverses et rollback » du frère (l.65). Aucune analyse de scalability nulle part (les deux pools, le TTL des sessions Redis, la croissance multi-tenant). Un lecteur qui vérifie ce document par en-têtes échoue sur la rubrique risques.

**🟠 Majeur — Le titre « ## Questions ouvertes » contredit son contenu : 5 des 6 puces déjà tranchées ou différées**
- Périmètre : l.151–158. Un lecteur de l'en-tête conclut à six décisions en suspens ; en fait l.153 (jamais pendant la migration = tranché), l.154 (tranché 2026-11-25, reliquat sous-choix), l.155 (tranché post-migration), l.156 (tranché conditionnel), l.157 (différé backlog), l.158 (tranché hors périmètre). Une seule puce (l.154) porte un vraiment ouvert.

**🟠 Majeur — Prose cassée à l'ancre A2 (l.149)**
- Citation exacte : « …dans un seul mouvement. Impact architectural ici) : les imports racine↔modules sont tolérés… ». Fragment télégraphique « Impact architectural ici) : ». L'ancre A2 a été vérifiée sur disque : `goldfish_review_report.md` l.5, l.31-38 contient bien A2/Option 1 « Tolérance transitoire explicite » — la cible résout mais le fragment de prose sous-mine l'ancrage.

**🟠 Majeur — Comportement critique sécurité reporté d'un commun accord à l'exécution (l.37+l.154)**
- « choix figé en exécution » : la norme exige les risques majeurs traités avant implémentation ; un choix de politique de sécurité reporté rend le document non ratifiable sur ce point (les garanties connexes tenant : détenteur, tests dédiés, trace Phase 3).

**🟠 Majeur — Provenance des décisions « (validation utilisateur) » sans ancrage vérifiable (l.29, l.31-l.37)**
- Sept décisions reposent sur une validation utilisateur revendiquée, sans lien vers artefact, extrait ou registre de revision ; dates sans référence d'artefact. Discipline de traçabilité problème↔décisions à appliquer ici aussi.

**🟡 Mineur — Numérotation de rubriques asymétrique (l.6 et l.151 non numérotées vs 1-4)**

**🟡 Mineur — Style de renvoi croisé incohérent (lien markdown l.3 vs chemins backticks l.37, l.66, l.121, l.123)**

**🟡 Mineur — Lignes 117 et 260 non réclamées par le mapping de scission (l.3)**

**🟡 Mineur — Taxonomie §3 englobe des items hors couplage (7, 9, 10)**

**🟡 Mineur — Anatomie module du préambule (l.8) diffère du layout (l.140) : `model.go` manquant**

#### Ce qui passe (global)

Pas de saut d'en-tête ; hierarchie correcte. Alternatives Rule satisfaite (4 vraies alternatives, coûts des choix explicites, attribution sans lien absente). Vérifications de renvois : TOUTES réussies sur disque (plan frère, Risques transverses l.65, P1.7 l.140, goldfish_review_report.md l.5/l.35 A2 confirmé, design-review/, remove-public-access.md, memory-bank/progress.md:56). Traçabilité interne décisionnelle cohérente (décisions 4/6/7 ↔ §4). Tableaux bien formés. **Bilan : 1 Bloqueur (sic 1), 4 Majeurs, 5 Mineurs.**

### egm-style (persona seul)

#### Majeurs M1-M6

- **M1 (l.4)** : métadonnées EGM redondantes — stages listés puis répétés « Step `feature_spec` (Stages 5-8) ». Garder la liste, couper la répétition.
- **M2 (l.33 vs l.44)** : « platform » (angl., masc.) vs « plateforme » (fr., fém.). Canon : « la plateforme » en prose, `platform` réservé aux références de package.
- **M3 (l.151-158)** : « Questions ouvertes » contient une entrée close (l.154) + duplications verbatim des Non-objectifs (l.156=l.24 ; l.158=l.23). Garder l.154 comme point d'inscription « résolu » avec son seul reliquat ; couper les répétitions.
- **M4 (l.124)** : point 10 énonce la même analyse 2× (intro 3 zones + re-démonstration). ~1/3 réductible ; « à statuts différents » = jargon.
- **M5 (l.123)** : paragraphe 10 lignes, synonymes de préservation en rotation (tel quel/verbatim/copie/à l'identique/telle quelles) — réserves (a)(b)(c) noyées dans la prose.
- **M6** : 5 libellés pour byte-exactness (l.18 « au byte près », l.37/124 « byte-exact », l.78 « **au byte près** (les tests les épinglent) », l.123 « copie verbatim »...). Canon recommandé : « inchangé au byte près » (contrats) + « recopié à l'identique » (action).

#### Mineurs N1-N15

- N1 l.8 « transvers » → « transversal » ; N2 l.36 « ses repository » → « ses repositories » ; N3 l.115 « que cartographiés » → « identiques à ceux cartographiés » ; N4 wiring/câblage mélangés (l.12, 42, 115-124) ; N5 l.41 « contrat HTTP doit rester vert » (vert = CI, pas contrat) ; N6 l.123 « précisément » ; N7 l.149 fragment A2 + virgule de fusion + « le 2026-11-25 » (article superflu) ; N8 l.37 « le retient » ambigu ; N9 l.124 « cf. fin de point » auto-référence ; N10 l.123/l.156 « sans effet » vs « inopérante » ; N11 l.12/l.67/l.120 même constat ×3 (connexions « et » vs « + ») ; N12 l.12/l.111 « multiplie » / « n'auront qu'à » non calibrés ; N13 l.119 « en central » ; N14 l.62 « Interdit à la main » → « édition manuelle interdite » ; N15 divers.

#### Ce qui passe (agrégat)

Aucune hyperbole marketing ; jugements portés par des tests nommés ; mises en garde avec décisions nommées (KEEP) ; tableaux denses en faits ; terminologie domaine (soft-consume strict, Unit of Work, interfaces côté consommateur, fenêtre transitoire, A2) cohérente de bout en bout. **Résumé : 0 Blocker, 6 Majors — racines communes = répétition/ligne 4, 12, 123-124, 151-158 — et 15 nits.**
