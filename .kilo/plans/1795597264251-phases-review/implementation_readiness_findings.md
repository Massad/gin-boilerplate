# Rapport Implementation Readiness Goldfish — `1795597264251-architecture-modulaire-par-domaine-plan-phases.md`

**Mode** : REVIEW (DetailedRange = l.6-168, en-tête `## Detailed Implementation Plan` l.6 ; l.1-5 = contexte de scission) · **Persona** : `egm-implementation-readiness` · **GoldfishContext** : design `1795597264251-architecture-modulaire-par-domaine-design.md` (décisions actées, invariants) · **Contrat d'exécution** : `1795627363688-lots-impl-migration-modulaire.md` (points d'audit G-1) — tous les constats code sont vérifiés dans le dépôt réel à la date du 2026-11-26.

## Synthèse

Le plan est exécutable dans sa quasi-totalité : le sweep d'ancrage complet (P0.x-P1.x et Phases 2-8) vérifie **tous** les ancrages file:line sauf un, les étapes P0.x-P1.x sont zéro-décision, et les critères de sortie sont mesurables. Deux exigences zéro-décision commandées par le contrat ne sont **pas** portées par le plan-phases :

1. **`pgdb.InTx` / `pgdb.WithinTx`** (livrable plateforme L4, consommateur L5) : les deux types n'existent pas dans `internal/db/pool.go` (seuls `Store.InTx` et `Store.WithTenant` — vérifié `pool.go:42-68`) et le plan se contente de les **nommer**, sans signature ni mécanisme de liaison queries↔tx. L'implémentateur devra designer l'interface lui-même → **Blocker**.
2. **Acquisition du client Redis par le module `auth`** (Phase 3) : P1.2 conserve les exports `GetRedis`/`InitRedis`, la cible prohibe les singletons, et le plan ne tranche pas. Le texte prêt existe (C8 de l'audit `1795684415717`) mais n'est pas dans le plan-phases → **Major**.

Trois other Majors : item Phase 8 périmé (`docs/test.md` — commit `581d770` vérifié), et le `TestMain` du paquet `middleware` (`middleware/main_test.go`, pin `REDIS_DB=2`) que la Phase 3 ne mentionne pas alors que sa logique « dossier supprimé si vide » se heurte à ce fichier. Les Nits couvrent le slogan de couverture (l.164), les ellipses de signature Phase 2 (l.19) et la forme de réponse de la recovery non épinglée (l.26).

**Points d'audit contractuels — verdicts** :
1. Spéc zéro-décision `pgdb.InTx`/`WithinTx` → **FINDING (Blocker)**, F1.
2. Sweep d'ancrage complet → **CONFIRMÉ** (1 seul drift trouvé : `docs/test.md`, F3).
3. Fraîcheur Phase 8 → **FINDING (Major)**, F3 (`docs/test.md` périmé ; artefacts racine `gin-boilerplate`/`e2edb.exe`/`tmp_line.txt` tous présents — vérifié `ls`).
4. Cohérence puce P3 soft-consume ↔ design décision 7 → **CONFIRMÉ** (détail bénin : F7, Nit).
5. Opérationnalisation « aucune phase ne baisse la couverture » → **FINDING (Nit)**, F5.
6. Spéc zéro-décision client Redis auth → **FINDING (Major)**, F2 (seul `models/auth.go` consomme `GetRedis()` hors tests — vérifié par grep *.go).
7. Propriétaire unique Redis → **CONFIRMÉ** (cohérent tel qu'écrit : repo sessions centralisé l.25, rien d'autre ne touche Redis en prod, règle exécutable via gate item 8 du lots plan l.73 + texte C8 qui la remonte dans la puce Risques Phase 3).

**Compte** : Blocker 1 · Major 3 · Nit 3.

---

## Findings

### F1 — `pgdb.InTx` / `pgdb.WithinTx` : nommés mais jamais spécifiés (audit G-1 point 1)
- **Sévérité** : Blocker
- **Périmètre** : Phase 4 « Interfaces » (l.35 : `CreateInTx(ctx, tx pgdb.InTx, name) (Tenant, error)`), Phase 5 « Interfaces » (l.41 : « register : `pgdb.WithinTx` pilote la Tx »), Phase 6 (l.46 « audit dans la même Tx ») ; evidence code : `internal/db/pool.go:42-68` — aujourd'hui `Store.InTx(fn func(tx pgx.Tx, q *Queries) error)` et `WithTenant(ctx, tenantID, fn func(q *Queries) error)` ; **aucun type `pgdb.InTx`, aucune fonction `pgdb.WithinTx`** dans le dépôt (grep vérifié).
- **Justification** : le contrat (lots plan l.21) exige une spécification « zéro-décision » précisant « comment les repos concrets lient leurs requêtes sqlc au tx sans exposer `*pgdb.Queries` ». Le plan-phases ne répond à **aucune** des questions de conception que l'implémentateur de L4 devra trancher seul : (a) quelle est la définition exacte de `pgdb.InTx` — méthodes portées, ou wrapper exposant un `*pgdb.Queries` lié au tx (méthode utilisée par `pool.go:50` `s.Queries.WithTx(tx))` ? ni paramètre ni retour de port ne peut porter `*pgdb.Queries` (règle absolue), le repository a pourtant besoin de requêtes liées pour appeler `q.CreateTenant`/`q.CreateUserOwner` ; (b) signature exacte de `pgdb.WithinTx` — le design (l.120) le décrit comme « wrapper de Store.InTx + WithTenant », mais `WithTenant` exige un `tenantID` : le register crée le tenant **par et** sans contexte tenant (tables plateforme, pool BYPASSRLS), et `models/admin.go:55-64` pilote la Tx d'audit sans RLS — quel appelant passe quoi ? ; (c) rôle du `tx pgx.Tx` brut dans `CreateInTx(ctx, tx pgdb.InTx, ...)` — est-ce `pgx.Tx` ou le wrapper ? ; (d) comment le repository `user` appelle `CreateUserOwner` sur le **même** InTx que le port `TenantProvisioner.CreateInTx` (Phase 5, l.42 « `CreateUserOwner` »). Exécuter L4 en l'état = laisser l'implémentateur définir une API plateforme structurante, exactement ce que la règle zéro-decision interdit.
- **Proposition concrète** : compléter la Phase 4 (l.35) par une spécification verrouillée, p.ex. :
  > « Livrable plateforme L4 dans `pool.go` manuscrit : `type InTx interface { Queries() *pgdb.Queries }` (la requête `*Queries` liée au tx reste consumable uniquement par le repository du module — jamais paramètre ni retour d'un port cross-module) ; `func (s Store) WithinTx(ctx context.Context, fn func(tx pgdb.InTx) error) error` (wrapper de `Store.InTx` sans contexte tenant — les opérations in-Tx visées ciblent le pool admin sans RLS) ; appel par le service : `store.WithinTx(ctx, func(tx pgdb.InTx) error { t, err := tenantRepo.CreateInTx(ctx, tx, name); ... })` ; le repository binding : `tenantRepo.CreateInTx` exécute `q.CreateTenant(ctx, ...)` sur `tx.Queries()`. »
  Le contenu est la décision à acter (egm-resolve, édition plan-phases) ; l'intitulé ci-dessus est une proposition, pas une alternative laissée ouverte.

### F2 — Acquisition du client Redis par le module `auth` : non tranchée (audit G-1 point 6)
- **Sévérité** : Major
- **Périmètre** : Phase 3 « Sources » (l.25 : repository sessions), « Transition » (l.27 : injection du `*auth.Service`), « Risques » (l.30) ; P1.2 (l.111 : P1.2 conserve les exports `InitRedis`/`GetRedis`/`RedisDBIndex`/`DefaultRedisDB`) ; evidence : seul `models/auth.go` consomme `db.GetRedis()` hors tests (grep *.go — 8 appels prod dans ce seul fichier ; les autres occurrences sont `db/db.go:51` (définition), `db/db_test.go` et `*_test.go`) ; règle cible : `.kilo/rules/go-architecture-modulaire.md` « Interdits »/DI explicite — singletons `GetRedis()` prohibés, anti-patterns hérités prohibés.
- **Justification** : la Phase 3 doit créer `auth/repository.go` qui parle à Redis ; le plan ne dit nulle part comment le repository obtient le client. Deux lectures possibles, aux conséquences opposées : injection par constructeur (alignée cible, recommandée par le contrat) ou poursuite du singleton `redisdb.GetRedis()` pendant la fenêtre. Le texte prêt existe **en dehors du plan-phases** (audit `1795684415717-audit-lots-impl-migration.md`, correctif C8, l.146 : `NewRepository(client *redis.Client)` — résolution au câblage dans `internal/app`, lecture du singleton confinée au composition root, dette tracée post-Phase 8) mais le plan-phases reste muet : premier consommateur = L3, décider sur place = une décision laissée à l'implémentateur.
- **Proposition concrète** : appliquer le texte C8 (réservé à la passe egm-resolve de G-1 — jamais en direct) dans la puce **Risques** de la Phase 3 (l.30) : « **Acquisition du client Redis : `NewRepository(client *redis.Client)` — le client est résolu au câblage dans `internal/app` (lecture du singleton `redisdb.GetRedis()` confinée au composition root pendant la fenêtre transitoire ; `InitRedis` reste à l'entrypoint `cmd/api`), dette tracée pour sortie post-Phase 8** ; seul le module `auth` accède aux clés de session (records UUID, index `user_sessions:*`, préfixe `admin:`) — import de `internal/platform/redisdb` interdit aux autres modules (`cmd/api`/`cmd/e2edb`/tests autorisés) ». Ce texte couvre aussi l'audit point 7 (propriétaire unique).

### F3 — Item Phase 8 périmé : `docs/test.md` a déjà été supprimé par le commit `581d770` (audit G-1 point 3)
- **Sévérité** : Major
- **Périmètre** : Phase 8 (l.59 : « traiter la suppression non commitée `docs/test.md` ») ; même affirmation périmée dans le design (l.73 : « `docs/test.md` supprimé (non commité, préexistant) ») et dans le lots plan (l.19, qui avait signalé « à faire confirmer par G-1 et corriger »).
- **Justification** : vérifié `git log` : le commit `581d770` « docs(test): remove test documentation » existe et `git status` est propre sur ce fichier ; `docs/test.md` est **absent** du disque. La « suppression non commitée » n'existe plus : l'étape demande à l'implémentateur de traiter un état qui n'a jamais lieu, contradiction entre plan et dépôt — pour un document source de vérité opérationnelle, tout énoncé factuel faux sur l'état initial décrédibilise la certification (le critère de sortie G-1 « éditions appliquées et revérifiées » exige la correction). À l'inverse, les trois artefacts racine visés par la même étape sont tous **toujours présents** (vérifié : `gin-boilerplate`, `e2edb.exe`, `tmp_line.txt`) — seule cette sous-partie est périmée.
- **Proposition concrète** : l.59 → « supprimer les artefacts racine (`gin-boilerplate`, `e2edb.exe`, `tmp_line.txt`) » et supprimer la mention `docs/test.md` (ou : « `docs/test.md` : déjà réglé par le commit `581d770` — vérifié 2026-11-26, aucune action »). Reporter la même correction au design l.73.

### F4 — `middleware/main_test.go` (TestMain pin Redis) : absent de la spécification Phase 3 alors qu'il bloque « dossier supprimé si vide »
- **Sévérité** : Major
- **Périmètre** : Phase 3 « Transition » (l.27 : « `middleware/` vidé de `auth.go` … → dossier supprimé si vide, sinon gardé pour les résidus ») et « Validation » (l.31 : « `middleware/auth_test.go` migré vers le module (7 tests) ») ; P1.4 (l.124 : « le répertoire `middleware/` garde `auth.go` + tests ») ; evidence : `middleware/main_test.go:14-19` — un `TestMain` qui pose `REDIS_DB=2` pour tout le binaire de test du paquet.
- **Justification** : la liste de migration de la Phase 3 ne mentionne que `middleware/auth.go` et `middleware/auth_test.go`. Le paquet `middleware` contient aussi `main_test.go` (pin d'environnement), que ni P1.4, ni la Phase 3, ni la Phase 7 (« suppressions de la phase : … `middleware/` … → suppression des répertoires », l.54) ne déplacent. Conséquence double : (a) après migration de `auth_test.go`, `middleware/` contient encore `main_test.go` — « dossier supprimé si vide » est faux, l'implémentateur devra improviser (garder un paquet rémanent sans test, ou déplacer le fichier, deux choix non spécifiés) ; (b) les tests migrés dans `internal/modules/auth` (y compris les tests dédiés soft-consume de l.26, qui touchent Redis) exigent leur propre `TestMain` épinglant `REDIS_DB=2` + `DB_NAME=artisan_test` (invariant déjà formulé en l.165) — le plan n'indique pas que le fichier à migrer est celui-ci. Sans cette mention, un exécutant naïf déplace `auth_test.go` seul : les tests du nouveau paquet tourneraient sans pin (index Redis 1 = sessions dev — violation directe de l'invariant index 1/2/3).
- **Proposition concrète** : l.31 et l.27 → « migration du paquet de test complet : `middleware/auth_test.go` (7 tests) **et `middleware/main_test.go` (TestMain, pin `REDIS_DB=2`) deviennent le socle de test du module `auth`** (le pin porte `REDIS_DB=2` + `DB_NAME=artisan_test`, cf. critères de sortie l.165) ; après cette migration, `middleware/` ne contient plus aucun fichier → suppression du répertoire en Phase 3 (l.27) plutôt qu'en Phase 7 ».

### F5 — « Aucune phase ne baisse la couverture » (l.164) : slogan non actable (audit G-1 point 5)
- **Sévérité** : Nit
- **Périmètre** : « Critères de sortie de chaque phase », charge de tests (l.164) ; redit en « Risques transverses » (l.70 : « chaque phase module embarque une suite `service_test.go` pur (critère de sortie) »).
- **Justification** : le corps de l.164 est actionnable (inventaire précis des tests déplacés et de la suite pur par module, avec cas nommés) ; la phrase finale « Aucune phase ne baisse la couverture » reste en revanche un énoncé invérifiable en tant que tel — le dépôt n'instrumente pas la couverture, et le contrat exige explicitement l'opérationnalisation qualitative (« chaque rapport `L<N>-gate.md` consigne l'inventaire des tests du lot — fonctions déplacées/ajoutées, exécution verte ; aucun seuil chiffré », lots plan l.34/l.75). Le plan-phases ne reprend nulle part cette convention de consignation : un exécutant qui ne lit que ce document pourrait chercher à mesurer un delta chiffré — exigence inventée que le contrat proscrit.
- **Proposition concrète** : l.164, remplacer la phrase : « **Charge de tests consignée** : le rapport de gate du lot (`L<N>-gate.md`) consigne l'inventaire des tests du lot (fonctions déplacées/ajoutées, exécution verte) — exigence qualitative, sans seuil chiffré (le dépôt n'instrumente pas la couverture). »

### F6 — Phase 2 : signatures d'interface elipsées et constructeur « -style » (l.19)
- **Sévérité** : Nit
- **Périmètre** : Phase 2 « Interfaces » (l.19 : `type ArticleRepository interface { Create(ctx, tenantID, userID, form) (uuid, error); One(...); All(...); Update(...); Delete(...) }` et « constructeur `NewHandler(svc *Service) gin.HandlerFunc`-style »).
- **Justification** : le corps exact des méthodes existe dans les fichiers sources (`models/article.go`, `controllers/article.go`) et les types circulent via les déplacements `dto.go`/`model.go` de la même phase, donc la décision reste dérivable ; mais un standard zéro-décision attend des signatures pleines (paramètres et retours typés), et « gin.HandlerFunc-style » laisse ouvert la forme du handler (struct avec méthodes vs fonctions paires).
- **Proposition concrète** : écrire les 5 signatures complètes (types dto/model de la phase) et remplacer « -style » par la forme : « `NewHandler(articles *Service) *Handler` (`Handler` porte les 5 méthodes HTTP routées dans `internal/app`) ».

### F7 — Forme de la réponse de la recovery non épinglée dans le plan (l.26)
- **Sévérité** : Nit
- **Périmètre** : Phase 3, puce soft-consume (l.26, « politique de rejeu ») ; désaccord partiel de granularité avec le design décision 7 (l.39 : « même forme 200 que le happy path »).
- **Justification** : la puce reprend fidèlement le mécanisme, le TTL, le carve-out et les tests ; elle ne répète pas (à juste titre) la totalité de la politique, mais le seul point de contrat HTTP touché par le carve-out — la forme de la réponse de la recovery — devrait être épinglé dans le plan, au même titre que les 401 encadrés (déjà cités deux fois). L'énoncé actuel (« une seule re-émission de récupération… état `recovered` encodé dans la valeur, jamais dans la réponse HTTP ») laisse implicite que la réponse est identique au happy path.
- **Proposition concrète** : l.26, après « re-émission de récupération » : « (réponse de même forme 200 que le happy path — cf. design décision 7 (ii) ; état `recovered` encodé dans la valeur…) ».

---

## Caveats — calibration

| Ligne | Formulation | Verdict | Action |
|---|---|---|---|
| l.26 | « rollback pré-Phase-3 borné aux consommations en grâce (re-login équivalent à une consommation normale) » | promesse d'équivalence non prouvable avant implémentation | **CALIBRATE** (rattrapé en cas d'édition F7/F2) → « re-login requis pour les sessions consommées en grâce » |
| l.97-99 | liste figée des 14 routes P0.3 | vérifiée contre `main.go:153-181` — égalité bidirectionnelle actable | KEEP |
| l.163 | « le test `tests/api_surface_test.go` vert sans éditer la liste attendue » | règle opérateur mesurable, exception bornée (« déplacement mécanique de helpers ») | KEEP |
| l.166 | « Régénération idempotente : sqlc, mocks, swagger → diffs vides » | commandes et cibles nommées (l.91-93, l.155) | KEEP |
| l.167 | « décompte `git ls-files` … **0 en sortie de Phase 7** » | critère chiffré unique et mesurable | KEEP |

---

## Sweep d'ancrage — résultats (audit G-1 point 2)

Tout vérifié contre le code réel, sauf F3. Échantillon exhaustif :

| Ancrage (plan) | Vérifié | Evidence |
|---|---|---|
| `ci.yml:139/157/166-168` (smoke `/v1/invoice`) | ✓ | `.github/workflows/ci.yml:139,157,166-168` |
| « les 14 routes API » de `main.go:148-182` | ✓ | compte exact `main.go:153-181` (14 routes `/v1`) |
| `main.go:16/18/20/21/22` (imports P1.1/P1.2/P1.5) | ✓ | `main.go:16,18,20,21,22` |
| `main.go:127-134` (routeur P1.7), `main.go:130` (forms usage unique P1.5), `main.go:56-77` (`runCreateAdmin`), `main.go:136-146` (init pools/Redis), `main.go:148-212` (routes) | ✓ | `main.go:127-134,130,56-77,136-146,148-212` |
| `models/auth.go:11`, 7 sites import `db` (P1.2) | ✓ | `models/auth.go:11`, `cmd/e2edb/main.go:45`, `tests/article_test.go:20`, `tests/login_test.go:18`, `controllers/auth_test.go:15`, `middleware/auth_test.go:12`, `main.go:18` |
| `sqlc.yaml:9`, `.mockery.yaml:9` + `dir` (P1.3) | ✓ | `sqlc.yaml:9`, `.mockery.yaml:9,11` |
| 11 sites import pgdb (P1.3) | ✓ | `main_test.go:10`, `models/{article,user,tenant,admin}.go`, `internal/testutil/{fixtures.go:16,postgres_test.go:11}`, `tests/article_test.go:22`, `tests/login_test.go:20`, `controllers/auth_test.go:16`, `internal/mocks/mock_Querier.go:8` |
| `ci.yml:88` (`git diff --exit-code internal/db`), `ci.yml:83` (build P1.8) | ✓ | `.github/workflows/ci.yml:88,83` |
| `Makefile:50/88/168`, `playwright.config.ts:51` (P1.8) | ✓ | `Makefile:50,88,168` ; `app/playwright.config.ts:51` |
| `SetupRouter` `tests/article_test.go:34-80`, `binding.Validator` l.39 | ✓ | `tests/article_test.go:34-80,39` |
| `internal/testutil/factory_test.go:86` (P1.5), `factory.go` non import `binding` | ✓ | `factory_test.go:86` |
| helpers `getUserID`/`getTenantID` `controllers/user.go:19-33` (l.21, l.136) | ✓ | `controllers/user.go:19-33` |
| `models/user.go:122` (ErrNoRows registre), `models/user.go:47` (`authModel`), `models/user.go:149` (`OneWithTenant`) | ✓ | `models/user.go:122,47,149` |
| `createTenantInTx` `models/tenant.go:120`, `tenantModel` l.99, `maxSlugAttempts` 99 | ✓ | `models/tenant.go:120,99,68` |
| audit dans la Tx `models/admin.go:55-64` | ✓ | `models/admin.go:55-64` |
| `DummyHashComparisons` `models/login.go:34-38`, oracle l.80, hash factice l.30 | ✓ | `models/login.go:32-38,80,30` |
| Refresh `controllers/auth.go:84-162`, `authModel` l.19 ; article l.18/99 ; middleware l.10 | ✓ | `controllers/auth.go:84-162,19` ; `controllers/article.go:18,99` ; `middleware/auth.go:10` |
| `reservéSlugs`/`TestAPISurface` (l.36) ; `GetUserForLogin`/`WithTenant`/`CreateUserOwner`/`GetUserByID` (l.40) ; `articles.sql:11,19` filtre `user_id` | ✓ | `models/tenant.go:43-58` ; `db/queries/users.sql:1,8,15,21` ; `db/queries/articles.sql:11,19` |
| artefacts racine (l.59) | ✓ présents | gin-boilerplate, e2edb.exe, tmp_line.txt |
| `docs/test.md` non commité (l.59) | ✗ périmé | commit `581d770` ; F3 |

---

## Tableau récapitulatif

| # | Sévérité | Périmètre (lignes cibles) | Objet | Action |
|---|---|---|---|---|
| F1 | **Blocker** | l.35, l.41, l.46 | `pgdb.InTx`/`pgdb.WithinTx` nommés, jamais spécifiés (audit pt 1) | spécification zéro-décision à écrire dans la Phase 4 |
| F2 | Major | l.25/l.27/l.30, l.111 | acquisition client Redis auth non tranchée (audit pt 6) | appliquer le texte C8 dans la puce Risques Phase 3 |
| F3 | Major | l.59 (+ design l.73) | `docs/test.md` périmé — déjà supprimé par commit `581d770` (audit pt 3) | retirer l'item, garder les artefacts racine |
| F4 | Major | l.27, l.31, l.124, l.165 | `middleware/main_test.go` (TestMain pin) hors spécification Phase 3 | migrer le pin avec le paquet de test, corriger « dossier supprimé si vide » |
| F5 | Nit | l.164 (et l.70) | slogan « couverture » non opérationnalisé (audit pt 5) | reformuler en consignation d'inventaire dans `L<N>-gate.md` |
| F6 | Nit | l.19 | signatures `ArticleRepository` elipsées, constructeur « -style » | écrire les signatures complètes et la forme du handler |
| F7 | Nit | l.26 | forme 200 de la réponse de recovery implicite (audit pt 4, cohérence confirmée) | épingler « même forme que le happy path » |

**Total : Blocker 1 · Major 3 · Nit 3.**

À l'issue des éditions (egm-resolve, portant sur le plan-phases uniquement) : F1 bloquerait L4/L5 tels qu'écrits ; F2-F4 sont des éditions bornées avec texte prêt ou fichier déjà identifié ; F5-F7 sont des précisions. Les points d'audit 2, 4 et 7 sont confirmés sans modification de fond.
