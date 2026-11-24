# Isolation REDIS_DB : dev (index 1), suites Go (2), cycle e2e (3)

**Statut** : document prêt pour exécution — l'étape suivante est
l'implémentation T1-T9 (3 phases, § Phasage), du ressort d'un agent
d'implémentation ; chaque phase se ferme sur son jalon T9 ; commit
uniquement sur demande explicite.

## Contexte

Faits établis le 24 nov. 2026 :

- `cmd/e2edb` fait un `FLUSHDB` sur l'index Redis **1** (const `redisDBIndex = 1`,
  `cmd/e2edb/main.go:54`) — le cycle e2e détruit les sessions dev.
- **`make test` détruit aussi les sessions dev** : `tests/article_test.go:914-928`
  (TestCleanUp) et `tests/login_test.go:267-279` suppriment **globalement**
  `user_sessions:*` et `admin:*` de l'index 1 — les sessions dev ont la même
  signature de clés ; les suites middleware/controllers ne suppriment que leurs
  propres clés mais écrivent quand même en 1.
- Tous les clients Redis du dépôt sont créés sur l'index 1 — l'app
  (`main.go:144`) et les trois suites de test (`middleware/auth_test.go:30`,
  `controllers/auth_test.go:34`, `tests/article_test.go:159`, chaque appel
  `db.InitRedis(1)`).

**Symptôme** : après `make test` ou `make e2e`, le développeur devait se
reconnecter à l'app de dev. **Bénéfice** : une session dev survit désormais
aux deux cycles.

## Objectifs et hors-objectifs

**Objectif** : l'index Redis 1 devient exclusivement dev ; les suites Go
travaillent sur l'index 2, le cycle e2e sur l'index 3. Aucun comportement
fonctionnel ne change (rotation atomique des refresh — script Lua
`consumeRefreshAuth`, `models/auth.go:262-277` — préfixes, TTL inchangés ;
seul l'index varie).

**Critère de succès** : le marqueur posé sur l'index 1 avant validation — clé
`admin:sentinel` (valeur `dev-session-marker`), choisie à dessein dans
l'espace de clés que les suites purgent (`admin:*`/`user_sessions:*`, pas une
clé hors motif) — subsiste après `make test` ET après `make e2e` : si une
suite repassait sur l'index 1, sa purge `admin:*` effacerait le marqueur et
le jalon échouerait au lieu d'être vert à tort — preuve rejouée par les
vérifications T9.

**Non-objectifs** : modifier la recette `make test` du Makefile, `.env` /
`.env_rename_me`, la doc Swagger, ou le comportement des sessions.

**Hors-objectif assumé** : `make test` purge toujours la base PostgreSQL dev
(`DELETE FROM public.tenant`, `tests/article_test.go:909`) — follow-up nommé :
base jetable `artisan_test`, symétrique d'`artisan_e2e`.

Risques et limites → § Risques / cas d'échec ; périmètre exclu → § Hors scope.

## Décisions de design (fixées)

Les identifiants (D1)-(D6) sont les décisions tranchées par la revue ; les
tâches et les risques s'y réfèrent.

| Sujet | Choix |
|---|---|
| Répartition des index | **1 = dev (défaut, `REDIS_DB` absent)**, **2 = suites Go** (posé par `TestMain`), **3 = cycle e2e** (préfixe de la recette `make e2e` + `webServer` Playwright) — deux index jetables distincts : le FLUSHDB e2e ne peut jamais croiser les sessions des suites Go, ni inversement |
| Nature de l'isolation (mécanisme) | Index logique sur une instance unique : mêmes identifiants, aucun ACL — protège des purges accidentelles, pas d'un accès malveillant (la doc officielle Redis qualifie les db logiques de *namespacing* : « Don't use them to run multiple unrelated applications in a single Redis instance ») |
| Nature de l'isolation (déploiement) | Ne jamais pointer `REDIS_DB` vers un index partagé avec une autre application ; en production la variable reste absente (défaut 1) |
| Résolution de l'index | Fonction pure partagée `db.RedisDBIndex(getenv) (int, error)` + `const db.DefaultRedisDB = 1` : valeur absente ou vide → `DefaultRedisDB` ; sinon entier trimé requis dans [0, 15] (`databases 16` = réglage par défaut des conteneurs dev/CI ; le `databases` effectif du serveur reste la référence) ; hors plage ou non numérique → erreur |
| Échec explicite | `InitRedis()` retourne l'erreur (le variadic `selectDB ...int` disparaît) ; `main.go` fait `log.Fatal`, style des pools (main.go:137-142) ; message préfixé `error: `, même forme que `config.ValidateSigningSecrets` |
| Rupture d'API | `InitRedis(selectDB ...int)` → `InitRedis() error` : rupture de signature pour tout importateur externe du module (aucun consommateur interne après le changement) |
| Un seul parseur | `cmd/e2edb` importe `github.com/Massad/gin-boilerplate/db` et passe son `getenv` injecté à `RedisDBIndex` → l'outil qui FLUSH et l'app du cycle e2e résolvent **le même index depuis le même env** ; la const `redisDBIndex` disparaît |
| Index des suites Go (D1) | Un `TestMain` par paquet concerné parmi les six paquets taggés `all` — `tests`, `middleware`, `controllers`, les seuls qui créent un client Redis — posant `REDIS_DB=2` au niveau processus — l'invariant de protection : le client Redis doit être créé tant que `REDIS_DB` est présent dans l'environnement du processus ; plus aucun `t.Setenv("REDIS_DB", …)`, plus aucune contrainte d'ordre avec `godotenv.Load` |
| Index du cycle e2e (D2) | `REDIS_DB=3` : préfixe sur l'appel `cmd/e2edb` de la recette `make e2e` + `REDIS_DB: "3"` dans le `webServer` de `app/playwright.config.ts` — la valeur n'est héritée ni du runner ni du `.env` |
| Garde `cmd/e2edb` (D3) | Refus de purger l'index dev quand `REDIS_DB` est absent (modèle `validateTarget`) ; résolution de l'index dans `run()` **avant** `recreateDatabase` ; `Usage:` et énumération des variables lues mis à jour |
| Garde CI (D4) | Une étape ajoutée : `! git grep -n "InitRedis(1)" -- '*.go'` (modèle « legacy drivers ») ; le reste de la CI inchangé |
| Littéraux d'index (D5) | Les valeurs 2 (`TestMain`) et 3 (chaîne e2e) sont écrites en dur ; leur accord est prouvé par les vérifications T9 |
| Documentation historique (D6) | Annotation datée, sans réécriture : `activeContext.md` l. 26/29 → « (à la date du 2026-11-23 ; l'isolation des index 1/2/3 est en vigueur depuis le plan REDIS_DB) » ; `.kilo/HANDOFF.md:27` édité en assumant explicitement le statut « section encore utilisée comme référence » ; aucune autre retouche d'historique |

## Ce qui ne change pas

- **Makefile cible `test`** : inchangée — les suites posent elles-mêmes leur
  index via `TestMain`, même pour `go test` lancé directement hors make.
- **`.env` / `.env_rename_me`** : aucune ligne `REDIS_DB` ajoutée — le défaut 1
  EST la config dev ; la variable n'existe pas dans l'environnement dev
  (après T8 elle est documentée dans AGENTS/README mais reste absente des
  fichiers `.env*`).
- **Swagger** : aucune annotation touchée → pas de `make generate_docs`.
- **CI** : une étape de garde ajoutée (D4), le reste inchangé — les suites
  posent elles-mêmes leur index ; le smoke du binaire reste sur l'index 1 par
  défaut.

## Alternatives rejetées

| Alternative | Motif |
|---|---|
| **RETENU : 1/2/3** | trois index — deux acteurs destructeurs globaux (FLUSHDB e2e, purges des suites) ne partagent pas un index : la prémisse « les commandes sont séquentielles, jamais concurrentes » n'est pas garantie ; l'entier supprime la classe entière d'interférence |
| Index partagé (2) + règle écrite | rejeté : un invariant documenté ne remplace pas une séparation d'index (une règle peut être violée, un entier non) |
| Purge ciblée (`user_sessions:*`/`admin:*`) au lieu du FLUSHDB | Ne protège pas les sessions dev : mêmes clés, même index — le FLUSHDB sur un index dédié aux tests reste le nettoyage déterministe correct |
| Conteneur Redis séparé pour e2e | Surdimensionné : un index Redis suffit, pas de port/volume/cycle de vie supplémentaire |
| `REDIS_DB` exporté par le Makefile pour `make test` | Contournable : `go test` lancé directement hors make n'aurait pas la variable ; `TestMain` pose l'index dans le processus de test lui-même |
| Parseur déporté dans `config/` | rejeté : `db` est le point d'accès unique au client Redis (32 lignes, go-redis déjà importé par l'outil), coût de transitivité nul |
| `InitRedis` silencieuse sur valeur invalide (repli 1) | Une faute de frappe de `REDIS_DB` dans le cycle e2e provoquerait une dérive silencieuse vers l'index dev, sans trace ; l'erreur explicite est le pattern du repo (`config.ValidateSigningSecrets`) |
| `InitRedis(getenv)` alignée sur l'injection du repo (`config`/`pgdb`) | rejeté : la pureté testable est déjà isolée dans `RedisDBIndex(getenv)` (même séparation que `run`/`validateTarget` dans e2edb) ; `InitRedis` reste l'unique point d'accroche du binaire principal, sans paramètre supplémentaire — asymétrie assumée à T1 |
| Dupliquer le parse dans `cmd/e2edb` | Deux parseurs = deux sources de vérité pour le même index = exactement la classe de bug corrigée |
| Détecteur de dérive relisant `Makefile`/`playwright.config.ts` depuis un test Go | Couple la suite à l'arborescence ; la sentinel T9 suffit (remède rejeté) |

## Phasage

Trois phases ; chaque phase laisse la base dans un état compilable :

1. **Phase 1 — cœur + appelants (T1-T6)** : `db/db.go`, `db/db_test.go`,
   `main.go`, les trois `main_test.go` + paquets de test, `cmd/e2edb`. La
   suppression du variadic rend `main.go` et les paquets de test non
   compilables tant que T3/T5/T6 ne sont pas posés — livrer T1-T6 en un seul
   bloc. Jalon : phase 1 de T9 verte.
2. **Phase 2 — chaîne e2e + garde CI (T7)**. Jalon : phase 2 de T9 verte.
   Entre les phases 1 et 2, `make e2e` est volontairement refusé : la garde
   D3 (T4) rejette `REDIS_DB` absent alors que le préfixe `REDIS_DB=3` n'est
   posé qu'à T7 — un échec fermé protecteur (sinon la recette purge
   l'index dev), qui ne doit pas être « réparé » en relâchant la garde. Le
   jalon de phase 1 (T9) vérifie cet échec explicitement.
3. **Phase 3 — documentation (T8)**. Jalon : phase 3 de T9 (diff relu).

## Tâches

### T1. `db/db.go` — résolution partagée + signature + défaut exporté

```go
// DefaultRedisDB is the Redis index used when REDIS_DB is absent or empty:
// the dev sessions. Tests and the e2e cycle pin their own disposable indexes.
const DefaultRedisDB = 1

// RedisDBIndex resolves the Redis index from REDIS_DB: an absent or empty
// value means DefaultRedisDB, anything else must be an integer within the
// default Redis range 0-15 (`databases 16` is the stock setting of the
// dev/CI containers; the server's `databases` directive remains
// authoritative).
func RedisDBIndex(getenv func(string) string) (int, error) {
	raw := strings.TrimSpace(getenv("REDIS_DB"))
	if raw == "" {
		return DefaultRedisDB, nil
	}
	index, err := strconv.Atoi(raw)
	if err != nil || index < 0 || index > 15 {
		return 0, fmt.Errorf("error: REDIS_DB %q must be an integer between 0 and 15", raw)
	}
	return index, nil
}

// InitRedis connects to Redis on the REDIS_DB index (default DefaultRedisDB).
func InitRedis() error {
	index, err := RedisDBIndex(os.Getenv)
	if err != nil {
		return err
	}
	RedisClient = _redis.NewClient(&_redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       index,
	})
	return nil
}
```

Imports : `strconv`, `strings`, `fmt` en plus de `os`/go-redis. Reste du
mécanisme inchangé : pas de ping (l'échec de connexion se manifeste au premier
usage, les tests pingent eux-mêmes). `InitRedis` lit `os.Getenv` directement
(consommateur de production) ; l'injection `getenv` n'existe que sur
`RedisDBIndex`, pour les tests et e2edb — asymétrie assumée. `DefaultRedisDB`
est utilisée par le défaut du parseur et par la garde de T4.

### T2. `db/db_test.go` — test unitaire pur (nouveau fichier, sans tag, package `db` interne)

Table-driven `require` (conventions AGENTS), `getenv` injecté : `""` → 1,
`"   "` → 1, `"2"` → 2, `" 2 "` → 2 (trim), `"0"` → 0, `"15"` → 15 (borne
haute valide), `"+2"` → 2, `"02"` → 2, `"-0"` → 0 (formes `Atoi` documentées
comme acceptées) ; erreurs pour `"-1"`, `"16"`, `"abc"`, `"1.5"`. Nom du
test : `TestRedisDBIndex` ; table `struct{name, value string; want int;
wantErr bool}` itérée avec `t.Run(name, …)`. Plus le chemin d'erreur
explicite, sans service :

```go
t.Setenv("REDIS_DB", "16")
require.Error(t, InitRedis(), "an out-of-range REDIS_DB must be rejected explicitly")
```

Plus le chemin nominal sans service (`NewClient` ne se connecte pas à la
création) : `t.Setenv("REDIS_DB", "3")` → `require.NoError(t, InitRedis())`
→ `require.Equal(t, 3, GetRedis().Options().DB, "the parsed index must reach
the client")`.

### T3. `main.go:144` — appel avec gestion d'erreur

```go
if err := db.InitRedis(); err != nil {
	log.Fatal(err)
}
```

### T4. `cmd/e2edb/main.go` — garde + usage + ordre de résolution (D3)

- Résolution de l'index **dans `run()`, avant `recreateDatabase`** (modèle
  `validateTarget`), via une fonction pure testable :

```go
func redisFlushIndex(getenv func(string) string) (int, error) {
	index, err := db.RedisDBIndex(getenv)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(getenv("REDIS_DB")) == "" && index == db.DefaultRedisDB {
		return 0, errors.New("REDIS_DB is unset: refusing to flush the dev Redis index (1)")
	}
	return index, nil
}
```

- `flushRedis(ctx, getenv, index)` reçoit l'index résolu en paramètre (garde
  déjà passée ; `getenv` reste pour `REDIS_HOST`/`REDIS_PASSWORD` ; `FLUSHDB`
  inchangé) ; supprimer la const `redisDBIndex` (lignes 52-54) et son
  commentaire ; import `github.com/Massad/gin-boilerplate/db` ajouté.
- Terminologie et messages (`cmd/e2edb/main.go` est le seul fichier Go du
  périmètre qui nomme un index Redis — règle T8) : doc de paquet (l. 1-3)
  « flushes the Redis database dedicated to the e2e sessions » → « flushes
  the Redis index named by `REDIS_DB` (the e2e sessions index; refuses the
  dev index 1 when `REDIS_DB` is unset) » ; godoc `flushRedis` (l. 145-146) →
  « flushRedis empties the Redis index given by the caller: the session keys
  of a previous run must not survive » ; messages runtime → l. 150
  « REDIS_HOST is empty: the e2e Redis index cannot be flushed », l. 159
  « Flushing the Redis index %d », l. 161 « flush the Redis index %d: %w ».
- En-tête du paquet : `Usage:` (l. 19-21) → `E2E_DB_NAME=artisan_e2e
  REDIS_DB=3 go run ./cmd/e2edb` ; ajouter `REDIS_DB` à l'énumération des
  variables lues (l. 23-27) ; préciser que la cible `e2e` du Makefile
  épingle 3.
- `cmd/e2edb/main_test.go` (périmètre élargi) : nouveaux cas `redisFlushIndex`
  — `REDIS_DB` absent → erreur de garde (sans service Redis) ; `"1"` explicite
  → 1 accepté (opt-in documenté) ; `"3"` → 3 accepté ; `"abc"` → erreur.

### T5. `tests/main_test.go` (nouveau) + `tests/article_test.go` — TestMain (D1)

- Nouveau `tests/main_test.go` (package `tests`, tag `all`, identique en
  mécanisme pour T6) :

```go
//go:build all

// TestMain pins the disposable Redis index for the whole test binary: the
// client must be created while REDIS_DB is present in the process
// environment (godotenv never overrides an already-set variable).
func TestMain(m *testing.M) {
	if err := os.Setenv("REDIS_DB", "2"); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
```

Imports par fichier : `fmt` (chemin d'erreur), `os`, `testing` — pas de
`require` (`TestMain` n'a pas de `*testing.T`).

- Plus d'insertion `t.Setenv` : `db.InitRedis(1)` (ligne 159) →
  `require.NoError(t, db.InitRedis(), "resolving the Redis database index")`
  **+** `require.Equal(t, 2, db.GetRedis().Options().DB, "the suite must run
  on the disposable Redis index 2")`.
- Couvre `tests/login_test.go` : même paquet, la suite est séquentielle et
  partage l'init de `TestIntDB` (ordre alphabétique de définition des
  fichiers) ; `go test -run TestCleanUp` seul laisse le client `nil` — le
  premier accès `GetRedis()` panique alors (préexistant, hors scope).
- Invariant (décision « Index des suites Go ») : le client Redis doit être
  créé tant que `REDIS_DB` est présent dans l'environnement du processus ;
  `TestMain` le garantit pour toute la durée du binaire de test. Aucune
  contrainte d'ordre, aucun commentaire d'ordre dans le code.

### T6. `middleware/main_test.go` + `controllers/main_test.go` (nouveaux) et ajustements

- Deux `main_test.go` avec le même `TestMain` que T5, tous deux avec le tag
  `//go:build all` et les imports `fmt`/`os`/`testing` :
  `middleware/main_test.go` (package `middleware_test`) et
  `controllers/main_test.go` (package `controllers`).
- `middleware/auth_test.go` — `middlewareTestSetup` (l. 21-34) :
  `db.InitRedis(1)` (l. 30) → `require.NoError(t, db.InitRedis(), "resolving
  the Redis database index")` + assertion `Options().DB == 2` (celle de T5) ;
  aucun `t.Setenv` à ajouter (le `TestMain` du paquet pose la variable).
- `controllers/auth_test.go` — **bloc de setup** de
  `TestRefreshOnlyConsumesMatchingOwner` (l. 27-38 ; la fonction s'étend
  jusqu'à l. 145) : `db.InitRedis(1)` (l. 34) → mêmes deux `require`. Aucun
  `t.Setenv` à ajouter (le `TestMain` du paquet pose la variable).

### T7. Chaîne e2e sur l'index 3 + garde CI (D2, D4)

- `Makefile`, recette `e2e` (l. 152-165) : la ligne
  `E2E_DB_ADMIN_USER=… go run -mod=readonly ./cmd/e2edb` est en
  **`Makefile:155`** (pas ~123-125 — ces lignes-là sont le DSN). Préfixe
  posé : `… E2E_DB_NAME=$(E2E_DB_NAME) REDIS_DB=3 go run -mod=readonly
  ./cmd/e2edb`. Le préfixe ne porte que sur cette commande ; l'outil ne
  charge pas `.env` (le préfixe écrase toute valeur héritée du `set -a`).
- `app/playwright.config.ts` webServer (l. 57) :
  `env: { PORT: E2E_PORT, DB_NAME: E2E_DB_NAME, REDIS_DB: "3" }` + une ligne
  de commentaire (l'index 3 isole les sessions e2e des suites Go en 2 et du
  dev en 1 ; la valeur n'est héritée ni du runner ni du `.env`).
- Les autres lignes de la recette e2e ne changent pas : le CLI
  `-create-admin` n'appelle pas `InitRedis` (retour anticipé,
  main.go:112-117), et le webServer Playwright porte le REDIS_DB de l'app e2e.
- **Garde CI (D4)** : `.github/workflows/ci.yml`, étape ajoutée à côté du
  grep « legacy drivers » (l. 84-88), sous la forme YAML exacte (un scalaire
  `run:` commençant par `!` serait parsé comme une balise) :

  ```yaml
        - name: Check no hardcoded Redis index
          run: |
            ! git grep -n "InitRedis(1)" -- '*.go'
  ```

  Filet partiel assumé : après T1 la forme `InitRedis(<entier>)` ne compile
  plus (rupture de signature, l'étape Build/Vet précédant la garde), et
  `git grep` ne voit que les fichiers suivis — la garde attrape donc les
  mentions littérales dans les sources suivies (commentaires compris) et un
  rollback partiel.

### T8. Documentation + mémoire (D6)

Terminologie unique : « index Redis N » partout ; « db N » uniquement dans les
commandes `redis-cli -n N` ; « base » proscrit pour un index. Périmètre de la
règle : les fichiers de documentation listés ci-dessous ET les
commentaires/messages de `cmd/e2edb/main.go` (retouches exactes listées à
T4 — c'est le seul fichier Go du périmètre qui nomme un index Redis).

- `AGENTS.md` :
  - ligne `make e2e` (« FLUSHDB Redis db 1 … purge les sessions de l'app de
    dev » → « FLUSHDB de l'index REDIS_DB (3 dans le cycle e2e) ; les
    sessions dev (index 1) ne sont plus purgées ») ; autres mentions
    équivalentes mises sur la même répartition (bullet E2E des Tests,
    bullet `cmd/e2edb` de la Structure — « purge de la base Redis 1 » →
    « purge de l'index REDIS_DB (3 dans le cycle e2e) »).
  - bullet `db/` de la Structure mis à jour (`RedisDBIndex`,
    `InitRedis() error`).
  - conventions de test (« Tests Redis ») : répartition index 1/2/3 +
    convention D1 — « toute création de client Redis dans un test passe par
    le `TestMain` du paquet (index 2 posé au niveau processus) ».
  - « Variables principales » (ligne ~83) : ajouter `REDIS_DB` (défaut 1) —
    **avertissement : ne pas la poser dans `.env`** (le Makefile exporte
    `.env` à toutes les recettes, `make run` suivrait).
- `README.md` : paragraphe env (près de `REDIS_PASSWORD`, ligne ~60-64) :
  `REDIS_DB` sélectionne l'index (défaut 1 = dev ; suites Go = 2 via
  `TestMain` ; cycle e2e = 3) — **ne pas la poser dans `.env`**. Section
  Testing : une phrase sur la répartition 1/2/3.
- `memory-bank/techContext.md` : ligne 22 (variables d'env : ajouter
  `REDIS_DB`) + ligne 41 (« la base 1 du Redis dédié ne contient plus que ce
  projet » → « index 1 = sessions dev, index 2 = suites Go, index 3 = cycle
  e2e (REDIS_DB) »).
- `memory-bank/activeContext.md` : cibles = **l. 26 et l. 29** (bloc « Session
  2026-11-23 »), pas la ligne 60 (joignabilité WSL). Annotation datée, sans
  réécriture : « (à la date du 2026-11-23 ; l'isolation des index 1/2/3 est
  en vigueur depuis le plan REDIS_DB) ».
- `memory-bank/progress.md` : liste « Ce qui fonctionne » — la ligne qui
  mentionne `FLUSHDB` Redis db 1 / `cmd/e2edb` (l. 19) devient fausse → mise
  à jour « FLUSHDB de l'index REDIS_DB (3 dans le cycle e2e) » ; + **entrée
  numérotée 23** de l'« Historique des étapes » (une phrase : l'isolation des
  index 1/2/3, avec les ancres `db/db.go`, `cmd/e2edb/main.go`, `Makefile`,
  `app/playwright.config.ts`). Les entrées historiques restantes ne sont pas
  retouchées (l'entrée 22 « session Redis db 1 » l. 50 reste vraie : le dev
  reste en index 1).
- `memory-bank/systemPatterns.md` : la mention `db.InitRedis` (l. 12) devient
  fausse après la rupture de signature → « `db.InitRedis() error` »
  (l'index vient de `REDIS_DB`).
- `.kilo/HANDOFF.md:27` : édité en assumant explicitement le statut « section
  encore utilisée comme référence » (« DB 1 dédiée à ce projet » → répartition
  index 1/2/3).

### T9. Validation (jalons par phase)

Les jalons ci-dessous vérifient en particulier les passages sensibles
introduits en dernière passe de correction — toute erreur s'y manifeste en
jalon rouge, pas en faux vert :

- message exact de la garde D3 → phase 1, jalon 4 : exit 1 avec « REDIS_DB
  is unset: refusing to flush the dev Redis index (1) » ;
- preuve D5 (une clé « e2e » n'est pas distinguable par nom) → phase 2,
  étape 5 : index 3 non vide **et** `dbsize` de l'index 2 identique avant
  et après `make e2e` ;
- cycle de vie de l'app dev (binaire, attente bornée, arrêt vérifié) →
  phase 2, étapes 4b/4c/4f : le PID tué est celui du serveur, le port
  9001 est confirmé libéré.

**Phase 1 (T1-T6)** — jalon :
1. `gofmt -l .` → vide ; `go vet -mod=readonly -tags=all ./...` → vert (le
   tag `all` couvre les trois `main_test.go`, invisibles d'un `go vet` sans
   tag).
2. `go test -mod=readonly -count=1 -race ./db` → vert (pur, sans services).
3. `make test` complet (tag `all`) → vert.
4. `make e2e` → **échec attendu à ce stade** : exit 1 avec le message de
   garde « REDIS_DB is unset: refusing to flush the dev Redis index (1) »
   (T7 pas encore posé — état voulu, documenté au § Phasage ; ne pas
   relâcher la garde D3 pour le « faire passer »). Le cycle redevient vert
   à la phase 2.

**Phase 2 (T7)** — toutes les commandes Redis passent par
`docker compose exec -T redis redis-cli …` (le binaire `redis-cli` n'est pas
provisionné sur l'hôte ; le Redis de dev est publié sur 127.0.0.1:16379, pas
6379 — une commande nu interrogerait une autre instance, faux négatif) :
1. `make services_up` puis marqueur :
   `docker compose exec -T redis redis-cli -n 1 set admin:sentinel dev-session-marker`.
2. Contrôle `go test` direct (propriété « hors make ») :
   `go test -mod=readonly -count=1 -tags=all -run TestIntDB ./tests` → vert —
   la preuve de l'index 2 hors make est l'assertion `Options().DB == 2`
   exécutée par ce test ; le contrôle
   `docker compose exec -T redis redis-cli -n 1 get admin:sentinel` → intact
   est un témoin de non-régression (cette sélection n'exerce aucune purge).
3. `make test` complet → vert (les assertions `Options().DB` de chaque suite
   y tournent) ; marqueur → **toujours intact** sur l'index 1 ; `-n 2 keys
   '*'` → résidus éventuels des suites uniquement (observation, pas critère).
4. App dev (preuve que dev vit en 1) — `dev-check@example.com` /
   `dev-check-password-0` sont des littéraux jetables de la preuve (l'upsert
   de `-create-admin` réécrit le mot de passe de ce compte, idempotent) ;
   lancer depuis un shell où `REDIS_DB` est **absent** (sinon l'app
   démarrerait sur un autre index sans erreur) :
   a. `go run . -create-admin -email dev-check@example.com -password dev-check-password-0`.
   b. Démarrage : `go build -mod=readonly -o /tmp/artisan-dev .` puis
      `/tmp/artisan-dev & pid=$!` depuis la racine (le binaire charge `.env`
      lui-même, main.go:97) — pas `make run` : derrière `make`/`go run` se
      cache un processus enfant binaire qu'un `kill` du PID parent ne tue pas
      de façon fiable ; avec le binaire, le PID capturé EST le serveur.
   c. Attente de disponibilité **bornée** avec détection de mort du
      processus (patron du smoke CI, ci.yml:118-129 — pas de boucle infinie) ;
      sonde **sans** `--fail` (toute réponse HTTP = prêt — l'app dev ne sert
      le SPA que si `app/dist` existe, un 404 resterait une réponse valide ;
      refus de connexion = pas prêt) ; branche d'échec explicite post-boucle :

      ```
      for i in $(seq 1 50); do curl -s -o /dev/null --max-time 2 http://127.0.0.1:9001/ && break; kill -0 "$pid" || { echo "le serveur est mort"; exit 1; }; sleep 0.2; done; curl -s -o /dev/null --max-time 2 http://127.0.0.1:9001/ || { echo "app non disponible après 50 tentatives"; exit 1; }
      ```
   d. `curl -s -o /tmp/dev-login.json -w '%{http_code}' -X POST -H
      'Content-Type: application/json' -d
      '{"email":"dev-check@example.com","password":"dev-check-password-0"}'
      http://127.0.0.1:9001/v1/login` → `200` ; puis
      `grep -q '"kind":"superadmin"' /tmp/dev-login.json` → vrai.
   e. `docker compose exec -T redis redis-cli -n 1 keys 'admin:*'` → la
      session superadmin (`admin:<uuid>` ; le marqueur `admin:sentinel`
      figure aussi, à écarter) vit en **1**.
   f. Arrêt **vérifié** avant l'étape 5 : `kill "$pid"` puis confirmer le
      port libéré —
      `curl --silent --max-time 2 http://127.0.0.1:9001/ >/dev/null && { echo "TOUJOURS UP"; exit 1; } || echo arrêté` —
      une app résiduelle réécrirait des sessions en 1 et occuperait le port
      9001.
5. `make e2e` → **3 scénarios CUJ + projet setup verts** (4 tests aujourd'hui :
   1 setup + 3 CUJ — le critère est le statut des scénarios, ne pas figer un
   compte) ; à lancer depuis un shell où `REDIS_DB` est **absent** (la valeur
   3 doit venir du préfixe de la recette, pas de l'environnement ; prérequis
   pnpm/chromium documentés dans AGENTS.md). Preuves d'index :
   `-n 3 keys '*'` → sessions e2e du run (c'est ELLE qui prouve que la
   chaîne e2e vit en 3) ; contrôle négatif de l'index 2 :
   `docker compose exec -T redis redis-cli -n 2 dbsize` **identique juste
   avant et juste après `make e2e`** — aucune écriture e2e en 2. Une clé
   « e2e » n'est pas distinguable par nom (mêmes formes de clés que les
   suites) : l'accord Makefile/Playwright (D5) est prouvé par la paire
   {3 non vide, dbsize(2) inchangé}, pas par un motif de clés. Marqueur →
   **toujours intact** sur l'index 1.
6. Nettoyage du marqueur :
   `docker compose exec -T redis redis-cli -n 1 del admin:sentinel`.

**Phase 3 (T8)** — jalon : `git status` + `git diff` relus ; pas de commit
(convention : commit sur demande explicite).

## Risques / cas d'échec

- **Création du client hors fenêtre `REDIS_DB`** (invariant défini à T5) :
  le client
  Redis doit être créé tant que `REDIS_DB` est présent dans l'environnement
  du processus ; `TestMain` le garantit pour toute la durée du binaire de
  test — une création tardive échouerait sur l'assertion `Options().DB` de la
  suite concernée.
- **`make e2e` lancé à la main sans la cible make** : la commande réelle est
  `E2E_DB_NAME=artisan_e2e REDIS_DB=3 go run ./cmd/e2edb`. Sans
  `E2E_DB_NAME`, l'outil refuse avant tout DROP ; sans `REDIS_DB`, la garde
  D3 refuse explicitement de purger l'index dev (échec explicite). Seule la
  cible make épingle 3.
- **Serveur à `databases` réduit** : l'index est accepté par le parseur puis
  `ERR DB index is out of range` à la première commande ; pas de ping (choix
  assumé) → échec au premier usage, erreur sans mention de `REDIS_DB`.
- **`REDIS_DB` mal choisi (valeur valide hors dépôt, ex. `0`)** : accepté par
  le parseur puis silencieux — l'index 0 n'est utilisé par rien ici, aucune
  commande ne le repousse. L'échec explicite couvre les valeurs *mal
  formées*, pas les valeurs *mal choisies* ; un `REDIS_DB` non reconnu par le
  dépôt déplace app/outil sans trace (nature namespacing, décision
  « Nature de l'isolation »).
- **Résidus sur les index jetables** : les suites nettoient leurs clés mais
  les index 2/3 peuvent garder des orphelins sans TTL — sans conséquence
  (jetables, jamais lus par le dev) ; le FLUSHDB e2e remet l'index 3 à zéro à
  chaque cycle.
- **Repli arrière** : révert git des fichiers ; rien à migrer (les index 2/3
  sont éphémères, l'index 1 n'a jamais bougé).

## Hors scope

- **CI** : une étape de garde ajoutée (D4), le reste inchangé.
- ACL Redis et mot de passe.
- scram PostgreSQL — follow-up documenté dans le plan Docker.
- Le pendant PostgreSQL de la classe de bug corrigée : la purge de la base
  dev par `make test` reste hors périmètre — follow-up nommé dans
  « Hors-objectif assumé » (base jetable `artisan_test`).
- Autres usages de bases Redis.
- Toute écriture hors allowlist (§ Allowlist).

## Allowlist

- **Créés :** `db/db_test.go`, `tests/main_test.go`, `middleware/main_test.go`,
  `controllers/main_test.go`
- **Modifiés :** `db/db.go`, `main.go`, `cmd/e2edb/main.go`,
  `cmd/e2edb/main_test.go` (cas de garde), `tests/article_test.go`,
  `middleware/auth_test.go`, `controllers/auth_test.go`, `Makefile`,
  `app/playwright.config.ts`, `.github/workflows/ci.yml` (une étape),
  `AGENTS.md`, `README.md`, `memory-bank/techContext.md`,
  `memory-bank/systemPatterns.md`, `memory-bank/activeContext.md`,
  `memory-bank/progress.md`, `.kilo/HANDOFF.md`
- **Interdits d'écriture :** `.env`, `.env_rename_me`, `docs/`, `go.mod`,
  `go.sum`, `app/src/**`, `app/e2e/**`, `internal/db/**`, `models/**`,
  `middleware/auth.go`, `controllers/*.go` (hors `auth_test.go`). Exception :
  les fichiers listés dans « Créés » restent autorisés (notamment
  `controllers/main_test.go` couvert par la liste Créés).
