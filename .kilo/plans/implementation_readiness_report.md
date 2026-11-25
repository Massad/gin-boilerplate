# Stage 9 — Implementation Readiness Test

Périmètre audité : DetailedRange 175..fin (Phase 0 P0.1-P0.4, Phase 1 P1.1-P1.8, §7 critères de sortie, questions ouvertes).
Contexte : GoldfishContext aucun ; vérification contre le code réel pour egm-codebase-details.
Reviewers : egm-implementation-readiness + egm-codebase-details (standard implementation-readiness-standard) + egm-style.

---

## Reviewer 1 — egm-implementation-readiness (Technical Lead)

# Rapport d'audit — DetailedRange (lignes 175-268), persona Implementation Readiness

Périmètre audité : §6 Phase 0 (P0.1-P0.4) et Phase 1 (P1.1-P1.8), §7 critères de sortie, questions ouvertes. Les références de lignes citées dans le plan ont été vérifiées contre le dépôt (ci.yml, main.go, Makefile, sqlc.yaml, .mockery.yaml, playwright.config.ts, tests/article_test.go, imports `db`/`internal/db`/`forms`/`config` : toutes exactes à la ligne près).

## Constats — Major

**M1 — [Major] P1.7 (lignes 232, 234) — `internal/app.Deps` sous-spécifié : décision de design laissée à l'implémenteur.**
Le plan dit « `type Deps struct{...}` » et ne nomme qu'un champ (`TemplateGlob string`) plus « env », forme indéterminée : champ booléen, champ string, ou lecture directe de `os.Getenv("ENV")` dans `NewRouter` ? Or la branche conditionnelle Swagger (`if os.Getenv("ENV") != "PRODUCTION"`, main.go:184) en dépend directement, et la façon dont les tests contrôlent cette branche (ENV posé via `t.Setenv` vs champ Deps) n'est pas tranchée. L'appel d'exemple dans `tests/article_test.go` (`app.NewRouter(app.Deps{TemplateGlob: "../public/html/*", ...})`) contient une ellipse : les autres champs à passer au runtime de test restent à inventer. Guidance : figer la liste exhaustive des champs de `Deps` (nom, type, valeur par défaut), trancher la gestion de ENV, et écrire l'appel de test complet sans « ... ».

**M2 — [Major] P1.5 (lignes 223-225) — liste des fichiers à modifier incomplète et validation absente.**
L'étape liste les appelants à migrer (`forms.Translate` → `validate.Translate`, branchement `binding.Validator`) mais omet `internal/testutil/factory_test.go:86`, qui fait aussi `binding.Validator = new(forms.DefaultValidator)` et doit passer à `validate.DefaultValidator`. De plus, après le basculement de main.go:130, l'import `forms` de main.go:20 devient inutilisé et doit être supprimé — non mentionné. Enfin P1.5 ne porte aucune ligne « Vert » ni commande de vérification propre (les autres étapes en ont une). Le standard exige une liste explicite create/modify/delete : la liste fournie laisse un point de compilation et de périmètre d'écriture à découvrir. Guidance : ajouter `internal/testutil/factory_test.go` et la suppression de l'import main.go:20, et une ligne de vérification (`go build/vet` + `go test ./internal/platform/validate/ ./internal/testutil/`).

**M3 — [Major] P1.6 (lignes 228-229) — signatures et messages de `httpx` non déterminés (Zero-Decision).**
`MapPgxError` est décrit seulement par ses mappings de codes (`ErrNoRows`→404, `23505`→409, `23503`→409, `42501`→500). Ni le type de retour (`int`, `(int, string)`, `(int, gin.H)`, `error` encodé ?), ni le message exact porté par chaque cas, ni le traitement des `*pgconn.PgError` hors des trois codes listés (500 générique ?) ni celui d'un `error` non-pgx ne sont fixés. `AbortMessage`/`JSONMessage` : distinction non définie (abort vs écriture simple), forme exacte du corps. Ce paquet devient l'API centrale d'erreur des phases 2+ : une signature inventée au fil de l'eau se propagera. Guidance : figer les signatures complètes et le message codé en dur de chaque branche (ex. messages repris à l'identique des constants actuels des controllers), et le comportement catch-all (500, message générique).

**M4 — [Major] §7 ligne 258 — critère de sortie contradictoire avec la feuille de route.**
« après la Phase N, aucun paquet racine (`controllers/`, `models/`, `forms/`, `middleware/`) n'importe un paquet `internal/modules/*` » est impossible à satisfaire pour les phases où il est évalué : la Phase 3 injecte `*auth.Service` dans `models/login.go`, `models/user.go`, `controllers/admin.go` (lignes 137, 139, 154) — des paquets racine qui importeront `internal/modules/auth`, par conception du plan lui-même. Le critère tel qu'écrit rend la phase 3 non clôturable. Par ailleurs « compteur de fichiers legacy décroissant » est énoncé sans métrique (comptage de quoi, évalué quand ?). Guidance : reformuler — l'import racine→modules est autorisé en transition, mesuré par un comptage explicite (ex. nombre de fichiers sous `controllers/`, `models/`, `forms/`, `middleware/`), et l'interdiction racine→modules n'est vérifiée qu'après suppression des répertoires (Phase 7) ; garder tel quel l'invariant « platform n'importe aucun module ».

## Constats — Nit

**N1 — [Nit] P1.7 ligne 232 — recopie « à l'identique main.go:127-212 » imprécise sur deux points.**
(1) La plage citée inclut les lignes 136-146 (init des pools pgx + Redis + `defer pgdb.Close()`) qui doivent rester dans `cmd/api` : le plan ne dit pas explicitement de les exclure ; une recopie littérale produirait une double initialisation ou dépendrait du contexte d'interprétation. (2) La liste des éléments à recopier omet `r.Use(gzip.Gzip(gzip.DefaultCompression))` (main.go:134), pourtant épinglé par le smoke CI (assert content-encoding). Guidance : préciser « recopier main.go:127-134 et 148-212, hors lignes 136-146 », et ajouter gzip à la liste.

**N2 — [Nit] P1.2 ligne 207 — le renommage des qualificateurs d'appel n'est pas énoncé.**
L'étape dit « MAJ imports (7 fichiers) » mais les sites appelants utilisent `db.InitRedis` / `db.GetRedis` / `db.RedisDBIndex` : le qualificateur doit devenir `redisdb.` (et `cmd/e2edb/main.go` aussi réécrit ses usages). Mécanique, mais à écrire pour rester zero-décision ; noter aussi le commentaire `internal/db` en `cmd/e2edb/main.go:200` (cosmétique, à mettre à jour avec P1.3).

**N3 — [Nit] Lignes 203, 208, 216 (et discipline ligne 177) — incohérence `-mod=readonly`.**
Plusieurs lignes « Vert » écrivent `go build ./... && go vet ./...` sans le flag, alors que la ligne de discipline (177) et la convention AGENTS.md (« toujours `-mod=readonly` sur les commandes go ») l'exigent. Harmoniser pour éviter qu'un copier-coller de la commande produise une écriture de go.sum/module graph.

**N4 — [Nit] P1.4 (lignes 218-220) — pas de vérification propre à l'étape.**
Aucune ligne « Vert » ni test cité (les 7 tests middleware restent dans `middleware/` avec `main_test.go`/`auth_test.go` — la seule vérification est build+vet global via la discipline commune). Acceptable, mais nommer explicitement la commande attendue (`go test ./middleware/... ./internal/platform/middleware/` ou l'équivalent `make test` différé) rendrait le checkpoint symétrique aux autres étapes.

**N5 — [Nit] P0.3 (lignes 192-194) — forme exacte des clés comparées non fixée.**
Le test extrait `{(méthode, chemin)}` de `docs/swagger.json` : le style des paramètres Swagger (`{id}`) diffère de la déclaration Gin (`:id`) et le préfixe `/v1` est implicite dans les deux sources ; « compare à la liste attendue » n'indique pas si c'est une égalité bidirectionnelle stricte. Ces micro-choix, mécaniques mais réels, devraient être figés (préciser : normalisation `:id`→`{id}`, préfixe `/v1/*app/`, comparaison d'ensembles égaux dans les deux sens) pour que deux sessions différentes écrivent le même test.

**N6 — [Nit] Question ouverte ligne 263 (invoice, Phase 8) — optionnalité sans critère.**
« Déplacement recommandé mais optionnel » n'est pas une décision : à l'entrée de la Phase 8, l'implémenteur devra trancher lui-même. Les autres questions ouvertes sont correctement bornées (« jamais pendant la migration », « après Phase 8 ») ; celle-ci seule laisse un choix pendant. Guidance : trancher maintenant (oui/non) ou définir le critère qui la déclenchera.

## Points vérifiés conformes (aucun constat)

Phase 0 : la dérive smoke (ci.yml:139/157/166-168 vs main.go sans routes invoice, 404 NoRoute sur `/v1*`) est réelle et le correctif P0.1 est complet et chiffré exactement ; P0.2 couvre la régénération idempotente des trois générateurs ; P0.3 est viable (le `TestMain` de `tests/` ne fait que poser REDIS_DB/DB_NAME en variables — aucun service requis ; la liste figée des 14 routes correspond exactement à main.go:148-182). Phase 1 : les 7 fichiers d'import `db` (P1.2) et les 12 sites d'import `internal/db` (P1.3) correspondent exactement à l'inventaire réel du dépôt (rien de manqué, rien de superflu) ; la séquence P1.2 (renommage redis) avant P1.3 (paquet `db` conservé) élimine correctement la collision de nom de paquet ; Makefile:50/88/168, ci.yml:83/88, playwright.config.ts:51 tous exacts ; P1.7+P1.8 dans le même commit préserve le caractère buildable de chaque commit ; chaque phase finit en état buildable (build/vet/make test, e2e en Phases 1 et 7). Les éléments obligatoires du standard (chemins de fichiers, étapes actionnables, plan de tests, prérequis) sont présents dans l'ensemble, avec les réserves M1-M4 ci-dessus.

---

## Reviewer 2 — egm-codebase-details (Pedantic Auditor, vérification code)

# Rapport d'audit Goldfish — « architecture-modulaire-par-domaine », DetailedRange (l. 175-268)

## Résumé de vérification

Toutes les affirmations vérifiables de la zone auditée ont été confrontées au code réel et sont **exactes** : `ci.yml:139/157/166-168` (probes `/v1/invoice` + `/v1/invoice/download`, fragment `INV-001`, bloc PDF), `ci.yml:83` (`go build -mod=readonly -o /tmp/gin-boilerplate .`), `ci.yml:88` (`git diff --exit-code internal/db`), garde `! git grep "InitRedis(1)"` (ci.yml:96) et garde « fuite helpers » (ci.yml:97-104) ; « 14 routes API » (14 opérations sur 13 chemins, conformes à `main.go:148-182` et `docs/swagger.json`, sans route invoice — NoRoute renvoie bien 404 JSON sur `/v1*` : le smoke est donc réellement rouge tel que décrit) ; `Makefile:50` (`@go run .`), `:88` (`swag init` racine), `:168` (`go run -mod=readonly . -create-admin`) ; `sqlc.yaml:9` (`out: internal/db`, package `db`, overrides uuid/timestamptz) ; `.mockery.yaml:9` et `dir: "{{.InterfaceDir}}/../mocks"` ; `db/db.go` (DefaultRedisDB=1, plage 0-15, exports InitRedis/GetRedis/RedisDBIndex, db_test.go présent avec rejets hors plage) ; e2edb refusant l'index 1 quand REDIS_DB absent ; `internal/db/pool.go` (Store/InTx/WithTenant `set_config(...,true)`/postgresDSN/InitTenant/InitAdmin/TenantStore/AdminStore/Close, singletons l.92-121), `querier.go` interface ; `internal/mocks/{mock_DBTX,mock_Querier}.go` ; `internal/testutil/{factory,factory_test,fixtures,postgres,postgres_test}.go` ; `main.go:16-28/56-77/127-146/148-182/184-212/130` ; `app/playwright.config.ts:51/54/57` (`go run .`, cwd `..`, REDIS_DB "3") ; `tests/article_test.go:20/22/34-80/39` ; `tests/login_test.go:18/20` ; `controllers/auth_test.go:15-16` ; `middleware/auth_test.go:12` (ainsi que :11 controllers, :14 models) ; `cmd/e2edb/main.go:45` ; `main_test.go:10` ; `internal/testutil/fixtures.go:16`, `postgres_test.go:11` ; `internal/mocks/mock_Querier.go:8` ; `controllers/user.go:19-33` (getUserID/getTenantID/getRole, bornes exactes) ; fichiers `config/`, `forms/{validator,errors}.go` (exports DefaultValidator, Translate, ValidationMessages, DefaultMessage), middleware (cors/request_id/auth + 2 tests), `docs/swagger.json` committé, `tests/main_test.go` (TestMain ne fait que poser l'env — P0.3 « pas de services requis » est correct au niveau logique) ; convention AGENTS.md « Risques » ; fichiers proposés (`internal/platform/*`, `internal/app`, `tests/api_surface_test.go`) traités comme créations, non audités.

## Constats

**1. Severity : Major — Affected Scope : §6 P1.5 (l. 224), inventaire des sites appelants `forms`.**
- `internal/testutil/factory.go` est cité comme site appelant `forms.Translate` (« internal/testutil/factory.go (usage de forms) ») : affirmation fausse sur le code. `factory.go` n'appelle ni `Translate` ni `DefaultValidator` ; il n'utilise que les structs de domaine `forms.RegisterForm`/`forms.CreateArticleForm` (factory.go:60, 71, 99, 102), qui restent dans `forms/` après la migration — l'instruction est à effet nul et le lecteur cherche une modification inexistante.
- Le site réellement cassant est omis : `internal/testutil/factory_test.go:86` fait `binding.Validator = new(forms.DefaultValidator)` (fichier sans build tag). Après le déplacement de `DefaultValidator` vers `validate`, ce fichier ne compile plus. `go build ./...` (premier verbe de la porte « Vert » de P1.5) ignore les fichiers `_test.go` ; seul `go vet` ou un `go test` du paquet le rattrapera. Le plan viole la règle zéro-décision : l'implémenteur doit découvrir lui-même ce troisième site (`main.go:130` et `tests/article_test.go:39` sont les deux seuls mentionnés).

**2. Severity : Nit — Affected Scope : §6 P1.7 (l. 234), remplacement de `SetupRouter` par `app.NewRouter`.**
Hypothèse non énoncée : `SetupRouter` actuel n'installe que `gin.Default()` (Logger + Recovery) et `LoadHTMLGlob` — ni CORS, ni RequestID, ni gzip, ni le NoRoute JSON 404, ni la route `/`. En recopiant `main.go:127-212`, `app.NewRouter` fait entrer ces middlewares et ces routes dans la surface HTTP de la suite `tests/`. Les assertions vérifiées semblent tolérantes (le seul 404 testé, `TestGetInvalidArticle`:449, vient du handler et non du NoRoute), mais le plan affirme « assertions non modifiées » / « zéro comportement changé » sans documenter cette différence ; une assertion future ou un effet de bord (en-têtes CORS, RequestID, compression) sera inexplicable pour un lecteur zéro-contexte.

**3. Severity : Nit — Affected Scope : §6 P1.7 (l. 232), énumération du contenu de `NewRouter`.**
L'énumération omet `r.SetTrustedProxies(nil)` (main.go:128). « à l'identique » couvre théoriquement, mais un implémenteur qui recopie strictement la liste des éléments mentionnés produira un routeur qui diverge de `main()` (message d'avertissement trusted proxies dans les logs de test). À ajouter explicitement.

**4. Severity : Nit — Affected Scope : §6 P1.2 (l. 206-208), « MAJ imports (7 fichiers) ».**
Sous-description : renommer le package `db` → `redisdb` ne demande pas seulement de mettre à jour les lignes d'import listées, mais aussi de réécrire tous les identifiants qualifiés `db.InitRedis`/`db.GetRedis`/`db.RedisDBIndex` dans les corps des 7 fichiers consommateurs (p.ex. controllers/auth_test.go:34-37, middleware/auth_test.go:30-33, cmd/e2edb/main.go, tests/article_test.go) et de choisir l'alias/prefix dans `main.go:18`. Le plan ne mentionne que les numéros de lignes d'import ; le travail réel est mécaniquement plus large et n'est déductible que par inférence.

---

## Reviewer 3 — egm-style (Audit DetailedRange)

# Rapport — Audit de style (Scope : ligne 175 → fin, §6, §7, questions ouvertes)

**1. Nit — Ligne 175 (en-tête §6)**
« PROPOSITION PRÊTE À L'EMPLOI » est une étiquette marketing (équivalent « turnkey »), pas une information technique. En majuscules, elle ajoute de l'emphase sans signal. Rédaction factuelle suggérée : « §6 — Instructions détaillées : Phase 0 et Phase 1 ». Les renvois aux lignes 121 et 125 (hors scope) suivront.

**2. Nit — Ligne 231 (P1.7, titre d'étape)**
« (l'étape structurante) » est un jugement subjectif d'importance. Le contenu de l'étape (création du composition root `internal/app`) parle de lui-même. Supprimer.

**3. Nit — Ligne 213 (P1.3)**
« MAJ imports (12 fichiers + mocks générés) » : l'énumération compte 11 fichiers (main.go, main_test.go, 4 sous models/, 2 sous testutil/, 2 sous tests/, controllers/auth_test.go) auxquels s'ajoutent les mocks, listés à part. Le chiffre ne correspond pas à la liste. Recalibrer le décompte.

**4. Nit — Lignes 223-225, 227-229, 232-235 (P1.5, P1.6, P1.7)**
P1.1, P1.2, P1.3 se terminent par une ligne « Vert : ... » ; P1.5, P1.6, P1.7 n'en ont aucune. Granularité de validation incohérente au sein de la même phase. Ajouter une ligne de vérification uniforme à chaque étape.

**5. Nit — Ligne 192 vs ligne 194 (P0.3)**
« pas de services requis à l'exécution » contredit la ligne 194 qui exige « base `artisan_test` déjà provisionnée ». `CALIBRATE` — l'affirmation « pas de services requis » n'est vraie qu'au niveau du fichier de test (logique), pas du paquet (TestMain). Réécrire : « la logique du test ne dépend d'aucun service ; l'exécution reste soumise au `TestMain` du paquet ».

**6. Nit — Ligne 241 (P1.8)**
« le test P0.3 le prouve » : « prouve » surqualifie (le test vérifie l'égalité doc↔liste, pas le comportement servi). `CALIBRATE` — « le vérifie ».

**7. Nit — Ligne 188 (P0.2)**
« (diff git vide sinon : committer la régénération avant de commencer) » est une phrasé télécomprimée. Réécrire en ordre d'exécution : « si une régénération produit un diff non vide, le committer avant de commencer la migration ».

**8. Nit — Ligne 258 (§7)**
« Compteur de fichiers legacy décroissant » introduit une métrique jamais définie ; la règle qui suit porte sur la direction des imports. Rétitrer : « Sens des imports racine → modules interdit après chaque phase ».

**9. Nit — Lignes 177, 187, 244-249, 253 (§6, §7)**
La batterie « build + vet + make test » répétée quatre fois sous quatre formes : garder §7 comme source unique générique et référencer depuis les phases (ou inversément) — deux listes parallèles divergeront.

**Synthèse** : aucun Blocker ni Major sur ce scope. Problèmes de forme : étiquette hyperbolique (en-tête §6), jugement subjectif (P1.7), quatre imprécisions de calibrage ou de décompte, incohérence de granularité de validation entre étapes P1.x, redondance des batteries de vérification.
