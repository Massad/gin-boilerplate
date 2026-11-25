# Migration incrémentale vers une architecture modulaire par domaine — Design

> Scission du plan original `1795597264251-architecture-modulaire-par-domaine.md` (268 lignes, trop volumineux pour le cycle egm-review) : ce document reprend les lignes 1-116 et 261-268 de l'original (117 et 260 : lignes vides — aucun contenu perdu). Le plan d'exécution (lignes 118-259 : feuille de route Phases 0-8, proposition P0-P1, critères de sortie) vit dans [1795597264251-architecture-modulaire-par-domaine-plan-phases.md](1795597264251-architecture-modulaire-par-domaine-plan-phases.md).
> Review EGM — Elephant-Goldfish Model, protocole de review multi-agent. Cycle `feature_spec` d'egm-review : 5 zoom-out initial, 6 test de compréhension, 7 review critique, 8 zoom-out intermédiaire ; rapports et reports dans `1795597264251-design-review/`. Les ancrages de lignes des rapports antérieurs (`zoom_out_report.md`, `goldfish_review_report.md`, racine de `.kilo/plans/`) suivent la numérotation de l'original.

## Préambule

**Cible** : `cmd/api/` (entrypoint + DI explicite), `internal/platform/` (socle technique transversal), `internal/modules/<domaine>/` (handler / service / repository / dto / model / tests). Chaque phase laisse `go build ./...`, `go vet ./...` et `make test` au vert. Aucune montée de version de dépendance n'est mélangée à la migration (go-redis/v7, jwt/v4 restent en l'état — follow-up possible).

### Problème

L'API fonctionne, mais son wiring est global : singletons de pools (`internal/db/pool.go:92-121`), client Redis global (`db/db.go`), vars de package `*Model` — aucun constructeur n'injecte de dépendance. La logique métier vit dans les handlers Gin (Refresh `controllers/auth.go:84-162`, Logout, admin Login), le vrai câblage HTTP n'est épinglé que par le smoke CI (étape de `.github/workflows/ci.yml` qui démarre l'application et sonde les routes HTTP en exigeant les codes attendus, p.ex. 200/PDF) et Playwright (`SetupRouter` duplique les routes de `main.go`), et l'ajout d'un domaine touche quatre paquets racine sans frontière d'import vérifiable. Coût de ne pas migrer : chaque évolution (membres/invitations, billing) ajoute par domaine du câblage manuel et un risque supplémentaire de régression d'isolement.

### Objectifs

- Testabilité : DI explicite, fakes de repository (`service_test.go` purs par module).
- Layout par domaine : `internal/modules/<domaine>/` — frontières d'import, onboarding, extensibilité RBAC/members.
- Contrat HTTP inchangé au byte près pendant toute la migration (sauf le carve-out acté de la décision 7 — chemin « rejeu pendant grâce » du soft-consume) ; `go build`, `go vet`, `make test` verts à chaque phase.

### Non-objectifs

- Montées de versions (go-redis v7→v9, jwt/v4→v5) : follow-up dédié post-Phase 8.
- RBAC câblé : la migration livre le point d'extension (module `auth`), pas les routes membres/invitations.
- CORS configurable : uniquement si une origine croisée devient réelle.
- Refresh httpOnly : backlog explicite, breaking pour le contrat token-in-JSON actuel.
- Homogénéisation de la casse des clés de réponse (`"Message"` vs `"message"`) : évolution d'API post-migration.
- Tout changement de comportement HTTP.

### Décisions structurantes actées (validation utilisateur)

> Provenance : validations recueillies interactivement pendant la rédaction (cycles egm-brainstorm du plan original) ; arbitrages A2/B10 des reviews antérieures : `goldfish_review_report.md` (racine de `.kilo/plans/`, Status: resolved) ; directives architecturales du 2026-11-25 : `1795597264251-design-review/user_directives_2026-11-25.md`.

1. Code sqlc généré : `internal/db` → `internal/platform/db` en **Phase 1** (sqlc.yaml, `.mockery.yaml`, vérif CI `git diff --exit-code` mis à jour dans le même commit).
2. `main.go` → `cmd/api/main.go` en **Phase 1** (Makefile `run`/`e2e`, `swag init -g`, build CI, `app/playwright.config.ts:51` webServer, `main_test.go` suivent dans le même commit).
3. Middlewares `TokenAuth()`/`AdminAuth()`/`RequirePermission` : dans le **module `auth`** (politique de domaine). La plateforme n'importe jamais les modules (aucun cycle).
4. Découpage identité : **4 modules** — `auth` (sessions/JWT, unified login, refresh, logouts), `user` (identité, register, login user), `tenant` (slug, provisioning), `admin` (backoffice, audit). Le module `auth` définit les ports `UserAuthenticator`/`AdminAuthenticator`, implémentés par `user` et `admin` au câblage ; `user`/`admin` consomment le `TokenService` d'`auth` **via une interface définie côté consommateur** (auth fournit l'implémentation, injectée dans `internal/app`) — aucune interface cross-module définie côté fournisseur (décision étendue aux services techniques, 2026-11-25). Pas de cycle.
5. Suppression de la duplication des routes : paquet **`internal/app`** (composition root : `NewRouter`), importé par `cmd/api` **et** par les tests d'intégration. `SetupRouter` (copie manuelle des routes dans `tests/article_test.go:34-80`) disparaît : la suite HTTP épingle le vrai câblage.
6. **Interdiction absolue (2026-11-25)** : `*pgdb.Queries` ne franchit jamais une frontière de module (ni en paramètre de port, ni en retour). Les échanges in-Tx inter-modules passent par un **Unit of Work explicite** (`pgdb.WithinTx`) et l'interface plateforme étroite `pgdb.InTx` ; chaque module expose ses repositories sous **interfaces explicites** (`ArticleRepository`, `UserRepository`, `TenantRepository`, `TokenRepository`, …) définies côté consommateur (règle §4).
7. **Soft-consume strict du refresh token (2026-11-25)** : la consommation ne supprime plus immédiatement — le record de session reste la **chaîne STRING existante**, adressée par l'UUID du refresh JWT (`Set(RefreshUUID, value)` — le token lui-même n'est jamais stocké, en clair ni haché ; `user_sessions:{userID}` reste l'index SET de révocation globale, pas le record) ; l'état est encodé dans la valeur (extension du format actuel « userID:tenantID:role », jamais l'inverse) : actif `userID:tenantID:role` (format inchangé — **aucune migration de schéma**) → consommé `consumed:<delivered|recovered>:userID:tenantID:role` avec TTL réduit à la grâce **fixée à 5 min** (min(grâce, TTL restant)) au lieu de DEL immédiat ; atomicité Lua inchangée (une seule opération atomique par consommation), ownership inchangé (préfixe userID de la valeur comparé avant toute consommation — `TestRefreshOnlyConsumesMatchingOwner` reste le filet). **Politique de rejeu (tranchée ici)** : (i) rejeu d'un jeton déjà délivré (`delivered=1`) → `401`, même forme de réponse que le chemin invalide actuel ; (ii) rejeu d'un jeton consommé non délivré (`delivered=0` : l'application a échoué avant émission de la nouvelle paire) pendant la grâce → **une seule** re-émission de récupération (même forme 200 que le happy path ; état `recovered` encodé dans la valeur du record, jamais dans la réponse HTTP), sinon `401` ; (iii) crash serveur entre consommation et lecture → échec fermé (`delivered` inconnu = non récupérable). **Analyse de menace** : un token volé rejoué pendant la grâce n'obtient une paire que si la première tentative légitime n'a rien délivré ; après délivrance légitime, rejeu → 401 — grâce fixe de 5 min pour borner l'exposition. **Cleanup** : l'éviction par TTL est primaire et suffisante (chaque record consommé expire au bout de sa grâce) ; le sweep in-process du module `auth` (ticker ~5 min) n'est jamais la dépendance de purge : il compte et journalise les records consommés (observabilité) et purge défensivement les reliques (ex. TTL non appliqué après incident Redis) ; le coût lifecycle du ticker est assumé (start/stop lié au service `auth`, testable). L'invariant de course reste : le sweep ne touche jamais un record dont la grâce est active ni un record actif. **Rollout/rollback** : les records actifs conservent le format exact actuel, donc un binaire pré-Phase-3 les lit sans changement ; seuls les records consommés (fenêtre ≤ 5 min) portent le format étendu — un rollback au binaire pré-Phase-3 n'affecte que les consommations en grâce (cas limite : re-login équivalent à une consommation normale). L'incident de session reste observable côté support (journalisé au marquage et à la récupération). Tests dédiés exigés : rejeu / grâce / recovery / fails-closed / sweep / unicité de la recovery. Tâche planifiée en Phase 3 (module `auth`), carve-out de contrat associé cf. [1795597264251-architecture-modulaire-par-domaine-plan-phases.md](1795597264251-architecture-modulaire-par-domaine-plan-phases.md), section « Risques transverses et rollback ».

### Alternatives considérées (rejetées)

1. **Gros mouvement unique vs migration incrémentale par phases** — rejeté : le contrat HTTP doit rester inchangé à chaque étape ; un remaniement global interdit le revert étape par étape et la review. Coût du choix incrémental : discipline de commits/branches et 9 phases à exécuter.
2. **Réorganisation légère des paquets racine (DI partielle, maintien de `controllers/`/`models/`) vs modules par domaine** — rejeté : n'apporte ni frontière d'import ni layout par domaine ; le wiring global de `main.go` subsisterait. Coût du choix modulaire : déplacements complets + recâblage au lieu d'une injection ponctuelle.
3. **Hexagonal complet (ports/repositories pour chaque domaine) vs service simple sur les requêtes sqlc** — rejeté (hexagonal complet) : un repository à implémentation unique n'apporte comme bénéfice mesurable qu'un fake de test. Retenu : interfaces définies côté consommateur, uniquement là où un fake est utile (`ArticleRepository`, `TenantProvisioner`, ports `auth`). Coût assumé : gabarit parfois mince (`dto.go`/`model.go`), accepté pour l'uniformité.
4. **Middlewares dans `internal/platform` vs module `auth`** — rejeté (platform) : `TokenAuth`/`AdminAuth`/`RequirePermission` portent une politique de domaine (claims, RBAC) ; la plateforme n'importe jamais les modules et ne doit pas dupliquer leur contrat de contexte. Coût : les consommateurs racine transitent par `internal/app` pendant la fenêtre transitoire (Phases 3-6).

---

## 1. Cartographie de l'existant

### Arborescence et responsabilités

| Chemin | Rôle actuel | Observations |
|---|---|---|
| `main.go` | Entrypoint + câblage global : dotenv, CLI `-create-admin` (`runCreateAdmin`, main.go:56), `ValidateSigningSecrets`, init des 2 pools pgx (main.go:137-142), `InitRedis` (main.go:144), déclaration **de toutes les routes** `/v1` (main.go:148-182), Swagger hors prod, SPA fallback (`serveSPAIndex`, `NoRoute`), statiques `/public`+`/assets`, run SSL | ~230 lignes, tout le wiring est global (aucune injection) |
| `main_test.go` | Test CLI `-create-admin` (tag `all`) | Sera déplacé vers `cmd/api/main_test.go` |
| `controllers/` | `auth.go` (login unifié + **Refresh portant toute la logique** : parse JWT, claims, `ConsumeRefreshAuth`, relecture user, ré-émission — lignes 84-162), `user.go` (login/register/logout), `article.go` (CRUD simple, sans logique supplémentaire), `admin.go` (login/tenants/status/logout) + `main_test.go` (TestMain REDIS_DB=2/DB_NAME=artisan_test), `auth_test.go` (propriétaire refresh) | Handlers = structs vides, dépendances = **vars de package** (`authModel` controllers/auth.go:19, `userModel` user.go:16, `articleModel` article.go:18, `adminModel` admin.go:15) |
| `models/` | `auth.go` (AuthModel : JWT + sessions Redis + script Lua refresh), `user.go` (Login, Register tenant+owner en une Tx, OneWithTenant), `login.go` (UnifiedLogin + bcrypt factice à coût constant), `tenant.go` (Slugify/ValidSlug + TenantModel), `admin.go` (AdminModel), `article.go` (ArticleModel CRUD + mapping réponse) + `tenant_test.go` (pur) | État global : `authModel` models/user.go:47, `tenantModel` tenant.go:99 |
| `forms/` | `validator.go` (DefaultValidator branché sur `binding.Validator`), `errors.go` (Translate/ValidationMessages), forms par domaine (`auth.go`, `user.go`, `article.go`, `admin.go`) | Partie générique → platform ; partie par domaine → dto de module |
| `middleware/` | `cors.go`, `request_id.go` (purs Gin) ; `auth.go` (TokenAuth/AdminAuth/RequirePermission — couplés à `models.AuthModel` via var, middleware/auth.go:10) | Split platform / module auth |
| `config/` | `secrets.go` (ValidateSigningSecrets) | → `internal/platform/config` |
| `db/` (racine) | `db.go` : client Redis global (`RedisClient`, `InitRedis`, `GetRedis`, `RedisDBIndex`) ; + `db/migrations/` (goose 000001-000003), `db/queries/` (sources sqlc : articles, users, tenants, admins) | `db/` restera sources SQL uniquement après migration |
| `internal/db/` | **Code sqlc GÉNÉRÉ** (`models.go`, `querier.go`, `*.sql.go`, `db.go`) + **`pool.go` manuscrit** (Store, Connect, InTx, `WithTenant` RLS via `set_config(...,true)`, postgresDSN, `InitTenant`/`InitAdmin`, singletons `tenantStore`/`adminStore` pool.go:92-121) + `pool_test.go` | Édition manuelle interdite (sauf pool.go) ; vérif CI `git diff --exit-code internal/db` |
| `internal/mocks/` | Mocks mockery commités (`DBTX`, `Querier`), config `.mockery.yaml` (dir `{{.InterfaceDir}}/../mocks`) | Suit le déplacement du paquet généré |
| `internal/testutil/` | `factory.go`, `fixtures.go` (testfixtures via pool admin), `postgres.go` (testcontainers opt-in) — importés uniquement par `*_test.go` (garde CI) | Inchangé, imports à renommer |
| `cmd/e2edb/` | CLI infra : drop/create base jetable + FLUSHDB index Redis (garde-fous anti-DROP) ; importe `db` (RedisDBIndex) | Inchangé (hors périmètre), import à renommer |
| `invoice/` | Bibliothèque PDF maroto, **sans route** (les routes démo ont été supprimées, cf. `.kilo/plans/1795284300000-remove-public-access.md`) | Domaine futur « billing » |
| `tests/` | Suite d'intégration (tag `all`) : `article_test.go` (940 l., cycle HTTP complet + `TestCrossTenantIsolation`:501 + `TestRLSFailClosed`:596, rôles, suspend/réactivation) + `login_test.go` (matrice unified login) + `main_test.go` (TestMain pin) ; **`SetupRouter` duplique les routes de main.go** | Le vrai câblage main() n'est couvert que par le smoke CI + Playwright |
| `docs/` | Swagger généré (`swag init`), commité | Régénérer à chaque déplacement d'annotations |
| `public/html/` | `index.html` (fallback SPA/démo), `404.html` | Inchangé |
| `app/` | Frontend Vite/React/TS ; `app/playwright.config.ts:51` webServer `go run .` (cwd racine, port 9100, REDIS_DB=3) | Commande à mettre à jour en Phase 1 |
| Racine | `gin-boilerplate` (binaire), `e2edb.exe`, `tmp_line.txt` : artefacts racine à supprimer (binaire de build, binaire Windows) ; `docs/test.md` supprimé (non commité, préexistant) | Nettoyage Phase 8 |

### Invariants à préserver pendant TOUTE la migration

- RLS échec fermé : contexte tenant réglé **uniquement** par `WithTenant` (transaction, `set_config(...,true)`) — jamais de `SET` session-wide.
- Pool admin (`app_admin BYPASSRLS`) pour plateforme (login, register, refresh, backoffice) ; pool tenant (`app_user`) + `WithTenant` pour le métier.
- Login unifié : coût bcrypt constant (hash factice `models/login.go:30`), pas d'énumération, politique « un email superadmin ne recouvre jamais un email user ».
- Réponses JSON : formes `gin.H{"message": ...}` et codes actuels **au byte près** (les tests les épinglent) — casse mixte `"Message"`/`"message"` dans controllers/article.go (p.ex. `"Message": "Article not found"` article.go:99) : conserver telle quelle lors du déplacement, homogénéiser seulement si les tests sont mis à jour dans la même phase.
- Env test : `TestMain` épinglant `REDIS_DB=2` + `DB_NAME=artisan_test` requis dans tout nouveau paquet de test d'intégration ; secrets de test via `t.Setenv` uniquement.
- SQL exclusivement dans `db/queries/*.sql` → sqlc ; jamais de SQL concaténé en Go.

## 2. Inventaire des domaines métier

| Domaine | Sources actuelles | Tables | Routes |
|---|---|---|---|
| **article** (pilote) | `controllers/article.go`, `models/article.go`, `forms/article.go`, `db/queries/articles.sql` | `article` | POST `/v1/article`, GET `/v1/articles`, GET/PUT/DELETE `/v1/article/:id` |
| **auth** (sessions) | `models/auth.go`, `controllers/auth.go` (Refresh), `middleware/auth.go`, `forms/auth.go`, sessions Redis (`user_sessions:*`, `admin:*`) | — (Redis) | POST `/v1/login`, POST `/v1/token/refresh`, GET `/v1/user/logout`, GET `/v1/admin/logout` |
| **user** (identité) | `models/user.go` (Login/Register), `controllers/user.go` (Login/Register), `forms/user.go`, `db/queries/users.sql` | `user` | POST `/v1/user/login`, POST `/v1/user/register` |
| **tenant** | `models/tenant.go` (Slugify/ValidSlug/Create/All/SetStatus), `db/queries/tenants.sql` | `tenant` | — (via register + admin) |
| **admin** (backoffice) | `models/admin.go`, `controllers/admin.go`, `forms/admin.go`, `db/queries/admins.sql` | `admin_user`, `admin_audit` | POST `/v1/admin/login`, GET `/v1/admin/logout`, GET `/v1/admin/tenants`, POST `/v1/admin/tenants/:id/status` |
| **invoice** (futur billing) | `invoice/` (lib sans HTTP ni DB) | — | — |

### Découpage « super-admin » vs « tenant » — état réel vérifié dans le code

La cible (API REST avec un pôle super-admin gérant les tenants + un pôle tenant avec des niveaux de droits) est déjà structurée dans l'existant, et la migration la préserve explicitement :

| Axe | Pôle super-admin (plateforme) | Pôle tenant (métier) |
|---|---|---|
| Routes | `/v1/admin/*` → module `admin` (Phase 6) | `/v1/article*`, `/v1/user/*` → modules métier |
| Pool PostgreSQL | `app_admin` BYPASSRLS (`AdminStore()`) | `app_user` + RLS (`TenantStore()` + `WithTenant`) |
| Tables | `tenant`, `admin_user`, `admin_audit` — **sans RLS** (plateforme) | `article`, `user` — RLS échec fermé sur `tenant_id` |
| Sessions | JWT secret `ADMIN_ACCESS_SECRET`, Redis préfixe `admin:`, 8h, sans refresh | JWT `ACCESS_SECRET`/`REFRESH_SECRET`, Redis + index `user_sessions:*`, 15min/7j |
| Isolation | Prouvée par les tests croisés (`TestAdminTokenOnTenantRoute`:807, `TestTokenAuthRejectsAdminToken`:104) | Prouvée par `TestCrossTenantIsolation`:501, `TestRLSFailClosed`:596 |

**Niveaux de droits — état réel : RBAC minimal non câblé** — le socle existe et est testé, mais aucune route ne l'utilise :
- `RequirePermission` + `PermManageMembers`/`PermManageTenant` existent (`middleware/auth.go:14-32`, owner = 2 permissions, member = 0) et sont testés (`middleware/auth_test.go:161`), mais **aucune route n'y est câblée** (grep : usages uniquement dans le middleware et son test) ;
- le rôle vit uniquement dans le claim JWT signé, **re-fraîchi depuis la base à chaque refresh** (jamais lu du client — `models/user.go:149` `OneWithTenant`) — invariant préservé Phase 5 ;
- la base garantit **un seul owner par tenant** (index unique partiel, épinglé par `TestOneOwnerConstraint`:659) ; aucun compte `member` n'est créable aujourd'hui (pas d'invitation, `CreateUserOwner` ne crée que des owners) ;
- le cloisonnement utilisateur-à-utilisateur dans un tenant est un cloisonnement **par propriétaire de ligne** (les requêtes articles filtrent `user_id = @user_id`, `db/queries/articles.sql:11,19`), pas par rôle.

**Point d'extension après migration** : le module `auth` porte le middleware `RequirePermission` et la map rôle→permissions (Phase 3) ; les futures invitations/membres (listées en backlog `memory-bank/progress.md:56`) se borneront à câbler `RequirePermission(...)` dans `internal/app` et à créer les comptes `member` côté module `user`. La migration ne fige pas le modèle de rôles : la map et le middleware restent l'unique point de politique RBAC.

## 3. Couplages, points de friction et zones de vigilance

1. **État global partout (pas de DI)** : mêmes singletons (`pgdb.TenantStore()`/`AdminStore()`, `db.GetRedis()`) et vars de package `*Model` identiques à ceux cartographiés en §1. Aucun constructeur n'injecte une dépendance → tests peu isolables (seuls `DBTX`/`Querier` sont mockables aujourd'hui).
2. **Logique métier dans les handlers Gin** : `controllers/auth.go:84-162` (Refresh : parse JWT + claims + consommation + relecture + ré-émission), `controllers/user.go:113-126` (Logout), `controllers/admin.go:28-54` (Login : création de tokens dans le handler). `models/` n'est pas plus « métier » que les controllers : la logique (parse, consommation, ré-émission) vit dans le handler, lignes citées à l'appui — la découpe actuelle ne déplace pas la logique, seulement les fichiers.
3. **Login unifié transverse** : `models/login.go:80` orchestre `UserModel`+`AdminModel`+`AuthModel` (3 domaines) + compteur atomique de comparaisons factices observé par les tests (`DummyHashComparisons`, models/login.go:38).
4. **Transaction inter-domaines** : le register crée tenant + owner dans UNE plateforme Tx (`models/user.go:111-130` appelle `createTenantInTx` de `models/tenant.go:120`). Cible (décision 6) : le service d'orchestration (user) pilote un **Unit of Work explicite** (`pgdb.WithinTx` — wrapper de `Store.InTx` + `WithTenant`) ; les deux repos exposent leurs opérations « in-tx » via des **interfaces explicites** (`TenantProvisioner.CreateInTx`, user `CreateOwnerTx`) portant l'interface plateforme étroite `pgdb.InTx` — `*pgdb.Queries` ne traverse aucune frontière de module.
5. **Pas de mapping d'erreurs central** : codes hardcodés par handler ; conventions AGENTS.md (pgx `ErrNoRows`→404, `23505`→409, `23503`→409, `42501`→500) non centralisées. `controllers/user.go:94` renvoie `err.Error()` brut au client (messages models — pas de fuite SQL, mais format non maîtrisé).
6. **Duplication du câblage de routes** : `SetupRouter` (`tests/article_test.go:34-80`) copie main.go à la main → le vrai câblage HTTP n'est épinglé que par le smoke CI et Playwright. **Résolu en Phase 1** par `internal/app`.
7. **Dérive CI préexistante (à corriger en Phase 0)** : `.github/workflows/ci.yml:139,157,166` sonde encore `/v1/invoice` + `/v1/invoice/download` (attendus 200/PDF) alors que ces routes ont été **supprimées** de `main.go` (NoRoute renvoie 404 sur `/v1*`) → le smoke step entre en conflit avec `main.go` actuel — échec au prochain run. Le plan de suppression initial (`.kilo/plans/1795284300000-remove-public-access.md`) n'a pas mis à jour `ci.yml`.
8. **Swagger dépend de l'emplacement des annotations** : `swag init` part de `main.go` ; chaque déplacement de controller/handler doit repasser `make generate_docs` et committer `docs/` dans la même phase.
9. **Contrat SPA single-origin — préservé (recopié à l'identique)** : le plan recopie `main.go:127-212` dans `internal/app` (fallback SPA `NoRoute` : 404 JSON sur `/v1*`/non-GET uniquement, `serveSPAIndex` no-cache, `/assets` `immutable`, `/public` statique, Swagger hors prod). Le smoke CI (gzip sur `/`, fragments SPA) et le cycle Playwright (`app/playwright.config.ts` webServer, origine unique `app/dist`, port 9100) épinglent ce contrat ; `make e2e` est rejoué aux Phases 1 et 7 — les deux seules phases qui changent l'entrypoint (`main.go` → `cmd/api`) et l'assemblage du routeur, donc la couche que seul le cycle Playwright observe (processus réel + SPA servis) ; toute régression de câblage des autres phases (middleware rebranchés en Phases 2-3 inclus) se manifeste comme une violation du contrat HTTP, couverte de bout en bout par la suite `tests/` qui pilote `app.NewRouter` (vrais handlers, middlewares, réponses) et par le smoke CI à chaque branche. **Trois réserves, à traiter après la migration** : (a) `Access-Control-Allow-Origin` est codé en dur `http://localhost` (`middleware/cors.go:10`) — sans effet en single-origin, mais à rendre configurable (`internal/platform/config`) dès qu'une origine croisée existera ; (b) les tokens voyagent dans le corps JSON (SPA les stocke côté client) — le passage « refresh httpOnly » est un backlog explicite hors périmètre de la migration (breaking pour le contrat actuel) ; (c) l'affectation globale `binding.Validator` (main.go:130) est recopiée à l'identique dans `NewRouter` — dette **remontée en tâche de backlog explicite** (relocalisation hors de `NewRouter` vers l'entrypoint ou `internal/platform/validate`, élimination requise en sortie de migration, Phase 8 — [plan-phases](1795597264251-architecture-modulaire-par-domaine-plan-phases.md), Phase 1 étape P1.7 « Dettes consignées »).
10. **Couverture actuelle (filet existant)** — couvre les invariants listés. Filet complet : suite HTTP `tests/` (cycle register/login/CRUD/isolation cross-tenant/RLS fail-closed/suspend/admin), matrice unified login (`tests/login_test.go`), propriété refresh (`controllers/auth_test.go`), 7 tests middleware (`middleware/auth_test.go`), CLI bootstrap (`main_test.go`), garde-fous e2edb, unitaires purs (`models/tenant_test.go`, `config/secrets_test.go`, `db/db_test.go`, `internal/db/pool_test.go`, `internal/testutil/*`). Trois statuts : le câblage réel main.go n'est pas couvert hors smoke/e2e (Phase 0/1 y répond) ; les formes de réponses exactes sont épinglées par les tests (à préserver au byte près) ; la consommation du refresh (`models/auth.go:262`) est atomique et ne supprime la clé que si elle existe **et** appartient au demandeur — un utilisateur ne peut pas consommer le refresh token d'un autre, propriété épinglée par `TestRefreshOnlyConsumesMatchingOwner`. Ce mécanisme évolue en **soft-consume strict** (décision 7 : état encodé dans la valeur STRING du record, adressé par l'UUID du refresh JWT — le token n'est jamais stocké ; grâce 5 min, sweep d'observabilité et de purge défensive, politique de rejeu tranchée dans le design, carve-out « rejeu pendant grâce » sous tests dédiés) — ownership et réponses happy-path/rejeu hors grâce inchangés au byte près.

## 4. Layout cible

```
cmd/api/main.go                  # dotenv, flags, secrets, pools, redis, app.NewRouter, run
cmd/e2edb/                       # inchangé (CLI infra)
internal/app/router.go           # NewRouter(deps) : table de routes + SPA/static/swagger/NoRoute
internal/platform/config/        # ← config/secrets.go
internal/platform/redisdb/       # ← db/db.go (client Redis, RedisDBIndex)
internal/platform/db/            # ← internal/db (sqlc généré + pool.go ; package db, alias pgdb)
internal/platform/middleware/    # ← middleware/{cors,request_id}.go (purs)
internal/platform/validate/      # ← forms/{validator,errors}.go (DefaultValidator, Translate)
internal/platform/httpx/         # responders, mapping pgx→HTTP, ParseUUIDParam (+ tests purs ; créé en Phase 2)
internal/platform/mocks/         # ← internal/mocks
internal/testutil/               # inchangé
internal/modules/article/        # handler.go, service.go, repository.go, dto.go, model.go, *_test.go
internal/modules/auth/           # sessions/JWT + TokenAuth/AdminAuth/RequirePermission + unified login/refresh + ports ; documente les clés de contexte (userID/tenantID/role)
internal/modules/tenant/         # slug + provisioning (+ backoffice data)
internal/modules/user/           # identité, register (Tx tenant+owner), login, implémente UserAuthenticator
internal/modules/admin/          # backoffice, audit, implémente AdminAuthenticator
internal/modules/invoice/        # ← invoice/ (Phase 8)
db/migrations/, db/queries/      # inchangés (sources SQL)
```

Règle de dépendance : `modules → platform` directement ; `modules → modules` EXCLUSIVEMENT via **interfaces définies côté consommateur** — la règle s'applique aux deux familles et `internal/app` les distingue au câblage : **ports d'orchestration** (`auth` : `UserAuthenticator`/`AdminAuthenticator`/`UserReader` ; `user` : son port de provisioning tenant) et **services techniques** (le port `TokenService` déclaré par `user`/`admin`, implémentation `*auth.Service` injectée ; repos interfaces par module). Aucune interface cross-module définie côté fournisseur (décision 4 étendue, 2026-11-25). `platform` n'importe jamais `modules` (invariant absolu, toute la migration) ; `*pgdb.Queries` ne franchit aucune frontière de module (décision 6 : Unit of Work `pgdb.WithinTx` + interface étroite `pgdb.InTx`). **Fenêtre transitoire Phases 3-6 — décision A2** (« A2 » est l'identifiant de l'arbitrage Blocker porté par `goldfish_review_report.md`, racine de `.kilo/plans/` — review antérieure du plan original, tranché le 2026-11-25 en gate interactif, Option 1 « Tolérance transitoire explicite ») : une interdiction immédiate des imports croisés aurait imposé de déplacer chaque paquet et chacun de ses importeurs dans un seul mouvement ; les répertoires producteurs et consommateurs migrent de façon décalée. Impact : les imports racine↔modules sont tolérés **dans les deux sens**, uniquement pour le code en cours de migration ; la fenêtre est refermée en Phase 7 (suppression des répertoires racine) et l'interdiction vérifiée par le critère de sortie (décompte `git ls-files` = 0). `internal/app` et `cmd/api` sont les seuls endroits qui importent tout à terme.

## 5. Risques et atténuations

- **Carve-out de contrat (décision 7)** : le seul chemin déviant de la migration est le « rejeu pendant grâce » du soft-consume ; sa politique est tranchée dans le design (rejeu après délivrance → 401, recovery unique si non délivré, échec fermé au crash) et l'ownership reste épinglé (`TestRefreshOnlyConsumesMatchingOwner` déplacé avec le module) — tests dédiés nouveaux exigés (rejeu/grâce/recovery/fails-closed/sweep/unicité de la recovery).
- **Fenêtre A2** : imports croisés transitoires (Phases 3-6) — atténués par l'invariant absolu « la plateforme n'importe jamais les modules » (toute la fenêtre) et refermés vérifiablement en Phase 7 : décompte `git ls-files` = 0 (critère de sortie).
- **Scalabilité — statu quo présumé et préservé** : la migration ne modifie ni les deux pools PostgreSQL (~`app_admin` BYPASSRLS, ~`app_user`), ni les TTL Redis des sessions actives, ni le profil de croissance multi-tenant ; le soft-consume borne les records consommés dans le temps (grâce 5 min, éviction par TTL primaire + sweep d'observabilité et de purge défensive) ; le Unit of Work réutilise la mécanique Tx existante (`Store.InTx` + `WithTenant`) sans introduire de transaction supplémentaire ni de nouvelle contention.
- **Détection de régression** : test contrat P0.3 (surface API bidirectionnelle), suite `tests/` inchangée (contrat de bout en bout via `app.NewRouter`), smoke CI à chaque branche, `make e2e` aux Phases 1 et 7 (argument explicite §3.9).
- **Exécution et rollback** : reportés à la rubrique [« Risques transverses et rollback » du plan-phases](1795597264251-architecture-modulaire-par-domaine-plan-phases.md) (branchement par phase, revert au dernier commit vert, jamais de `make test`/`make e2e` en parallèle).

## 6. Décisions et reports

> Arbitrages tranchés et reports hors migration ; l'inventaire complet des non-objectifs reste au Préambule (« Non-objectifs »).

- Montée de go-redis v7 → v9 et jwt/v4 → v5 : **jamais pendant la migration** — follow-up dédié après Phase 8.
- Perte de session après échec de relecture post-consommation du refresh : **tranché 2026-11-25** (soft-consume strict, décision 7) — la politique de rejeu est désormais fixée dans le design (délivrance → 401, recovery unique si non délivré, fails-closed) : **aucun reliquat ouvert restant**.
- Reports et dettes hors migration : casse des clés de réponse (`"Message"`/`"message"`), CORS configurable, refresh httpOnly, câblage RBAC — détaillés dans « Non-objectifs » ; **normalisation des payloads d'erreur client** (`err.Error()` brut, `controllers/user.go:94` → répondants `httpx` + erreurs de domaine typées, un changement coordonné avec la SPA, post-Phase 8 — backlog issu du Stage 5, Persona 2 F4).
- Déclinaisons consignées (closent le registre des reviews antérieures) : extraction immédiate de l'origine CORS (déclinée au gate Stage 5 — reportée jusqu'à un cas croisé réel) ; gabarits `dto.go`/`model.go` parfois minces (coût assumé, alternative 3).
