# comprehension_report.md — Stage 6 : Comprehension Test

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md` (207 lignes)
**Périmètre :** FeatureRange 1..47 · DetailedRange 48..207 (paraphrase sur l'ensemble)
**Reviewer :** egm-comprehension — Naive Reader (contexte frais, lecture naïve)

---

## What this proposes / Ce que le document propose

Le document propose de séparer les espaces Redis du projet en deux indexes logiques : la base Redis **1** devient la base **exclusive du développement** (sessions de l'app lancée par `make run`), tandis que la base **2** accueille à la fois les tests Go et le cycle e2e. Concrètement, il faut ajouter une fonction de parsing pure et partagée `db.RedisDBIndex(getenv)` qui lit `REDIS_DB` (absent ou vide ⇒ 1 ; sinon entier dans [0,15], sinon erreur), faire porter `InitRedis()` par cette résolution au lieu d'un variadic `selectDB ...int`, faire échouer bruyamment l'app et l'outil `cmd/e2edb` sur une valeur invalide, faire porter la valeur `2` par les fichiers de test eux-mêmes (`t.Setenv("REDIS_DB","2")`), par la recette `make e2e` (préfixe `REDIS_DB=2` sur l'appel `cmd/e2edb`) et par le `webServer` de `app/playwright.config.ts`, puis mettre à jour la documentation (AGENTS.md, README.md, memory-bank, HANDOFF) et valider par une série de commandes (`gofmt`, `go vet`, `go test ./db`, `make test`, `make e2e`) incluant un « marqueur » Redis destiné à prouver que la purge e2e n'atteint plus la base de dev.

## How the existing system works, per this document / Comment le système existant fonctionne, d'après ce document

Aujourd'hui, `db.InitRedis(selectDB ...int)` ouvre un client go-redis sur l'index 0 par défaut, et tous les appelants passent explicitement `1` : l'app (`main.go:144`), les suites `middleware`/`controllers`, et la suite `tests` (`tests/article_test.go:159`). En parallèle, `cmd/e2edb` porte une constante `redisDBIndex = 1` et lance un `FLUSHDB` sur cet index à chaque cycle e2e ; les suites Go, de leur côté, ne font pas de `FLUSHDB` mais suppriment globalement les clés `user_sessions:*` et `admin:*` (TestCleanUp et TestUnifiedLoginCleanUp) — les mêmes noms de clés que ceux utilisés par l'app de dev, sur la même base. Le document pose donc comme invariant que toute écriture ou purge de session des tests doit se faire sur un index distinct de celui du dev, sans changer aucun comportement métier (rotation atomique des refresh, préfixes de clés, TTL), et sans toucher aux recettes `make test`, à la CI, au `.env`, ni aux annotations Swagger.

---

## Findings

### 1. [Major] T7 — ancrage de ligne faux dans la recette `make e2e` du Makefile
**Portée :** T7 (l. 133-142).
**Écart :** le document indique « appel e2edb (lignes ~121-122) ». Dans le Makefile réel, l'appel est en **ligne 155** ; les lignes 121-122 contiennent le `migrate_down` de goose. L'écart est de 34 lignes : un lecteur qui suit l'ancre à la lettre modifie le mauvais bloc. Corriger en 155 ou désigner la cible par son nom (`e2e`).

### 2. [Major] T9 — les commandes `redis-cli` de validation ne sont pas exécutables telles qu'écrites
**Portée :** T9, étapes 3 à 6 (l. 173-183).
**Écart :** l'étape 3 utilise `docker compose exec -T redis redis-cli -n 1 set sentinel …` (dans le conteneur), mais les étapes 4-6 utilisent `redis-cli -n 1 get sentinel` etc. **sans hôte ni port**. Le Redis de dev est publié sur `127.0.0.1:16379` (pas 6379) : un `redis-cli` nu sur l'hôte interroge le mauvais port et renvoie une réponse vide — le lecteur conclura à tort que « l'isolation a échoué ». Le document ne dit ni où exécuter `redis-cli` (hôte vs conteneur) ni quel port viser. La recette de validation est le cœur de la preuve : elle doit être reproductible au caractère près.

### 3. [Major] Décisions/Alternatives — l'hypothèse « jamais concurrents » est posée comme un fait, sans règle ni risque
**Portée :** l. 26, l. 44, section Risques (l. 187-201).
**Écart :** la justification de l'index unique repose sur « les commandes sont séquentielles » — propriété du monde énoncée nulle part comme règle du dépôt, absente de la section Risques. Rien n'interdit de lancer `make e2e` pendant un `make test` : le `FLUSHDB` e2e porterait alors sur la base 2 et effacerait les sessions des suites Go en cours (401 aléatoires, flaky). Le lecteur ne saura pas qu'il franchit une règle non écrite.

### 4. [Major] Risques — l'invariant réel n'est pas celui décrit, et l'effet de bord de `t.Setenv` n'est pas expliqué
**Portée :** Risques, 1re puce (l. 188-192) ; décision l. 30 ; T5 (l. 115-122).
**Écart :** le document présente l'ordre `t.Setenv` **avant** `godotenv.Load` comme le point critique — mais cet ordre n'est pas l'invariant (godotenv n'écrase rien ; `t.Setenv` posé après donnerait le même résultat). L'invariant réel est « le client Redis doit être créé pendant que `REDIS_DB=2` est présent dans l'environnement du processus ». Le document ne dit nulle part que **`t.Setenv` restaure la valeur précédente à la fin du test** : dans le paquet `tests`, `REDIS_DB=2` n'existe que pendant `TestIntDB`. Conséquence non signalée : un nouveau fichier de test du même paquet qui appellerait `db.InitRedis()` après `TestIntDB` résoudrait l'index par défaut (1) — précisément la « dérive silencieuse vers db 1 » que la ligne 28 déclare impossible. Le document donne une garantie plus large que ce qu'il n'explique.

### 5. [Major] T5 — la protection du paquet `tests` repose sur un prérequis d'ordre d'exécution jamais énoncé
**Portée :** T5 (l. 115-122) : « même paquet, la suite est séquentielle et partage l'init de `TestIntDB` ».
**Écart :** l'affirmation est vraie mais inexpliquée : `tests/login_test.go` n'initialise rien (il consomme `db.GetRedis()`), seul `TestIntDB` crée le client. Le document ne dit pas : (a) quels tests doivent tourner avant les autres, (b) que l'ordre de Go est l'ordre de définition, (c) que `go test -run TestX ./tests` seul laisse le client `nil` (panique), (d) que `-p 1` du Makefile sérialise les paquets. Rejouer une seule suite (usage courant) tombe sur une panne non expliquée par le document.

### 6. [Nit] Contexte — « validation fenêtre 5 » est un terme d'équipe non défini
**Portée :** l. 5.
**Écart :** renvoie à un artefact de workflow interne non défini ni dans le document ni dans le dépôt. À supprimer ou expliciter.

### 7. [Nit] Risques — la commande citée dans le risque « e2e à la main » ne peut pas produire le flush décrit
**Portée :** Risques, 2e puce (l. 193-196).
**Écart :** lancé tel quel, `go run ./cmd/e2edb` s'arrête **avant** le flush : l'outil refuse un `E2E_DB_NAME` vide (`main.go:106`) et l'ordre du code est `validateTarget` → `recreateDatabase` → `flushRedis`. Le scénario suppose implicitement `E2E_DB_NAME` déjà exporté — condition non mentionnée. La commande de référence d'un risque doit être celle qui déclenche réellement le risque.

### 8. [Nit] Décisions — le domaine accepté par le parseur (0..15) n'est pas entièrement mappé, et la borne 15 suppose une config Redis particulière
**Portée :** l. 27 ; T1 (l. 56-66) ; T2 (l. 88-92).
**Écart :** seules les bases 1 et 2 ont un rôle défini ; 0 et 3..15 restent sans signification. La borne 15 n'est vraie que si le serveur a `databases 16` : sur un Redis configuré avec `databases 1` (usage production/managé), `REDIS_DB=2` est accepté par le parseur puis échoue côté client avec une erreur qui ne mentionne pas `REDIS_DB`. Le document ne décrit pas ce cas.

### 9. [Nit] Objectif — « rotation atomique » est un mécanisme projet non défini
**Portée :** l. 20.
**Écart :** renvoie au script Lua `consumeRefreshAuth` (`models/auth.go:262-277`) sans l'expliquer ni le relier, alors qu'il sert d'argument central (« rien ne change sauf l'index »). Un lecteur naïf ne peut pas vérifier la promesse.

### 10. [Nit] T9 étape 5 — la validation du login dev n'est pas reproductible telle qu'écrite
**Portée :** T9, étape 5 (l. 178-181).
**Écart :** « `POST /v1/login` → 200 » sans URL/port (app dev sur 9001), sans corps de requête, sans commande curl. Le lecteur doit reconstituer l'étape depuis `controllers/auth.go` et le `.env`.

### 11. [Nit] Contexte/Objectif — le symptôme visible par l'utilisateur n'est jamais énoncé ; « base » vs « index » dérive
**Portée :** l. 7-20 ; T8.
**Écart :** le document ne dit jamais ce que le développeur **voit** (re-login requis après `make test`/`make e2e`) — c'est pourtant le bénéfice de l'Objectif. Par ailleurs le plan écrit indifféremment « db Redis 1 », « base Redis 1 », « base Redis REDIS_DB » pour un **index** logique (0-15), ambiguïté que T8 propage dans trois fichiers de documentation.

---

**Synthèse :** aucun Blocker — toutes les sections sont paraphrasables et leur enchaînement cohérent. Les constats se concentrent sur la reproductibilité (ancre T7, commandes redis-cli T9), les prérequis implicites (invariant `t.Setenv`/`TestIntDB`, ordre d'exécution, hypothèse de non-concurrence) et des termes d'équipe non définis.
