# Fenêtre de résolution — corrections de la revue Goldfish sur le plan REDIS_DB

**Objet :** appliquer au plan `../1795535384311-redis-db-isolation.md` toutes les corrections issues de la revue Goldfish consolidée. Cette fenêtre est auto-contenue : un agent **neuf** (aucun historique) lit ce brief + les fichiers listés, exécute **strictement** le périmètre, prouve les critères de sortie.

**Périmètre : UNIQUEMENT le fichier plan** `\\wsl.localhost\Debian\home\sylvaindeb\projects\go-artisan-services\.kilo\plans\1795535384311-redis-db-isolation.md` (réécriture des sections selon le programme ci-dessous). Aucun fichier source, aucune doc, aucun rapport n'est modifié.

**Entrée (rejouer avant de commencer) :** le plan existe (207 lignes, titre « Isolation REDIS_DB… ») ET les 5 rapports existent dans ce dossier (`consolidated_review_report.md`, `zoom_out_report.md`, `comprehension_report.md`, `critic_report.md`, `intermediate_zoom_out_report.md`, `implementation_readiness_report.md`).

**Lecture obligatoire avant d'éditer :** le plan en entier + `consolidated_review_report.md` en entier (la liste maîtresse des findings M-01..M-12 + nits). Les autres rapports sont consultables pour le détail d'un finding (référencés par trace).

**Décisions D1-D6 — FIXÉES, ne pas rediscuter :**

| # | Décision |
|---|---|
| D1 | **`TestMain` par paquet taggé `all`** posant `REDIS_DB=2` au niveau processus + assertion d'index par suite. Plus aucun `t.Setenv("REDIS_DB", …)`, plus aucune contrainte d'ordre avec `godotenv.Load` |
| D2 | **Index 3 pour la chaîne e2e** (préfixe make + webServer Playwright) ; 2 = suites Go ; 1 = dev (défaut). Décision utilisateur |
| D3 | **Garde active dans `cmd/e2edb`** : refus de purger l'index de dev quand `REDIS_DB` est absent ; résolution de l'index dans `run()` **avant** `recreateDatabase` ; `Usage:` et énumération des variables mis à jour |
| D4 | **Garde CI ajoutée** : `! git grep -n "InitRedis(1)" -- '*.go'` (modèle legacy drivers) — la décision « CI inchangée » devient « une étape de garde ajoutée, le reste inchangé » ; Hors scope amendé en conséquence |
| D5 | **Littéraux d'index reconnus** : les valeurs 2 (TestMain) et 3 (chaîne e2e) restent des littéraux nus, reconnus explicitement dans le plan ; leur accord est prouvé par les vérifications T9 (étapes 7/8 ci-dessous). Pas de constante Go, pas de détecteur de dérive |
| D6 | **Sections historiques annotées datées** : `activeContext.md:26/29` reçoivent une annotation « (à la date du 2026-11-23 ; l'isolation db 1/db 2/db 3 est en vigueur depuis le plan REDIS_DB) » sans réécriture ; `HANDOFF.md:27` est édité en assumant explicitement le statut « section encore utilisée comme référence » |

**Faits vérifiés (ground truth, docs officielles via Context7 — injecter tels quels aux re-auditeurs, ne pas relire le module cache) :** go-redis `Client.Options() *Options` retourne les options passées à `NewClient` (champ `DB`) ; godotenv `Load` ne surcharge jamais une variable déjà posée (seul `Overload`) ; les db logiques Redis sont du *namespacing* (« Don't use them to run multiple unrelated applications in a single Redis instance »), `SELECT` hors plage → `ERR DB index is out of range`.

---

## Programme de corrections (à appliquer intégralement au plan)

### P1. Décisions de design — table réécrite
Ligne « Répartition des index » devient : **« 1 = dev (défaut, `REDIS_DB` absent), 2 = suites Go (posé par `TestMain`), 3 = cycle e2e (préfixe de la recette `make e2e` + `webServer` Playwright) — deux index jetables distincts : le FLUSHDB e2e ne peut jamais croiser les sessions des suites Go, ni inversement »**. La ligne « Échec loud » devient « Échec explicite » (même forme que `config.ValidateSigningSecrets`, préfixe `error: `). Supprimer « inconditionnel », « jamais concurrents », « l'app qu'il sert » (→ « l'app du cycle e2e »), « hériteront » (→ « poseront eux-mêmes »). Ajouter deux lignes **« Nature de l'isolation »** : (a) index logique sur une instance unique, mêmes identifiants, aucun ACL — protège des purges accidentelles, pas d'un accès malveillant (la doc officielle Redis qualifie les db logiques de *namespacing*) ; (b) ne jamais pointer `REDIS_DB` vers un index partagé avec une autre application ; en production la variable reste absente (défaut 1). Ajouter une ligne « Rupture d'API » : `InitRedis(selectDB ...int)` → `InitRedis() error` = rupture de signature pour tout importateur externe du module. Ligne « Un seul parseur » : ajouter au tableau des rejets « parseur dans `config/` — rejeté : `db` est le point d'accès unique au client Redis (32 lignes, go-redis déjà importé par l'outil), coût de transitivité nul ». Reconnaissance D5 : une ligne « les littéraux 2 et 3 sont écrits en dur ; leur accord est prouvé par les vérifications T9 ».

### P2. Contexte / Objectif — restructure FeatureRange (M-10, M-07, nits)
- Supprimer « (validation fenêtre 5 + lecture du code) » (garder les références fichier:ligne, auto-vérifiables) ; énoncer le **symptôme utilisateur** (« le développeur devait se reconnecter après `make test` ou `make e2e` ») et le **bénéfice** (« une session dev survit désormais aux deux cycles ») ; remplacer « rotation atomique » par « rotation atomique des refresh (script Lua `consumeRefreshAuth`, `models/auth.go:262-277`) ».
- Insérer une section **`## Objectifs et hors-objectifs`** entre Contexte et Décisions : objectif + critère de succès (la sentinel T9) + non-objectifs déplacés hors de la table des décisions (recette `test` inchangée, `.env` inchangé, Swagger inchangé, comportement des sessions inchangé) + **une ligne « Hors-objectif assumé : `make test` purge toujours la base PostgreSQL dev (`DELETE FROM public.tenant`, `tests/article_test.go:909`) — follow-up nommé : base jetable `artisan_test`, symétrique d'`artisan_e2e` »** (M-07).
- Scinder la table : `## Décisions de design (fixées)` (architecture/câblage) et `## Ce qui ne change pas` (constats d'inaction).
- Table des alternatives : ajouter une ligne **« RETENU : 1/2/3 »** en tête ; **inverser la ligne « Deux index (test=2, e2e=3) »** qui devient le choix retenu (motif : deux acteurs destructeurs globaux ne partagent pas un index — la prémisse « jamais concurrents » n'est pas garantie) ; ajouter une ligne « Index partagé (2) + règle écrite — rejeté : un invariant documenté ne remplace pas une séparation d'index (une règle peut être violée, un entier non) ».
- En fin de FeatureRange, un renvoi : « Risques et limites → § Risques / cas d'échec ; périmètre exclu → § Hors scope ».

### P3. T1 `db/db.go` (M-11, nits)
Code inchangé sur le mécanisme, mais : message d'erreur → `fmt.Errorf("error: REDIS_DB %q must be an integer between 0 and 15", raw)` ; commentaire godoc → « the default Redis range 0-15 (`databases 16` is the stock setting of the dev/CI containers; the server's `databases` directive remains authoritative) ». Ajouter `const DefaultRedisDB = 1` (exportée, utilisée par le défaut du parseur et par la garde D3).

### P4. T2 `db/db_test.go`
Cas : `""` → 1, `"   "` → 1, `"2"` → 2, `" 2 "` → 2, `"0"` → 0, `"15"` → 15 (frontière haute valide), `"-1"`/`"16"`/`"abc"`/`"1.5"` → erreurs ; formes `Atoi` documentées comme acceptées (`"+2"` → 2, `"02"` → 2, `"-0"` → 0 — cas de test inclus). `package db` (interne). **Ajouter le test du chemin d'erreur loud (sans service) :**
```go
t.Setenv("REDIS_DB", "16")
require.Error(t, db.InitRedis(), "an out-of-range REDIS_DB must fail loudly")
```

### P5. T5/T6 — remplacer le mécanisme `t.Setenv` par `TestMain` + assertions (D1, M-03)
- **Trois nouveaux fichiers** `main_test.go` (tag `//go:build all`) : `tests/main_test.go` (package `tests`), `middleware/main_test.go` (package `middleware_test`), `controllers/main_test.go` (package `controllers`), chacun :
```go
//go:build all

// TestMain pins the disposable Redis index for the whole test binary: the
// client must be created while REDIS_DB is present in the process
// environment (godotenv never overrides an already-set variable).
func TestMain(m *testing.M) {
	os.Setenv("REDIS_DB", "2")
	os.Exit(m.Run())
}
```
- T5 : plus d'insertion `t.Setenv` — `db.InitRedis(1)` → `require.NoError(t, db.InitRedis(), "resolving the Redis database index")` **+** `require.Equal(t, 2, db.GetRedis().Options().DB, "the suite must run on the disposable Redis index 2")`. Conserver la note séquencement en l'explicitant : ordre de définition des fichiers alphabétique ; `go test -run TestCleanUp` seul laisse le client `nil` (préexistant, hors scope).
- T6 : idem (`middlewareTestSetup`, bloc de setup de `TestRefreshOnlyConsumesMatchingOwner` l. 27-38 — préciser « bloc de setup », la fonction s'étend jusqu'à l. 145).
- **Invariant partout** (T5, Risques, décisions) : « le client Redis doit être créé tant que `REDIS_DB` est présent dans l'environnement du processus ; `TestMain` le garantit pour toute la durée du binaire de test, indépendamment de `godotenv.Load` (qui n'écrase jamais une variable déjà posée). » **Supprimer toute contrainte d'ordre et son commentaire de code.**

### P6. T4 `cmd/e2edb` — garde + usage + ordre (D3, M-06)
- Résolution de l'index **dans `run()`, avant `recreateDatabase`** (modèle `validateTarget`), via une fonction pure testable :
```go
func redisFlushIndex(getenv func(string) string) (int, error) {
	index, err := db.RedisDBIndex(getenv)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(getenv("REDIS_DB")) == "" && index == db.DefaultRedisDB {
		return 0, errors.New("REDIS_DB is unset: refusing to flush the dev Redis database (index 1)")
	}
	return index, nil
}
```
`flushRedis` reçoit l'index résolu (garde déjà passée, `FLUSHDB` inchangé).
- `cmd/e2edb/main_test.go` (périmètre **élargi**, voir allowlist) : cas `redisFlushIndex` — absent → erreur de garde (sans Redis) ; `"1"` explicite → 1 accepté (opt-in documenté) ; `"3"` → 3 accepté ; `"abc"` → erreur.
- En-tête : `Usage:` → `E2E_DB_NAME=artisan_e2e REDIS_DB=3 go run ./cmd/e2edb` ; ajouter `REDIS_DB` à l'énumération des variables lues ; préciser que la cible `e2e` du Makefile épingle 3.

### P7. T7 — chaîne e2e sur l'index 3 (D2)
- Ancre corrigée : **`Makefile`, recette `e2e` (l. 152-165), la ligne `E2E_DB_ADMIN_USER=… go run -mod=readonly ./cmd/e2edb` est en `Makefile:155`** (pas ~121-122 — ces lignes sont le DSN). Préfixe à poser : `… E2E_DB_NAME=$(E2E_DB_NAME) REDIS_DB=3 go run -mod=readonly ./cmd/e2edb`. Préciser : le préfixe ne porte que sur cette commande ; l'outil ne charge pas `.env` (le préfixe écrase toute valeur héritée du `set -a`).
- `app/playwright.config.ts` webServer env → `env: { PORT: E2E_PORT, DB_NAME: E2E_DB_NAME, REDIS_DB: "3" }` + commentaire (l'index 3 isole les sessions e2e des suites Go en 2 et du dev en 1 ; la valeur n'est héritée ni du runner ni du `.env`).

### P8. T8 — documentation (D6, M-01, nits)
- `activeContext.md` : cibles = **l. 26 et l. 29** (bloc « Session 2026-11-23 »), pas la ligne 60 (joignabilité WSL). Application D6 (annotation datée, pas de réécriture).
- `HANDOFF.md:27` : édité en assumant explicitement « section encore utilisée comme référence ».
- `progress.md` : destination précisée — liste « Ce qui fonctionne » + **entrée numérotée 23** de l'« Historique des étapes ».
- `AGENTS.md` : ajouter `REDIS_DB` aux « Variables principales » (défaut 1 ; **avertissement : ne pas la poser dans `.env`** — le Makefile exporte `.env` à toutes les recettes, `make run` suivrait) ; bullet `db/` de la Structure mis à jour (`RedisDBIndex`, `InitRedis() error`) ; mise à jour des mentions « FLUSHDB Redis db 1 » → « FLUSHDB de l'index REDIS_DB (3 dans le cycle e2e) » ; convention D1 : « toute création de client Redis dans un test passe par le `TestMain` du paquet (index 2 posé au niveau processus) ».
- Terminologie unique dans T8 et partout : **« index Redis N »** (« db N » uniquement dans les commandes `redis-cli -n N` ; « base » proscrit pour un index).

### P9. T9 — validation exécutable (M-02, M-04, D5)
Réécrire T9 en 3 phases (M-08) avec ces étapes :
1. **Phase 1 (T1-T6)** — jalon : `gofmt -l .` vide ; `go vet -mod=readonly ./...` vert ; `go test -mod=readonly -count=1 -race ./db` vert (pur) ; `make test` complet vert.
2. **Phase 2 (T7)** — étapes, **toutes via `docker compose exec -T redis redis-cli …`** (le binaire `redis-cli` n'est pas provisionné sur l'hôte ; le Redis de dev est publié sur 16379, pas 6379) :
   - `make services_up` puis `docker compose exec -T redis redis-cli -n 1 set sentinel dev-session-marker` ;
   - **contrôle `go test` direct** (la propriété « hors make ») : `go test -mod=readonly -count=1 -tags=all -run TestIntDB ./tests` → vert ; `-n 1 get sentinel` → intact ;
   - `make test` complet → vert (les assertions `Options().DB` de chaque suite y tournent) ; `-n 1 get sentinel` → **toujours intact** ; `-n 2 keys '*'` → résidus des suites uniquement (observation, pas critère) ;
   - app dev : `go run . -create-admin -email dev-check@example.com -password <mot de passe dev>` (idempotent) ; `make run` **en arrière-plan** ; `curl -s -X POST -H 'Content-Type: application/json' -d '{"email":"dev-check@example.com","password":"<mot de passe dev>"}' http://127.0.0.1:9001/v1/login` → 200 `kind=superadmin` ; `docker compose exec -T redis redis-cli -n 1 keys 'admin:*'` → la session vit en **1** ; stop de l'app ;
   - `make e2e` → **3 scénarios CUJ + projet setup verts** (ne pas figer « 4/4 ») ; `-n 1 get sentinel` → **toujours intact** ; `-n 3 keys '*'` → sessions e2e du run (preuve que la chaîne e2e vit en 3) ; `-n 2 keys '*'` → **aucune clé e2e** (preuve de l'accord Makefile/Playwright, D5).
3. **Phase 3 (T8)** — jalon : `git status` + `git diff` relus ; pas de commit (convention).

### P10. Risques / Hors scope
- Risques : supprimer la puce « ordre `t.Setenv`/`godotenv.Load` » (remplacée par l'invariant P5) ; réécrire la puce « e2e à la main » : la commande réelle est `E2E_DB_NAME=artisan_e2e REDIS_DB=3 go run ./cmd/e2edb` (sans `E2E_DB_NAME`, l'outil refuse avant tout DROP) ; la garde D3 fait que `REDIS_DB` absent = refus explicite ; ajouter une puce **« serveur à `databases` réduit »** (index accepté par le parseur puis `ERR DB index is out of range` à la première commande, pas de ping → échec au premier usage, erreur sans mention de `REDIS_DB`).
- Hors scope : amendé — « CI : une étape de garde ajoutée (D4), le reste inchangé » ; ajouter le pendant PostgreSQL (M-07, formulation P2) ; scinder la ligne ambiguë (« ACL Redis et mot de passe ; scram PostgreSQL — follow-up documenté dans le plan Docker ; autres usages de bases Redis : hors périmètre »).

### P11. Nits résiduels (une passe)
« Échec loud » → « échec explicite » partout · doublons purgés (argument « mêmes clés, même db » **une seule fois**, dans la table des alternatives ; argument « dérive silencieuse » une seule fois, idem ; justification godotenv **une seule fois**, dans le commentaire de `TestMain`) · « pas d'héritage implicite » → « la valeur n'est héritée ni du runner ni du `.env` » · `-race` aligné sur la convention (T9 phase 1) · message d'assertion `InitRedis` → « resolving the Redis database index » · ligne « Hors scope » F-11 scindée (P10).

### P12. Allowlist (remplace la liste implicite du plan ; M-08)
- **Créés :** `db/db_test.go`, `tests/main_test.go`, `middleware/main_test.go`, `controllers/main_test.go`
- **Modifiés :** `db/db.go`, `main.go`, `cmd/e2edb/main.go`, `cmd/e2edb/main_test.go` (cas de garde), `tests/article_test.go`, `middleware/auth_test.go`, `controllers/auth_test.go`, `Makefile`, `app/playwright.config.ts`, `.github/workflows/ci.yml` (une étape), `AGENTS.md`, `README.md`, `memory-bank/techContext.md`, `memory-bank/activeContext.md`, `memory-bank/progress.md`, `.kilo/HANDOFF.md`
- **Interdits d'écriture :** `.env`, `.env_rename_me`, `docs/`, `go.mod`, `go.sum`, `app/src/**`, `app/e2e/**`, `internal/db/**`, `models/**`, `middleware/auth.go`, `controllers/*.go` (hors `auth_test.go`)

### P13. README.md (T8, complément)
Paragraphe env : `REDIS_DB` sélectionne l'index (défaut 1 = dev ; suites Go = 2 via TestMain ; cycle e2e = 3) ; **ne pas la poser dans `.env`**. Section Testing : une phrase sur la répartition 1/2/3.

## À NE PAS appliquer (remèdes rejetés par le Stage 8 — toute réintroduction est une erreur)
Namespace `E2E_REDIS_DB` complet + constantes jumelles (R1) · parseur déporté dans `config/` ou `internal/redisconf` (R2 — ligne de rejet ajoutée au tableau, P1) · détecteur de dérive relisant Makefile/playwright depuis un test Go (R3) · purge scopée des suites Go sur leurs seules clés (R4 — follow-up nommé, pas ici) · 4e index, `PING` au boot (R5).

## Sortie (critères)
1. Chaque finding consolidé M-01..M-12 est résolu dans le plan **ou explicitement rejeté avec motif** ; la passe nits P11 est appliquée.
2. **Re-audit (cap 1)** : dispatcher deux reviewers frais (`explore`, lecture seule) sur le DetailedRange du plan corrigé — personas `\\wsl.localhost\Debian\home\sylvaindeb\projects\kilo\agents\implementation_readiness.md` et `…\codebase_details.md`, standard `…\projects\kilo\resources\document_standards\implementation-readiness-standard.md`, ground truth Context7 (bloc ci-dessus) injecté dans les prompts, interdiction de lire le module cache, rapports en français. Critère de sortie : **0 Major bloquant l'exécution** (nits tolérés) ; sinon corriger et re-auditer une dernière fois (cap).
3. `open_plan` sur le plan corrigé + `plan_exit` ; message final : chemin du plan, verdict du re-audit, décisions appliquées (D1-D6 avec D2=index 3).

**Rappel de protocole : ne jamais dépasser le périmètre (le fichier plan uniquement) ; un commit n'intervient que sur demande explicite de l'utilisateur.**
