# Rapport Codebase Details Goldfish — `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`

**Mode** : REVIEW (DetailedRange = document entier, lignes 6-168 ; l.1-5 = contexte d'en-tête) · **Persona** : `egm-codebase-details` · **GoldfishContext** : design `1795597264251-architecture-modulaire-par-domaine-design.md` (l.6-l.168, décision 7 l.39 lue en intégralité). Standard appliqué : `implementation-readiness-standard.md` (règle zéro-décision). Révision croisée du contrat d'exécution `1795627363688-lots-impl-migration-modulaire.md` (l.21, l.29-36, l.49-54, l.73, l.75) uniquement pour la formulation lots/gates.

## Synthèse

Le pas-à-pas P0.x-P1.x est remarquablement fiable : le sweep d'ancrage complet (l.106-158 et Phases 2-8, l.17-63) contre le code réel n'a trouvé **aucun chemin de fichier erroné** et **aucune erreur de ligne** (le disclaimer l.76 de péremption des numéros est en plus correctement posé). Les trois dérives factuelles détectées portent sur la **sémantique** (conflit Phase 3/Phase 6 sur l'admin Login), l'**état du dépôt** (`docs/test.md`) et des **lacunes de spécification** (Unit of Work `pgdb.InTx`/`pgdb.WithinTx`, DTO partagé `TenantPayload`, acquisition du client Redis) — exactement les points d'audit que le contrat du Stage 9 (l.29-36) exige de ce review.

Verdicts sur les 7 points d'audit obligatoires du contrat (l.29-36) :

1. **Spec zéro-décision `pgdb.InTx`/`pgdb.WithinTx`** — ❌ **Finding** (Major, F1) : le fait est confirmé (`internal/db/pool.go:42-68` ne contient que `Store.InTx` l.44 et `WithTenant` l.59 ; aucune interface `InTx`, aucune fonction `WithinTx`), et la plan-phases ne spécifie ni la signature ni le mécanisme de liaison requêtes↔tx.
2. **Sweep d'ancrage complet** — ✅ **Confirmé** : tous les ancrages P0.x-P1.x et Phases 2-8 vérifiés corrects (voir « Rélevé d'ancrage » ci-dessous) ; 0 chemin faux, 0 ligne erronée ; 1 incohérence sémantique inter-phases (F2) et 3 nits de précision (F6-F9).
3. **Fraîcheur Phase 8** — ❌ **Finding** (Major, F4) : `docs/test.md` est supprimé **et commité** (commit `581d770` « docs(test): remove test documentation », vérifié par `git show --stat` : 1 fichier, 4- deletions ; `Test-Path docs/test.md` = False à HEAD) — la « suppression non commitée » de l.59 (et l.73 du design) est périmée. Les artefacts racine `gin-boilerplate`, `e2edb.exe`, `tmp_line.txt` sont, eux, **confirmés présents** à la racine (ls).
4. **Cohérence de la puce Phase 3 soft-consume (l.26) ↔ design décision 7 (l.39)** — ✅ **Confirmé** : formats (`consumed:<delivered|recovered>:<value>` ≡ `consumed:<delivered|recovered>:userID:tenantID:role` — la valeur EST le triplet), grâce 5 min `min(grâce, TTL restant)`, politique de rejeu (délivré → 401 même forme ; non délivré → recovery unique encodée dans la valeur ; crash → échec fermé), TTL primaire / sweep ~5 min d'observabilité + purge défensive, invariant de course, journalisation au marquage et à la récupération, tests dédiés (mêmes six), rollout/rollback borné aux consommations en grâce : tout est aligné au contenu près. Le mécanisme réel que le soft-consume remplace a été vérifié (`models/auth.go:262-277` : script Lua DEL immédiat si ownership OK) ; le carve-out annoncé correspond bien au seul chemin modifié.
5. **Opérationnalisation de « aucune phase ne baisse la couverture » (l.164)** — ✅ **Confirmé** : l.164 énumère des livrables de tests objectifs par phase (tests purs déplaçés, service_test.go pur à fake avec cas discriminés nommés par phase), ce qui donne au gate lots un inventaire vérifiable ; le rapport de gate (inventaire : fonctions déplacées/ajoutées, exécution verte) et l'absence de seuil chiffré restent posés dans le contrat d'exécution (lots plan l.75, déjà appliqué) — le plan-phases ne re-spécifie pas, conformément à son rôle.
6. **Spec zéro-décision de l'acquisition du client Redis (Phase 3)** — ❌ **Finding** (Major, F5) : la puce Risques (l.30) ne dit rien de l'acquisition ; les faits amont sont confirmés (production : seul `models/auth.go` consomme `db.GetRedis()` — 11 sites l.119-354 ; `middleware/auth.go` ne touche Redis qu'au travers des méthodes `AuthModel`). Texte correctif **C8 prêt et vérifié** dans `1795684415717-audit-lots-impl-migration.md` (§C8, cible l.30 de la plan-phases : `NewRepository(client *redis.Client)` + lecture du singleton confinée au composition root + dette tracée) — tous ses faits amont ont été re-vérifiés exacts.
7. **Propriétaire unique Redis** — ✅ **Confirmé et cohérent** : fait vérifié — en production, les seules lectures/écritures de clés de session (records UUID Set/Get/Del, index `user_sessions:%v` l.260/232/247, préfixe `admin:` l.359-361, script Lua) sont dans `models/auth.go` ; `main.go:144` appelle `InitRedis` (pas `GetRedis`), `cmd/e2edb` n'utilise que `RedisDBIndex` (l.132) avec son propre client (`cmd/e2edb/main.go:174-178`) ; tous les autres usages de `GetRedis` sont dans `*_test.go`. La règle est codifiée côté contrat (gate checklist item 8, lots plan l.73) ; la plan-phases devient explicitement cohérente dès que C8 est appliqué (sa proposition inclut l'interdiction d'import `redisdb` hors `auth`/wiring/tests).

**Compte** : Blocker 0 · Major 5 · Nit 4.

**Verdict** : conforme pour exécution après résolution par la passe egm-resolve de G-1 — aucun Blocker ; les 5 Majors sont des éditions bornées de la plan-phases (4 d'entre elles ont déjà un texte corrigeable prêt ou trivial), aucune ne fragilise l'architecture ni le contrat HTTP.

---

## Findings

### F1 — Phase 4/5 : l'Unit of Work `pgdb.InTx`/`pgdb.WithinTx` est livré sans spécification zéro-décision
- **Sévérité** : Major
- **Périmètre** : Phase 4 Interfaces (l.35), Phase 5 Interfaces (l.41), Contrat L4 (lots plan l.50) ; baseline réelle vérifiée : `internal/db/pool.go:42-68` (seuls `Store.InTx` l.44 et `Store.WithTenant` l.59 ; ni `pgdb.InTx` ni `pgdb.WithinTx` n'existent aujourd'hui)
- **Justification** : le plan pose le nom (`CreateInTx(ctx, tx pgdb.InTx, name) (Tenant, error)`, « `pgdb.WithinTx` pilote la Tx ») mais laisse **trois décisions de conception à l'implémenteur** : (a) la signature Go exacte de l'« interface plateforme étroite `pgdb.InTx` » ; (b) la sémantique de `pgdb.WithinTx` (signature, retour, commit/rollback, relation exacte avec `Store.InTx` — le design l.120 dit « wrapper de `Store.InTx` + `WithTenant` » sans fixer la forme) ; (c) **le mécanisme par lequel un repository concret lie ses requêtes sqlc à la transaction sans exposer `*pgdb.Queries`** — précisément ce que le contrat exige d'arbitrer (lots plan l.21 : « G-1 doit exiger une spécification zéro-décision »). Les encodages viables divergent (interface exposant `Queries()` — la valeur `*Queries` reste alors confinée à l'usage interne du repository appelant ; interface portant les opérations nécessaires ; interface `Exec/Query` générique), et le mauvais choix fait échouer la clause gate 1 (aucun `*pgdb.Queries` franchissant une frontière de module, lots plan l.66). Point l.164 couvre `pool_test.go` mais le plan n'exige pas de cas spécifiques pour le nouveau code (contrat L4 les mentionne : « + couverture pool_test.go »).
- **Proposition concrète** (éditer l.35 et l.41 — au choix de l'Elephant, mais UNE forme fixée) : spécifier dans la plan-phases, p.ex. `pkg db // alias pgdb` : `type InTx interface { Queries() *Queries }` (l'`*Queries` lié au tx n'est obtenu qu'à l'intérieur du code de module, jamais paramètre/retour d'un port) ; `func WithinTx(ctx context.Context, s Store, fn func(InTx) error) error` (wrapper strict de `Store.InTx`, aucun `set_config` supplémentaire pour le pool admin) ; documenter que `TenantProvisioner.CreateInTx(ctx, tx pgdb.InTx, name)` et `CreateOwnerTx(ctx, tx pgdb.InTx, ...)` appellent `tx.Queries()` en interne, et ajouter à l.164 les cas `pool_test.go` exigés pour L4 (commit après fn OK, rollback après erreur, aucun `SET` session-wide).

### F2 — Phase 3 (l.25) contredit Phase 6 (l.46) : l'admin Login change de module propriétaire
- **Sévérité** : Major
- **Périmètre** : Phase 3 Sources (l.25, « `controllers/admin.go` Logout/Login → migrés dans `handler.go` du module ») ↔ Phase 6 Sources (l.46, « `models/admin.go` + `controllers/admin.go` → `handler.go` (Login/Tenants/SetTenantStatus/Logout) ») ; design §2 (l.89 : routes auth = `/v1/login`, `/v1/token/refresh`, `/v1/user/logout`, `/v1/admin/logout` — `/v1/admin/login` est côté admin, l.92)
- **Justification** : l.25 fait migrer l'admin **Login** dans l'handler du module `auth` en Phase 3, alors que le design (§2, l.92) et la Phase 6 (l.46, L6 du contrat l.52) l'assignent au module `admin`, avec `TenantReader`/`AdminAuthenticator` côté admin. La contradiction est réelle et non cosmétique : un exécutant naïf de la Phase 3 déplace l'admin Login, L6 voit son périmètre déplacé et sa validation (« suite admin » l.49) se trouve dans le mauvais module ; le câblage `RequirePermission`/ports de l.53 s'en trouve ambigu.
- **Proposition concrète** : corriger l.25 en « `controllers/user.go` Logout et `controllers/admin.go` **Logout** → migrés dans `handler.go` du module » et préciser : « l'admin Login reste dans `controllers/admin.go` (consommant l'`*auth.Service` injecté) jusqu'à sa migration en Phase 6 (l.46) — création des tokens admin = service `auth` conformément au design §2, mais l'handler appartient au module `admin` ».

### F3 — DTO partagé `TenantPayload` : placement cross-module non tranché, conflit potentiel avec la checklist gate 2
- **Sévérité** : Major
- **Périmètre** : Phase 4 model.go (l.34 : « `model.go` (Tenant, TenantPayload, AllTenantResponse) »), Phase 5 (l.40 : types « User/UserLoginResponse », dont le service Login retourne `(User, TenantPayload, Token, error)` l.41), Phase 7 port (l.53 : `UserAuthenticator (Login(form) (User, TenantPayload, Token, error))`) ; état réel : `TenantPayload` est défini dans **`models/user.go:27-31`**, pas dans `models/tenant.go`
- **Justification** : l.34 fait entrer `TenantPayload`, défini dans `models/user.go` (Phase 5), dans le `model.go` du module `tenant` (Phase 4). Après coup, le service `user` (l.41) et l'auth lui-même via le port `UserAuthenticator` (l.53) référencent ce type : soit `user`/`auth` importent le type du module `tenant` (couplage module→module **hors** liste certifiée des interfaces consommatrices — FAIL des gate 2, lots plan l.67, « toute abstraction nouvelle hors design = FAIL »), soit chaque module redéfinit son propre payload (duplication + conversions non spécifiées). Le plan ne tranche rien : l'implémenteur décidera. (Le fait anchor est en même temps une dérive : le plan attribue le type à `models/tenant.go` alors qu'il vit dans `models/user.go:27-31`.)
- **Proposition concrète** : trancher en plan-phases une seule des options et la documenter à l.34/l.41, p.ex. (recommandé) : « `TenantPayload` reste un type **défini et porté par chaque consommateur** : le port `UserAuthenticator` (module `auth`) définit sa propre struct tenant-payload, le module `user` définit la sienne, `tenant` expose les colonnes brutes `GetTenant*` en retour de son repository ; aucune importation de type cross-module » — ou (alternatif) : amender la règle §4/design pour lister les types partagés autorisés et l'écrire dans la checklist gate 2. Dans tous les cas : corriger l.34 pour refléter la localisation actuelle (`models/user.go:27-31`).

### F4 — Phase 8 (l.59) : la « suppression non commitée `docs/test.md` » est périmée — la suppression est commitée (`581d770`)
- **Sévérité** : Major
- **Périmètre** : Phase 8 (l.59, « traiter la suppression non commitée `docs/test.md` ») ; état réel vérifié : commit `581d770` « docs(test): remove test documentation » présent dans l'historique (`git log --oneline`), `git show 581d770 --stat` = 1 fichier, 4 deletions sur `docs/test.md` (ajouté dans `7a68626`) ; `docs/test.md` inexistant à HEAD (`Test-Path` False ; arbre propre, aucun fichier supprimé non commité)
- **Justification** : l'instruction demande à l'implémenteur de Phase 8 de « traiter » une suppression uncommitée qui n'existe plus — lecture confuse pour un lecteur neuf (il cherchera un état de git qui n'est plus le sien) et contredit l'état vérifié du dépôt. Le lots plan (l.19) avait déjà le doute (« vraisemblablement déjà résolu par `581d770` … à faire confirmer par G-1 ») — confirmation : oui, la dérive est réelle. Le design (l.73) porte la même affirmation périmée mais vit hors du DetailedRange de cette review.
- **Proposition concrète** : remplacer l.59 par « supprimer les artefacts racine (`gin-boilerplate`, `e2edb.exe`, `tmp_line.txt`) — toujours présents en racine le 2026-11-26 ; `docs/test.md` : dérive résolue par le commit `581d770` (suppression commitée, fichier absent à HEAD) — aucune action du lot ». Retirer toute mention de `docs/test.md` de la checklist d'exécution.

### F5 — Phase 3 : l'acquisition du client Redis par le repository sessions n'est spécifiée nulle part (texte correctif C8 prêt)
- **Sévérité** : Major
- **Périmètre** : Phase 3 Risques (l.30, qui ne mentionne que le script Lua, les secrets et les middlewares) ; design l.39 et règles cible `.kilo/rules/go-architecture-modulaire.md` (singletons prohibés, `GetRedis()` cité) ; fait amont vérifié : en production, seul `models/auth.go` consomme `db.GetRedis()` (11 sites, l.119-354) — `middleware/auth.go` ne touche au client qu'au travers de `AuthModel`
- **Justification** : P1.2 conserve les exports `InitRedis`/`GetRedis` (l.111) ; sans spécification, l'implémenteur de Phase 3 doit choisir entre appeler `redisdb.GetRedis()` depuis le module (singleton, interdit par la cible) ou injecter le client — choix de conception interdit par le standard appliqué (zéro-décision). Le texte C8 (`1795684415717-audit-lots-impl-migration.md`, §C8) fixe l'option recommandée et ses faits amont sont **tous confirmés** ce jour.
- **Proposition concrète** : appliquer C8 à l.30 par la passe egm-resolve de G-1 (jamais en direct) : « acquisition du client Redis : `NewRepository(client *redis.Client)` — résolution au câblage dans `internal/app` (lecture `redisdb.GetRedis()` confinée au composition root pendant la fenêtre transitoire ; `InitRedis` reste à `cmd/api`), dette tracée post-Phase 8 ; seul le module `auth` accède aux clés de session — import de `internal/platform/redisdb` interdit aux autres modules (`cmd/api`/`cmd/e2edb`/tests autorisés) ». Côté contrat, l'exigence est déjà portée par la checklist gate 8 (lots plan l.73).

### F6 — P0.3 (l.98) : le `TestMain` du paquet `tests` n'« exige » ni PostgreSQL ni la base provisionnée — il épingle seulement les variables d'environnement
- **Sévérité** : Nit
- **Périmètre** : P0.3 (l.98 : « l'exécution passe par le `TestMain` du paquet `tests`, qui exige PostgreSQL actif et la base `artisan_test` provisionnée ») ; état réel : `tests/main_test.go:16-25` (TestMain) ne fait que `os.Setenv("REDIS_DB", "2")` et `os.Setenv("DB_NAME", "artisan_test")` avant `m.Run()` — aucune connexion, aucune assertion d'infra
- **Justification** : `TestAPISurface` (lecture de `docs/swagger.json` + comparaison bidirectionnelle) est correctement conçu « sans requête applicative » et s'exécuterait donc **sans PostgreSQL** ; l'énoncé inverse est une affirmation fausse sur le comportement du code actuel (le lecteur croit dépendre du prérequis de provisionnement pour ce test précis, ou croit que le TestMain refusera de tourner sans DB). Le prérequis réel (provisionnement de `artisan_test` par `make migrate_up`, P0.2 l.85) reste juste pour la suite dans son ensemble.
- **Proposition concrète** : reformuler l.98 : « le fichier `TestAPISurface` ne fait aucune requête applicative ; le `TestMain` du paquet `tests` épingle `REDIS_DB=2` + `DB_NAME=artisan_test` (tests/main_test.go:16-25, sans connexion). L'exécution reste faite après P0.2 (provisionnement), conformément à la discipline de la suite. »

### F7 — Clé de contexte `adminID` (posée par `AdminAuth`) absente du contrat de clés documenté par le plan
- **Sévérité** : Nit
- **Périmètre** : Phase 2 Risques (l.21) et P1.6 (l.136 : « les clés de contexte (`userID`, `tenantID`, `role`) sont posées par `TokenAuth`… ») ; Phase 6 (l.46) ; état réel : `middleware/auth.go:86` (`c.Set("adminID", adminID.String())`) consommé par l'handler `SetTenantStatus` (`controllers/admin.go:87`, `c.GetString("adminID")`)
- **Justification** : le plan documente uniquement le triplet `userID`/`tenantID`/`role` comme contrat de clés du module `auth` (le design l.143 répète l'omission) ; la clé `adminID` existe dans le code, traverse la même frontière de module (`auth` → `admin`) et échappera au futur lecteur : au moment de la Phase 6, l'handler admin migré doit continuer à lire `adminID` posé par `AdminAuth` du module `auth`. Comportement préservé mécaniquement, seule la documentation manque.
- **Proposition concrète** : compléter l.21/l.136 et la phrase « qui documente leur contrat » en « les clés `userID`, `tenantID`, `role` (TokenAuth, middleware/auth.go:51-53) et `adminID` (AdminAuth, middleware/auth.go:86 — consommée par l'handler `SetTenantStatus`) ».

### F8 — En-tête Phase 1 (l.13) : le décompte « 9 `git mv` » omet les deux déplacements de `main.go`/`main_test.go` de P1.7
- **Sévérité** : Nit
- **Périmètre** : Phase 1 objet (l.13 : « 9 `git mv` … + 1 création ») ↔ P1.7 (l.142 : `git mv main.go cmd/api/main.go` + `git mv main_test.go cmd/api/main_test.go`)
- **Justification** : décompte vérifié : config (1), db.go+db_test.go (2), internal/db+internal/mocks (2), cors.go+request_id.go (2), validator.go+errors.go (2) = 9, plus les 2mv de `main.go`/`main_test.go` en P1.7 = 11 déplacements réels. Les livrables mesurables du lot L1 (lots plan l.47 : « 9 `git mv` + 1 création ») reporteront un nombre faux dans le rapport de gate.
- **Proposition concrète** : écrire l.13 « 9 `git mv` de paquets + 2 déplacements d'entrées (`main.go`, `main_test.go` → `cmd/api/`, P1.7) + 1 création (`internal/app`) », ou harmoniser à « 11 `git mv` + 1 création » partout (l.13, lots plan l.47).

### F9 — P1.6/Phase 2 : copier systématiquement `getRole` produirait du code mort (aucun appelant aujourd'hui)
- **Sévérité** : Nit
- **Périmètre** : Phase 2 Risques (l.21) et P1.6 (l.136 : « les accesseurs `getUserID`/`getTenantID`/`getRole` restent locaux aux handlers consommateurs (recopiés depuis `controllers/user.go:19-33`) ») ; état réel : `getRole` (`controllers/user.go:31-33`) n'a **aucun appelant** dans le dépôt (grep : seule la définition)
- **Justification** : l'instruction, lue littéralement, ferait recopier un accesseur actuellement inutilisé dans le handler article (et éventuellement d'autres) — Go ne l'interdit pas, mais c'est du code mort introduit par le plan lui-même, en contradiction avec l'esprit « pas de code mort sans appelant » (P1.6 l.135) que le plan invoque pour reporter `httpx`.
- **Proposition concrète** : préciser l.21/l.136 : « recopier uniquement les accesseurs appelés par le handler ; `getRole` (`controllers/user.go:31-33`) est aujourd'hui sans appelant — le recopier seulement si le handler en a besoin ».

---

## Rélevé d'ancrage (sweep complet P0.x-P1.x, l.106-158, et Phases 2-8, l.17-63)

**Ancrages vérifiés corrects — synthèse par fichier** (tous numéros de ligne exacts, échantillon représentatif vérifié à la lecture du fichier réel) :

- `main.go` : l.16 (import config), l.18 (import db), l.20 (import forms, usage unique l.130), l.21 (alias pgdb), l.22 (middleware), l.56-77 (`runCreateAdmin`), l.127-134 (bloc routeur : gin.Default l.127, SetTrustedProxies l.128, binding.Validator l.130, CORS/RequestID l.132-133, gzip l.134), l.136-146 (pools + Redis), l.144 (`db.InitRedis()`), l.148-182 (14 routes `/v1`), l.148-212 (routes + Swagger + LoadHTMLGlob + `/public` + `/assets` immutable + `serveSPAIndex` + NoRoute) — tout exact.
- `controllers/` : auth.go l.19 (var authModel), l.84-162 (Refresh, logique complète dans l'handler) ; user.go l.16, l.19-33 (accessoires), l.113-126 (Logout 200 « Successfully logged out »), l.94 (`err.Error()` brut) ; article.go l.18, l.99 (`"Message": "Article not found"`) ; admin.go l.15, l.28-54 (Login crée les tokens dans l'handler, l.43-51) — tout exact.
- `models/` : auth.go l.11 (import db), l.119-127 (records + index SET), l.260 (index), l.262 (script Lua, DEL immédiat — baseline que le soft-consume remplace, cohérente), l.359-361 (préfixe `admin:`), `os.Getenv` (secrets) ; user.go l.47, l.122 (`ErrNoRows` register), ordre Login (lecture user → status 403 → bcrypt, l.64-77) ; login.go l.30 (hash factice), l.34-38 (compteur + `DummyHashComparisons`), l.80 (orchestration) ; tenant.go l.99, l.120 (`createTenantInTx`), max 99 tentatives (l.68, l.125), 23505 (l.135), `reservedSlugs` (l.43-58) ; admin.go l.55-64 (audit dans la même Tx que `UpdateTenantStatus`) — tout exact.
- `middleware/` : auth.go l.10 (var authModel), TokenAuth (l.37-56, clés posées l.51-53), RequirePermission (l.60-68), AdminAuth (l.72-89, `adminID` posé l.86) ; cors.go l.10 (origine codée en dur) — exact.
- `db/db.go` : exports `RedisClient`/`DefaultRedisDB`/`RedisDBIndex`/`InitRedis`/`GetRedis` tous existants (l.13-51) ; `db/db_test.go` présent ; garde CI `! git grep "InitRedis(1)"` réelle (ci.yml l.96) — exact.
- `internal/db/pool.go` : `Store`/`Connect`/`InTx` (l.44)/`WithTenant` (l.59)/`postgresDSN`/`InitTenant`/`InitAdmin`/singletons tenantStore/adminStore (l.92-121) ; **absence confirmée** de `pgdb.InTx`/`pgdb.WithinTx` — exact (baseline du point d'audit 1).
- `sqlc.yaml:9` (out internal/db), `.mockery.yaml:9,11` (paquet + template `../mocks`), `internal/mocks/mock_DBTX.go`+`mock_Querier.go` présents — exact.
- `internal/testutil/` : `factory.go` sans `pgdb` ni `binding.Validator` (conforme « pas touché », l.130) ; `factory_test.go:86` (`binding.Validator = new(forms.DefaultValidator)`) exact ; `fixtures.go` importe `pgdb` ; `postgres_test.go` (tag `all`) importe `pgdb` ; `postgres.go` sans `pgdb` — exact.
- `.github/workflows/ci.yml` : l.83 (build smoke), l.88 (`git diff --exit-code internal/db`), l.96 (garde `InitRedis(1)`), l.139 (route list contenant `/v1/invoice` (+`/v1/invoice/download`)), l.157 (fragment `b"INV-001"`), l.166-168 (bloc PDF) ; dérive smoke préexistante confirmée (routes absentes de `main.go` : NoRoute 404 sur `/v1*', l.206-212) ; 14 routes API réelles ↔ liste P0.3 (l.97) conformes ; `docs/swagger.json` existant, chemins `/article/{id}`, `/admin/tenants/{id}/status` (préfixe `/v1` via BasePath — la normalisation prévue est la bonne) — exact.
- `Makefile` : l.50 (`@go run .`), l.88 (generate_docs, plain `swag init` à remplacer par `-g cmd/api/main.go`), l.168 (`go run -mod=readonly . -create-admin`) — exact.
- `app/playwright.config.ts` : l.51 (`command: "go run ."`), l.54 (cwd "..") — exact.
- `cmd/e2edb/main.go` : l.45 (import db, `RedisDBIndex` l.132), l.200 (commentaire « (internal/db) ») — exact.
- `main_test.go` : l.10 (import pgdb, tag `all`) — exact.
- `tests/` : article_test.go l.20 (db), l.22 (pgdb), l.34-80 (`SetupRouter`), l.39 (binding.Validator) ; login_test.go l.18/20 ; main_test.go TestMain (see F6) — exact (F6 sauf).
- `controllers/auth_test.go` l.15-16, `TestRefreshOnlyConsumesMatchingOwner` (l.27) ; `middleware/auth_test.go` l.12, **7 tests** comptés (l.37, 77, 104, 121, 139, 161, 195) — exact.
- `config/` : secrets.go + secrets_test.go présents ; exports `DefaultValidator`/`Translate`/`ValidationMessages`/`DefaultMessage` réels — exact.
- `forms/` : validator.go, errors.go, 4 fichiers domaine (auth/user/article/admin) — exact ; aucun `forms.Translate` direct dans `*_test.go` (conditional l.130 correct).
- Références de plans existantes : `.kilo/plans/1795284300000-remove-public-access.md`, `1795597264251-architecture-modulaire-par-domaine.md` (l.3), `zoom_out_report.md`, `goldfish_review_report.md` (l.4), fixtures `tests/fixtures/`, `memory-bank/progress.md` — tous existants.
- Racine : `gin-boilerplate`, `e2edb.exe`, `tmp_line.txt` présents (Phase 8, l.59 confirmé).
- Aucune autre occurrence de `go run .` à traiter (grep : Makefile:50 et playwright.config.ts:51 seulement ; `go run -mod=readonly . -create-admin` Makefile:168 couvert par P1.8).

**Dérives relevées individuellement** :
| Cible | Réel | Gravité | Finding |
|---|---|---|---|
| l.25 : « `controllers/admin.go` Logout/Login » migrés en Phase 3 | design §2 l.92 + l.46 : l'admin Login appartient au module `admin` (Phase 6) | Major | F2 |
| l.34 : `TenantPayload` listé dans les sources `models/tenant.go` du module tenant | défini dans `models/user.go:27-31` ; cross-module non tranché (l.41, l.53) | Major | F3 |
| l.59 : « suppression non commitée `docs/test.md` » | commitée (`581d770`) ; fichier absent à HEAD ; artefacts racine présents (partie confirmée) | Major | F4 |
| l.98 : TestMain « exige PostgreSQL actif » | `tests/main_test.go:16-25` : Setenv seulement, aucune connexion | Nit | F6 |
| l.21/l.136 : clés de contexte `userID`/`tenantID`/`role` | il manque `adminID` (middleware/auth.go:86 → controllers/admin.go:87) | Nit | F7 |
| l.13 : « 9 `git mv` » | 9 + 2 (`main.go`/`main_test.go`, P1.7 l.142) | Nit | F8 |
| l.21/l.136 : recopier `getUserID`/`getTenantID`/`getRole` | `getRole` sans appelant (controllers/user.go:31-33) | Nit | F9 |

---

## Tableau récapitulatif

| # | Sévérité | Périmètre cible | Objet |
|---|---|---|---|
| F1 | Major | l.35, l.41 (Phase 4/5) ; lots l.50, l.66 | Spec zéro-décision `pgdb.InTx`/`pgdb.WithinTx` + liaison requêtes↔tx |
| F2 | Major | l.25 ↔ l.46 ; design l.92 | Admin Login : Phase 3 vs Phase 6, module propriétaire contradictoire |
| F3 | Major | l.34, l.40, l.53 | DTO partagé `TenantPayload` : placement cross-module non tranché (gate 2) |
| F4 | Major | l.59 | Phase 8 : `docs/test.md` — suppression commitée `581d770`, instruction périmée |
| F5 | Major | l.30 ; C8 prêt (audit l.133-147) | Acquisition client Redis du repo sessions non spécifiée (texte C8 vérifié) |
| F6 | Nit | l.98 | TestMain `tests` n'exige pas PostgreSQL (Setenv seulement) |
| F7 | Nit | l.21, l.136, l.46 | Clé de contexte `adminID` omise du contrat de clés |
| F8 | Nit | l.13 ↔ l.142 | Décompte `git mv` Phase 1 (9 vs 11) |
| F9 | Nit | l.21, l.136 | `getRole` sans appelant : recopie systématique = code mort |

**Total : Blocker 0 · Major 5 · Nit 4**

---

*Persona egm-codebase-details — Goldfish context frais, lecture seule du code et des plans, rapport unique écrit dans `codebase_details_findings.md`. Chaque fait code-dépendant de ce rapport a été vérifié en lisant le fichier réel cité (chemins UNC `\\wsl.localhost\Debian\...`) ; l'état git a été vérifié par `git log/show/status` (aucune modification du dépôt).*