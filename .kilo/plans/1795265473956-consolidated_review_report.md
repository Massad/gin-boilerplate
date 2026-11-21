# Consolidated Review Report — Pipeline EGM complet

**Target:** `.kilo/plans/1795265473956-frontend-stack-design-draft.md` (318 lignes)
**Mode:** Consolidated report (Stages 5→6→7→8→9 sans checkpoints)
**Date:** 2026-11-21
**Reviewer Status global:** egm-zoom-out ✔, egm-inventive-catalyst ✔, egm-comprehension ✔, egm-design ✔, egm-structural ✔, egm-style (×2) ✔, egm-implementation-readiness ✔, egm-codebase-details ✔ — aucun reviewer « not run », aucune reprise nécessaire (1 reprise Stage 6 pour complétude, résultat conforme à la persona).

## Verdict

**Le document n'est pas prêt pour l'implémentation en l'état.** Deux constats bloquants, vérifiés par trois reviewers indépendants et par une vérification grep directe du dépôt :

1. **La couche backend qu'il planifie est déjà livrée.** Cookies httpOnly (`SetRefreshCookie`/`SetAdminCookie`/`Clear…`, `models/auth.go:19-74,365-394`), refresh lu dans le cookie (`controllers/auth.go:31-35,102-105`), login/logout tenant et admin (`controllers/user.go:68,135`, `controllers/admin.go:55,142`), servage SPA + `/assets` (`main.go:76-214`), tests d'assertions cookies déjà écrits (« tests/article_test.go:144,364,789,834 », `controllers/auth_test.go:96-181`, `middleware/auth_test.go:214,243`). Les items backend 1-11 du plan décrivent du travail terminé.
2. **Une cible d'édition invalide invalide le plan** : item 6 supprime `./forms/auth.go` qui n'existe pas ; le struct `Token` visé est dans `models/auth.go:114-118` et est encore utilisé par `UserModel.Login` (`models/user.go:67`).

Le travail réellement restant : **frontend `/app` (items 12-32) et intégration monorepo (items 33-38)**. Le document doit être recalé sur l'existant (constat d'existant au lieu de « modifications à faire »), corriger ses contrats (le vrai contrat articles renvoie `id` string UUID et les clés `results`/`data` — pas `id int64` / `articles`), puis retravailler ses choix discutables (CSRF, compromis httpOnly, Sentry v1) et sa langue (titres anglais, franglais, allégations non calibrées).

## Tâches consolidées par priorité

### 🔴 Blocker

* **B1 — Recaler le document sur le backend déjà livré** (items 1-11 obsolètes, « Modifications backend v1 » → « constat d'existant ; vérification uniquement » ; ordre d'exécution L236 et Verification Plan backend L304-305 à réécrire ; corriger au passage constantes `RefreshCookieName`/`AdminCookieName`, condition `Secure` = `ENV=PRODUCTION` **ou** `SSL=TRUE`, références `main.go` 76-214). [S5 Zoom-out, S8, S9 Readiness+Codebase, grep direct]
* **B2 — Corriger/supprimer l'item 6** (`forms/auth.go` inexistant ; `models/auth.go:114-118` encore utilisé par `UserModel.Login`). [S5, S9]

### 🟡 Major

* **M1 — Assumer le changement de contrat auth** : qualifier `login`/`refresh`/`admin login`/`logout` comme endpoints modifiés en v1 (L13 « sans modification du backend », L102 « Aucune modification de l'API /v1 » sont faux), régénérer swagger. [S5 Inventive]
* **M2 — Figer le contrat API avant `types.ts`** : formes réelles (id **string** UUID, clés `results`/`data`/`message`), types depuis le contrat OpenAPI et non « d'après les contrôleurs ». [S5, S9 Codebase+Style]
* **M3 — CSRF** : documenter les modes de défaillance de `SameSite=Strict`+`Path` (dégradation Lax, requêtes émises par XSS avec cookie attaché) ; évaluer jeton double-submit ou contrôle `Origin` sur les mutateurs. [S5 Inventive]
* **M4 — Réévaluer le compromis httpOnly** : chiffrer la complexité client que l'hybride impose (boucle refresh, single-flight, purge) face au 100 % httpOnly + `/me`. [S5 Inventive]
* **M5 — Production sans `app/dist`** : échec fermé au lieu du fallback silencieux vers le template démo. [S5 Inventive]
* **M6 — Sentry en v1 (YAGNI)** : optionnaiser (env) ou différer en v1.5 ; garder Error Boundaries + logs. [S5]
* **M7 — Isolation des e2e** : tenant dédié/par-run + nettoyage, sinon la base se pollue entre runs. [S5 Inventive]
* **M8 — Remonter l'altitude du plan d'implémentation** : architecture/contrats/flux au lieu des champs et chemins ; listes fichiers/champs dans une task list séparée. [S5]
* **M9 — Risques du servage statique + alternative absente** : analyser charge/point de défaillance unique, ajouter et justifier l'alternative NGINX/CDN. [S7 Design]
* **M10 — Langue et libellés des sections** : titres en français, « Reason for Rejection » → « Motif du rejet », « Open Questions » ne liste que des décisions : renommer ou poser de vraies questions. [S7 Style]
* **M11 — Jargon et franglais** : définir à la première occurrence (single-origin, chunk lazy, single-flight, claims, queryKeys, proxy serve-side, toaster, RHF…) ou traduire. [S7 Style, S9 Style]
* **M12 — Décision httpOnly répétée 4 fois** : énoncer une fois (L47 ou L228), renvoyer ailleurs. [S7 Style]
* **M13 — Allégations non calibrées** : « la plus grande fenêtre XSS restante » (L206 — Bloqueur en Stage 7), « garantit le smoke test » (L297), qualificatifs marketing (L13,126,170,180,216) → faits mesurables. [S7 Style, S9 Style]
* **M14 — Documentation hors alignement avec le livré** : section Authentication du `README.md` encore en mode tokens JSON, `memory-bank/projectbrief.md:17` + `activeContext.md:19` (repo séparé, cookie « à faire »), `AGENTS.md` sans commandes front ; qualifier l'état de `docs/` (item 11). [S9 Codebase]
* **M15 — Verification Plan imprecis** : étape init. front absente (Node/pnpm non fixés, L305-306), étape sécurité « cookie visible httpOnly » contradictoire (L307) → critères exécutables (Application > Cookies : attributs `httpOnly`/`Secure`/`Path` ; aucune clé `token*` dans localStorage). [S9 Readiness+Style]

### 🟢 Nit

* **N1** — Claims JWT décodés utilisés pour le contexte UI : n'en tirer l'affichage qu'après un refresh validé. [S5]
* **N2** — Versions front récentes (React 19, Vite, RR v7, Tailwind v4) : justifier ou verrouiller. [S5]
* **N3** — Références de ligne backend approximatives (L53, L64, L133) : corriger ou passer aux références symboliques. [S5]
* **N4** — Alternative « Sentry auto-hébergé » annoncée (L106) mais absente des Alternatives 1-15. [S5]
* **N5** — Risque de l'access admin 8 h : quantifier/documenter l'acceptation. [S7 Design]
* **N6** — Résilience de la rotation Redis : décrire le comportement en cas d'échec (dépendance front→401→login). [S7 Design]
* **N7** — Intitulé « Non-Functional Requirements » couvre des risques : renommer/aligner au standard. [S7 Structural]
* **N8** — Em dash / encodage (mojibake possible). [S7 Structural]
* **N9** — Gras décoratif (L19-125) : réserver aux alertes critiques. [S7 Style]
* **N10** — Remplissage et métaphores : « de façon invisible », « session Redis est morte », « ce harnes » (L7), « verts »/« de bout en bout » (L304-306), phrase tautologique L236, champ « Type/Scope » L244, « non-invasive » + PII (`tenant_id` dans Sentry, L287), 406 non documenté (L288). [S7+S9 Style, S9 Readiness]

## Matrice de traçabilité (constats bruts → tâches)

| Rapport stage | Constats bruts | Tâches consolidées |
| --- | --- | --- |
| `zoom_out_report.md` (S5, 15 constats) | Z1 blocker déjà implémenté ; Z2 contradiction API ; Z3 forms/auth.go+Token ; Z4 altitude ; Z5 CSRF ; Z6 zod mirror ; Z7 types/contrat ; Z8 fallback démo ; Z9 httpOnly ; Z10 Sentry ; Z11 e2e isolation ; Z12 claims ; Z13 versions ; Z14 lignes ; Z15 sentry-alt | B1, M1, B2, M8, M3, M11(zod: voir M2/gate CI), M2, M5, M4, M6, M7, N1, N2, N3, N4 |
| `comprehension_report.md` (S6) | 0 constat (paraphrase correcte, conservée) | — |
| `critic_report.md` (S7, 14 constats) | C1 XSS superlatif (Bloqueur S7) ; C2 infra statique ; C3 NGINX/CDN ; C4 titres EN/Open Questions ; C5 jargon ; C6 httpOnly ×4 ; C7 adjectifs ; C8 admin 8h ; C9 Redis ; C10 NFR naming ; C11 em dash ; C12 gras ; C13 remplissage ; C14 redondances | M13, M9, M9, M10, M11, M12, M13, N5, N6, N7, N8, N9, N10, N10 |
| `intermediate_zoom_out_report.md` (S8) | 0 nouveau ; re-vérification du blocker backend | B1 (confirmation) |
| `implementation_readiness_report.md` (S9, ~20 constats) | R1 items 1-11 obsolètes ; R2 item 6 ; R3 contrat articles ; R4 vérif front inapplicable ; R5 README Authentication ; R6 memory-bank/AGENTS/docs ; R7 étape sécurité ambiguë ; R8 « garantit » ; nits (statuts par item, docs sync, single-flight, franglais, PII, 406, métaphores CI, metadata) | B1, B2, M2, M15, M14, M14, M15, M13, N10 |

N.B. — Le drift zod↔`binding` (Z6) est conservé via M2/M11 (gate CI de dérive à évaluer dans la même passe contrat).

## Index des rapports de stages

* `.kilo/plans/1795265473956-zoom_out_report.md` (Stage 5, 15 constats)
* `.kilo/plans/1795265473956-comprehension_report.md` (Stage 6, 0 constat)
* `.kilo/plans/1795265473956-critic_report.md` (Stage 7, 14 constats)
* `.kilo/plans/1795265473956-intermediate_zoom_out_report.md` (Stage 8, 0 nouveau)
* `.kilo/plans/1795265473956-implementation_readiness_report.md` (Stage 9, 20 constats)
