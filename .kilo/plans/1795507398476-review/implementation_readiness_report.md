# implementation_readiness_report.md — Stage 9 (Implementation Readiness Test)

**Target:** DetailedRange 42..267 (Tâches 1-7, Tests, Validation, Risques, Hors scope)
**Reviewers:** egm-implementation-readiness + egm-codebase-details + egm-style
**Standards:** implementation-readiness-standard.md (×2), persona rubric (style)
**GoldfishContext:** the 10 baseline files (+ pool.go, testutil/postgres.go, migration 000002, playwright.config.ts, .kilo/HANDOFF.md, go.mod)

---

## Verdict global : **FAIL** au regard du standard Implementation-Readiness (Zero-Decision Rule)

Les tasks 1, 2, 3, 4, la Validation et la section Tests ne sont pas exécutables telles qu'écrites ; les tasks 5, 6 (sous réserve) et 7 le sont.

### Readiness par task

| # | Task (lignes) | Verdict | Bloquants |
|---|---|---|---|
| 1 | docker-compose.yml (44-84) | **NOT READY** | Volume `pgdata` ↔ postgres:18-alpine incompatible (IR#1) ; nits #10 (name:), #11 (tags flottants) |
| 2 | Makefile (86-131) | **NOT READY** | DSN migrate_up sans `.env` (IR#2) ; effet CI non assumé (IR#3) ; snippet perd `install_goose` (codebase M1) |
| 3 | .env/.env_rename_me (133-141) | **NOT READY** | SSL non tranché (IR#4) ; baseline factuellement fausse + lignes exactes manquantes (IR#5) |
| 4 | README.md (143-153) | **NOT READY** | SSL (IR#4) ; natif vs ports figés (IR#6) — le « remplacement » du passage export détruit la seule garantie de DSN |
| 5 | AGENTS.md (155-164) | READY WITH FIXES | ne pas reproduire l'auto-suffisance ni omettre SSL |
| 6 | ci.yml (166-171) | READY WITH FIXES | prérequis compose s'applique à la CI ; tags flottants |
| 7 | memory-bank (173-178) | READY | étendre à progress.md + HANDOFF + ligne entière techContext |
| — | Tests (180-212) | READY WITH FIXES | bascule testcontainers skip→run ; auto-suffisance revendiquée fausse |
| — | Validation (214-244) | **NOT READY** | étapes 4 (#2), 6 (#4), 9 (#1) non exécutables en l'état |

### Conformité au standard Implementation-Readiness
- **Zero-Decision Rule : violée** (DSN de migrate_up, valeur SSL, lignes exactes par fichier, chemin dev gagnant).
- **Target Scope & File Boundary : satisfaite** — liste de fichiers complète, aucune modification code applicatif.
- **Actionable Steps : partielle** — Task 1 YAML, Task 2 recipes et Task 6 tags exécutables ; tasks 3/4/5/7 + Validation laissent des décisions.
- **Test Plan : présent** mais trois étapes non vérifiables en l'état.
- **Phasing & Checkpoints : faible** — liste plate, ordre implicite (appliquer la Task 3 avant les Tasks 1-2 casse la session).

---

## egm-implementation-readiness — findings

### 🔴 1. Volume PostgreSQL incompatible avec `postgres:18-alpine` (L60-61, L38, L237-238, L253-254)
`18/alpine3.23/Dockerfile` : PGDATA déplacé de `/var/lib/postgresql/data` vers `/var/lib/postgresql/18/docker`, `VOLUME /var/lib/postgresql` ; `docker compose down` supprime le volume anonyme → données/rôles migrés perdus, contradiction avec « Persistance : volumes nommés » et l'étape 9. **Guidance :** monter `pgdata:/var/lib/postgresql` ou `PGDATA: /var/lib/postgresql/data` + re-valider l'étape 9 sur un volume réellement attaché.

### 🔴 2. `make migrate_up` sans sourcing `.env` (L119-127, L148-150, L198-201, L222-224)
`Makefile:92` construit le DSN depuis des variables **make** ; ne sourcent `.env` que `test` (:30) et `e2e` (:121-127) ; README:69-74 documente aujourd'hui `set -a; . ./.env; set +a` avant migrate_up — que la Task 4 *remplace*. La commande du plan échoue (`host='' port='' dbname=''`). **Guidance :** (a) garder l'export README/Validation + défauts make (`DB_HOST ?= 127.0.0.1`, `DB_PORT ?= 15432`, `DB_NAME ?= golang_gin_db`) ; ou (b) faire sourcer `.env` par migrate_up/down — dans ce cas utiliser les variables **shell** `$$DB_PORT` (GNU make n'enlève pas les guillemets, `DB_PORT="15432"` L136 est la première valeur guillemetée du dépôt).

### 🟡 3. CI modifiée silencieusement (L129-131)
`ci.yml:74` appelle `make migrate_up` → le runner tire postgres:18-alpine + redis:8-alpine, publie 15432/16379, attend le healthcheck, dans un job à `timeout-minutes: 15` (ci.yml:18). **Guidance :** retirer services_up de migrate_up/down ou conditionner ; chiffrer le surcoût si maintenu.

### 🟡 4. SSL jamais tranché (L133-141, L145-153, L227-228)
`.env_rename_me:3` SSL=TRUE → `sslmode=require` (pool.go:74-77, e2edb/main.go:183-185) vers un conteneur sans TLS ; `main.go:216-228` RunTLS exige `./cert/myCA.cer` absent du repo (.gitignore:14). **Guidance :** fixer SSL=FALSE explicitement, ou documenter le prérequis certificats ; l'étape 6 curl doit préciser la valeur SSL supposée.

### 🟡 5. Baseline `.env_rename_me` factuellement fausse (L140-141)
`.env_rename_me:8-9` `DB_HOST=localhost` / `DB_PORT=5432` (non guillemeté) vs `.env` : `127.0.0.1` / `"5432"` → « inchangé » et « fichiers identiques » faux. Donner les lignes exactes par fichier ; dire si DB_HOST est normalisé.

### 🟡 6. Natif vs ports figés (L145-147) — désigner le chemin gagnant (voir IR#1-zoom-out, comprehension#1).

### 🟢 7-14
- **Tests (L182-190) et étape 8 :** bascule skip→run du test testcontainers (postgres_test.go:18 → postgres.go:18 `postgres:16-alpine`) et son pull d'image non documentés.
- **Étape 2 (L217-219) :** Compose ≥ v2.17 ; nommer le fichier du remède.
- **Étapes 5/7 (L225-231) :** fixer un compte superadmin unique email/mot de passe partagé.
- **Task 1 / risques :** `name: artisan-dev` dans le compose (collision clones/worktrees) + symptôme « port is already allocated ».
- **Images flottantes :** figer (`postgres:18.7-alpine3.23`, `redis:8.8-alpine3.23` au 2026-11-24) ou assumer les majeures flottantes + corriger la mention « 8.12 ».
- **Task 7 :** ajouter memory-bank/progress.md (AGENTS.md:87) et .kilo/HANDOFF.md:27.
- **Phasage :** repérer les phases (1 stack → 2 make → 3 config → 4 docs/CI → 5 mémoire) avec état stable par phase.
- **Risques :** le sort de l'ancienne base native `golang_gin_db` (goose v2) jamais énoncé — « abandonnée, rien à importer » ou procédure.

---

## egm-codebase-details — findings

### 🔴 B1 — L129-171 : « CI pas affectée… jamais via make » faux (ci.yml:74 `make migrate_up`, :82 `make generate_sqlc`) — avec le prérequis, chaque run CI démarre la pile compose. « Rien d'autre ne change » (task 6) contredit le code.

### 🔴 B2 — L148-150, L198-201, L222-224 : bootstrap migrate_up sans export `.env` — inexécutable (détail Makefile:92, goose ne lit pas `.env`, README:69-74 documente l'export aujourd'hui).

### 🟡 M1 — L122-127 : le snippet **supprime `install_goose`** de `migrate_up: install_goose` / `migrate_down: install_goose` (Makefile:95,100) — appliqué tel quel, le binaire goose n'est jamais installé et la recette échoue sur `$(GOOSE_BIN)`. Forme correcte : `migrate_up: install_goose services_up`.

### 🟡 M2 — L133-141 : SSL/édition de la Task 1 (voir IR#4) — TLS du serveur, RunTLS, certs, README:159-166 `generate-certificate.sh`.

### 🟡 M3 — L63 : le healthcheck sans `-h` valide la socket Unix que le serveur temporaire d'initdb (`listen_addresses=''`) accepte déjà → `--wait` peut relâcher avant que 5432 n'accepte le TCP ; le repo connaît le piège (attente de la **seconde** occurrence du log readiness dans internal/testutil/postgres.go:41-47). **Guidance :** `pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db`.

### 🟡 M4 — L46-84 : aucun `name:` → collision déterministe entre clones (README.md:38 clone = gin-boilerplate) ; risque absent de la section Risques.

### 🟡 M5 — L166-171, L203-205, L242-244 : le troisième PostgreSQL épinglé (`testutil/postgres.go:18` postgres:16-alpine, test postgres_test.go:19 tag all qui va passer de skip à run) jamais mentionné dans la section Tests ni la validation.

### 🟢 N1-N6
- L140 : `DB_HOST=127.0.0.1` vrai seul pour `.env:8` (template = localhost) ; style mixte quoted/non-quoted dans le template.
- L182 : « Unitaires… aucun port » inexact — `cmd/e2edb/main_test.go:73-81` fige DB_PORT=5432 dans sa fixture (impact nul, phrase fausse).
- L36/170-171 : miroir CI approximatif (compose a `-d golang_gin_db`, CI non) ; seul `redis-cli ping` correspond exactement.
- L173-178 : périmètre memory-bank incomplet (HANDOFF.md:27, activeContext.md:60, techContext.md:41 ligne entière juste pour l'isolation FLUSHDB).
- L211-212 : localhostForwarding de WSL2 cité sans contrôle de bind ni emplacement du démon.
- L242-244 : « CI (tags identiques) » — seuls les majors correspondent.

### Catégories vides (vérifié exact, code) : toutes les références citées existent et sont exactes (InitRedis(1) ×4, go-redis v7, garde-fous e2edb, pool_test trust, ci.yml:21/25/34/122, vite.config.ts:20, ancres README). YAML valide ; aucun fichier compose existant dans le repo ; aucun autre helper de test impacté.

---

## egm-style — findings

### 🔴 B1 — L129-131 : phrase fondatrice de « Rien d'autre ne change » (L170) et « validée par construction » (L243) — factuellement fausse (ci.yml:74/82). **CALIBRATE** + garde-fou CI ( variable de skip, ou goose direct).
### 🔴 B2 — L122-127/148-150/199-201/222-224 : migrer exige l'export ; donner la commande exacte (`set -a; . ./.env; set +a; make migrate_up DB_USER=postgres DB_PASS=postgres`) ou faire sourcer la cible ; la cible e2e (Makefile:125) est le contre-exemple qui repasse tout explicitement.
### 🔴 B3 — L133-141 : le delta réel n'est pas « ports uniquement » — .env_rename_me DB_HOST=localhost, DB_PORT non quoté, SSL=TRUE vs SSL=FALSE du .env : tout `cp .env_rename_me .env` neuf échoue à se connecter (sslmode=require + RunTLS certs). **CALIBRATE** — delta par fichier ou bornage au seul .env existant.

### 🟡 M1-M7
- **M1 (L180-213)** : testcontainers PG 16, skip→run, « Hors scope » L266-267 affirme que seule la recette test reçoit un prérequis — **CALIBRATE**.
- **M2 (L20-22/29/204-205/242-244)** : parité = majeures seulement (alpine vs Debian, + PG16) ; « validée par construction » → **CALIBRATE** (supprimer la formule).
- **M3 (L211-212)** : wslrelay nommé sans installation ni vérification, démon Docker localisé non établi — **CALIBRATE ou DROP**.
- **M4 (L248-249 vs L218-219)** : « remède documenté » sans tâche documentaire qui le reçoit — **RELOCATE** vers la puce Prerequisites de la Task 4 (docker + usermod + re-login).
- **M5 (L173-178)** : liste incomplète (techContext:41 ligne entière, HANDOFF.md:27, activeContext.md:60) — **CALIBRATE**.
- **M6 (L119-120 vs L147)** : « la seule source dev » vs install native ; sort des données PG 15 natives jamais énoncé — **CALIBRATE**.
- **M7 (L246-256)** : « make run échoue bruyamment » ne couvre que run — les quatre cibles sont concernées — **CALIBRATE**.

### 🟢 N1-N8
N1 healthcheck miroir approximatif (CALIBRATE) · N2 « no-op 1-2 s » et « Compose ≥ 2.17 » non mesurés/ancrés · N3 « pgx/v5 supporte PG 18 » sans version (go.mod: v5.13.0), « go-redis v7 » sans version (v7.4.1, go.mod:10), PG 16 absent de la phrase de compatibilité · N4 remède de collision nomme deux fichiers (compose + .env) vs trois (.env_rename_me aussi) · N5 la puce « Vite absent » autoréférentielle sans décision lecteur — **DROP** · N6 « style underscore du fichier » → « noms de cibles en snake_case comme les cibles existantes » · N7 trust énoncé trois fois (table, Tests, Risques) — garder le mécanisme une fois (L189-190) et la conséquence L257-260 · N8 attribution ambiguë des ports squattés dans le commentaire compose (6379 = honcho ou teable ? CALIBRATE par port).

### Catégories vides : aucune hyperbole ni langage promotionnel ; langues mélangées: aucune (artefacts anglais conformes, corps français) ; pas de doublon structurel de la section Tests (elle porte des preuves).

### Recoupements avec stages 5-8 : B1 (compose en CI) déjà vu Stage 5 ; M1 (PG16/matrice) vu Stage 5 ; M5 vu Stage 5/zoom-out #6 ; M6 vu comprehension #1 ; M3 vu Stage 7. Les autres (B2, B3, N1, N2, N4-N8) sont spécifiques à cette passe.
