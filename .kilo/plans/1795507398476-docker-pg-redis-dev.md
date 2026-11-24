# Stack dev Docker : PostgreSQL + Redis (ports 15432 / 16379)

## Contexte

Aucun PostgreSQL/Redis joignable sur 5432/6379 : les ports sont tenus par les
stacks Docker voisines (5432 = honcho-database-1, 6379 = honcho-redis-1,
9000 = teable, 8000 = honcho-api, 8180 = searxng, 3000/2222 = gogs,
42345 = teable-db) et aucun service natif ne tourne dans le WSL Debian. L'API
tourne sur 9001 (`.env`), Vite a son port configuré sur 5173 (inoccupé au moment
de l'inventaire).

Inventaire des ports vérifié le **24 nov. 2026** (ss WSL + netstat Windows) —
libres des deux côtés : 5433, 6380, 9001, 9002, 5173, **15432, 16379**. Ollama
tient le 11434 côté Windows. Le patron local (chaque repli conventionnel s'est
fait squatter) motive des ports distinctifs — disposition complète en
« Alternatives rejetées ».

**Décision utilisateur** : conteneurs PostgreSQL sur **127.0.0.1:15432** et
Redis sur **127.0.0.1:16379**, publiés en loopback uniquement.

**Sort des données existantes** : l'ancienne base native `golang_gin_db`
(goose v2, PG 15.19 natif) est abandonnée — aucune donnée à importer n'a été
identifiée ; si des données locales devaient être conservées, les exporter
avant `make services_up` (pg_dump natif) puis les importer dans le conteneur.

## Objectif

`docker-compose.yml` à la racine, deux services (images officielles
`postgres:18-alpine` / `redis:8-alpine`), lancés individuellement
(`make postgres_up`, `make redis_up`) ou ensemble (`make services_up`) via
make, et démarrés automatiquement par `make run`, `make test` et `make e2e`
(prérequis `services_up` sur ces trois cibles uniquement — **jamais sur les
migrations**, que la CI invoque aussi).

**Périmètre assumé** : livrable = stack dev reproductible possédée par le
projet (compose suivi en git) ; l'install native reste joignable en
re-pointant `DB_PORT`/`REDIS_HOST` dans `.env` (documenté au README).
**La CI est alignée sur les mêmes images** : `ci.yml` passe ses deux tags de
services à `postgres:18-alpine`/`redis:8-alpine` (tâche 7) — parité
littérale : une seule chaîne d'image par dépendance, utilisée partout
(compose, CI, testutil éphémère). La parité porte sur les images ; les ports
CI restent ceux du runner (5432/6379 dans son propre conteneur de service,
sans conflit), le local garde 15432/16379.

## Décisions de design (fixées)

| Sujet | Choix |
|---|---|
| Image PostgreSQL | `postgres:18-alpine` |
| Image Redis | `redis:8-alpine` |
| Choix de version | Vérifié le 24 nov. 2026 via [endoflife.date/postgresql](https://endoflife.date/postgresql) : PG 18 (18.7, 14 mois de point releases, EOL nov. 2030) — PG 19.0 sorti le 19 nov. 2026 = trop frais (aucun point release) ; [endoflife.date/redis](https://endoflife.date/redis) : Redis 8 = ligne standard actuelle (7.4 en maintenance only depuis mai 2025) |
| Parité CI | **Oui, littérale** — `ci.yml` aligné sur les mêmes tags que le compose (`postgres:18-alpine`/`redis:8-alpine`, tâche 7) et le testutil éphémère ; une seule chaîne d'image par dépendance |
| Ports hôtes | `127.0.0.1:15432:5432`, `127.0.0.1:16379:6379` |
| Superuser PG | `postgres/postgres` — épinglé par la cible `e2e` du Makefile (`E2E_DB_ADMIN_*`, Makefile:122) ; `cmd/e2edb` ne fait que lire ces variables |
| Base PG | `POSTGRES_DB=golang_gin_db` (créée au premier boot du conteneur) |
| Rôles `app_user`/`app_admin` | Créés par la migration goose 000002 (inchangé) — jamais en dur dans le compose ; `make migrate_up` est le contrat du schéma |
| Auth PG | `POSTGRES_HOST_AUTH_METHOD: trust` — exigé par `pool_test.go` (la suite de parse DSN connecte des mots de passe volontairement cassés ; sous scram, échec 28P01 sans skip — mécanisme détaillé en Tests). Miroir de la CI et de l'ancien PG natif dev. Contrepartie assumée : voir Risques (vecteur WSL2) et Alternatives (scram = follow-up) |
| SSL dev | `SSL=FALSE` — stack compose sans TLS (le `.env` local l'est déjà, le template est aligné) ; l'API dev répond sur `http://localhost:9001`. TLS natif hors stack : `generate-certificate.sh` au README |
| Redis | Configuration par défaut de l'image (snapshots RDB, volume `/data`) — pas de mot de passe (`REDIS_PASSWORD` vide dans `.env`) |
| Healthchecks | `pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db` (sonde TCP : évite le faux positif sur la socket initdb) et `redis-cli ping` — équivalents CI, sonde TCP en plus côté PG |
| Readiness | `docker compose up -d --wait` dans chaque cible make — **exige Compose ≥ v2.17** |
| Persistance | Volumes nommés `pgdata` **monté sur `/var/lib/postgresql`** (disposition PG 18+ : l'image place PGDATA dans `/var/lib/postgresql/18/docker` et **refuse** un montage sur `/var/lib/postgresql/data`) et `redisdata` ; reset via `docker compose down -v` (documenté, pas de cible dédiée) |
| Projet compose | `name: artisan-dev` — un seul projet compose quel que soit le nom du clone/worktree, sinon les conteneurs orphelins (redémarrés automatiquement) tiennent les ports avec un message trompeur |
| Redémarrage | `restart: unless-stopped` (exposition pérenne assumée — voir Risques) |
| Sync des ports | Ports en dur dans compose + `.env` + template (interpolation refusée — voir Alternatives) avec **détecteur** : cible `check_ports` appelée par `services_up`, échoue si compose et `.env` divergent |

## Alternatives rejetées

| Alternative | Rejetée parce que |
|---|---|
| Services natifs WSL sur 5433/6380 | Plus de services natifs fiables dans ce WSL ; pattern local de squat des replis conventionnels ; pas de reproductibilité pour un nouveau développeur |
| Interpolation compose (`${DB_PORT}`) / split `REDIS_PORT` | Changer le contrat `REDIS_HOST=host:port` documenté (AGENTS.md, README) coûte deux lecteurs Go pour un confort ; la dérive réelle est détectée par `check_ports` |
| scram + `requirepass` Redis | Exige le refactor hors-ligne de `pool_test.go` (changement de code test hors périmètre) ; **follow-up possible** — risque trust documenté en attendant (Risques) |
| testcontainers-only (pas de compose) | Démarrage à chaque commande, pas d'état persistant entre sessions |
| Volumes bind vs nommés | Volumes nommés : reset propre (`down -v`) sans toucher l'arbre git |
| alpine vs debian | alpine retenu (empreinte), adopté partout (compose, CI, testutil) — un seul flavour, plus de delta entre environnements |
| Garder la CI en `postgres:15`/`redis:7` (report) | Écarté : le bump coûte deux lignes de tags et rend la parité littérale (une seule chaîne d'image partout) |

## Tâches

**Phase 1 — Stack** (état stable : `docker compose config -q` passe)

### 1. Créer `docker-compose.yml` (racine)

```yaml
# Dev dependency stack: PostgreSQL + Redis for go-artisan-services.
# Host ports deliberately avoid local ranges squatted by other stacks
# (5432/6379 = honcho, 9000/42345 = teable, 3000/2222 = gogs, ...).
# KEEP IN SYNC with .env: DB_PORT / REDIS_HOST — checked by `make check_ports`.
name: artisan-dev
services:
  postgres:
    image: postgres:18-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_HOST_AUTH_METHOD: trust
      POSTGRES_DB: golang_gin_db
    ports:
      - "127.0.0.1:15432:5432"
    volumes:
      - pgdata:/var/lib/postgresql
    shm_size: 256m
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db"]
      interval: 5s
      timeout: 5s
      retries: 10
    restart: unless-stopped
  redis:
    image: redis:8-alpine
    ports:
      - "127.0.0.1:16379:6379"
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 10
    restart: unless-stopped

volumes:
  pgdata:
  redisdata:
```

**Phase 2 — Makefile** (état stable : `make services_up` démarre une stack saine)

### 2. Makefile — sourcing `.env`, cibles, prérequis

En tête du fichier (après le bloc SHELL) — source unique de configuration, le
même `.env` que l'app (godotenv) ; **valeurs sans guillemets** (GNU make ne
les retire pas, contrairement à godotenv/shell/compose) :

```make
## Single configuration source: the same .env the application reads.
-include .env
export
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_NAME ?= golang_gin_db
```

(Les `?=` respectent l'environnement : la CI définit ces variables dans son
bloc env, un `.env` vide — `touch .env`, ci.yml:66 — ne les écrase pas.)

Nouvelles cibles (noms en snake_case, comme les cibles existantes) ; ajouter
au `.PHONY` : `postgres_up redis_up services_up services_down check_ports`.

```make
## START THE DEV POSTGRESQL CONTAINER (waits for health)
postgres_up:
	@docker compose up -d --wait postgres

## START THE DEV REDIS CONTAINER (waits for health)
redis_up:
	@docker compose up -d --wait redis

## START BOTH DEV CONTAINERS
services_up: check_ports
	@docker compose up -d --wait

## STOP THE DEV CONTAINERS (named volumes preserved; add -v to wipe)
services_down:
	@docker compose down

## FAIL IF COMPOSE AND .ENV DISAGREE ON PUBLISHED PORTS
check_ports:
	@pg="$(DB_PORT)"; redis="$$(printf '%s' "$(REDIS_HOST)" | sed 's/.*://')"; \
	grep -q "\"127.0.0.1:$${pg}:5432\"" docker-compose.yml || { echo "drift: compose postgres != DB_PORT ($$pg)" >&2; exit 1; }; \
	grep -q "\"127.0.0.1:$${redis}:6379\"" docker-compose.yml || { echo "drift: compose redis != REDIS_HOST ($$redis)" >&2; exit 1; }
```

Prérequis sur les cibles consommatrices — **run, test et e2e uniquement**
(jamais les migrations : la CI les invoque aussi, `ci.yml:74`) :

```make
run: services_up
	@echo -e "🚀 Running the application..."
	@go run .

test: services_up
e2e: services_up
```

Les cibles de migration **gardent leur prérequis existant** (pas de
`services_up`) :

```make
migrate_up: install_goose
migrate_down: install_goose
```

`make run`/`make test`/`make e2e` démarrent les conteneurs avant d'agir ;
`make migrate_up` lit le même `.env` que l'app via `-include` (plus d'export
manuel nécessaire) et fonctionne dès que la stack tourne (l'étape 3 de la
validation la précède). Le sous-make récursif de `e2e` retrouve la stack via
le prérequis `e2e: services_up`.

**Phase 3 — Configuration** (état stable : `make migrate_up` réussit de bout en bout)

### 3. `.env` et `.env_rename_me` — lignes exactes, sans guillemets

**`.env_rename_me`** (template — valeurs déjà non guillemetées) :

```
DB_HOST=127.0.0.1
DB_PORT=15432
REDIS_HOST=127.0.0.1:16379
SSL=FALSE
```

**`.env`** (dev réel — dé-guillemeter les valeurs `DB_*`, sans effet pour
godotenv/shell/compose, indispensable pour `-include .env`) :

```
DB_HOST=127.0.0.1
DB_PORT=15432
REDIS_HOST=127.0.0.1:16379
```

(`SSL=FALSE` y est déjà.) Aucune autre variable touchée.

### 4. `internal/testutil/postgres.go` — alignement local (1 ligne)

`postgres:16-alpine` → `postgres:18-alpine` (ligne 18). **Seule modification de
fichier source du plan** : le conteneur éphémère opt-in doit exercer la même
version que la stack (il passe de skip à run dès que Docker existe localement —
voir Tests).

**Phase 4 — Documentation** (état stable : un nouveau dev démarre en trois commandes)

### 5. README.md

- **Prerequisites** (nouvelle puce) : « Docker avec Compose ≥ v2.17 — si
  `docker ps` refuse : `sudo usermod -aG docker <user>` puis re-login » ;
  option recommandée `make services_up` ; **install native toujours valide à
  condition de re-pointer `DB_PORT`/`REDIS_HOST` dans `.env`** (ex. 5432 /
  127.0.0.1:6379).
- **Setup flow** : `make services_up` puis `make migrate_up DB_USER=postgres
  DB_PASS=postgres` — le sourcing `.env` est automatique (l'ancien passage
  `set -a; . ./.env; set +a` est retiré au profit du `-include` du Makefile).
- **Flux SSL** : stack dev sans TLS, `SSL=FALSE`, l'API répond sur
  `http://localhost:9001` (TLS natif via `generate-certificate.sh` inchangé).
- **Section tests** : `make services_up` comme moyen standard.
- **Retour arrière** (nouvelles lignes) : `docker compose down` + retirer les
  prérequis `services_up` (révert git du Makefile) + re-pointer
  `DB_PORT=5432`/`REDIS_HOST=127.0.0.1:6379` pour retrouver une stack native.

### 6. AGENTS.md

- **Commandes** : ajouter `postgres_up`, `redis_up`, `services_up`,
  `services_down`, `check_ports` ; prérequis `services_up` sur `run`,
  `test`, `e2e` (jamais les migrations) ; **contrat du schéma : `services_up`
  démarre le cluster, `make migrate_up` (superuser) crée le schéma — les
  cibles qui démarrent les conteneurs ne dispensent pas de la première
  migration** ; reset destructeur `docker compose down -v` (efface données et
  sessions, re-login requis).
- **Environnement & sécurité** : dev = conteneurs Docker 127.0.0.1:15432/16379
  (compose `name: artisan-dev`), superuser postgres/postgres (conteneur dev
  uniquement), rôles applicatifs créés par la migration 000002, `SSL=FALSE`.

### 7. `.github/workflows/ci.yml` — parité des tags

Dans le job `test` (bloc `services`) : `image: postgres:15` →
`postgres:18-alpine`, `image: redis:7` → `redis:8-alpine`. Rien d'autre ne
change : env, étapes, `POSTGRES_HOST_AUTH_METHOD: trust` (déjà en place) —
le healthcheck CI garde `pg_isready -U postgres` sans `-d` car la base
`golang_gin_db` y est créée par psql **après** le démarrage du service
(ci.yml:70), contrairement au compose où `POSTGRES_DB` la crée au boot.

**Phase 5 — Mémoire du projet**

### 8. memory-bank + HANDOFF

- `techContext.md` : remplacer la **ligne 41 entière** (« PG 15.19 natif
  127.0.0.1:5432 … Redis 127.0.0.1:6379 … DB 1 = cache teable ») par l'état
  réel : stack dev docker PostgreSQL 18 / Redis 8 (ports 15432/16379), API
  sur 9001, Vite 5173, la base 1 du Redis dédié ne contient plus que ce
  projet.
- `activeContext.md` : remplacer les ports périmés (5432/6379) par les nouveaux.
- `progress.md` : état de session (convention AGENTS.md:87).
- `.kilo/HANDOFF.md` : la note « PG 15.19 natif » devient la stack docker.

## Tests : ports, bases et commandes (vérifiés dans le code)

- **Contrat de démarrage** : `services_up` = cluster sain ; `make migrate_up`
  (superuser) = schéma ; `run`/`test`/`e2e` démarrent les conteneurs mais ne
  dispensent pas de la première migration — un volume neuf sans migration se
  manifeste à la première requête, pas au boot (les pools ne font que pinger).
- **Unitaires** (sans tag) : aucune dépendance runtime à un port ou une base.
  `invoice` exige `pdftotext` (poppler) — **vérifié présent le 24 nov. 2026
  (v25.03.0)** : attendu vert ; sur un poste sans poppler, l'échec est à
  signaler comme environnement, non régression.
- **Intégration** (`make test`, tag `all`) : le Makefile lit le même `.env`
  que l'app → conteneurs 15432/16379, base `golang_gin_db`, **Redis db 1**
  (`InitRedis(1)` partout : app `main.go:144`, `middleware/auth_test.go:30`,
  `controllers/auth_test.go:34`, `tests/article_test.go:159`).
- **`pool_test.go` (parse DSN)** : la boucle connecte des mots de passe
  volontairement cassés ("", espaces, quotes) — c'est l'oracle du parseur pgx ;
  le contrôle aux vraies identifiants réussit sous scram, donc pas de skip :
  **trust est exigé** (sous scram, la boucle échoue en 28P01). Sous trust, la
  boucle prouve le parsing du DSN, pas le round-trip du mot de passe — limite
  assumée (voir Alternatives : scram + refactor hors-ligne = follow-up).
- **Testcontainers** (`internal/testutil/postgres_test.go`, tag `all`, opt-in) :
  passe de **skip** (Docker injoignable) à **run** (pull postgres:18-alpine)
  dès que la stack Docker est en place — comportement nouveau de `make test`,
  attendu.
- **E2E** (`make e2e`) : `cmd/e2edb` lit `DB_HOST`/`DB_PORT`/`REDIS_HOST` de
  `.env` (→ conteneurs) et le superuser épinglé par la cible `e2e` du Makefile
  (`E2E_DB_ADMIN_USER/PASS=postgres/postgres`) = conteneur ; le cycle recrée
  `artisan_e2e` de façon autonome (drop/create FORCE, garde-fous anti-DROP,
  migrations dans le cycle) — **aucune commande de génération à ajouter**.
  `FLUSHDB` Redis db 1 → conteneur dédié : corrige un bug réel inter-stacks
  (le FLUSHDB e2e purgeait le cache d'une autre application sur le Redis
  partagé — techContext l'enregistrait comme « DB 1 = cache teable ») ; l'effet
  de bord « purge des sessions dev » reste, isolé.
- **Compatibilité app** : exigence repo « PostgreSQL ≥ 13 »
  (`gen_random_uuid()`) satisfaite ; RLS/policies/triggers/`set_config`
  stables 15→18 ; pgx/v5 (v5.13.0, go.mod:14) supporte PG 18 ; client go-redis
  v7 (v7.4.1, go.mod:10) compatible Redis 8 en RESP2 (EVAL, rotation
  atomique — commandes inchangées).
- **Accès depuis Windows** : 15432/16379 sont miroités par le
  localhost-forwarding WSL2 (observé sur les ports docker existants — netstat,
  PID wslrelay) ; vérifié à l'étape 7 de la validation, pas seulement affirmé.

## Validation (dans l'ordre)

1. `docker compose config` — YAML valide ; `make check_ports` passe (compose
   et `.env` d'accord).
2. Pré-vol — **vérifié le 24 nov. 2026** : `sylvaindeb` ∈ groupe `docker`
   (gid 990), `docker ps` répond sans sudo, Docker Compose **v5.7.0** (≥ v2.17
   requise pour `--wait`). Le remède reste documenté au README Prerequisites
   pour d'autres postes (`sudo usermod -aG docker <user>` + re-login).
3. `make postgres_up` puis `make redis_up` individuellement :
   `docker compose ps` → `healthy` ; second appel idempotent.
4. `make migrate_up DB_USER=postgres DB_PASS=postgres` → applique
   000001-000003, crée les rôles ; le sourcing `.env` est celui du Makefile
   (`-include`).
5. `go run . -create-admin -email admin@example.com -password <dev-pass>`
   → prouve le pool admin à travers 15432.
6. `make run` → conteneurs healthy + app sur `:9001` (http, `SSL=FALSE`) ;
   `curl http://127.0.0.1:9001/v1/articles` → 401 « Please login first ».
7. Login e2e manuel (POST /v1/login, `admin@example.com` + **le même mot de
   passe qu'à l'étape 5**) → session écrite dans le Redis conteneur (vérif
   `docker compose exec redis redis-cli -n 1 keys '*'` — l'app utilise la
   db 1, `main.go:144`) ; depuis Windows, vérifier l'accès à
   `127.0.0.1:15432` (localhost-forwarding).
8. `make test` complet (tag `all`) — inclut `TestPostgresDSNPasswords` (passe
   grâce à trust) et désormais le test testcontainers (run au lieu de skip,
   pull postgres:18-alpine). `./invoice` est attendu vert (poppler vérifié
   présent, pdftotext 25.03.0) ; si un poste sans poppler échoue sur ce seul
   paquet, le signaler comme environnement et non régression. Tout doit
   passer.
9. `make services_down` → stop ; `make services_up` → relance, données
   conservées (les sessions survivent — volume `pgdata` réellement attaché à
   `/var/lib/postgresql`).
10. Optionnel : `make e2e` cycle complet contre les conteneurs (valide le
    prérequis `e2e: services_up` et la recréation autonome d'`artisan_e2e`).
11. Preuve de parité : le `make test` local de l'étape 8 tourne sur PG 18 +
    Redis 8 — les mêmes tags d'images que la CI (`ci.yml` aligné, tâche 7) ;
    le push suivant le confirme sur le runner.

## Risques / cas d'échec

- **Docker daemon arrêté** ou Compose < v2.17 : `make run`/`make test`/
  `make e2e` échouent bruyamment avant d'agir — comportement voulu, remède
  dans le README (Prerequisites).
- **Collision future de port** (une stack arrive sur 15432/16379) : compose
  échoue au bind avec un message explicite ; remède : changer le port dans
  les **trois** fichiers — `docker-compose.yml`, `.env`, `.env_rename_me` —
  en sync.
- **Dérive compose/`.env`** : détectée par `make check_ports` (exit 1 avec le
  détail), appelé par `services_up` avant chaque démarrage — pas seulement un
  commentaire.
- **`down -v`** : destructif (base + sessions Redis) — documenté, pas de
  cible make dédiée pour éviter les accidents de complétion.
- **Trust auth dev** : le localhost-forwarding WSL2 expose les ports publiés
  aux processus Windows locaux (scripts, autres distros) — un PostgreSQL
  superuser et un Redis sans mot de passe y sont joignables, et
  `restart: unless-stopped` pérennise l'exposition après chaque reboot.
  Assumé pour une machine de dev (même posture que l'ancien PG natif et que
  la CI) ; le basculement scram + requirepass est documenté en Alternatives
  comme follow-up (refactor hors-ligne de `pool_test.go` requis). Ne jamais
  réutiliser ce compose pour un environnement exposé.
- **Second clone/worktree** : `name: artisan-dev` garantit un seul projet
  compose — sinon les conteneurs orphelins du premier checkout (redémarrés
  automatiquement) tiendraient les ports avec un message trompeur
  (« port is already allocated »).

## Hors scope

Dockerfile applicatif, conteneurisation de l'app/prod, TLS/ACL Redis, mots de
passe forts en dev (superuser `postgres` + trust volontaires, dev local
uniquement), recettes internes des cibles `test`/`e2e` (seuls des prérequis
make sont ajoutés, les recettes ne changent pas).
