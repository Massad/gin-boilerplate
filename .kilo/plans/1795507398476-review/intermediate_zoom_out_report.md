# intermediate_zoom_out_report.md — Stage 8 (Intermediate Zoom Out)

**Target:** `.kilo/plans/1795507398476-docker-pg-redis-dev.md` (267 lines, full file)
**Persona:** egm-zoom-out — Skeptical Architect · Mode REVIEW (context: Stage 5, 6, 7 reports + GoldfishContext + pool.go, testutil/postgres.go, migration 000002, .gitignore)

---

## Findings

### 🟡 F1 — Major — NEW : un problème de poste, des livrables de projet
**Scope :** L3-16, L18-23, L119-127, L143-164, L166-171
Le Contexte décrit un conflit de ports local à un poste WSL, mais les livrables sont des engagements de projet (compose suivi en git obligatoire pour run/test/e2e/migrate, README et AGENTS.md réécrits, CI touchée) dont l'objectif n'est jamais énoncé comme tel. Le README documente aujourd'hui un parcours sans Docker (README:27-31, :67-90) et AGENTS.md:76-83 est le contrat d'environnement du projet — le plan le modifie sans le dire.
**Mitigations (exclusives, avec leur prix) :** (a) « stack dev reproductible pour tous » → compose = unique provisioning (CI comprise), ports dans `.env` interpolés, bootstrap schema dans la cible, Docker = défaut du README ; (b) « débloquer cette machine » → aucun fichier suivi ne change : deux `docker run` locaux + `.env` local suffit ; CI/README/AGENTS intacts.

### 🟡 F2 — Major — NEW : `.env` devient le contrat d'un seul projet, mais il est lu par quatre parseurs aux règles de guillemets différentes
**Scope :** L40, L133-141 (L136 `DB_PORT="15432"`), Makefile:30/:91-92/:95-102, .env_rename_me
godotenv, le shell POSIX et docker compose retirent les guillemets ; GNU make ne le fait pas — le correctif du gap `migrate_up` (source `.env`) transformerait `DB_PORT="15432"` en la chaîne littérale `"15432"` et le DSN goose cesserait de contenir un port. Le plan introduit la première valeur guillemetée de l'histoire du dépôt, précisément dans le fichier qui devient la source de vérité de quatre consommateurs.
**Mitigation :** valeurs sans guillemets + une seule lecture make (`-include .env` + `export` + défauts `?=` pour le `.env` vide de la CI) — ferme simultanément le 🔴 Stage 5/F1, aligne make/app/compose, 0 changement Go.

### 🟡 F3 — Major — NEW (réutilisation) : le maillon manquant (« le schéma fait partie du contrat ») a déjà son modèle dans le dépôt
**Scope :** L37, L119-127, L180-212
`internal/testutil/postgres.go:33-77` sait déjà « démarrer un PostgreSQL + goose + créer les rôles » ; `cmd/e2edb` est le précédent d'un outil d'infra gardé et testé. Le plan ne mentionne ni l'un ni l'autre.
**Mitigation (b) :** assumer et ÉCRIRE le contrat : `services_up` = cluster, `make migrate_up` (corrigé) = schéma — dans README/AGENTS et les étapes de validation.

### 🟡 F4 — Major — NEW (mitigation) : « parité dev = CI » ne peut devenir vrai que si la CI provisionne via le compose
**Scope :** L21-22, L29, L166-171, L242-244
La CI garde son bloc `services:` (`ci.yml:19-41`), son healthcheck et son `CREATE DATABASE` (`ci.yml:70`) — « mêmes tags » n'est ni « mêmes majors » ni parité ; quatre chemins de provisionnement coexistent (compose, services CI, testcontainers postgres.go:18, e2edb). Une parité déclarative manuelle ne survit à aucun changement d'image/healthcheck/port.
**Mitigation (mutuellement exclusive) :** la CI provisionne via compose (`docker compose up -d --wait` + `exec -T psql`, bloc `services:` supprimé) → parité structurelle ; OU renoncer à la claim et écrire « mêmes majors, provisioning différent » (Stage 5/F13 honnête).

### 🟡 F5 — Major — NEW (mitigation) : `trust` présenté comme contrainte de design, alors qu'elle vient d'un test qui confond parse-DSN et auth serveur
**Scope :** L34, L257-260, L264-266
`pool_test.go:35-49` utilise le serveur comme oracle d'un comportement de parseur ; un test hors-ligne (`pgx.ParseConfig` + `cfg.Password == password`) prouverait la même chose, la connexion de contrôle restant le seul besoin serveur. `internal/db/pool.go:73-90` construit le DSN de façon déterministe ; testutil tourne déjà en scram avec les mêmes rôles.
**Mitigation (≈10 lignes + 1 ligne compose) :** pool_test hors-ligne + retirer `POSTGRES_HOST_AUTH_METHOD` (scram par défaut, mots de passe migration 000002:8/14) ; CI prête (`PGPASSWORD=postgres`, ci.yml:55). La décision L34 tombe et le risque L257-260 perd sa substance.

### 🟡 F6 — Major — NEW : le plan livre un nouvel artefact d'infrastructure sans vérification automatisée
**Scope :** L44-84, L180-212, L214-244
Rien ne vérifie la stack elle-même (validité compose, cohérence compose/`.env`, healthcheck prouvant une base utilisable) et la vérification repose sur un script manuel de 11 étapes. AGENTS.md:52/:72-74 rend les tests obligatoires et fournit le modèle (`cmd/e2edb/main_test.go`) ; le risque de dérive des ports (L250-252) est documenté sans détecteur armé.
**Mitigation (~15 lignes) :** `docker compose config -q` en étape CI + cible `make check_ports` comparant les ports publiés du compose aux valeurs de `.env`.

### 🟢 F7 — Nit — NEW : le nom du projet compose dérive du répertoire → worktrees et clones multiples orphelinent des conteneurs qui gardent les ports
**Scope :** L38, L39, L103-105
Sans `name:`, Compose v2 dérive le nom du répertoire (`gin-boilerplate` selon le clone du README:38, worktrees…) : un second checkout crée un autre projet/volume, les conteneurs du premier (`restart: unless-stopped`) gardent 15432/16379 → échec de bind trompeur. **Mitigation :** `name: artisan-dev` en tête du compose.

### 🟢 F8 — Nit — NEW : un plan qui change le contrat du dépôt doit documenter son « undo »
**Scope :** L38, L103-105, L143-153, L253-254
`services_down` éteint, ne re-pointe pas : aucune procédure de retour arrière (retirer les prérequis, re-pointer `DB_PORT`/`REDIS_HOST` sur du natif). **Mitigation :** 5 lignes « retour arrière » dans le README.

---

## Jeu de mitigations qui change l'architecture (par levier)

1. **M1** — une seule lecture de `.env` (`-include .env` + `export` + `?=` defaults) → ferme le 🔴 Stage 5/F1, aligne make/app/compose.
2. **M2** — les ports sont des données : `DB_PORT` + `REDIS_PORT` (split via `net.JoinHostPort`, 2 lignes Go), compose `${DB_PORT:-5432}` — la dérive L250-252 disparaît par construction.
3. **M3** — une seule source de provisioning : CI via compose (bloc `services:` supprimé) OU suppression de la claim de parité.
4. **M4** — `trust` hors repo : refactor hors-ligne pool_test + compose scram.
5. **M5** — un seul propriétaire du schéma : `services_up` = cluster, `migrate_up` = schéma, écrits explicitement.
6. **M6** — deux vérifications permanentes : `docker compose config -q` en CI + `make check_ports`.
7. **M7** — `name: artisan-dev` dans le compose.

**À ne pas faire (YAGNI) :** pas de nouvel outil d'infra ; pas de split REDIS_HOST sans REDIS_PORT ; pas de cibles make au-delà de services_up/down ; si M3 refusé, ne plus parler de parité.

## Jugement architectural résiduel

Après quatre étapes, le risque résiduel se concentre en un point : **le plan fait de la machine l'unité du changement tout en écrivant le projet comme si le projet était l'unité du changement.** Les deux exigences encore dures — un contrat de schéma et une posture d'authentification — ont déjà des réponses dans le dépôt (goose + `internal/testutil/postgres.go` ; `pool_test.go` + les mots de passe dev de la migration 000002). Réduit à M1 + M2 + M4 + M5, le plan tient en ~60 lignes et **retire** de la surface architecturale au lieu d'en ajouter. Exécuté tel quel, il livre en plus une rétrogradation d'authentification permanente, une dépendance dure à Docker dans un dépôt promettant un parcours sans Docker, et une CI qui démarre une seconde pile inutilisée.
