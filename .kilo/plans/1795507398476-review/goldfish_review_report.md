# Goldfish Annotated Review Report

**Status: resolved** — 2026-11-24 : « auto accept » utilisateur avec exclusion CI
(alignement ci.yml reporté, parité non revendiquée) ; 13 PROPOSE FIX appliquées
au plan (B-01..B-05, M-01, M-02, M-04..M-09, N-03/05/06/08 ; volets CI de M-03
retirés — testcontainers doc conservée, testutil bump conservé), 1 CONTEST
bypassé, 5 DECLINE enregistrés, 0 ESCALATE. Voir B-03 : le retrait de
`services_up` des migrations est conservé (protège la CI au lieu de la modifier).
**Post-gate 2026-11-24 15:28** : l'utilisateur a réintégré l'alignement CI —
M-03 est désormais appliqué **en entier** au plan (tâche 7 : tags
`postgres:18-alpine`/`redis:8-alpine` dans ci.yml, parité littérale
compose/CI/testutil).

**Target:** `.kilo/plans/1795507398476-docker-pg-redis-dev.md` (`full file`, 267 lignes)
**FindingsList:** `consolidated_review_report.md` ( même dossier) + 5 rapports bruts
**Catégories attribuées :** 14 PROPOSE FIX · 1 CONTEST · 5 DECLINE · 0 ESCALATE

---

## Prioritized Findings & AI Suggestions

### [B-01] [egm-zoom-out, egm-implementation-readiness] - Severity: Blocker
*   **Affected Scope:** L38, L52, L60-61 (Décisions Persistance / Task 1 compose) ; Validation 3 et 9
*   **Discrepancy Description:** Le montage `pgdata:/var/lib/postgresql/data` est incompatible avec `postgres:18-alpine` : depuis PG 18, l'image place PGDATA dans `/var/lib/postgresql/18/docker`, déclare `VOLUME /var/lib/postgresql`, et son entrypoint refuse de démarrer devant un montage sur `/var/lib/postgresql/data`.
*   **Rationale:** Le conteneur meurt (exit 1, « unused mount/volume »), `--wait` n'atteint jamais `healthy` ; même contourné, les données iraient dans un volume anonyme — contradictoire avec la décision Persistance et l'étape 9.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Vérifié par deux reviewers indépendants contre le Dockerfile officiel 18/alpine ; le correctif est la recommandation du message d'erreur de l'image elle-même.
*   ***Proposed Edit:***
    *   Location: Task 1, bloc compose, ligne du volume postgres (L60-61) + Décisions L38
    *   Change: `- pgdata:/var/lib/postgresql/data` → `- pgdata:/var/lib/postgresql` ; noter dans la décision Persistance : « volume monté sur `/var/lib/postgresql` (disposition PG 18+, l'image refuse `/var/lib/postgresql/data`) ».

### [B-02] [egm-inventive-catalyst, egm-implementation-readiness, egm-codebase-details, egm-style] - Severity: Blocker
*   **Affected Scope:** L119-127 (prérequis), L133-141 (Task 3), L148-150 (Task 4), L198-201 (Tests), L222-224 (Validation 4)
*   **Discrepancy Description:** `make migrate_up` construit son DSN depuis des variables make jamais alimentées : aucune cible de migration ne source `.env` (seuls `test` et `e2e` le font), la commande « auto-suffisante » échoue avec `host='' port='' dbname=''` ; de plus le snippet proposé **supprime le prérequis `install_goose`** (Makefile:95/100).
*   **Rationale:** Makefile:92 ; README:69-74 documente aujourd'hui l'export que la Task 4 *remplace* sans substitution ; par ailleurs GNU make ne retire pas les guillemets alors que godotenv/shell/compose les retirent — `DB_PORT="15432"` (L136) serait la première valeur guillemetée lue par make.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Triple vérification convergente ; le correctif ferme le gap par construction (un seul point de lecture) et rend le template et le `.env` cohérents (déjà non guillemeté côté template).
*   ***Proposed Edit:***
    *   Location: Task 2 (Makefile) + Task 3 (.env/.env_rename_me) + Validation 4 + Tests L198-201
    *   Change: (1) Prérequis : `migrate_up: install_goose` / `migrate_down: install_goose` (services_up retiré — voir B-03). (2) En tête du Makefile, après le bloc SHELL : `-include .env`, `export`, puis `DB_HOST ?= 127.0.0.1`, `DB_PORT ?= 5432`, `DB_NAME ?= golang_gin_db` (les `?=` respectent l'env CI qui définit déjà ces variables). (3) Task 3 : valeurs **sans guillemets** (`DB_PORT=15432`, `REDIS_HOST=127.0.0.1:16379`) et **dé-guillemeter les valeurs DB_* existantes** du `.env` (sans effet pour godotenv/shell/compose, indispensable pour `-include`). (4) Validation 4 reformulée : `make migrate_up DB_USER=postgres DB_PASS=postgres` fonctionne après `make services_up` (étape 3), le sourcing étant assuré par le Makefile. (5) Supprimer « auto-suffisante » au profit de « lit le même `.env` que l'app ».

### [B-03] [unanime — 9 reviewers] - Severity: Blocker
*   **Affected Scope:** L129-131 (claim CI), L122-127 (prérequis), L166-171 (Task 6), L242-244 (Validation 11)
*   **Discrepancy Description:** « La CI n'est pas affectée : elle appelle go test et goose directement, jamais via make » est faux — `ci.yml:74` appelle `make migrate_up` ; avec le prérequis `services_up`, chaque run CI tirerait la pile compose (2 conteneurs inutilisés, pulls, dépendance Compose v2) dans un job à timeout 15 min.
*   **Rationale:** ci.yml:74 et :82 (`make generate_sqlc`) ; la phrase fonde « Rien d'autre ne change » (L170) et « validée par construction » (L243).
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Fait vérifié par tous ; le retrait du prérequis des cibles de migration (a) rend la claim CI **vraie** à nouveau, (b) préserve le bootstrap dev (run/test/e2e démarrent les conteneurs, l'étape 3 de validation précède déjà l'étape 4).
*   ***Proposed Edit:***
    *   Location: L122-127 (snippet prérequis), L129-131, L242-244
    *   Change: Prérequis conservés sur `run`, `test`, `e2e` uniquement ; `migrate_up`/`migrate_down` gardent `install_goose`. Remplacer L129-131 par : « La CI appelle `make migrate_up` (ci.yml:74) — sans prérequis compose sur les cibles de migration, elle reste inchangée. » L242-244 : remplacer « validée par construction » par « les mêmes tags d'images sont exercés localement et en CI ; le push suivant le confirme ».

### [B-04] [egm-zoom-out, egm-implementation-readiness, egm-codebase-details, egm-style] - Severity: Blocker
*   **Affected Scope:** L133-141 (Task 3), L143-153 (Task 4), L222-228 (Validation 5-6)
*   **Discrepancy Description:** `SSL=TRUE` (template) n'est jamais traité : il force `sslmode=require` vers un conteneur sans TLS et `RunTLS` exige des certificats gitignorés — tout `cp .env_rename_me .env` neuf échoue à se connecter ; les étapes 5/6 de validation tombent.
*   **Rationale:** .env_rename_me:3 ; internal/db/pool.go:74-77 ; main.go:216-228 ; .gitignore:14 ; le compose ne provisionne aucun TLS.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** La stack dev compose est sans TLS par design ; la posture doit être explicite, et le delta réel par fichier (DB_HOST, quotes) doit être énoncé puisque le compose comment fait de `.env` un contrat.
*   ***Proposed Edit:***
    *   Location: Task 3 + tableau Décisions (nouvelle ligne) + Task 4 + Validation 6
    *   Change: (1) Task 3 : ajouter `SSL=FALSE` au template (le `.env` local est déjà FALSE) + donner les lignes exactes **par fichier** : `.env_rename_me` → `DB_HOST=127.0.0.1` (normalisé, ferme la nuance `::1`), `DB_PORT=15432`, `REDIS_HOST=127.0.0.1:16379`, `SSL=FALSE` ; `.env` → mêmes valeurs (dé-guillemetées). (2) Décisions : ligne « SSL dev » = `SSL=FALSE` (stack compose sans TLS ; TLS natif documenté au README via generate-certificate.sh). (3) Task 4 : le flux README mentionne SSL=FALSE et `http://localhost:9001`. (4) Validation 6 : préciser « http (SSL=FALSE) ».

### [B-05] [egm-design, egm-structural, egm-style — unanime Stage 7] - Severity: Blocker
*   **Affected Scope:** L3-41 (Contexte/Objectif/Décisions — FeatureRange entier)
*   **Discrepanpy Description:** Aucune section « Alternatives considérées » : le feature-spec-standard échoue un document qui propose un chemin unique sans alternatives évaluées et rejetées pour des raisons documentées ; les alternatives vérifiées (5433/6380) ne sont jamais dispositionnées.
*   **Rationale:** Standard : « fails this standard if it proposes a single path without proving that other viable designs were considered and rejected for specific, documented reasons ».
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Un lecteur humain a besoin de ce registre (pourquoi 15432/16379, pourquoi pas natif, pourquoi trust encore) : plusieurs reviewers ont contesté des choix « fixés » précisément faute de disposition écrite. L'arbitrage fondateur (problème de poste vs livrable de projet — Stage 8 F1) doit aussi être énoncé.
*   ***Proposed Edit:***
    *   Location: Nouvelle section « ## Alternatives rejetées » entre « Décisions de design » et « Tâches »
    *   Change: (0) Préambule : « Livrable assumé : stack dev reproductible possédée par le projet (compose suivi en git) — pas un dépannage local jetable ; l'install native reste joignable en re-pointant `.env` (documenté au README). » (a) Services natifs WSL sur 5433/6380 — rejeté : plus de services natifs fiables dans ce WSL, pattern local de squat des replis conventionnels, pas de reproductibilité pour les nouveaux développeurs. (b) Interpolation compose `${DB_PORT}` / split `REDIS_PORT` — rejeté : changer le contrat `REDIS_HOST=host:port` documenté (AGENTS.md, README) coûte deux lecteurs Go pour un gain de confort ; la dérive est instead détectée par `make check_ports` (M-06). (c) scram + `requirepass` Redis — refusé pour ce périmètre : exige le refactor hors-ligne de `pool_test.go` (changement de code test hors scope) ; enregistré comme follow-up possible, risque trust documenté (M-04). (d) testcontainers-only — rejeté : démarrage à chaque commande, pas d'état persistant entre sessions. (e) volumes bind vs nommés — rejeté : les volumes nommés se reset proprement (`down -v`) sans toucher l'arbre git. (f) alpine vs debian — alpine retenu (taille), parité assurée en alignant les tags partout (M-03).

### [M-01] [egm-design, egm-inventive-catalyst] - Severity: Major
*   **Affected Scope:** L23, L33, L37, L119-127, L180-212
*   **Discrepancy Description:** `services_up` ne livre qu'un cluster vide : le schéma (rôles, tables) n'existe qu'après `make migrate_up` — absent du contrat, l'échec n'apparaît qu'à la première requête (log.Fatal tardif).
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** Le contrat doit être écrit, pas implicite ; l'auto-migration silencieuse est refusée (superuser requis par 000002 — régression sécurité).
*   ***Proposed Edit:*** Location: Tests (L180+) + README Task 4 + AGENTS Task 5. Change: phrase de contrat : « `services_up` démarre le cluster ; `make migrate_up` (superuser) crée le schéma ; `run`/`test`/`e2e` démarrant les conteneurs ne dispensent pas de la première migration. » + mention dans AGENTS.md Commandes.

### [M-02] [egm-comprehension, egm-zoom-out, egm-implementation-readiness, egm-style, Stage 8 F1/F8] - Severity: Major
*   **Affected Scope:** L143-153 (Task 4), L133-141, Risques
*   **Discrepanpy Description:** « L'install native restant valide » contredit `.env` figé sur 15432/16379 ; le sort de l'ancienne base native n'est jamais énoncé ; aucune procédure de retour arrière.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: Task 4 (README). Change: (1) « L'install native reste valide à condition de re-pointer `DB_PORT` et `REDIS_HOST` dans `.env` (ex. 5432 / 127.0.0.1:6379). » (2) Une ligne Contexte : « l'ancienne base native `golang_gin_db` (goose v2) est abandonnée — aucune donnée à importer » (à ajuster si l'utilisateur décide autrement). (3) Section « Retour arrière » : retirer les prérequis services_up, re-pointer `.env`, `docker compose down` (les fichiers suivis sont tous réversibles par git).

### [M-03] [egm-inventive-catalyst F6, egm-style M2, egm-codebase-details M5] - Severity: Major
*   **Affected Scope:** L29, L166-171, L203-205, L242-244, Tests L180-212, Validation 8
*   **Discrepancy Description:** « Parité » revendiquée mais trois versions de PostgreSQL coexistent (compose 18-alpine, CI 18, testutil **16-alpine**) et deux flavours ; le test testcontainers (tag `all`) passe de skip à **run** dès que Docker existe — non documenté.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** La parité ne peut être une claim qu'avec un seul tag partout ; la bascule skip→run est un vrai changement de comportement de `make test` (pull réseau postgres:16).
*   ***Proposed Edit:*** Location: Task 6 + nouvelle tâche + Tests + Validation 8. Change: (1) CI tags → `postgres:18-alpine` / `redis:8-alpine` (identiques au compose). (2) Nouvelle tâche 6bis : `internal/testutil/postgres.go:18` : `postgres:16-alpine` → `postgres:18-alpine` (1 ligne — la parité devient littérale). (3) Tests : ajouter « le test testcontainers (opt-in, tag all) passe de skip (Docker absent) à run (pull postgres:18-alpine) dès la stack en place ». (4) Validation 8 : ajouter l'exception connue testcontainers à côté de poppler.

### [M-04] [egm-design #4/#5, egm-inventive-catalyst F2/F3, Stage 8 M4] - Severity: Major
*   **Affected Scope:** L34, L257-260 (Risques trust), L264-266
*   **Discrepancy Description:** La justification trust ne couvre pas le vecteur WSL2 localhost-forwarding (superuser + `app_admin BYPASSRLS` joignables de tout processus local) et `restart: unless-stopped` pérennise l'exposition au-delà des sessions de dev.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** La justification **fonctionnelle** de trust est vérifiée exacte par les reviewers (mécanisme 28P01 de pool_test.go, miroir de la CI et de l'ancien PG natif dev) — je ne conteste pas le choix, mais le document doit porter le modèle de menace complet ; le basculement scram exige le refactor pool_test (hors périmètre → alternatives B-05c).
*   ***Proposed Edit:*** Location: L34 (décision) + Risques L257-260. Change: ajouter au risque trust : « vecteur : le localhost-forwarding WSL2 expose les ports publiés aux processus Windows locaux (scripts, autres distros) ; `restart: unless-stopped` maintient l'exposition après reboot — assumé pour une machine de dev, le basculement scram+requirepass est documenté en Alternatives comme follow-up (refactor `pool_test.go` hors-ligne requis). »

### [M-05] [egm-design #3, egm-codebase-details M3] - Severity: Major
*   **Affected Scope:** L36, L63 (healthcheck compose), L170-171
*   **Discrepancy Description:** `pg_isready -U postgres -d golang_gin_db` sans `-h` sonde la socket Unix acceptée par le serveur temporaire d'initdb → healthy possible avant l'écouteur TCP ; « miroir CI » approximatif (ci.yml:29 sans `-d`).
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: Task 1 healthcheck + Décisions L36. Change: `test: ["CMD-SHELL", "pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db"]` ; L36 : « healthchecks équivalents CI, sonde TCP pour PG (l'image refuse les montages data et initdb écoute d'abord sur socket) ».

### [M-06] [egm-zoom-out Z12, catalyst F8, design #6, Stage 8 M2/M6] - Severity: Major
*   **Affected Scope:** L40, L49, L250-252
*   **Discrepancy Description:** Les ports vivent en trois fichiers (compose, .env, template) avec « un commentaire » comme seul garde-fou ; le risque de dérive est listé sans détecteur.
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Justification:*** L'interpolation est refusée en Alternatives (B-05b) pour cause de contrat `REDIS_HOST` ; le détecteur reste le correctif minimal du risque énoncé.
*   ***Proposed Edit:*** Location: Task 2 (Makefile) + Risques L250-252. Change: nouvelle cible `check_ports` : échoue si les ports publiés par `docker compose config` diffèrent de `DB_PORT`/`REDIS_HOST` du `.env` ; appelée par `services_up` ; le remède de collision référencé passe à trois fichiers (compose, `.env`, `.env_rename_me`).

### [M-07] [Stage 8 F7, codebase M4, IR#10] - Severity: Major
*   **Affected Scope:** L46-84 (compose)
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: tête du compose. Change: `name: artisan-dev` (évite les collisions clones/worktrees avec `restart: unless-stopped`).

### [M-08] [Z6, style M5, codebase N4, IR#12] - Severity: Major
*   **Affected Scope:** L173-178 (Task 7)
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: Task 7. Change: étendre aux fichiers périmés : `techContext.md` **ligne 41 entière** (dont la clause « DB 1 = cache teable » devenue fausse — c'est elle qui justifiait l'isolation FLUSHDB), `memory-bank/progress.md` (obligatoire AGENTS.md:87), `memory-bank/activeContext.md` (ports 5432/6379), `.kilo/HANDOFF.md` (« PG 15.19 natif… »).

### [M-09] [structural #6, style N2/N4, IR #8/#9/#13, catalyst F10] - Severity: Major
*   **Affected Scope:** L10-13, L29, L217-219, L225-231, L248-249
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** Location: transversal. Change: dates absolues (« vérifié le 24 nov. 2026 », lien https://endoflife.date/postgresql ) ; « Compose ≥ v2.17 » partout où « v2 » apparaît (L37 décision, L217 validation) ; le remède docker (daemon, groupe `docker` + re-login) **atterrit** dans la puce Prerequisites de la Task 4 ; compte superadmin unique partagé étapes 5/7 (`admin@example.com`, même mot de passe) ; supprimer « no-op ~1-2 s » (non mesuré) ; marqueurs de phase dans les Tâches (1 stack → 2 make → 3 config → 4 docs → 5 CI → 6 mémoire, état stable entre phases).

### [N-03] [Z5, design nit#2] - Severity: Nit
*   **Discrepancy Description:** Contradiction « command stock » (décision) vs aucun `command:` dans le compose.
*   ***AI Suggestion:*** PROPOSE FIX (vaut le coup : contradiction interne)
*   ***Proposed Edit:*** L35 : « Redis | configuration par défaut de l'image (snapshots RDB, volume `/data`) — pas de mot de passe (`REDIS_PASSWORD` vide) ».

### [N-05] [structural #9/#8, style] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX
*   ***Proposed Edit:*** L29 : scinder la cellule Images en quatre lignes (Image PG / Image Redis / Choix de version — dates absolues + lien endoflife.date / Parité CI) — une décision par ligne.

### [N-06] [catalyst F12] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (une ligne, classique « could not resize shared memory segment »)
*   ***Proposed Edit:*** Service postgres du compose : `shm_size: 256m`.

### [N-08] [style S7/S9 nits] - Severity: Nit
*   ***AI Suggestion:*** PROPOSE FIX (subset bon marché)
*   ***Proposed Edit:*** (a) justification trust énoncée **une fois** (garder le mécanisme L189-190, renvoyer depuis L34 et L257-260) ; (b) **DROP** la puce « Vite absent au moment de l'écriture » (autoréférentielle, sans décision) ; (c) « vraies credentials » → « vraies identifiants » ; (d) « style underscore du fichier » → « noms de cibles en snake_case, comme les cibles existantes » ; (e) commentaire compose : attribuer les ports squattés un par un (6379 : honcho-redis ; 9000 : teable ; …).

---

## CONTEST (1)

### [N-01] [egm-zoom-out #8] - Severity: Nit — « Quatre cibles make là où deux suffisent »
*   ***AI Suggestion:*** CONTEST
*   ***Justification:*** Exigence explicite enregistrée dans le document : L22-23 « lancés **individuellement** ou ensemble via make » — les cibles `postgres_up`/`redis_up` sont le livrable demandé par l'utilisateur, pas du YAGNI. Le coût (2 lignes .PHONY + 2 lignes docs) est assumé. Évidence vérifiable dans le fichier cible lui-même.

## DECLINE (enregistrés, non corrigés)

1. **« Unitaires … aucun port » (codebase N2)** — la fixture e2edb fige un DSN constant hors ligne : aucune dépendance runtime ; l'intention de la phrase (pas de serveur requis) tient.
2. **`services_reset` avec CONFIRM (F11)** — le plan refuse délibérément une cible destructive à portée de complétion ; `down -v` documenté suffit.
3. **Bascule scram/requirepass immédiate (catalyst F2)** — refactor de `pool_test.go` = changement de code test hors périmètre ; enregistrée en Alternatives comme follow-up (B-05c/M-04).
4. **Interpolation compose (F8/Z12)** — refusée pour cause de contrat `REDIS_HOST` documenté ; `check_ports` couvre le risque réel (M-06) ; consignée en Alternatives.
5. **`restart: "no"` (design #5)** — `unless-stopped` assumé pour la praticité dev ; le risque est documenté (M-04) au lieu d'être contourné.

## ESCALATE (0)

Aucun finding ne dépend d'un savoir hors fichiers : les arbitrages (choix des ports, objectif du livrable) sont déjà des décisions de la conversation que le document peut enregistrer lui-même (B-05).
