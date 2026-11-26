# Goldfish Annotated Review Report

**Target:** `1795597264251-architecture-modulaire-par-domaine-plan-phases.md` (`full file` — DetailedRange l.6-168)
**Findings sources:** `consolidated_report.md` (C1-C25) ← `style_findings.md` + `implementation_readiness_findings.md` + `codebase_details_findings.md`
**Règle de portée** : leséditions portent sur le plan-phases uniquement (contrat lots l.37). Le texte prêt C8 (audit `1795684415717` §C8) s'applique DANS CETTE PASSE uniquement. Hors périmètre, notés : design l.73 (`docs/test.md` — pour egm-drift-report de clôture), lots l.47 (harmonisation « 9 git mv » — mini-édition directe à faire approuver séparément).

Status: resolved

## Prioritized Findings & AI Suggestions

### [C1] Unit of Work `pgdb.InTx`/`pgdb.WithinTx` jamais spécifiés [`readiness F1 Blocker` ≡ `details F1`] - Severity: Blocker
*   **Affected Scope:** l.35, l.41 (evidence : `internal/db/pool.go:42-68` — seuls `Store.InTx` l.44 et `WithTenant` l.59, receivers valeur)
*   **Discrepancy Description:** le contrat exige une spécification zéro-décision (lots l.21) ; le plan nomme les types sans signature ni mécanisme de liaison requêtes/sqlc↔tx ni couverture pool_test L4.
*   **Rationale:** signature, relation avec `Store.InTx`, absence de `set_config` tenant, mode de liaison sans exposition `*pgdb.Queries`, appel `CreateUserOwner` sur le même `InTx` — tout reste à décider par l'implémenteur L4/L5.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** vérifié en lecture (pool.go) ; l'option retenue ci-dessous est la seule compatible avec la règle absolue « `*pgdb.Queries` ne franchit pas une frontière de module » tout en restant un wrapper strict du code existant.
*   ***Proposed Edit:***
    *   Location: l.35 (fin de la puce Interfaces de Phase 4) — insertion
    *   Change : ajouter : « **Spécification zéro-décision de l'Unit of Work (livrable plateforme L4, `pool.go` manuscrit)** : `type InTx interface { Queries() *Queries }` — la `*Queries` liée au tx est obtenue uniquement à l'intérieur du code du module exécutant (jamais paramètre ni retour d'un port cross-module) ; `func (s Store) WithinTx(ctx context.Context, fn func(tx InTx) error) error` — wrapper strict de `Store.InTx` (Begin/Rollback/Commit identiques, aucun `set_config` supplémentaire : les opérations in-Tx du register ciblent le pool admin BYPASSRLS sans contexte tenant) ; binding des repos : `tenantRepo.CreateInTx` exécute `tx.Queries().CreateTenant(ctx, ...)`, `userRepo.CreateOwnerTx` exécute `tx.Queries().CreateUserOwner(ctx, ...)` **sur le même `InTx`** passé par `pgdb.WithinTx` au service `user` appelant. Aucune autre forme d'API : besoin divergent = STOP + amendement egm-lite. Cas `pool_test.go` exigés pour L4 : commit après `fn` OK, rollback après erreur, jamais `SET` session-wide. » et remplacer dans la même puce « orchestration par `pgdb.WithinTx` » par « orchestration par `pgdb.WithinTx` (spec ci-dessous) ».
    *   Location: l.41 (Phase 5 Interfaces) — remplacement
    *   Change : « déclare le port `TenantProvisioner` (register : `pgdb.WithinTx` pilote la Tx, décision 6) » → « déclare le port `TenantProvisioner` (register : le service appelle `pgdb.WithinTx(ctx, func(tx pgdb.InTx) error { ... })` — spec complète en Phase 4, décision 6) ».

### [C2] Acquisition du client Redis non tranchée — texte prêt C8 [`readiness F2` ≡ `details F5`] - Severity: Major
*   **Affected Scope:** l.30 (Risques Phase 3)
*   **Discrepancy Description:** P1.2 conserve les exports `GetRedis`/`InitRedis` (l.111), la cible prohibe les singletons, le plan ne spécifie pas comment `auth/repository.go` obtient son client.
*   **Rationale:** seul `models/auth.go` consomme `db.GetRedis()` en production (11 sites, l.119-354 — vérifié) ; le texte C8 de l'audit est prêt, verbatim.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** le contrat réserve cette édition à la passe egm-resolve de G-1 (cette passe) ; recommandation constructeur injecté alignée sur la cible architecturale.
*   ***Proposed Edit:***
    *   Location: l.30 — remplacement intégral de la puce par le texte C8 verbatim (audit `1795684415717-audit-lots-impl-migration.md` §C8, bloc « Par ») : la puce « Risques » Phase 3 devient : « **Risques** : sémantique du script Lua (ownership) — couverte par `TestRefreshOnlyConsumesMatchingOwner` (le déplacer au module) ; secrets lus via `os.Getenv` dans le service — injecter un `Secrets` struct au constructeur (DI explicite, testable `t.Setenv`→variables) ; **acquisition du client Redis : `NewRepository(client *redis.Client)` — le client est résolu au câblage dans `internal/app` (lecture du singleton `redisdb.GetRedis()` confinée au composition root pendant la fenêtre transitoire ; `InitRedis` reste à l'entrypoint `cmd/api`), dette tracée pour sortie post-Phase 8 (cible `.kilo/rules/go-architecture-modulaire.md` : singletons prohibés) ; seul le module `auth` accède aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — import de `internal/platform/redisdb` interdit aux autres modules (`cmd/api`/`cmd/e2edb`/tests autorisés : wiring, purge infra, nettoyages)** ; middlewares deviennent des constructeurs paramétrés (`TokenAuth(svc)`) → routes mises à jour uniquement dans `internal/app` + `tests/` (via app.NewRouter, rien à changer). »

### [C3] `middleware/main_test.go` hors spécification Phase 3 [`readiness F4`] - Severity: Major
*   **Affected Scope:** l.27, l.31, l.54 (evidence : `middleware/main_test.go:14-19` — TestMain, pin `REDIS_DB=2` seulement)
*   **Discrepancy Description:** « dossier supprimé si vide » est faux (main_test.go reste) ; les tests migrés dans le module exigent le pin.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** fait vérifié (lecture du fichier) ; sans migration du pin, un exécutant naïf laisse les tests du module sur l'index Redis 1 (sessions dev) — violation de l'invariant 1/2/3.
*   ***Proposed Edit:***
    *   Location: l.27 — remplacement de la fin de puce « `middleware/` vidé de `auth.go` (CORS/RequestID partis en Phase 1) → dossier supprimé si vide, sinon gardé pour les résidus. » par : « `middleware/` vidé en migrant **le paquet de test complet** : `auth_test.go` (7 tests) **et `main_test.go` (TestMain, pin `REDIS_DB=2`, middleware/main_test.go:14-19) migrent avec le module** et deviennent le socle de test du module `auth` (tout nouveau test du module épingle son environnement selon les critères de sortie : `REDIS_DB=2` + `DB_NAME=artisan_test`) ; `middleware/` ne contient plus aucun fichier → **suppression du répertoire en Phase 3**. »
    *   Location: l.31 — remplacement : « `middleware/auth_test.go` migré vers le module (7 tests), suite `tests/` verte. » → « `middleware/auth_test.go` (7 tests) **et `middleware/main_test.go`** (TestMain, pin `REDIS_DB=2`) migrés vers le module, suite `tests/` verte. »

### [C4] Contradiction Phase 3/Phase 6 : l'admin Login [`details F2` ; intègre style F10 « idem »] - Severity: Major
*   **Affected Scope:** l.25 ↔ l.46 ; design §2 l.92 (evidence : `controllers/user.go:126` et `controllers/admin.go:139` — même message 200)
*   **Discrepancy Description:** l.25 fait migrer `controllers/admin.go` Logout/Login en Phase 3 alors que design §2 + l.46 assignent l'admin Login au module `admin` (Phase 6) ; « admin idem » ambigu en l'état.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** contradiction réelle cross-phases ; le message de logout vérifié permet d'épingler la forme exacte.
*   ***Proposed Edit:***
    *   Location: l.25 — remplacement de la fin de puce « `controllers/user.go` Logout et `controllers/admin.go` Logout/Login → migrés dans `handler.go` du module (logout user 200 « Successfully logged out », admin idem). » par : « `controllers/user.go` Logout et `controllers/admin.go` **Logout** → migrés dans `handler.go` du module (logout user et admin : 200 + `{"message": "Successfully logged out"}` — vérifié `controllers/user.go:126`, `controllers/admin.go:139`) ; l'admin **Login** reste dans `controllers/admin.go` (consumant le `*auth.Service` injecté) jusqu'à sa migration en Phase 6 — l'handler appartient au module `admin` (design §2 : route `/v1/admin/login`). »

### [C5] DTO `TenantPayload` : placement cross-module non tranché [`details F3`] - Severity: Major
*   **Affected Scope:** l.34, l.41, l.53 (evidence : `TenantPayload` défini dans `models/user.go:27-31`)
*   **Discrepancy Description:** l.34 liste le type dans `model.go` du module `tenant` ; l.41/l.53 l'utilisent via ports. Sans tranchage : import type cross-module (= FAIL gate 2) ou duplication non spécifiée ; l'ancre l.34 est en plus fausse.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** option retenue (zero-decision) : types payload définis/portés par chaque consommateur, aucune importation cross-module — la seule compatible avec les règles de dépendance sans amendement du design.
*   ***Proposed Edit:***
    *   Location: l.34 — remplacement « `model.go` (Tenant, TenantPayload, AllTenantResponse) » par « `model.go` (Tenant, AllTenantResponse) — `TenantPayload` n'appartient PAS au module tenant (il est défini dans `models/user.go:27-31`, voir l.41) ».
    *   Location: l.41 — insertion en fin de puce : « **Type `TenantPayload`** : défini et porté par chaque consommateur — le port `UserAuthenticator` (l.53) définit sa propre struct payload tenant, le module `user` garde la sienne dans son `model.go`, le module `tenant` expose les données tenant en retour de son repository ; **aucune importation de type cross-module** (sinon FAIL checklist gate 2). »
    *   Location: l.53 — remplacement « `UserAuthenticator` (`Login(form) (User, TenantPayload, Token, error)`) » par « `UserAuthenticator` (`Login(form) (User, TenantPayload, Token, error)` — ici `TenantPayload` = struct définie par le port, voir l.41) ».

### [C6] Phase 8 : `docs/test.md` périmé [`readiness F3` ≡ `details F4` ≡ `style F11`] - Severity: Major
*   **Affected Scope:** l.59 (evidence : commit `581d770` « docs(test): remove test documentation », fichier absent à HEAD ; artefacts racine présents)
*   **Discrepancy Description:** « traiter la suppression non commitée » demande de traiter un état qui n'existe plus ; les 3 artefacts racine, eux, sont toujours présents.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** factuel ; le même énoncé périmé existe au design l.73 — hors DetailedRange : note pour egm-drift-report de clôture, pas d'édition du design certifié.
*   ***Proposed Edit:***
    *   Location: l.59 — remplacement de « ; traiter la suppression non commitée `docs/test.md`. » par « — toujours présents le 2026-11-26 ; `docs/test.md` : dérive résolue par le commit `581d770` (suppression commitée, fichier absent à HEAD, vérifié 2026-11-26) — aucune action du lot. »

### [C7] Bandeaux d'introduction l.3-4 [`style F1`] - Severity: Major
*   **Affected Scope:** l.3-4
*   **Discrepancy Description:** méta-légende de scission/review redondante avec l'en-tête du design ; l'essentiel (où est le design) noyé.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** le contenu supprimé reste disponible dans les rapports de review (`style_findings.md` F1 conserve la note de numérotation des ancres) — RELOCATE sans perte.
*   ***Proposed Edit:***
    *   Location: l.3-4 — remplacement des deux blockquotes par un seul : « > Plan d'exécution Phases 0-8 ; le design (objectif, décisions actées, invariants, layout cible) vit dans [1795597264251-architecture-modulaire-par-domaine-design.md](1795597264251-architecture-modulaire-par-domaine-design.md). » (la note « les ancrages de lignes des rapports antérieurs suivent la numérotation de l'original » est RELOCATÉE dans `1795597264251-phases-review/consolidated_report.md` § traçabilité lors de l'application).

### [C8-lot] Mur de texte soft-consume l.26 [`style F2` + readiness F7 + caveats CALIBRATE ×3 + style F14 ; C25 intégré] - Severity: Major
*   **Affected Scope:** l.26 (contenu inchangé, structure recalée ; baseline vérifiée `models/auth.go:262-277`)
*   **Discrepancy Description:** ~600 mots en un paragraphe dupliquent la politique du design ; la forme 200 de la recovery n'est pas épinglée ; 3 formulations à recalibrer.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** réécriture structurelle avec épinglage explicite et calibrages du style persona ; aucun changement de politique.
*   ***Proposed Edit:***
    *   Location: l.26 — remplacement intégral de la puce par la version sous-puces : « **Soft-consume strict du refresh token (décision actée 2026-11-25, design décision 7 — politique fixée dans le design, non recopiée ici)** :
    - **Record** : chaîne STRING inchangée adressée par l'UUID du refresh JWT (`user_sessions:{userID}` = index SET de révocation, pas le record ; **aucune migration de schéma**, format actif « userID:tenantID:role » inchangé) ; état encodé dans la valeur (consommé « consumed:<delivered|recovered>:<value> ») ; atomicité Lua et ownership inchangés (préfixe userID comparé avant consommation — `TestRefreshOnlyConsumesMatchingOwner` déplacé avec le module).
    - **TTL** : plus de DEL immédiat — record consommé = TTL réduit à la grâce **5 min** (min(grâce, TTL restant)) ; éviction TTL primaire ; sweep in-process du module (ticker ~5 min) d'observabilité + suppression des records consommés expirés que le TTL n'a pas encore évacués (invariant de course écrit dans le design).
    - **Rejeu** : jeton déjà délivré (`delivered`) → `401` (même forme de réponse que le chemin invalide actuel) ; jeton consommé non délivré pendant la grâce → **une seule** re-émission de récupération, **réponse de même forme 200 que le happy path** (état `recovered` encodé dans la valeur, jamais dans la réponse HTTP ; sinon `401`) ; crash serveur entre consommation et lecture → échec fermé.
    - **Rollout/rollback** : records actifs au format inchangé (binaire pré-Phase-3 compatible) ; seuls les records consommés (≤ 5 min) portent le format étendu ; rollback pré-Phase-3 borné aux consommations en grâce (les sessions consommées en grâce sont perdues : re-login requis).
    - **Observabilité** : incident de session journalisé (au marquage et à la récupération — index `user_sessions:*`).
    - **Tests exigés** : rejeu / grâce / recovery **unique** / fails-closed / sweep / unicité de la recovery.
    - **Carve-out** : seul le chemin « rejeu pendant grâce » peut différer du comportement actuel — happy path et rejeu après délivrance inchangés au byte près (cf. Risques transverses). »

### [C9] Fenêtre A2 ×3 [`style F3`] - Severity: Major
*   **Affected Scope:** l.28 (occurrence complète conservée), l.68, l.167
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** l.28 porte déjà la version complète ; chaque doublon est une source de dérive.
*   ***Proposed Edit:***
    *   Location: l.68 — remplacement de « la fenêtre transitoire racine↔modules (Phases 3-6, décision A2 — tolérance transitoire explicite, cf. design §4) est actée et vérifiée par le critère de sortie (décompte `git ls-files`). » par « la fenêtre transitoire racine↔modules est couverte par la tolérance transitoire (voir Phase 3) et le critère de sortie (décompte `git ls-files`). »
    *   Location: l.167 — remplacement du libellé « **Fenêtre transitoire et compteur de fichiers legacy (décision A2 — tolérance transitoire explicite ; définition complète et impact : §4 du design)** : pendant les Phases 3-6, les imports racine↔`internal/modules/*` sont tolérés **dans les deux sens** pour le code en cours de migration (fenêtre actée — voir Phase 3). » par « **Fenêtre transitoire et compteur de fichiers legacy** : définition complète en Phase 3 (tolérance transitoire). »

### [C10] Carve-out redit à l.67 [`style F4`] - Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:***
    *   Location: l.67 — remplacement de « **Unique carve-out acté** : le soft-consume du refresh (Phase 3) modifie le comportement du chemin de rejeu pendant la grâce — un jeton non délivré retrouvé pendant la grâce bénéficie de la recovery unique (nouveau chemin, tests dédiés dans le module `auth`) ; le rejeu après délivrance et le happy path restent inchangés au byte près, assertions existantes inchangées. » par « **Unique carve-out acté** : le soft-consume du refresh — chemins, politique et tests dédiés au § Phase 3. »

### [C11] Duplication helpers + report `httpx` (intègre details F9 `getRole` et details F7 `adminID`) [`style F5` + `details F9`, `details F7`] - Severity: Major
*   **Affected Scope:** l.13, l.18, l.21 ↔ l.135-136 (evidence : `getRole` sans appelant, `controllers/user.go:31-33` ; `adminID` posé `middleware/auth.go:86`, consommé `controllers/admin.go:87`)
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** canon unique en P1.6 ; recopier systématiquement `getRole` introduirait du code mort contredisant le principe « pas de code mort sans appelant » invoqué par le plan lui-même ; le contrat de clés documenté doit inclure `adminID`.
*   ***Proposed Edit:***
    *   Location: l.13 — remplacement « 9 `git mv` (config, redisdb, sqlc/pool, mocks, middleware pur, validate) + 1 création (`internal/app` ; `httpx` différé en Phase 2) » par « 9 `git mv` de paquets (config, redisdb, sqlc/pool, mocks, middleware pur, validate) + 2 déplacements d'entrées (`main.go`, `main_test.go` → `cmd/api/`, P1.7) + 1 création (`internal/app` ; `httpx` : voir P1.6) » (intègre aussi C18).
    *   Location: l.18 — remplacement de « (premier appelant réel — reporté depuis P1.6) » par « (canon : justification et contenu visés en P1.6) ».
    *   Location: l.21 — remplacement de « — pas dans `httpx` : les clés de contexte (`userID`, `tenantID`, `role`) sont posées par `TokenAuth` (module `auth`), qui documente leur contrat ; le platform n'en duplique pas la source de vérité. » par « — pas dans `httpx` (canon en P1.6). »
    *   Location: l.136 — remplacement de « - Pas de `identity.go` : les accesseurs `getUserID`/`getTenantID`/`getRole` restent locaux aux handlers consommateurs (recopiés depuis `controllers/user.go:19-33`) — les clés de contexte (`userID`, `tenantID`, `role`) sont posées par `TokenAuth` (module `auth`), qui documente leur contrat ; le platform n'en duplique pas la source de vérité. » par « - Pas de `identity.go` ; `httpx` créé en Phase 2, au premier appelant réel — pas de code mort sans appelant. Les accesseurs restent locaux aux handlers consommateurs : **recopier uniquement les accesseurs appelés par le handler** (`getRole`, `controllers/user.go:31-33`, est aujourd'hui sans appelant — ne le recopier que si le handler en a besoin). Les clés de contexte sont posées par le module `auth`, qui documente leur contrat : `userID`, `tenantID`, `role` (TokenAuth, middleware/auth.go:51-53) et `adminID` (AdminAuth, middleware/auth.go:86 — consommée par l'handler `SetTenantStatus`, controllers/admin.go:87) ; le platform n'en duplique pas la source de vérité. »

### [C12-lot] Dette `binding.Validator` ×3 [`style F6` (+ style F9 sur l.140)] - Severity: Major
*   **Affected Scope:** l.60, l.131, l.140
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** canon unique en P1.7(a) ; l.140 porte aussi le calibrage « acceptable → toléré ».
*   ***Proposed Edit:***
    *   Location: l.60 — remplacement de la puce par : « - **Sortie de dette tracée — `binding.Validator` (backlog explicite, directive 2026-11-25)** : traiter la dette P1.7(a) (canon unique) ; élimination requise à la sortie de migration. »
    *   Location: l.131 — remplacement de « passera à `validate.DefaultValidator` — même comportement (tag `binding`, règle `fullName`). » par « passe à `validate.DefaultValidator` — même comportement (tag `binding`, règle `fullName`) ; la dette (relocalisation) est canonisée en P1.7(a). »
    *   Location: l.140 — remplacement de « — acceptable en transitionnel, à déplacer vers l'entrypoint/`platform/validate` en sortie de migration ; » par « — toléré en transition, déplacé vers l'entrypoint/`platform/validate` et éliminé en sortie de migration (Phase 8) ; »

### [C13-lot] Clarté du risque slug l.36 [`style F7`] - Severity: Major
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:***
    *   Location: l.36 — remplacement de la puce par : « - **Risques** : boucle de collision slug (`23505`, max 99 tentatives) — tests purs à déplacer. Pendant la migration : conserver le comportement actuel d'échec d'épuisement des 99 tentatives (erreur discriminée + code HTTP actuels), déjà épinglé par la suite `tests/`. Après la migration : écrire la spécification cible dans le `service_test.go` pur. `reservedSlugs` contient des racines de routes : le test contrat (mécanique `TestAPISurface`, P0.3) échoue si une racine de route réelle manque dans la liste — à compléter d'un commentaire pointant vers la liste. »

### [C14] Slogan couverture l.164 [`readiness F5`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Nit mais fix pertinent : le contrat exige l'opérationnalisation qualitative ; sans elle un exécutant pourrait inventer un seuil.
*   ***Proposed Edit:***
    *   Location: l.164 — remplacement de « Aucune phase ne baisse la couverture. » par « **Charge de tests consignée** : le dépôt n'instrumente pas la couverture — exigence qualitative : le rapport de gate du lot (`L<N>-gate.md`) consigne l'inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte) ; aucun seuil chiffré. »

### [C15] Signatures `ArticleRepository` elliptiques l.19 [`readiness F6`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** signatures vérifiées sur `models/article.go:76-152` — calquage direct, zéro interprétation.
*   ***Proposed Edit:***
    *   Location: l.19 — remplacement de la puce par : « - **Interfaces** (côté service, consommateur) : `type ArticleRepository interface { Create(ctx context.Context, tenantID, userID uuid.UUID, form forms.CreateArticleForm) (uuid.UUID, error); One(ctx context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error); All(ctx context.Context, tenantID, userID uuid.UUID) ([]Result, error); Update(ctx context.Context, tenantID, userID, id uuid.UUID, form forms.CreateArticleForm) error; Delete(ctx context.Context, tenantID, userID, id uuid.UUID) error }` — signatures calquées sur `models/article.go:76-152` (les types de la phase circulent via `dto.go`/`model.go`) ; implémentée par `repository.go` sur `pgdb.Store` ; constructeur `NewHandler(articles *Service) *Handler` — `Handler` porte les 5 méthodes HTTP routées dans `internal/app` —, câblé dans `internal/app`. »

### [C16] TestMain `tests` : « exige PostgreSQL » inexact l.98 [`details F6`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** fait vérifié (`tests/main_test.go:16-25` : setenv seulement) ; TestAPISurface ne requiert réellement pas de DB.
*   ***Proposed Edit:***
    *   Location: l.98 — remplacement de « l'exécution passe par le `TestMain` du paquet `tests`, qui exige PostgreSQL actif et la base `artisan_test` provisionnée. » par « le `TestMain` du paquet `tests` épingle `REDIS_DB=2` et `DB_NAME=artisan_test` (tests/main_test.go:16-25 — setenv seulement, sans connexion). L'exécution reste faite après P0.2 (provisionnement), conformément à la discipline de la suite. »

### [C17] Clé de contexte `adminID` omise [`details F7`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (via l'édition C11, l.136 et l.21 — édit partagé, pas d'édition séparée)
*   ***Justification:*** le type traverse la même frontière de module (`auth` → `admin`) ; seule la documentation manque.

### [C18] Décompte `git mv` l.13 [`details F8`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (via l'édition C11, l.13) — le complément lots plan l.47 (« 9 `git mv` + 1 création ») est HORS portée de cette passe : harmonisation facultative à faire approuver séparément (gouvernance contrat).
*   ***Justification:*** décompte réel : 9 paquets + 2 entrées (`main.go`, `main_test.go`, P1.7).

### [C19] « le plus complexe, en dernier » l.51 [`style F8`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:***
    *   Location: l.51 — titre devient : `### Phase 7 — Module `auth` : collage final (dépend des phases 3-6)`.

### [C20] Subjectifs « vrai »/« acceptable » + disclaimer file:line [`style F9` + `style F13` (l.76) ; l.143] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (l.140 traité dans C12-lot)
*   ***Proposed Edits:***
    *   Location: l.76 — remplacement de la phrase de discipline : « ne jamais lancer `make test`/`make e2e`/`go test` en parallèle dans le même checkout (convention AGENTS.md « Risques »). Les numéros de ligne (`main.go:16`, `models/auth.go:11`, etc.) décrivent l'état initial du dépôt — ils deviennent obsolètes dès la première étape validée ; la liste des fichiers fait foi, la compilation verte est le vrai contrôle. » par « ne jamais lancer `make test`/`make e2e`/`go test` en parallèle dans le même checkout (convention AGENTS.md « Risques »). Discipline commune à TOUTES les phases (les Phases 2-7 utilisent aussi des refs `file:line`, p.ex. `models/auth.go` ici, `controllers/admin.go:55-64`, `models/login.go:34-38`) : les numéros décrivent l'état initial du dépôt et deviennent obsolètes dès la première étape validée ; la liste des fichiers fait foi, la compilation verte fait foi. »
    *   Location: l.143 — remplacement de « la suite épingle désormais le vrai câblage. » par « la suite épingle le câblage de production. »

### [C21] « idem » ambigu l.25 [`style F10`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (via l'édition C4, l.25 — édit partagé : forme précise vérifiée `controllers/user.go:126` / `controllers/admin.go:139`)

### [C22] Parenthèse paraphrase l.22 [`style F12`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:***
    *   Location: l.22 — remplacement de « suite `tests/` verte sans modification des assertions (preuve de non-régression), » par « suite `tests/` verte sans modification des assertions, ».

### [C23] Disclaimer file:line mal scopé l.76 [`style F13`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (via l'édition C20, l.76)

### [C24] Règle d'emphase transverse [`style F14`] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (portée limitée : appliquée aux lignes réécrites par cette passe — l.26 structurée en sous-puces (C8-lot), parenthèses multi-clauses converties là où le texte est retouché ; pas de re-écriture globale du document)

### [C25] Forme 200 recovery + caveats CALIBRATE l.26 [`readiness F7` + caveats] - Severity: Nit
*   ***AI Suggestion:*** DECLINE (traité par construction dans l'édition C8-lot — compté une seule fois)

## Notes hors périmètre (aucune édition TargetFile)

- **Design l.73** : même mention `docs/test.md` périmée (design certifié, intouchable ici) → à traiter par `egm-drift-report` en clôture L8 (contrat l.103).
- **Lots plan l.47** : livrable L1 « 9 `git mv` + 1 création » à harmoniser éventuellement en « 9 `git mv` de paquets + 2 déplacements d'entrées (P1.7) + 1 création » — mini-édition directe du contrat, hors éditions resolve, à faire approuver séparément.
