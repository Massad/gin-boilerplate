# Stage 6 — Comprehension Test

Périmètres audités : FeatureRange 1..174, DetailedRange 175..fin.
Contexte : aucun (aucun GoldfishContext, lecture du document seul — test d'auto-containment).
Reviewer : egm-comprehension (Naive Reader).

---

## What this proposes

Le document propose une migration incrémentale du code d'une API Go vers une architecture modulaire par domaine : un entrypoint `cmd/api/` avec injection de dépendances explicite, un socle technique transversal `internal/platform/`, et des modules métier `internal/modules/<domaine>/` (handler, service, repository, dto, tests). La migration se fait en huit phases (Phase 0 filet de sécurité, puis déplacements par domaine : platform, article, auth, tenant, user, admin, collage final auth, nettoyage), avec une règle stricte : le comportement HTTP ne doit pas changer d'un byte, chaque phase laisse build/vet/tests verts, et les Phases 0-1 sont détaillées pas à pas (commits, fichiers, lignes à modifier). Des questions ouvertes (montées de versions de libs, CORS, refresh httpOnly, RBAC) sont explicitement exclues du périmètre.

## How the existing system works, per this document

Le système actuel est un API Gin monolithique avec wiring global dans `main.go` (~230 lignes) : deux pools PostgreSQL (tenant `app_user` sous RLS, plateforme `app_admin` BYPASSRLS), Redis pour les sessions JWT, et du code structuré en paquets racine `controllers/`/`models/`/`forms/`/`middleware/` reliés par des variables de package globales sans DI. Le multi-tenant repose sur RLS échec fermé via `WithTenant` (`set_config(...,true)` en transaction), le login est unifié (tenant puis repli superadmin, à coût bcrypt constant), et l'isolation est prouvée par des tests croisés et une suite HTTP d'intégration qui duplique malheureusement les routes (`SetupRouter`) — duplication que le plan résout via `internal/app.NewRouter`. Les routes, tables et tests sont décrits par domaine (article, auth, user, tenant, admin, invoice), et une dérive CI préexistante (smoke sondant des routes supprimées) est identifiée à corriger en Phase 0.

---

## Findings

**1. Severity : Major**
- **Affected Scope** : §3 point 10 / Phase 3 (lignes 23, 91, 136) — *« script Lua de refresh »*.
- **Discrepancy & Rationale** : le document mentionne à trois reprises un script Lua gérant le refresh des sessions (`models/auth.go:262`), sa sémantique d'« ownership », et son déplacement en Phase 3 comme risque, mais n'explique nulle part ce que fait ce script ni ce que recouvre « ownership » ici — la seule définition offerte est le nom d'un test (`TestRefreshOnlyConsumesMatchingOwner`). Un lecteur naïf ne peut pas évaluer le risque annoncé ni vérifier qu'une re-implémentation ou un déplacement le préserve : c'est un prérequis de compréhension manquant pour la phase la plus critique du point de vue des sessions.

**2. Severity : Major**
- **Affected Scope** : Phase 7, Risques (ligne 164) — *« `kind` de réponse »*.
- **Discrepancy & Rationale** : `kind` est décrit comme un élément du contrat de réponse du login unifié à préserver, mais n'est défini nulle part dans le document (ni ce qu'il discrimine, ni ses valeurs possibles). Le lecteur ne peut pas comprendre quel contrat est engagé ni ce que « conserver le `kind` » signifie concrètement, alors que c'est une propriété projet spécifique et testée.

**3. Severity : Nit**
- **Affected Scope** : Phase 8, Note factuelle (ligne 170).
- **Discrepancy & Rationale** : *« le prompt mentionne `.kilocode/rules/` »* fait référence à un prompt externe non identifiable depuis le document seul ; un lecteur frais ne sait pas de quel prompt il s'agit ni pourquoi cette note existe. La conclusion (les règles vivent dans `AGENTS.md`) est claire, mais la justification reste un résidu de conversation invisible.

**4. Severity : Nit**
- **Affected Scope** : Phase 7, Risques (ligne 164) — *« jamais deux répertoires sondés »*.
- **Discrepancy & Rationale** : phrase ambiguë (« répertoires » pour désigner vraisemblablement les deux pistes d'authentification user puis admin du login unifié). Le concept est inférable du reste du document, mais le terme « répertoires » est trompeur.

**5. Severity : Nit**
- **Affected Scope** : §7 Critères de sortie (ligne 258) — *« Compteur de fichiers legacy décroissant »*.
- **Discrepancy & Rationale** : l'intitulé promet une métrique (un compteur décroissant), mais le critère qui suit ne définit aucune métrique ni aucun comptage : c'est uniquement une règle de direction d'imports (pas d'imports inverses). Le lecteur ne peut pas savoir comment mesurer la décroissance annoncée.

**6. Severity : Nit**
- **Affected Scope** : Phase 1 (lignes 126, 199) vs P1.8 (ligne 237).
- **Discrepancy & Rationale** : incohérence interne de comptage — la Phase 1 annonce *« ≈ 8 commits »* pour 8 étapes P1.1–P1.8, mais P1.8 déclare *« même commit que P1.7 »*, ce qui donne 7 commits. De même, *« 8 déplacements + 2 créations »* (ligne 126) n'est pas re-dérivable des étapes listées sans re-comptage arbitraire des fichiers.
