# Stage 8 — Intermediate Zoom Out

Périmètre : fichier complet (ré-audit haute altitude après Stages 5-7).
Contexte : zoom_out_report.md + comprehension_report.md + critic_report.md fournis au reviewer ; GoldfishContext : aucun.
Reviewer : egm-zoom-out (standard zoom-out-standard).

---

# Rapport Zoom-Out Intermédiaire — « Migration incrémentale vers une architecture modulaire par domaine »

Méthode : re-audit du document complet (v. 268 lignes) à haute altitude, après les Stages 5-7. Vérifications code refaites : `internal/platform/`, `internal/app`, `internal/modules/` n'existent pas encore ; aucun helper central de réponse (`MapPgxError`, `ParseUUIDParam`, `AbortMessage`) n'existe dans le dépôt — `httpx` reste une création nouvelle, pas une duplication. Aucun code de la migration n'a été appliqué entre le Stage 5 et maintenant.

## (a) Statut des constats du Stage 5

Les six constats du Stage 5 sont tous **toujours valables, aucun n'a été traité** dans le document actuel :

1. **[Major] YAGNI `httpx` créé sans appelant** (P1.6, l.227-230 : « Aucun appelant prod dans cette phase ») — toujours écrit tel quel. Recommandation inchangée : créer le paquet en Phase 2, au moment de la première adoption réelle.
2. **[Major] `httpx/identity.go` dans platform lit le contrat de contexte du module auth** (l.228) — toujours présent. Le placement sémantiquement correct reste le module `auth` (propriétaire des clés posées par `TokenAuth`).
3. **[Nit] Cinq fichiers par module sans critère de fusion de couches** (§4, l.107-112) — inchangé.
4. **[Nit] Extension de tests mêlée au refactoring à comportement inchangé** (§7, l.255) — inchangé.
5. **[Nit] Numéros de ligne d'imports dans P1.2/P1.3** (l.207, 213) — inchangés, fragiles dès la première étape.
6. **[Nit] Motivation (problème racine) absente de l'en-tête** (l.3 vs §3) — l'ouverture reste la cible, pas le pourquoi.

## (b) Constats nouveaux révélés par les Stages 6-7

**7. [Major] Transition Phase 3→7 non spécifiée : contradiction entre la règle de dépendance et le mécanisme d'injection**
- Scope : Phase 3 (l.137-139 : consommateurs « à l'ancienne monde » reçoivent `*auth.Service` via internal/app) vs règle l.116 (« platform n'importe jamais modules ») et l.258 (« aucun paquet racine n'importe internal/modules en sens inverse »).
- Révélé par le Stage 7. À haute altitude, c'est le seul endroit du plan où la règle architecturale affichée devient inapplicable en l'état : entre les Phases 3 et 7, `models/login.go`, `models/user.go`, `controllers/*` (paquets racine) doivent obtenir le service du module `auth` — soit en violant la règle d'import racine→module, soit en changeant des signatures, et le document ne choisit pas. Le mécanisme de collage transitoire (tolérance temporaire ? signature `UnifiedLogin(form, ts)` ?) doit être tranché dans le plan, sinon chaque exécuteur improvisera différemment et la règle l.258 perdra sa valeur de garde. Alternative la plus simple : autoriser explicitement et temporairement l'import racine→module jusqu'à la Phase 7, avec suppression vérifiée en sortie de Phase 7.

**8. [Major] Ports du module `auth` sous-spécifiés et séquencés à contre-temps**
- Scope : Phase 7 (l.161-162), Phases 5-6 (l.150, 156 : « implémente UserAuthenticator / AdminAuthenticator »).
- Révélé par le Stage 7. Deux défauts d'altitude : (a) le port `UserAuthenticator.Login(form)` ne couvre pas la relecture user exigée par Refresh (rôle/status re-fraîchi de la base) — aucun port ni signature ne le définissent, alors que c'est une interface transversale du module le plus couplé ; (b) les Phases 5-6 déclarent implémenter des ports définis à la Phase 7, deux phases plus tard : la déclaration d'implémentation ne peut être câblée qu'au moment où le port existe. Spécifier les deux ports (authentification + lecture user pour refresh) une fois, à la Phase 3 (naissance du module auth), et décaler la mention d'implémentation aux Phases 5-6 comme « implémentent le port défini en Phase 3 » supprime l'ambiguïté sans coût.

**9. [Nit] Cadence e2e justifiée par un claim faux (calibration, pas architecture)**
- Scope : §3.9 (l.90 : « les deux seules phases qui touchent au câblage servi à la SPA »).
- Confirmé par le Stage 7 (F1). Les Phases 2-6 modifient aussi `internal/app`. En pratique le risque reste couvert par le smoke CI rejoué à chaque branche, d'où Nit ici — mais la justification écrite est inexacte et devrait être reformulée (entrypoint + structure de router), car un document de référence ne doit pas fonder une décision de gate sur une prémisse fausse.

**10. [Nit] Incohérences de comptage et définitions manquantes locales**
- Scope : Phase 1 (« ≈ 8 commits », l.199 vs P1.8 « même commit que P1.7 » → 7) ; `kind` (l.164, jamais défini) ; sémantique du script Lua (l.91, 136, l'ownership renvoyé seulement à un nom de test).
- Révélés par le Stage 6. Détails de niveau bas, pas de défaut d'altitude ; à corriger mécaniquement.

Aucun autre problème d'altitude nouveau détecté : le séquencement des phases reste sain (article pilote, auth primitives puis collage final), le layout cible est stable, les invariants de sécurité sont inchangés, et le code cible vérifié n'introduit aucune réinvention de roue.

## (c) Compromis déguisés

**Aucun compromis déguisé accepté.** Les arbitrages reportés dans « Questions ouvertes » (l.261-268 : RBAC non câblé, CORS configurable conditionnel, refresh httpOnly en backlog, montées de versions post-migration, homogénéisation de casse post-migration) sont tous des reports conditionnels et explicites avec leur déclencheur — pas des impasses. La conservation délibérée de la casse mixte des clés JSON (invariant l.45) est un choix de contrat épinglé par les tests, correctement argumenté. Le point 9 ci-dessus est la seule calibration à relier à une décision (la cadence e2e), et il est corrigible en une phrase.

## Bilan

Le document tient toujours le standard Zoom-Out à l'échelle du tout. Les deux Majeurs du Stage 5 (httpx créé sans appelant, identity.go mal placé) restent ouverts et non traités. Deux nouveaux Majeurs issus des stages d'exploration (mécanisme de transition Phase 3→7, ports auth sous-spécifiés et mal ordonnancés) concernent le même volet : la zone d'interface du module `auth` entre les Phases 3 et 7, précisément la partie du plan la plus critique, reste la moins spécifiée. Les tenir dans une seule révision — httpx différé + ports de la Phase 3 + transition racine→module explicite — suffirait à lever les quatre.
