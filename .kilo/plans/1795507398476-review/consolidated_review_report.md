# consolidated_review_report.md — Goldfish Review consolidée

**Document cible :** `.kilo/plans/1795507398476-docker-pg-redis-dev.md` (267 lignes)
**Semantic split :** FeatureRange 1..41 · DetailedRange 42..267
**Preset :** Consolidated report — 5 stages sans checkpoints
**Reviewers :** egm-zoom-out ×2 (S5, S8), egm-inventive-catalyst (S5), egm-comprehension (S6), egm-design (S7), egm-structural (S7), egm-style ×2 (S7, S9), egm-implementation-readiness (S9), egm-codebase-details (S9) — **9 auditeurs en contextes frais, 0 not-run**
**Rapports bruts :** `zoom_out_report.md`, `comprehension_report.md`, `critic_report.md`, `intermediate_zoom_out_report.md`, `implementation_readiness_report.md` (même dossier)

---

## Verdict global

**FAIL contre les standards appliqués** — mais un FAIL sain : le plan résout le bon problème (plus de PG/Redis natif, ports squattés), ses justifications centrales sont vérifiées exactes contre le code (voir §Validé sans réserve), et les livrables sont bien listés. Ce qui échoue : **deux commandes du plan ne s'exécutent pas telles qu'écrites**, une tierce affirmation structurante (« la CI n'est pas affectée ») est fausse, la compatibilité postgres:18-alpine n'est pas assurée par le YAML proposé, et quatre décisions laissées à l'implémenteur violent la Zero-Decision Rule. Le plan devient implementation-ready **après les 5 blockers ci-dessous** — dont aucun ne remet en cause l'architecture choisie (le plus important, B-01, est une ligne de YAML).

## Readiness par task (Stage 9)

| Task | Verdict | Bloquant |
|---|---|---|
| 1 docker-compose.yml | NOT READY | B-01 |
| 2 Makefile | NOT READY | B-02, B-03 |
| 3 .env/.env_rename_me | NOT READY | B-04 |
| 4 README.md | NOT READY | B-04 |

---

## 🔴 Blockers — consolidés

### B-01 — Le volume PostgreSQL est incompatible avec `postgres:18-alpine`; le compose ne démarre pas
- **Trace :** Stage 5 (Z#1, L38/52/60-61) · Stage 9 IR#1 (codebase Dockerfile 18/alpine3.23) · impacte Validation 3/9 et la décision Persistance
- **Fait :** depuis PG 18, l'image place les données dans `/var/lib/postgresql/18/docker`, déclare `VOLUME /var/lib/postgresql` et son entrypoint **refuse de démarrer** devant un montage sur `/var/lib/postgresql/data` (`docker_error_old_databases` → exit 1) ; `--wait` n'atteint jamais `healthy` ; même contourné, les données iraient dans un volume anonyme (contradictoire avec `services_down` préservant les volumes, down -v, redémarrages).
- **Correctif :** `pgdata:/var/lib/postgresql` ou `PGDATA: /var/lib/postgresql/data` en environment ; re-valider l'étape 9 sur un volume réellement attaché.

### B-02 — `make migrate_up` (bootstrap « auto-suffisant ») échoue : le DSN est construit depuis des variables make et aucune cible de migration ne source `.env`
- **Trace :** Stage 5 (catalyst F1, L140-141/149/198-201/222-224) · Stage 8 F2 (quatre parseurs du `.env`, quotes) · Stage 9 IR#2, codebase B2-M1, style B2
- **Fait :** `Makefile:92` lit `$(DB_HOST)/$(DB_PORT)/$(DB_NAME)` ; seuls `test` (:30) et `e2e` (:121-127) sourcent `.env` ; README:69-74 documente aujourd'hui l'export — que la Task 4 *remplace* sans garantie de substitution → `host='' port='' dbname=''`.
- **Correctif (un seul point de lecture, fermé par construction) :** `DB_HOST ?= 127.0.0.1` / `DB_PORT ?= 15432` / `DB_NAME ?= golang_gin_db` + `-include .env` + `export`; valeurs **sans guillemets** dans `.env`/`.env_rename_me` (GNU make ne déquote pas ; la Task 5/IR introduit la première valeur guillemetée du dépôt) ; OU l'export shell dans les recipes avec les variables **shell** ($$). B-03 est le corollaire : sans ça, la CI qui appelle `make migrate_up` (ci.yml:74) hériterait du fix → voir C-01.

### B-03 — « La CI n'est pas affectée : … jamais via make » est factuellement faux → décision à prendre sur le prérequis de `migrate_up`
- **Trace :** unanimement (Z2, catalyst F5, design#2, structural#2, style S7-maj#2, Stage 8 F4, IR#3, codebase B1, style B1) — `ci.yml:74` (`make migrate_up`) et `:82` (`make generate_sqlc`)
- **Fait :** avec `migrate_up: services_up`, chaque run CI tire postgres:18-alpine+redis:8-alpine, publie 15432/16379, attend le healthcheck (job à timeout 15 min) pour une pile inutilisée ; « Rien d'autre ne change » (task 6) et « validée par construction » (L242-244) tombent avec.
- **Correctif :** choisir explicitement — (a) prérequis **retiré** de migrate_up/migrate_down (bootstraps dev : run/test/e2e gardent le leur) + reformuler « la CI exécute make migrate_up » ; ou (b) assumé en CI avec coût chiffré ; ou (c) conditionné (`SERVICES ?= 1` posé à 0 côté CI). Le snippet proposé doit aussi **garder `install_goose`** : `migrate_up: install_goose services_up` (codebase M1 — le snippet actuel supprime le prérequis d'installation du binaire goose).

### B-04 — SSL jamais tranché : le template `SSL=TRUE` casse le flux complet d'un clone neuf
- **Trace :** Z3, IR#4, codebase M2, style B3, structural#11
- **Fait :** `SSL=TRUE` → `sslmode=require` (pool.go:74-77, e2edb:183-185, Makefile:91-92) vers un conteneur **sans TLS** ; `main.go:216-228` RunTLS exige `./cert/myCA.cer/key` (gitignored) ; les étapes 5/6 de Validation tombent.
- **Correctif :** `SSL=FALSE` assumé pour la stack dev (dans `.env_rename_me` + le tableau de décisions), ou documenter le prérequis certificats ; donner le **delta réel** des lignes éditées par fichier (template ≠ `.env` : DB_HOST=localhost vs 127.0.0.1, quotes).

### B-05 — FeatureRange : pas de section « Alternatives considérées » (feature-spec-standard)
- **Trace :** unanimement (design 🔴, structural 🔴, style 🔴 — le seul rejet évalué est PG 19)
- **Correctif :** ajouter une section « Alternatives rejetées » : (a) services natifs sur 5433/6380 ; (b) interpolation compose (`${DB_PORT}`) / split `REDIS_PORT` ; (c) scram + requirepass (coût application nul : les variables sont déjà lues — pool.go:82, db/db.go:15) ; (d) testcontainers-only ; (e) volumes nommés vs bind ; (f) alpine vs debian.
- **NB — recalibration connexe (Stage 8 F1)** : trancher si le livrable est « stack reproductible pour le projet » ou « dépannage local » — deux objectifs incompatibles partagent le plan ; l'arbitrage calibre l'Objectif, la liste de fichiers, les touches CI et le README.

---

## 🟡 Majors — consolidées

| # | Constat | Correctif | Trace |
|---|---|---|---|
| M-01 | `services_up` n'assure que le cluster ; le **schema** manque → `make run` échoue tardivement sur volume neuf (roles manquants, log.Fatal main.go:137). | Contrat explicite écrit (services_up=cluster, migrate_up=fait par B-02) + mention README/AGENTS/Validation OU bootstrap goose in-repo (modèle testutil/postgres.go) ; alternative : check version goose au boot avec message actionnable (interdit l'auto-migration silencieuse : superuser requis). | design#1, Catalyst F7, Structure F3, IR#2 |
| M-02 | Contradiction natif/conteneurs : `.env` fige 15432/16379, README dit « install native restant valide » ; sort de l'ancienne base `golang_gin_db` jamais énoncé. | Désigner le chemin gagnant ; quick-start natif qui re-pointe `DB_PORT`/`REDIS_HOST` ; 1 ligne « ancienne base abandonnée » ; section « undo » 5 lignes. | comprehension#1, Z7, F4 IR#6, stylisation M6, Stage 8 F8, IR#14 |
| M-03 | « Parité dev = CI » : alpine (dev) vs debian (CI) vs Objectif sans suffixe ; `testutil/postgres.go:18` PostgreSQL 16 jamais mentionné ; test containers goes **skip→run** dès que Docker existe ; l'étape 11 « validée par construction » surestime. | Une seule chaîne d'image ou wording « mêmes majors » ; aligner le pin testutil ; mentioner le basculement du test testcontainers dans Tests+étape 8 ; supprimer « validée par construction ». | Z4, F6, F13/N3, comprehension#2, style M2, codebase M5/N6, IR#7 |
| M-04 | `trust` : le modèle de menace ne couvre pas le mirroring WSL2 (superuser + `app_admin BYPASSRLS` exposés, scripts postinstall…) ; `restart: unless-stopped` pérennise l'exposition. | Documenter le vecteur OU passer à scram + requirepass via refactor hors-ligne pool_test (`pgx.ParseConfig` + connexion de contrôle — Stage 8 M4 : ~10 lignes, décisions L34 et Risques L257-260 tombent) ; redémarrage `no` ou procédure off. | design#4/#5, F2/F3, Stage 8 M4/M5, compréhension implicit |
| M-05 | Healthcheck : sans `-h`, `pg_isready -U postgres -d golang_gin_db` peut valider la socket initdb (`listen_addresses=''`) → `--wait` peut relâcher avant TCP ; « miroir CI » approxim (ci.yml:29 sans `-d`). | `pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db` ; aligner la CI ou énoncer le delta. | design#3, codebase M3, stylisation N1, comprehension#5 |
| M-06 | Ports doublement codés (compose + `.env` + template = 3 sites pour le remède) : dérive documentée mais non détectée. | Interpolation `${DB_PORT:-15432}` ; OU split `REDIS_PORT` (2 lignes Go, `net.JoinHostPort`) ; OU cible `make check_ports` + `docker compose config -q` en CI (Stage 8 M6). | Z12, F8, design#6, Stage 8 M2/M6/F6 |
| M-07 | Pas de `name:` compose → collisions clones/worktrees (README:38 = gin-boilerplate) avec `restart: unless-stopped` ; ports gardés par l'orphelin. | `name: artisan-dev` (1 ligne) + symptôme « port is already allocated » dans les Risques. | Stage 8 M7, codebase M4, IR#10 |
| M-08 | Documentation-mémoire incomplète : techContext:41 (ligne Redis entière = justification de l'isolation FLUSHDB), HANDOFF.md:27, activeContext.md:60, progress.md (obligatoire, AGENTS.md:87). | Élargir la Task 7 à ces fichiers. | Z6, style M5, codebase N4, IR#12 |
| M-09 | Ancrages manquants : Compose ≥ 2.17 pour `--wait` ; remède Docker « à documenter » sans tâche réceptrice (RELOCATE dans README Prerequisites : docker + usermod + re-login) ; dates absolues (« ce jour », « il y a 5 jours », endoflife.date) ; compte superadmin unique partagé entre les étapes 5/7 ; ordre de phase des tâches. | Ajouts ciblés dans les tâches 4/5 + Validation. | structural#6, style N4/N2, IR#8/#9/#13, Catalyst F10 |

---

## 🟢 Nits groupées (polish, non-bloquants)

1. **Redis** : retirer « command stock » ou ajouter `command:` ; sémantique de persistance explicite ; `requirepass` coûte nul (REDIS_PASSWORD déjà lu) — Z5, design#2.
2. **Miroirs/exactitudes** : « miroir exact de la CI », « tags identiques », « no-op 1-2 s » non mesuré, « wslrelay » non défini/non vérifié, « pgx/go-redis sans version » (ancrer v5.13.0 / v7.4.1), « unitaires aucun port » (fixture e2edb figée mais impact nul) — CALIBRATE.
3. **Wording** : « vraies credentials »→« credentials » ; « style underscore »→« noms de cibles en snake_case » ; « non découpable proprement »→ contrainte réelle (Compose n'interpole pas de sous-chaînes) ; « démarrer proprement »→ symptôme observable ; DROP la puce « Vite absent ».
4. **Doublons à réduire** : justification trust 3× (garder le mécanisme une fois) ; caveat trust dupliqué (RELOCATE vers Risque) ; cellule Images à scinder ; compose comment attribuer les squatters port par port.
5. **Confort** : `shm_size: 256m` (PG) ; échec Docker pré-empté (check `docker info`/compose version + remède) ; `services_reset` avec CONFIRM ; remède de collision nommant les 3 fichiers.

## ✅ Validé sans réserve par les reviewers (à préserver)

- Justification du **trust** : mécanisme 28P01 de `pool_test.go` vérifié exact (contrôle réussit sous scram, variations doivent se connecter) ; CI pose la même variable.
- Ancres **`InitRedis(1)`** ×4 exactes ; **e2e autonome** (e2edb lit tout du `.env`, aucun ajout nécessaire) ; gardes-fous e2edb.
- Compat **Redis 8 + go-redis v7 en RESP2** (EVAL/rotation) et **migrations stables PG 15→18** (roles plain login, RLS FORCE intacte).
- Choix des **ports 15432/16379 en loopback** et du **Redis sans password** conformes au dépôt.
- **Playwright** : `app/playwright.config.ts` lit le `.env` racine — aucun changement front requis.
- Le plan **corrige un vrai bug inter-stacks** (FLUSHDB partagé) — argument à valoriser (Stage 5 opportun #1).

## Matrice de traçabilité brute → consolidée (extraits)

| Brut | Consolidé |
|---|---|
| Z#1 · IR#1 | B-01 |
| Catalyst-F1 · Stage 8-F2 · IR#2 · CB-B2 · ST-B2 | B-02 |
| Z#2 · F5 · design#2 · structural#2 · style-S7 · IR#3 · CB-B1 · ST-B1 | B-03 |
| Z#3 · IR#4 · CB-M2 · ST-B3 · structural#11 | B-04 |
| design🔴 · structural🔴 · style🔴 | B-05 |
| Stage 8-F1 | (recalibre) B-05 + M-02 |
| design#1 · F7 · Stage 8-M3 | M-01 |
| compreh#1 · Z#7 · F4 · IR#6 · ST-M6 · Stage 8-M8 | M-02 |
| Z#4 · F6 · F13 · compreh#2 · ST-M2 · CB-M5 · IR#7 | M-03 |
| design#4/#5 · F2/F3 · Stage 8-M4 | M-04 |
| design#3 · CB-M3 · ST-N1 · Stage 5 F9 · compreh#5 | M-05 |
| Z#12 · F8 · design#6 · Stage 8-M2/M6 | M-06 |
| Stage 8-M7 · CB-M4 · IR#10 | M-07 |
| Z#6 · ST-M5 · CB-N4 · IR#12 | M-08 |
| structural#6 · ST-N2/N4 · IR#8/#9/#13 · F10 | M-09 |

---

## Prochaines étapes possibles

1. **egm-resolve** sur l'ensemble du document (classer PROPOSE FIX vs CONTEST, appliquer les correctifs approuvés) — les constats B-01..B-05 et les majors M-01..M-09 ont chacun un correctif prescript.
2. Édition manuelle du plan (les correctifs sont listés ici), puis re-audit (max 1 re-pass selon le protocole).
3. Conclure sans remédiation en acceptant les findings résiduels comme trade-offs documentés.
