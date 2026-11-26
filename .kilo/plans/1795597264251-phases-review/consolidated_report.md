# Rapport consolidé — G-1 (Stage 9, egm-review Step `implementation`)

**Cible** : `1795597264251-architecture-modulaire-par-domaine-plan-phases.md` (DetailedRange = l.6-168) · **GoldfishContext** : design `…-architecture-modulaire-par-domaine-design.md` · **Date** : 2026-11-26 · **Rapports sources** : `style_findings.md` (déjà livré), `implementation_readiness_findings.md`, `codebase_details_findings.md` — consolidation de l'orchestrateur (synthesizer), déduplication cross-personas ci-dessous.

## Verdict d'ensemble

Le pas-à-pas P0.x-P1.x est vérifié **exact** (sweep d'ancrage complet : 0 chemin faux, 0 ligne erronée ; 1 seul état périmé : `docs/test.md`). La conformité soft-consume l.26 ↔ design décision 7 est **confirmée** sur le contenu. Trois exigences **zéro-décision** portées par le contrat (lots plan l.21, l.29-36) ne sont pas encore dans le plan-phases : la spécification `pgdb.InTx`/`pgdb.WithinTx` (Blocker pour L4/L5), l'acquisition du client Redis (texte C8 prêt hors plan), et le placement du DTO `TenantPayload` (risque de FAIL gate 2). Le reste est éditable borné.

**Décompte consolidé dédupliqué : Blocker 1 · Major 12 · Nit 12** — 25 items distincts (14 doublons cross-personas fusionnés).

## Verdicts des 7 points d'audit obligatoires (lots plan l.30-36)

| # | Point | Verdict | Preuve/finding |
|---|---|---|---|
| 1 | Spec zéro-décision `pgdb.InTx`/`pgdb.WithinTx` | ❌ FINDING — **Blocker** | C1 (`pool.go:42-68` : seuls `Store.InTx`/`WithTenant`) ; unanimité readiness+F1(details) |
| 2 | Sweep d'ancrage complet | ✅ CONFIRMÉ | relevés des deux personas : 0 chemin faux, 0 ligne erronée, 1 état périmé (`docs/test.md`) + 1 incohérence sémantique (admin Login, C4) + nits |
| 3 | Fraîcheur Phase 8 (`docs/test.md` vs `581d770`) | ❌ FINDING — Major | C6 : commit `581d770` vérifié, fichier absent à HEAD ; artefacts `gin-boilerplate`/`e2edb.exe`/`tmp_line.txt` **présents** (ls) |
| 4 | Cohérence puce P3 soft-consume ↔ design décision 7 | ✅ CONFIRMÉ (contenu) | formats, grâce 5 min, rejeu, carve-out, sweep, tests alignés (baseline `models/auth.go:262-277` vérifiée) ; calibrages de forme : C13 (style F2 + readiness F7) |
| 5 | Opérationnalisation « aucune phase ne baisse la couverture » | ✅/⚠️ | l.164 actionnable (inventaires nommés) ; consignation qualitative de gate à épingler : C12 (readiness F5, Nit) ; inventaire + 0 seuil déjà dans le contrat (lots l.75) |
| 6 | Spec zéro-décision acquisition client Redis (Phase 3) | ❌ FINDING — Major | C2 (unanimité readiness+details) ; texte C8 prêt dans `1795684415717-audit-lots-impl-migration.md` §C8, faits amont re-vérifiés ; passe egm-resolve = la session où il s'applique |
| 7 | Propriétaire unique Redis | ✅ CONFIRMÉ | seul `models/auth.go` touche aux clés de session en prod (`main.go:144` = `InitRedis`, `cmd/e2edb` = `RedisDBIndex` + client propre) ; règle codifiée (gate item 8, lots l.73) ; plan cohérent dès C8 appliqué |

## Findings consolidés (dédupliqués)

### 🔴 Blocker

**C1 — Unit of Work `pgdb.InTx`/`pgdb.WithinTx` : nommés, jamais spécifiés** *(readiness F1 Blocker ≡ details F1 Major)*
- Périmètre : l.35 (Phase 4 Interfaces), l.41 (Phase 5), l.46 ; baseline `internal/db/pool.go:42-68`.
- Le contrat (lots l.21) exige une spécification zéro-décision : signature exacte de `InTx`, sémantique de `WithinTx` (relation avec `Store.InTx`, pas de `set_config` tenant), mécanisme de liaison requêtes sqlc↔tx sans exposer `*pgdb.Queries` à la frontière d'un port, appel `CreateUserOwner` sur le MÊME `InTx` que `TenantProvisioner.CreateInTx`. Plus cas `pool_test.go` exigés pour L4 (commit OK / rollback erreur / jamais `SET` session-wide).

### 🟡 Major (12)

**C2 — Acquisition du client Redis du repository sessions : non tranchée** *(readiness F2 ≡ details F5)* — appliquer le **texte C8** (audit `1795684415717` §C8, verbatim) à l.30 ; couvre aussi le point d'audit 7.

**C3 — `middleware/main_test.go` absent de la Phase 3** *(readiness F4)* — l.27/l.31/l.124/l.54 : le TestMain pin `REDIS_DB=2` doit migrer avec le paquet de test ; sans lui : « dossier supprimé si vide » faux + tests du module `auth` sans pin (violation index 1/2/3). Suppression du répertoire en Phase 3 plutôt qu'en Phase 7.

**C4 — Contradiction Phase 3/Phase 6 : l'admin Login** *(details F2)* — l.25 fait migrer `controllers/admin.go` **Logout/Login** en Phase 3 alors que design §2 (l.92) + l.46 assignent l'admin Login au module `admin` (Phase 6). Corriger l.25 : Logout seulement + renvoi explicite.

**C5 — DTO `TenantPayload` : placement cross-module non tranché** *(details F3)* — l.34 le liste dans `model.go` du module `tenant` alors qu'il est défini dans `models/user.go:27-31` (vérifié) ; l.41/l.53 le consument via ports. Sans tranchage : import cross-module type→FAIL gate 2 ou duplication non spécifiée. Trancher une option unique (recommandé : type défini/porté par chaque consommateur, aucune importation cross-module) + corriger l'ancre l.34.

**C6 — Phase 8 l.59 : `docs/test.md` périmé** *(readiness F3 ≡ details F4 ≡ style F11)* — suppression **commitée** (`581d770`, vérifié) ; artefacts racine toujours présents. L.59 : retirer l'item, garder les 3 artefacts. Même affirmation au design l.73 — hors DetailedRange : **noter pour le drift-report de clôture**, ne pas éditer le design certifié.

**C7 — Bandeaux d'introduction l.3-4** *(style F1)* — méta redondante avec l'en-tête design ; compression + RELOCATE de la note « numérotation des rapports antérieurs » vers le dossier de review.

**C8-lot — Mur de texte soft-consume l.26** *(style F2 + readiness F7 + 3 caveats CALIBRATE + style F14 sur cette ligne)* — réécrire en sous-puces (contenu inchangé), y épingler la forme 200 de la réponse de recovery (« même forme que le happy path — design décision 7 »), appliquer les 3 recalibrages (rollback borné → « re-login requis » ; observabilité → « journalisé… index `user_sessions:*` ; sweep » ; « purge défensive des reliques » → « suppression des records consommés expirés que le TTL n'a pas encore évacués »), règle d'emphase.

**C9 — Fenêtre A2 ×3 (l.28/68/167)** *(style F3)* — une seule occurrence complète (l.28), renvois aux autres.

**C10 — Carve-out redit l.67** *(style F4)* — compression + renvoi Phase 3.

**C11 — Duplication helpers + report `httpx` (l.13/18/21 ↔ l.134-136)** *(style F5)* — canon unique en P1.6 ; renvois ailleurs. (Intègre details F9 : `getRole` sans appelant — recopier seulement les accesseurs appelés.)

**C12-lot — Dette `binding.Validator` ×3 (l.60/131/140)** *(style F6)* — canon unique en P1.7(a) ; l.60 renvoi à P1.7(a). Édition l.60 : remplacer « traiter la suppression non commitée `docs/test.md` » est couvert par C6 ; cet item cible le renvoi de dette dans le même bloc Phase 8.

**C13-lot — Clarté du risque slug l.36** *(style F7)* — réécriture pendant/après migration + frontière explicite.

### 🟢 Nit (12)

- **C14** *(readiness F5)* — l.64 (fin) : remplacer le slogan par « consignation de l'inventaire dans `L<N>-gate.md` — exigence qualitative, aucun seuil chiffré ».
- **C15** *(readiness F6)* — l.19 : 5 signatures complètes `ArticleRepository` + forme handler explicite (`NewHandler(articles *Service) *Handler`).
- **C16** *(details F6)* — l.98 : le TestMain du paquet `tests` épingle les variables (sans connexion) — reformulation.
- **C17** *(details F7)* — l.21/l.136/l.46 : compléter le contrat de clés avec `adminID` (middleware/auth.go:86 → controllers/admin.go:87).
- **C18** *(details F8)* — l.13 : « 9 `git mv` de paquets + 2 déplacements d'entrées (P1.7) + 1 création » ; harmonisation optionnelle du lots plan l.47 (hors éditions resolve : à arbitrer).
- **C19** *(style F8)* — l.51 : « (le plus complexe, en dernier) » → « (dépend des phases 3-6) ».
- **C20** *(style F9)* — l.76 « le vrai contrôle » → « fait foi » ; l.140 « acceptable en transitionnel » → « toléré en transition » ; l.143 « épingle désormais le vrai câblage » → « épingle le câblage de production ».
- **C21** *(style F10)* — l.25 : « admin idem » → forme précise (à confronter au code réel avant application : user Logout `200 {"message":"Successfully logged out"}` vérifié `controllers/user.go:113-126` ; forme admin à vérifier).
- **C22** *(style F12)* — l.22 : supprimer la parenthèse paraphrase « (preuve de non-régression) ».
- **C23** *(style F13)* — l.76 : RELOCATE du disclaimer file:line pour couvrir les Phases 2-7.
- **C24** *(style F14)* — règle d'emphase transverse : gras réservé invariants/carve-outs ; parenthèses >1 clause en sous-puces. (Appliquée aux lignes éditées par cette passe ; pas de re-écriture globale.)
- **C25** *(fusion l.26, intégré à C8-lot)* — la forme 200 de la recovery + caveats CALIBRATE : traités DANS l'édition C8-lot (compte unique). (Item conservé pour traçabilité ; édit via C8-lot seulement.)

## Matrice de traçabilité (raw → consolidé)

| Source | Finding | Consolidé | Statut |
|---|---|---|---|
| readiness F1 (Blocker) | details F1 (Major) | C1 | doublon fusionné |
| readiness F2 | details F5 | C2 | doublon fusionné |
| readiness F3 | details F4, style F11 | C6 | triple fusionné |
| readiness F4 | — | C3 | unique |
| readiness F5 | — | C14 | unique |
| readiness F6 | — | C15 | unique |
| readiness F7 | style F2, style F14, caveats CALIBRATE l.26 | C8-lot | fusion bundle l.26 (C25) |
| details F2 | — | C4 | unique |
| details F3 | — | C5 | unique |
| details F6 | — | C16 | unique |
| details F7 | — | C17 | unique |
| details F8 | — | C18 | unique (lots l.47 optionnel) |
| details F9 | style F5 | C11 | fusion |
| style F1 | — | C7 | unique |
| style F3 | — | C9 | unique |
| style F4 | — | C10 | unique |
| style F6 | — | C12-lot | unique |
| style F7 | — | C13-lot | unique |
| style F8/F9/F10/F12/F13 | — | C19/C20/C21/C22/C23 | uniques |
| style caveats KEEP (5) | — | aucun édit | confirmés |

**Points d'audit sans édit requis** : 2 ✓ (sweep), 4 ✓ (contenu), 5 ✓ à C14 près, 7 ✓ à C2 près.

**Note RELOCATE (style F1, appliquée)** : la mention retirée de l'ancien bandeau l.4 du plan-phases reste valable — « Les ancrages de lignes des rapports antérieurs (`zoom_out_report.md`, `goldfish_review_report.md`, racine de `.kilo/plans/`) suivent la numérotation du plan ORIGINAL scindé (`1795597264251-architecture-modulaire-par-domaine.md`) ». Elle est conservée ici faute de place ailleurs dans le dossier de review.

## Résolution

Passe **egm-resolve** (contrat lots l.37 : les éditions portent sur le plan-phases) → rapport `goldfish_review_report.md` dans ce répertoire, classification par finding (PROPOSE FIX / CONTEST / ESCALATE / DECLINE), gates interactifs, puis revérification par relecture du plan-phases renuméroté. Texte prêt référencé : C8 (audit `1795684415717` §C8, verbatim, interdit d'application directe hors cette passe).
