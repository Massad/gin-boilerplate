# implementation_readiness_report.md — Stage 9 : Implementation Readiness Test

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md`
**Périmètre :** DetailedRange 48..207 (Tâches T1-T9, Risques, Hors scope)
**Reviewers :** egm-implementation-readiness, egm-codebase-details, egm-style — contextes frais
**Standards appliqués :** Implementation Readiness Standard (readiness, codebase-details) ; aucun (style)
**Note méthodo :** faits bibliothèques vérifiés via Context7 (docs officielles) par l'orchestrateur et injectés — aucune lecture du module cache.

---

## A. egm-implementation-readiness — Pragmatic Tech Lead

**Verdict : le plan n'est pas exécutable « tel quel » au sens du standard.** Techniquement solide sur le cœur (parseur, signature, câblage de l'index), mais quatre décisions structurantes restent ouvertes (protection permanente de l'index, partage db 2 tests/e2e, usage/garde de l'outil e2edb, découpage en phases), deux ancres de fichiers sont erronées sur des opérations d'écriture, et le Test Plan ne prouve pas l'objectif annoncé (l'index effectif). **Aucun Blocker** ; les défauts Major empêchent une exécution correcte « au premier essai » sans interprétation.

### M1 — [Major] T7 : ancre de ligne fausse sur le Makefile (écriture sur la mauvaise recette)
Plan l. 135-136 vise « lignes ~121-122 » ; l'appel `cmd/e2edb` est en **`Makefile:155`** (recette `e2e` l. 152-165) ; les lignes 121-122 sont dans `install_goose`/DB_DSN. Corriger l'ancre (ou supprimer les numéros, nommer la cible : « la seule ligne de la recette `e2e` qui invoque `go run -mod=readonly ./cmd/e2edb` »). La commande de remplacement proposée par le plan est, elle, exacte (le préfixe `REDIS_DB=2` en position d'affectation de commande bat une valeur exportée par `.env`).

### M2 — [Major] T8 : ancre `activeContext.md` fausse + contradiction interne sur les sections historiques
La ligne 60 (joignabilité WSL) est sans rapport ; les énoncés qui deviennent **faux** sont `activeContext.md:26` (« `FLUSHDB` Redis db 1 ») et `:29` (« le `FLUSHDB` purge la base Redis 1 ») — tous deux dans la section historique « Session 2026-11-23 » que le plan interdit de retoucher. L'implémenteur doit choisir entre laisser une affirmation fausse, l'annoter, ou la réécrire : le plan ne tranche pas. Décision requise : (a) corriger les ancres sur 26/29 ; (b) trancher explicitement (amender avec annotation datée, ou déplacer) ; (c) l'invariant à écrire : « le client Redis doit être créé pendant que `REDIS_DB=2` est présent dans l'environnement du processus ».

### M3 — [Major] T5 + Risques : l'invariant d'ordre `t.Setenv`/`godotenv.Load` est faux et le plan demande de l'écrire dans le code
Fait établi (docs godotenv) : `Load` n'écrase jamais une variable déjà posée — l'ordre est **indifférent**. Le plan encode dans le dépôt un invariant faux et l'enseigne comme vrai dans un commentaire de code exigé. Réécrire l'invariant réel (« `REDIS_DB=2` présent dans l'environnement **au moment où le client Redis est créé** »), supprimer la contrainte d'ordre de la section Risques, faire porter le commentaire de code sur ce seul fait.

### M4 — [Major] Décision non actée : la portée réelle de `t.Setenv` n'est pas énoncée, le mécanisme de protection permanent n'est pas choisi
`t.Setenv` restaure la valeur à la fin du test ; dans le paquet `tests`, `REDIS_DB=2` n'existe que pendant `TestIntDB` — la suite marche par séquencement (ordre alphabétique des fichiers), non par garantie. Un nouveau fichier de test appelant `db.InitRedis()` après `TestIntDB` retomberait sur l'index 1 **sans erreur**. Zero-Decision violé : `TestMain` par paquet, helper commun, ou assertion d'index — le plan doit acter UNE option. Recommandation : **`TestMain` par paquet taggé `all`** (`tests`, `middleware`, `controllers`) posant `REDIS_DB=2` au niveau processus, **plus** l'assertion permanente d'index (M5) ; si le `t.Setenv` est retenu, formuler la règle exacte et l'ajouter à AGENTS.md comme convention.

### M5 — [Major] Aucun test ne prouve l'index effectif : le plan livre un test du parseur, pas un test du câblage
Rien ne vérifie que le client réellement ouvert par chaque suite pointe sur db 2. Une régression banale restaure le bug **sans signal** ; la validation T9 est manuelle, non rejouable par la CI. AGENTS.md impose « toute nouvelle feature livrée avec ses tests ». Ajouter après chaque `db.InitRedis()` : `require.Equal(t, 2, db.GetRedis().Options().DB, "…")` (API confirmée par les docs go-redis) dans `TestIntDB`, `middlewareTestSetup`, `TestRefreshOnlyConsumesMatchingOwner`. Et trancher la garde CI anti-dérive `! git grep -n "InitRedis(1)" -- '*.go'` (modèle legacy drivers, `ci.yml:88`) — attention : elle contredit « Hors scope : CI inchangée », c'est une décision explicite à acter, pas un détail.

### M6 — [Major] Test Plan incomplet : le mode d'échec « loud » (décision centrale) n'est pas testé
Deux comportements nouveaux non couverts : `db.InitRedis()` retourne une erreur sur `REDIS_DB` invalide ; `cmd/e2edb` propage l'erreur au lieu de purger. Ajouter un cas sans service :
```go
t.Setenv("REDIS_DB", "16")
require.Error(t, db.InitRedis(), "an out-of-range REDIS_DB must fail loudly")
```
et côté outil un test vérifiant qu'une valeur invalide renvoie une erreur **avant** tout `FLUSHDB` (modèle `TestValidateTarget`).

### M7 — [Major] Partage de l'index 2 tests Go/e2e : prémisse « les commandes sont séquentielles » non garantie, mode de défaillance non décrit
Deux acteurs destructeurs et globaux sur le même index, mêmes formes de clés : `FLUSHDB` e2e ↔ purges `user_sessions:*`/`admin:*` des suites (et `TestCleanUp` purgerait la session `admin:<uuid>` du superadmin e2e → 401 Playwright). L'hypothèse « jamais concurrents » a déjà failli dans ce dépôt. La section Risques ne mentionne ce scénario à aucun endroit. Trancher : **Option A** index distinct e2e (`REDIS_DB=3`, Makefile + playwright.config.ts) — supprime la classe entière ; **ou Option B** index partagé assumé + invariant écrit dans AGENTS.md (« ne jamais exécuter `make test` et `make e2e` simultanément ») + garde-fou dans l'outil (M9).

### M8 — [Major] T9 : les commandes de preuve Redis ne sont pas exécutables telles quelles
Étape 3 correcte (`docker compose exec -T redis redis-cli …`), mais les étapes 4-6 utilisent `redis-cli` **sans hôte ni port** : le Redis de dev est publié sur `127.0.0.1:16379`, le port 6379 de l'hôte appartient à une autre stack → la commande interrogerait une autre instance ou échouerait → **faux négatif possible sur la preuve d'isolation**. Préciser aussi : pas de mot de passe (aucun `-a`), et l'attendu de l'étape 4 formulé comme observation. Fixer la forme : `docker compose exec -T redis redis-cli -n N …` partout, ou `redis-cli -h 127.0.0.1 -p 16379 -n N …` (port issu de `REDIS_HOST`).

### M9 — [Major] `cmd/e2edb` : la ligne `Usage:` de l'en-tête reste le mode d'emploi découvert, et elle est destructive pour les sessions dev
Après T4, la commande d'usage documentée (`E2E_DB_NAME=artisan_e2e go run ./cmd/e2edb`, `main.go:21`) résout `REDIS_DB` absent → index 1 → FLUSHDB de la base des sessions dev. Le plan ne décide ni si la ligne `Usage:` doit porter `REDIS_DB=2`, ni si l'outil doit refuser de purger l'index dev. Trancher : mettre à jour `Usage:` en `E2E_DB_NAME=artisan_e2e REDIS_DB=2 go run ./cmd/e2edb` **et** décider de la garde active (refuser l'index 1 sauf opt-in explicite, sur le modèle testé de `validateTarget`) — seul moyen de rendre « db 1 = dev exclusivement » une garantie du code et non un commentaire.

### M10 — [Major] Phasage/checkpoints absents, et l'ordre T1→T9 casse la compilation entre les tâches
T1 (nouvelle signature) casse la compilation de `main.go:144` et des trois paquets de test jusqu'à T3/T5/T6 ; le dépôt ne redevient vert qu'après T6, sans checkpoint énoncé. Aucune allowlist consolidée de fichiers. Reformuler en trois phases : **Phase 1 — cœur + tous les appelants** (T1-T6), jalon `gofmt`/`go vet`/`go test ./db`/`make test` verts ; **Phase 2 — chaîne e2e** (T7), jalon `make e2e` + sentinel db 1 intacte ; **Phase 3 — documentation/mémoire** (T8), jalon `git status`/`git diff` relus (T9.7). Allowlist proposée :
- **Créés :** `db/db_test.go`
- **Modifiés :** `db/db.go`, `main.go`, `cmd/e2edb/main.go`, `tests/article_test.go`, `middleware/auth_test.go`, `controllers/auth_test.go`, `Makefile`, `app/playwright.config.ts`, `AGENTS.md`, `README.md`, `memory-bank/techContext.md`, `memory-bank/activeContext.md`, `memory-bank/progress.md`, `.kilo/HANDOFF.md`
- **Interdits d'écriture :** `cmd/e2edb/main_test.go` (sauf garde active décidée → cas ajoutés), `.env`, `.env_rename_me`, `.github/workflows/ci.yml` (sauf décision D-garde-CI), `docs/`, `go.mod`, `go.sum`, `app/src/**`, `app/e2e/**`, `internal/db/**`, `models/**`, `middleware/auth.go`, `controllers/**` (hors `auth_test.go`)

### Nits (readiness)
- **N1** Message d'erreur sans le préfixe `error: ` du modèle cité (`config/secrets.go:20`) — trancher une fois.
- **N2** Terminologie « base Redis N » vs « index Redis N » — retenir « index », « db N » réservé aux commandes.
- **N3** `progress.md` : destination de l'entrée non précisée (liste « Ce qui fonctionne » + entrée numérotée **23** de l'« Historique des étapes », qui s'arrête à 22).
- **N4** T2 : package (`db` vs `db_test`) et forme de l'injection (closure sur `map[string]string`) non spécifiés.
- **N5** T7 : préciser que le préfixe `REDIS_DB=2` ne porte que sur la commande `cmd/e2edb` (les autres lignes de la recette restent sur l'environnement hérité — correct, mais à dire).
- **N6** CI l. 33 : « hériteront » → chaque suite **pose** la variable ; verdict de fond correct (CI tenable : `databases 16`, pas de `REDIS_DB`, smoke en db 1).
- **N7** Bénéfice utilisateur jamais énoncé : après le changement, `make test` et `make e2e` ne déconnectent plus les sessions de l'app de dev (plus de re-login). Une phrase dans l'Objectif.

### Éléments vérifiés conformes
Diagnostic exact (4 sites + const, aucun cinquième appelant, `main_test.go` non impacté) ; ancres exactes partout ailleurs (`main.go:112-117/137-142/144`, `cmd/e2edb/main.go:52-54`, tous les setups de test, `playwright.config.ts:57`, AGENTS 9/33/59/68/83, README 60-64/157, techContext 22/41, HANDOFF 27) ; pas de conflit d'import `cmd/e2edb` → `db` (ni cycle ni collision de noms) ; préfixe Makefile correct ; webServer env fusionné avec `process.env` → db 2 atteinte ; CI tenable sans modification ; Swagger correct ; `db/db_test.go` créable sans conflit.

---

## B. egm-codebase-details — Pedantic Quality Auditor

**Verdict : plan techniquement solide sur le fond (le mécanisme est correct et l'objectif est atteint), mais incomplet sur trois points structurants :** (a) deux ancres fausses, (b) l'invariant central (« les tests n'écrivent jamais en db 1 ») n'est ni garanti par le code ni testé, (c) le phasage laisse le dépôt non compilable entre T1 et T6. **Aucun énoncé de la portée ne décrit un comportement que le code n'a pas : 0 Blocker.** 9 Major, 12 Nit.

### M1 — [Major] Ancre `Makefile` factuellement fausse (T7, l. 135-136)
Appel réel `Makefile:154-155` ; les lignes 121-122 contiennent `DB_SSLMODE`/`DB_DSN` (section DSN). Écart de 33 lignes, hors recette `e2e`. Corriger en `Makefile:154-155` (recette :152-165).

### M2 — [Major] Ancre `activeContext.md` fausse + affirmation devenue fausse laissée en l'état (T8)
Les énoncés qui deviennent faux : `:26` et `:29` (section « Session 2026-11-23 »). `activeContext.md` est le fichier d'état **courant** (« mettre à jour à CHAQUE session », l. 3) : laisser `:29` intact y propage une affirmation fausse. L'instruction « ne pas retoucher les sections historiques » entre en conflit avec la correction requise : trancher (amender, annoter « obsolète depuis le plan REDIS_DB », ou déplacer).

### M3 — [Major] La portée du `t.Setenv` n'est pas documentée : l'invariant « jamais de db 1 » n'est garanti que par accident
Dans le paquet `tests`, `REDIS_DB=2` n'existe que **pendant** `TestIntDB` ; le client global conserve `DB=2` pour la suite — ça marche aujourd'hui par hasard de séquencement (ordre alphabétique : `article_test.go` avant `login_test.go`). Tout `db.InitRedis()` appelé après `TestIntDB` résout l'index 1 — la dérive silencieuse que le plan qualifie d'impossible. L'invariant réel : « `REDIS_DB=2` présent dans l'environnement **au moment où le client Redis est créé** ». Formuler et choisir le mécanisme (helper, `TestMain`, assertion d'index).

### M4 — [Major] Aucun test automatisé ne prouve l'index effectif ; l'objectif du plan n'est couvert que par une validation manuelle
Le plan livre un test du **parseur**, aucun du **câblage**. Une régression banale restaure le bug sans test rouge. Remède minimal, une ligne par paquet : `require.Equal(t, 2, db.GetRedis().Options().DB, "…")` après chaque `db.InitRedis()`. Si une garde CI anti-`InitRedis(1)` est souhaitée, l'ajouter à T9 **explicitement** — « Hors scope » (l. 206) interdit tout changement de CI : décision à trancher, pas à laisser à l'implémenteur.

### M5 — [Major] L'hypothèse « `databases 16` » n'est couverte par aucun cas d'échec documenté
Vraie pour `redis:8-alpine` (dev `docker-compose.yml:25-36` sans `command:`, CI `ci.yml:33-41`), fausse pour un Redis managé réduit : `REDIS_DB=2` accepté par le parseur (échec loud muet) puis `ERR DB index is out of range` à la première commande — et sans ping (choix assumé), l'app démarre et casse à la première opération, avec une erreur qui ne mentionne ni `REDIS_DB` ni l'index. La section Risques ne couvre pas ce cas : l'énoncer (au minimum dans le message d'erreur ou le commentaire de la fonction).

### M6 — [Major] La valeur « 2 » est dupliquée dans la chaîne e2e sans source de vérité ni contrôle de dérive
Deux acteurs doivent partager le même index : l'outil qui purge (préfixe Makefile) et l'app qui écrit (`webServer.env`) — deux littéraux « 2 » dans deux fichiers, sans constante partagée ni garde. Une divergence réintroduit le bug par une simple édition de littéral. Remèdes au choix : constante Go, variable make unique transmise aux deux, ou contrôle de dérive rejouable. Le plan doit trancher (il ne peut pas laisser les deux littéraux nus sans le dire) — noter que le Stage 8 juge le détecteur-by-test disproportionné : la reconnaissance explicite + la sentinel T9 suffisent, à acter dans le texte.

### M7 — [Major] Le phasage laisse le dépôt non compilable de T1 à T6
Après T1 : `main.go:144` ne compile plus (« too many arguments » + « value not used »). Après T3 : `main.go` compile mais les trois paquets de test restent rouges jusqu'à T6. Un implémenteur qui lance `go build` entre deux tâches croira à une régression. Regrouper T1+T3(+T5+T6) en une unité vérifiable, ou poser le checkpoint explicitement.

### M8 — [Major] L'invariant de non-concurrence `make test`/`make e2e` n'est énoncé nulle part dans la portée
Deux acteurs destructeurs partagent l'index 2 après le changement. La parenthèse « les commandes sont séquentielles » (l. 26) n'est pas un invariant opposable — et la prémisse a déjà failli. Risques : FLUSHDB e2e ↔ sessions des suites Go (401 flaky) ; `TestCleanUp` ↔ session superadmin e2e (401 Playwright). Énoncer la règle dans AGENTS.md ou isoler les deux acteurs (index distinct).

### M9 — [Major] Le refus de purger l'index de dev n'est garanti par aucun code, alors que le dépôt dispose d'un motif de garde-fou testé
`REDIS_DB=1` est parfaitement valide au sens du parseur : l'« échec loud » ne couvre que les valeurs *mal formées*, pas *mal choisies*. Le seul rempart prévu est un commentaire — alors que `validateTarget` (testé dans `main_test.go:13-45`) est le motif exact. Adopter la garde (une ligne + test) ou l'écarter explicitement comme alternative rejetée : décision ouverte aujourd'hui, le standard exige zéro décision ouverte.

### Nits (codebase-details)
- **N1** T2 : borne valide `15` non couverte ; formes acceptées par `Atoi` (`+2`, `02`, `-0`) non tranchées alors que le message dit « integer between 0 and 15 ».
- **N2** Message d'erreur sans préfixe `error: ` (convention `config/secrets.go:20`).
- **N3** T2 : package du fichier non spécifié (`db` recommandé).
- **N4** Message d'assertion « initializing the Redis client » trompeur (pas de ping : seule la résolution de l'index est validée) → « resolving the Redis database index ».
- **N5** T4 : l'énumération des variables lues (`main.go:23-27`) et la ligne `Usage:` (`:21`) doivent inclure `REDIS_DB` — sinon l'en-tête devient faux par omission.
- **N6** T8 : le bullet « Structure » d'AGENTS.md pour `db/` (« init Redis (`InitRedis`/`GetRedis`) », `AGENTS.md:28`) doit mentionner `RedisDBIndex` et la nouvelle signature.
- **N7** T9 : uniformiser `redis-cli` sur `docker compose exec -T redis …` ; étape 5 : URL/port (9001) et corps JSON manquants.
- **N8** Terminologie instable « base/db/index » — fixer « index Redis N », propager dans T8.
- **N9** T5 : invariant de séquencement (ordre alphabétique) à écrire ; noter `go test -run TestCleanUp` seul → client `nil` → panic (préexistant).
- **N10** Bénéfice utilisateur jamais énoncé (les sessions dev survivent) — une ligne dans T8/l'Objectif.
- **N11** `REDIS_DB` posé dans `.env` basculerait silencieusement le dev (le Makefile exporte `.env` à toutes les recettes) : ajouter l'avertissement AGENTS « ne pas la poser dans `.env` ; le dev reste sur le défaut 1 ».
- **N12** Même sans garde active, le choix doit être écrit comme rejet d'alternative, pas laissé en constat de risque.

### Vérifications confirmées (0 correction requise)
Appelants `InitRedis` : exactement 4, aucun cinquième ; aucune autre construction de client Redis dans le dépôt (`db/db.go:22` + `cmd/e2edb/main.go:152`) → le changement d'index est complet ; ancres exactes (liste complète re-vérifiée, seule exception `activeContext.md:60`) ; mécanisme e2e cohérent (préfixe inline > `.env` exporté ; `webServer.env` fusionné avec `process.env` ; `migrate_up` et `-create-admin` n'ouvrent aucun client Redis) ; « 4/4 » e2e correct (1 setup + 3 CUJ) ; `databases 16` vrai pour dev et CI ; CI tenable sans modification ; convention d'erreur citée existe (`ValidateSigningSecrets(getenv)`) ; couverture T8 des mentions « db 1 » exhaustive ; follow-up plan Docker vérifié (l. 58/74/381/423-424) ; frontière de fichiers exploitable ; `db/db_test.go` bien un fichier à créer.

---

## C. egm-style — Technical Copy Editor (DetailedRange)

**16 constats : 4 Major, 12 Nit.** Aucun Blocker de forme ; le texte est sec et factuel — aucun marketing, aucun remplissage. Trois familles : ancres fausses, garanties plus larges que le mécanisme, décisions opérationnelles non tranchées (redis-cli hôte vs conteneur).

### Major
- **F-01** T7 l. 135 : « ~121-122 » → réel `Makefile:155` (l. 121-122 = DSN/`install_goose`). Le tilde n'excuse pas 33 lignes d'écart vers une autre section. Remplacer par `Makefile:154-155` + préciser où se pose le préfixe (après le `set -a`).
- **F-02** T8 l. 162-163 : `activeContext.md:60` sans rapport ; l'énoncé devenu faux est `:29`, dans le bloc historique que l'instruction interdit de toucher — l'instruction ne peut pas être suivie littéralement. Nommer la cible par sa section ou créer une section datée ; expliciter le sort de `:29`.
- **F-03** Risques l. 188-192 + T5 l. 117-119 : (a) la contrainte d'ordre est inerte (godotenv `Load` n'écrase jamais ; `t.Setenv` écrase l'env du processus — l'ordre est indifférent) : **DROP** l'avertissement d'ordre, il enseignerait une règle inexistante ; (b) l'invariant réel (client créé tant que `REDIS_DB=2` est présent ; `t.Setenv` restaure à la fin de `TestIntDB`) n'est pas énoncé — destinataire concret : l'auteur du prochain fichier de test du paquet `tests`.
- **F-04** T9 étapes 3-6 : deux formes d'invocation `redis-cli` sans dire laquelle fait foi ; le binaire `redis-cli` est fourni par l'image du stack, pas par le poste — la lecture littérale échoue sur une commande introuvable. Uniformiser sur `docker compose exec -T redis redis-cli …` ; préciser qu'à l'étape 5 `make run` doit être mis en arrière-plan par l'opérateur.

### Nit
- **F-05** T1 : borne 15 = défaut `databases 16` des images du stack, pas un invariant serveur — **CALIBRATE** (« within the default Redis range 0-15 (`databases 16` des images du stack) ; le `databases` effectif de l'instance reste la référence »).
- **F-06** T1/T3 : asymétrie d'injection — le dépôt injecte `getenv` partout (`InitTenant`, `InitAdmin`, `ValidateSigningSecrets`, `runCreateAdmin`, `run`) ; le plan crée `RedisDBIndex(getenv)` puis fige `os.Getenv` dans `InitRedis()`. Aligner (`InitRedis(getenv …) error`) ou assumer le choix en une phrase.
- **F-07** T4 : ajouter `REDIS_DB` à l'énumération des variables lues (`main.go:23-27`), sinon l'en-tête devient faux par omission.
- **F-08** T4 : ordre d'exécution `recreateDatabase` **puis** `flushRedis` (`main.go:90-95`) — une erreur de parsing remonterait **après** la destruction de la base e2e. Préciser que la résolution de l'index a lieu dans `run()` avant `recreateDatabase` (modèle `validateTarget`).
- **F-09** T6 l. 129 : la plage 27-38 est le bloc de setup, pas la fonction (`TestRefreshOnlyConsumesMatchingOwner` s'étend l. 27-145) — préciser « bloc de setup ».
- **F-10** T8 l. 166-167 : `HANDOFF.md` est un compte rendu daté et clos — le plan y édite une ligne historique alors qu'il interdit de le faire pour `activeContext.md` : même nature, traitement opposé. Assumer l'édition (« section encore utilisée comme référence ») ou s'en dispenser.
- **F-11** Hors scope l. 205-206 : la parenthèse « follow-up documenté du plan Docker » (exacte, vérifiée l. 58/74/381/423-424) se lit comme portant sur les deux sujets adjacents ; « autres bases Redis pour d'autres usages » est vague — scinder et préciser.
- **F-12** T9 étape 2 : convention du repo = `-race` sur les tests unitaires (`AGENTS.md:57`) — aligner ou signaler l'omission.
- **F-13** T2 : package du fichier non fixé — préciser (`db` interne, cohérent avec `cmd/e2edb/main_test.go`).
- **F-14** T7 l. 31/139 : « pas d'héritage implicite » ambigu — ce qui est exclu : l'héritage d'une **valeur** `REDIS_DB` depuis l'environnement/`.env`, pas l'héritage d'environnement en soi.
- **F-15** Terminologie : quatre formulations pour le même objet (« db 1 », « base Redis 1 », « index 1 », « REDIS_DB (2 dans le cycle e2e) ») ; collision réelle avec les deux bases PostgreSQL — fixer « index Redis N » et propager dans T8.
- **F-16** Redondances : justification godotenv répétée (T5 + Risques — garder T5, renvoi croisé côté Risques) ; épinglage `REDIS_DB=2` rappelé trois fois dans T8.

### Vérifications sans réserve
Toutes les autres ancres re-vérifiées exactes (tableau complet dans le rapport brut) ; les trois items de « Risques » sont chacun adossés à une décision opérateur identifiable — fond **KEEP**, seule la formulation de l'ordre est à recaler (F-03).
