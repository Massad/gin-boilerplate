# zoom_out_report.md — Stage 5 (Initial Zoom Out)

**Target:** `.kilo/plans/1795507398476-docker-pg-redis-dev.md` (267 lines, full file)
**Reviewers:** egm-zoom-out (Skeptical Architect) + egm-inventive-catalyst
**Mode:** REVIEW — Consolidated pipeline, Stage 5 of 5-stage flow
**GoldfishContext:** AGENTS.md, Makefile, ci.yml, pool_test.go, db/db.go, main.go, cmd/e2edb/main.go, .env_rename_me, README.md, memory-bank/techContext.md

---

## Reviewer 1 — egm-zoom-out (Skeptical Architect)

**Verdict : FAIL.** L'objectif du plan est clair et la partie « trust / e2e / ports » est solide, mais le `docker-compose.yml` proposé **ne démarre pas** avec `postgres:18-alpine`, et trois affirmations du document contredisent les fichiers du dépôt.

### 1. 🔴 Blocker — le montage du volume Postgres est incompatible avec `postgres:18-alpine`
**Lignes :** 38, 52, 60-61 (impacte les étapes de validation 220-221 et 237-238)
**Description :** Monter le volume nommé `pgdata` sur `/var/lib/postgresql/data` est la disposition des images ≤ PG 17 ; depuis PG 18 l'image place ses données dans `/var/lib/postgresql/18/docker` et déclare `VOLUME /var/lib/postgresql`, et son point d'entrée **refuse de démarrer** quand un montage existe sur `/var/lib/postgresql/data`.
**Justification :** le Dockerfile officiel `18/alpine3.23` pose `ENV PGDATA /var/lib/postgresql/18/docker` + `VOLUME /var/lib/postgresql`, et `docker-entrypoint.sh` (`docker_setup_env` → `docker_error_old_databases`) ajoute `/var/lib/postgresql/data (unused mount/volume)` puis `exit 1` → le conteneur meurt, `docker compose up -d --wait` n'atteint jamais `healthy`, et les étapes 3 à 10 tombent toutes ; même en contournant ce garde-fou, les données partiraient dans un volume anonyme et non dans `pgdata`, ce qui annule la décision « Persistance » (L38) et l'étape 9 (« données conservées »).
**Correctif :** monter le volume nommé sur `/var/lib/postgresql` (comme le recommande le message d'erreur de l'image) ou poser `PGDATA=/var/lib/postgresql/data` dans l'environnement du service.

### 2. 🟡 Major — la CI est affectée, contrairement à ce qu'affirme le plan
**Lignes :** 122-127, 129-131 (et 166-171)
**Description :** « La CI n'est pas affectée : elle appelle `go test` et goose directement, jamais via make » est faux, la CI appelle bien `make migrate_up`.
**Justification :** `.github/workflows/ci.yml:74` exécute `make migrate_up DB_USER=postgres DB_PASS=postgres` → le nouveau prérequis `services_up` fera démarrer la pile compose sur le runner (une seconde paire PostgreSQL/Redis inutile sur 15432/16379, temps runner perdu) et fera échouer l'étape « Configure disposable test services » tant que le point 1 n'est pas corrigé.
**Correctif :** retirer `services_up` du prérequis de `migrate_up`, ou assumer explicitement le démarrage compose dans la CI.

### 3. 🟡 Major — `SSL=TRUE` (template livré) est incompatible avec le conteneur, et le plan n'en parle pas
**Lignes :** 133-141, 143-153, 206-210, 222-228
**Description :** Le plan livre `.env_rename_me` sans modifier `SSL=TRUE` et réécrit le parcours d'installation du README, alors que le conteneur PostgreSQL alpine n'active pas TLS.
**Justification :** `.env_rename_me:3` vaut `SSL=TRUE`, `Makefile:91` en déduit `sslmode=require` pour goose, `internal/db/pool.go:75` et `cmd/e2edb/main.go:184` font de même pour les DSN applicatifs, or le Dockerfile `18/alpine3.23` ne génère aucun certificat et laisse `ssl = off` → un nouveau dev qui suit le README (`.env` sourcé) ne peut plus migrer, et `main.go:216-225` lance `RunTLS` avec `cert/myCA.cer`/`cert/myCA.key`, absents d'un clone frais (`cert/` est gitignoré, `.gitignore:14`) donc l'app se termine sans message ; le dépôt pose déjà `SSL: "FALSE"` pour son propre conteneur éphémère (`internal/testutil/postgres.go:90`).
**Correctif :** aligner le template sur `SSL=FALSE` (stack dev sans TLS) et l'inscrire dans le tableau de décisions.

### 4. 🟡 Major — le bump CI `15→18` / `7→8` est hors problème et sans justification de risque
**Lignes :** 29, 166-171, 206-210
**Description :** La montée de version des images de la CI n'est pas nécessaire au problème traité (plus de PG/Redis natif) et est présentée comme une simple « parité dev = CI », sans évaluation de risque ni réflexion sur les versions supportées.
**Justification :** dev et CI n'échangent rien (seul le tag change) et le seul point réellement couplé est le test de parsing DSN qui exige `trust` (`ci.yml:25`, `internal/db/pool_test.go:15-32`) et fonctionne sur toute majeure ; pendant ce temps le README annonce toujours « PostgreSQL 13+ » (`README.md:30`) alors qu'aucun environnement ne teste plus de majeure < 18, et le dépôt contient déjà une troisième image épinglée (`postgres:16-alpine`, `internal/testutil/postgres.go:18`) que le plan ignore.
**Correctif :** justifier la montée ou la reporter ; ne pas la présenter comme de la parité.

### 5. 🟡 Major — le tableau de décisions et le YAML se contredisent sur Redis
**Lignes :** 35 vs 68-79
**Description :** Le tableau annonce un « command stock » pour Redis, mais le service `redis:` du compose ne contient aucun `command:`.
**Justification :** sans `command`, le conteneur démarre avec la configuration par défaut de l'image (snapshots RDB par défaut) — le document se contredit sur la sémantique de persistance qu'il prétend avoir fixée (le chemin `/data` du volume `redisdata`, lui, est correct).
**Correctif :** ajouter la commande ou supprimer la mention du tableau.

### 6. 🟡 Major — la mise à jour de `techContext.md` oublie la note qui devient fausse
**Lignes :** 174-178
**Description :** La tâche 7 ne remplace que les notes « PG 15.19 natif » et « port 9000 occupé », laissant en place la note Redis du même fichier qui décrit un partage de base désormais inexistant.
**Justification :** `memory-bank/techContext.md:41` affirme « Redis 127.0.0.1:6379 sans password (DB 1 = cache teable — cleanup de tests ciblé sur `user_sessions:*`/`admin:*`) », alors qu'après le plan Redis est dédié (16379) et que la base 1 ne contient plus que les sessions du projet — ce que le plan confirme lui-même (L196-197) ; laisser l'ancienne phrase entretient une information périmée.
**Correctif :** remplacer toute la ligne « Environnement local », pas seulement deux fragments.

### 7. 🟡 Major — « l'install native restant valide » n'est plus vrai après le changement de ports
**Lignes :** 143-153 (notamment L147), 133-141
**Description :** Le README doit continuer à présenter l'installation native de PostgreSQL/Redis comme une voie valide, alors qu'après le passage à `DB_PORT=15432` / `REDIS_HOST=127.0.0.1:16379` l'application ne dialogue plus avec une instance native.
**Justification :** `.env` et `.env_rename_me` ne pointeront plus que sur les conteneurs, donc une installation native sur 5432/6379 n'est plus jointe qu'à condition de rééditer `.env` — la formulation prévue par le plan est trompeuse.
**Correctif :** préciser « valide uniquement si vous adaptez `DB_PORT`/`REDIS_HOST` », ou assumer le conteneur comme source unique.

### 8. 🟢 Nit — quatre cibles make là où deux suffisent
**Lignes :** 86-106 (et 88)
`postgres_up` et `redis_up` ne sont que des variantes de `services_up`. Chaque cible coûte une entrée `.PHONY` (Makefile:11), une ligne AGENTS.md, une ligne README et une ligne de mémoire de projet.

### 9. 🟢 Nit — le plan décrit un `DB_HOST` que le template ne contient pas
**Lignes :** 140
« `DB_HOST=127.0.0.1` inchangé » ne correspond pas au template (`DB_HOST=localhost`, `.env_rename_me:8`) ; un hôte résolvant d'abord en `::1` prend un refus avant repli — à aligner ou documenter.

### 10. 🟢 Nit — version minimale de Compose non précisée
**Lignes :** 217-219
« compose v2 requise pour `--wait` » sans version minimale ; sur un poste ancien l'échec est opaque au lieu de dire « Compose ≥ X requis ».

### 11. 🟢 Nit — le miroir WSL→Windows est affirmé sans vérification
**Lignes :** 211-212
« 15432/16379 seront miroités côté Windows » présenté comme un fait, sans étape de validation, et dépend de l'endroit où tourne le démon Docker.

### 12. 🟢 Nit — duplication des ports évitable pour PostgreSQL
**Lignes :** 40, 49, 135-138
Le plan refuse toute substitution `${...}`, mais pour PostgreSQL `${DB_PORT:-15432}` supprimerait un des deux points de synchronisation (l'argument REDIS_HOST `host:port` ne vaut que pour Redis).

### Vérifié sans réserve (aucun constat)

- **Auth `trust`** : justification exacte (`pool_test.go:35` mots de passe invalides, contrôle 21-32 ne se déclenche qu'après échec d'auth ; CI pose la même variable `ci.yml:25`).
- **Références `InitRedis(1)`** : exactes (main.go:144, middleware/auth_test.go:30, controllers/auth_test.go:34, tests/article_test.go:159).
- **Cycle e2e autonome** : `cmd/e2edb/main.go:23-27, 169-174` ; « FLUSHDB db 1 isolé » exact.
- **Base `golang_gin_db`** : cohérente avec `pool_test.go:48`.
- **Images** : `postgres:18-alpine` (18.7) et `redis:8-alpine` existent ; critère « PG 19 trop frais » cohérent.
- **Ports loopback** : publication `127.0.0.1:15432/16379` = bon réflexe ; Redis sans mot de passe conforme à `REDIS_PASSWORD=`.

---

## Reviewer 2 — egm-inventive-catalyst

**Mode:** REVIEW (strategic audit). Trois arbitrages acceptés comme des lois — (a) « la base doit ignorer les mots de passe », (b) « chaque cible make doit exiger Docker », (c) « le port doit être écrit deux fois ». Deux affirmations factuelles fausses sur le repo, une commande documentée non fonctionnelle.

### 🔴 F1 — La commande de migration documentée ne peut pas fonctionner telle qu'écrite
**Lignes :** 140-141 (§3), 149 (§4), 198-201 (§Tests), 222-224 (Validation 4)
Le plan propose un `make migrate_up DB_USER=postgres DB_PASS=postgres` « auto-suffisant », mais `migrate_up` ne source jamais `.env` : seul `test` (Makefile:30) et `e2e` (Makefile:121,124,127) sourcent `.env`. `Makefile:92` construit le DSN depuis `$(DB_HOST)/$(DB_PORT)/$(DB_NAME)` → la commande brute donne `host='' port='' dbname=''` et goose échoue, malgré le démarrage des conteneurs par le prérequis `services_up`.
**Résolution proposée (Self-Service)** : rendre le sourcing uniforme — ajouter le prologue `set -a && . ./.env && set +a` à `migrate_up`/`migrate_down`, ou poser des défauts `DB_HOST ?=` / `DB_PORT ?=` / `DB_NAME ?=`. Tolérer un `.env` vide (la CI fait `touch .env`, ci.yml:66).

### 🟡 F2 — « Trust » présenté comme une loi de nature, alors que seul le design d'un test l'impose — et trust désarme ce test
**Lignes :** 34, 257-260, 264-266
`pool_test.go:25-49` utilise des **connexions réelles** comme oracle d'un comportement de **parseur** : sous `trust`, les quatre mots de passe mutés se connectent quel que soit le round-trip du mot de passe ; pendant ce temps `internal/testutil/postgres.go:39` démarre un `postgres:16-alpine` **sans** trust (scram par défaut) et se connecte correctement avec les rôles créés par la migration 000002.
**Résolution proposée (Extraction)** : découpler l'assertion parseur du mode d'auth du serveur — `pgx.ParseConfig(dsn)` hors-ligne (`cfg.Password == password`) + une connexion de contrôle réelle pour le round-trip. Alors le compose peut passer en `scram-sha-256` avec les mots de passe dev existants, et « trust » disparaît du repo (y compris `ci.yml:25`).
**Liabilités :** scram impose un mot de passe à chaque `psql`/`docker exec` ad hoc ; l'assertion hors-ligne ne prouve pas un connect de bout en bout (mitigé par la connexion de contrôle).

### 🟡 F3 — L'argument « loopback only » ne couvre pas l'exposition Windows que le plan introduit lui-même
**Lignes :** 15-16, 30, 211-212, 257-260
Le compose bind `127.0.0.1:15432/16379` mais L211-212 annonce un « mirroring automatique » côté Windows via un relay dont l'adresse de bind n'est pas contrôlée par le compose — un PostgreSQL superuser sans mot de passe et un Redis sans auth peuvent être joignables hors de la machine.
**Résolution proposée (Beforehand Cushioning)** : ne pas faire porter une garantie d'authentification par une garantie de transport : corriger F2 (scram) et documenter explicitement le bind du relay.

### 🟡 F4 — « Chaque cible auto-suffisante » contredit le README du plan lui-même (« install native reste valide »)
**Lignes :** 116-127, 145-147
Avec `services_up` en prérequis dur de `run`/`test`/`e2e`/`migrate_up`/`migrate_down`, Docker devient obligatoire : `make test` ne démarre plus sans daemon — la voie native ne survit que pour `go run .` et `go test` directs.
**Résolution proposée (Self-Service + cushioning)** : prérequis **conditionnel** — `services_up` sonde d'abord le endpoint configuré et n'appelle `docker compose` que si rien ne répond ; par contre le comportement devient implicite (l'opérateur ne voit plus lequel a répondu).

### 🟡 F5 — « La CI n'est pas affectée » est factuellement faux : `ci.yml:74` appelle `make migrate_up`
**Lignes :** 129-131
Chaque run CI démarrerait la pile compose (double pile PG/Redis inutilisée, dépendance à Compose v2, pulls d'images) alors que la vraie base CI est son service container sur 5432.
**Résolution proposée (Segmentation)** : bootstrap conteneurs sur les cibles dev uniquement (`run`/`test`/`e2e`), laisser `migrate_up`/`migrate_down` libres, ou le conditionner à une variable explicite posée par la CI.

### 🟡 F6 — « Parité dev = CI prouvée par construction » est inexacte : trois versions de PostgreSQL coexistent et deux flavours d'image
**Lignes :** 29, 166-171, 203-205, 242-244
`internal/testutil/postgres.go:18` épinglait encore `postgres:16-alpine` (testcontainers opt-in — qui se mettra à EXÉCUTER dès que Docker existe localement) ; compose = alpine, CI = debian ; CI crée la base à la main (`ci.yml:70`) alors que le compose passe par `POSTGRES_DB` — deux chemins de bootstrap différents.
**Résolution proposée (consistency)** : une seule chaîne d'image partout (`postgres:18-alpine` incluant ci.yml et testutil) + provision CI identique au compose ; attention au testcontainers qui doit alors être rejoué (AGENTS.md:74 interdit les skips silencieux).

### 🟡 F7 — « make run est auto-suffisant » est faux sur un volume neuf, et l'échec arrive tard et confusément
**Lignes :** 38-39, 116-120, 237-238
`main.go:137-141` ne ping que les pools (`pool.go:35`) — un ping réussi ne dit rien du schéma ; avec `restart: unless-stopped` + volume `pgdata`, un schéma manquant ou périmé remonte à la première requête, pas au boot.
**Résolution proposée (Self-Service honest)** : échec explicite plutôt qu'automatisme — soit `run` dépend de `migrate_up` (goose épinglé, idempotent), soit un check de version goose au boot qui sort avec « run make migrate_up » ; noter la contrainte : la migration 000002 crée des rôles cluster, l'auto-migration exigerait un superuser dans `.env` — régression de sécurité que le plan évite.
**Résolution proposée** : check explicite, pas auto-migration silencieuse.

### 🟡 F8 — Le port écrit deux fois par design, et la raison du refus d'interpolation ne tient pas
**Lignes :** 40, 49, 136-137, 250-252
Compose lit le `.env` voisin pour l'interpolation : `"127.0.0.1:${DB_PORT:-5432}:5432"` réutilise le `DB_PORT` que lit l'app (`pool.go:84`, `e2edb/main.go:190`, `Makefile:92`) sans changement Go ; pour Redis, la séparation n'est que deux lignes de Go (`db/db.go:14`, `e2edb/main.go:148`).
**Résolution proposée (Intermediary)** : `.env` propriétaire des nombres — compose interpole `${DB_PORT}`, `REDIS_HOST` devient hôte seul + `REDIS_PORT` (`net.JoinHostPort`) ; si Go hors scope, au minimum une cible `make check-ports` qui échoue quand compose et `.env` divergent — elle ferme exactement le risque listé L250-252.
**Liabilités :** interpolation = dépendance au `.env` gitigné (fichier absent ⇒ fallback silencieux et `docker compose config` valide les mauvaises valeurs) ; le split `REDIS_HOST` change le contrat documenté (AGENTS.md:81, README.md:59-60) et doit toucher les deux lecteurs Go + le template.

### 🟢 F9 — Les healthchecks ne sont pas le « miroir CI » revendiqué
**Lignes :** 36, 63, 75, 170-171
Compose : `pg_isready -U postgres -d golang_gin_db` ; CI : `pg_isready -U postgres` (ci.yml:29) — le compose prouve la base applicative, la CI non. Garder le check compose plus strict et aligner la CI (avec le `CREATE DATABASE` avant), ou énoncer la différence.

### 🟢 F10 — L'échec Docker est la stratégie d'erreur : « échoue bruyamment » pris comme diagnostic acceptable
**Lignes :** 218-219, 248-249
Un check préalable `docker info` + version compose avec remède en deux lignes (start Docker / plugin compose v2) coûte peu et épargne la recherche.

### 🟢 F11 — Le reset destructeur n'a pas d'entrée documentée
**Lignes :** 38, 103-105, 253-254
Proposition : `services_reset` refusant de tourner sans token explicite (`make services_reset CONFIRM=wipe`), sinon imprime la commande exacte.

### 🟢 F12 — `shm_size` absent du service PostgreSQL
**Lignes :** 50-67
L'image livre `/dev/shm` par défaut (64 MB) — symptôme classique « could not resize shared memory segment » sur workers parallèles. `shm_size: 256m` : une ligne, sans inconvénient.

### 🟢 F13 — La step de preuve surestime ce qu'un run local vert prouve
**Lignes :** 242-244
`Makefile:34` tourne sans `-race`, `ci.yml:100` avec ; bases et provisioning diffèrent — reformuler « les mêmes majeures sont exercées » ou aligner les flags.

### Opportunités inventives (non-défauts)

1. **Le conteneur corrige silencieusement un vrai bug inter-stacks — le dire.** `techContext.md:41` (« DB 1 = cache teable ») + AGENTS.md:9,58 : chaque cycle e2e flushait le cache d'une autre application sur le Redis partagé. Le conteneur dédié supprime ce danger (L191-197) mais le plan ne le présente que comme « l'effet de bord reste, isolé ». Le nommer serait l'argument le plus fort du changement de ports.
2. **Segmentation optionnelle dev/test des données :** `make test` purge la base de dev `golang_gin_db` (AGENTS.md:57, tests/article_test.go:164) — désormais dans un volume persistant. Base de test séparée = follow-up (hors scope d'un plan de stack ; préexistant).
3. **Le compose comme foyer du contrat dev complet** : un profil compose pour une base de test jetable, laissant la base applicative intacte.

### Vérifié exact (aucune finding)

- Les 4 sites `InitRedis(1)` cités (main.go:144, middleware/auth_test.go:30, controllers/auth_test.go:34, tests/article_test.go:159).
- `pool_test.go` exige bien un serveur trust-auth (contrôle Pool_test.go:25-32 réussit avec le vrai mot de passe, puis les mots de passe mutés doivent aussi se connecter).
- Aucune commande de génération à ajouter pour e2e : Makefile:119-132 (e2edb → migrate_up → -create-admin → build → Playwright) et `e2edb` lit `E2E_DB_ADMIN_*` épinglé postgres/postgres (Makefile:122) = superuser du compose.
- Surface Redis utilisée par l'app (GET/DEL/SET/SREM/SMEMBERS/KEYS + EVAL/EVALSHA `models/auth.go:262-277`) supportée par Redis 8 en RESP2 avec go-redis v7 (`db/db.go:6`).
- Migrations stables PG 15→18 (`gen_random_uuid()`, `timestamptz`, `set_updated_at()`, RLS FORCE + NULLIF 000002:70-78 ; rôles plain login, BYPASSRLS seulement pour app_admin).

---

## Synthèse orchestrator (Stage 5)

- **2 🔴 Blockers distincts** : (1) compose incompatible postgres:18-alpine (montage `/var/lib/postgresql/data`) — Z1 ; (2) `migrate_up` sans sourcing `.env` → la commande de bootstrap du plan échoue — F1.
- **Accords inter-reviewers** : CI affectée via `make migrate_up` (Z2 = F5) ; « parité » à reformuler/nuancer (Z4 ≈ F6, F13) ; install native vs cibles docker-only à arbitrer (Z7 ≈ F4) ; duplication des ports contestée (Z12 ≈ F8).
- **Divergences** : sur trust, le zoom-out valide la justification « exigée par pool_test.go », le catalyseur conteste que la contrainte soit auto-infligée et propose un refonte du test (hors scope d'un plan de stack, mais à noter pour la discussion de remédiation).
