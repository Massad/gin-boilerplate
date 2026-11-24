# critic_report.md — Stage 7 (Critic Review)

**Target:** FeatureRange 1..41 (title, `## Contexte`, `## Objectif`, `## Décisions de design (fixées)`)
**Reviewers:** egm-design (design_critic) + egm-structural (structural) + egm-style (style)
**Standards:** feature-spec-standard.md (design + structural), persona-defined rubric (style)
**GoldfishContext:** the 10 confirmed baseline files.

---

## Shared 🔴 Blocker (raised independently by all three reviewers)

### Alternatives evaluation absent — violation of the feature-spec-standard
- **design_critic (🔴, L3-16, L25-41):** les 9 décisions sont posées « fixées » sans évaluation/rejet d'aucune alternative. Le standard conditionne la validation à cette évaluation. Le context documente un PG 15.19 natif installé (techContext.md:41) : l'alternative « services natifs sur 5433/6380 » est réelle, non évaluée. **Recommandation :** section « Alternatives rejetées » couvrant (a) natif WSL sur 5433/6380, (b) split REDIS_HOST/REDIS_PORT pour interpolation, (c) scram + requirepass (coût nul : `pool.go:82`, `db/db.go:15` lisent déjà DB_PASS/REDIS_PASSWORD), (d) format du healthcheck TCP.
- **structural (🔴, L10-13, L25-40):** le standard échoue le doc s'il propose un chemin unique sans preuve qu'autres designs viables ont été considérés/rejetés pour des raisons documentées. Les alternatives vérifiées libres (5433, 6380, L11) ne sont jamais dispositionnées — les seuls rejets sont des à-côtés d'une ligne (PG 19 « trop frais » L29, « ${...} » refusé L40).
- **style (🔴, L25-41):** le tableau fige 13 choix sans rejet documenté des alternatives viables — alpine vs non-alpine, compose vs testcontainers existant (`internal/testutil/postgres.go`), compose vs natif, volumes nommés vs bind mounts.

## egm-design (design_critic) — findings restants

### 🟡 1. Cibles make auto-suffisantes en conteneurs, pas en schema
**L23, 33, 37** — `services_up` démarre des conteneurs vides ; rôles et tables n'existent qu'après `make migrate_up`, absent de la chaîne. `main.go:137-142` log.Fatal si rôle absent ; `pool_test.go:28-30` skip silencieux sur 28000 (interdit par AGENTS.md:74). **Recommandation :** migration idempotente dans `services_up`, ou vérification avec message actionnable.

### 🟡 2. La chaîne make couplée à Docker : la CI n'est pas « sans impact »
**L29, 37** — `ci.yml:74` appelle bien `make migrate_up` (et `:82` generate_sqlc) : chaque run CI lancerait un second PG/Redis (pulls inclus). **Recommandation :** opt-in (`SERVICES ?= 1` désactivable côté CI) ou corriger l'affirmation.

### 🟡 3. `--wait` sans borne + volume `pgdata` non versionné par major : blocage infini
**L29, 37, 38** — un `pgdata` d'un major antérieur refuse le serveur (« data directory incompatible ») ; healthcheck ne passe jamais ; `--wait` attend indéfiniment. **Recommandation :** volume versionné (`pgdata18`) et/ou `--wait-timeout` + diagnostic pointant vers `down -v`.

### 🟡 4. Modèle de menace de `trust` sous-estimé pour l'environnement cible
**L31, 34** — la boucleback WSL2 est miroirée vers Windows (localhost forwarding) ; `trust` autorise tout rôle tout mot de passe : le superuser et `app_admin BYPASSRLS` (000002:14) exposés à tout processus local (scripts npm `postinstall`…) contournant l'isolation RLS. `POSTGRES_PASSWORD: postgres` devient décoratif sous trust. **Recommandation :** documenter le vecteur WSL2, ou scram + requirepass (les mots de passe existent déjà ; justification fonctionnelle de trust pour pool_test.go est correcte et bien fondée).

### 🟡 5. `restart: unless-stopped` + volumes persistants : exposition permanente
**L38, 39** — après chaque reboot, superuser sans mot de passe et store de sessions renaissent, sans limite. **Recommandation :** `restart: "no"` pour une stack dev, ou mot de passe + procédure « stack dev off ».

### 🟡 6. Double source de vérité sur les ports, sans contrôle de dérive
**L30, 40** — compose interpole nativement le `.env` racine que l'app lit déjà ; le plan reconnaît la dérive comme mode d'échec (L250-252) sans aucun moyen de détection. **Recommandation :** `${DB_PORT:-15432}` / `${REDIS_PORT:-16379}` ou une cible de vérification.

### 🟡 7. Tags d'images flottants, contraires au pinning du dépôt
**L29** — le dépôt épingne tous ses outils (Makefile:36-86) ; ici le moteur de base du dev change entre deux `make run` sans commit. **Recommandation :** épingler `postgres:18.7-alpine` / `redis:8.12-alpine` (ou digest), CI alignée.

### 🟢 Nits (design_critic)
1. **L21-22 vs 29** — Objectif annonce `postgres:18`/`redis:8` (debian) vs tableau alpine ; parité limitée au major, écart de flavour non analysé.
2. **L35, 38** — Redis sans mot de passe non arbitré (requirepass à coût nul via DB déjà lues) ; persistance Redis non spécifiée (aucun `command:`/`--save`/`--appendonly`/`--maxmemory` prévu).
3. **L36, 37** — le healthcheck sonde la socket du serveur de bootstrap (`listen_addresses=''` pendant init) et peut déclarer healthy avant l'existence de la base ; proposer `pg_isready -h 127.0.0.1 -U postgres -d golang_gin_db`.
4. **L10-13** — inventaire de ports non re-vérifiable et périmé sans détection.

### Conformité au feature-spec-standard (range 1..41)
- Problem Statement ✅ ; Proposed Architecture ✅ ; Goals & Non-Goals ⚠️ partiel (non-goals hors range) ; Alternatives Considered ❌ absent ; Risks & Mitigations ⚠️ partiel (registre hors range, mais ne couvre ni Redis sans mot de passe, ni auto-restart, ni blocage --wait/volume obsolète, ni couplage CI↔Docker, ni premier démarrage sans rôles, ni tags flottants).

---

## egm-structural (structural) — findings restants

### 🔴 1. No "Alternatives Considered" — voir Blocker partagé.

### 🟡 2. L'invariant « parité dev = CI » affiché trois fois, contredit par les choix figés
**L21-22, L29, L36** — Objectif nomme debian, décision tranche alpine, CI (ci.yml:21,34) tourne en non-alpine 15/7 ; healthcheck diffère (ci.yml:29 vs décision) — le lecteur ne sait pas quel côté est normatif ; la parité est infalsifiable telle qu'écrite.

### 🟡 3. Objectif (L23) vs Readiness (L37) : rayon d'action divergent
« make run » dans l'Objectif vs « chaque cible make » dans la Readiness — lus littéralement : `install_swag`, `install_sqlc`, `generate_docs` (Makefile:40-93) tireraient Docker.

### 🟡 4. Goals & Non-Goals insatisfaits dans le range
**L18-23** — objectifs sans non-goals ni critères de succès dans le range (le « Hors scope » L262-267 et « Validation » L214-244 sont hors range) — tous deux obligatoires selon le standard.

### 🟡 5. La table de décisions : unique record, sans colonne artefact, mêlant décision/risque/mitigation
**L25-40** — 13 lignes mêlent trois classes sans colonne de mapping vers le fichier touché ni marque décision-vs-risqué (L34 fusionne choix, exposition et mitigation dans une cellule) — les exigences "Proposed Architecture" et "Risks & Mitigations" n'ont pas d'ancre en range.

### 🟡 6. Décisions conditionnées par des expressions temporelles relatives non ancrées
**L10, L29** — « vérifié ce jour » (sans date absolue), « sorti il y a 5 jours » (sans date) : l'inventaire de ports qui conditionne tout le design et le rejet de PG 19 deviennent invérifiables pour un implémenteur ultérieur.

### 🟢 Nits (structural)
1. **L6** — Contexte attribue 6379 à honcho alors que techContext.md:41 enregistre le partage DB 1 avec le cache teable (le plan lui-même base son claim d'isolation sur ce fait, L196-197).
2. **L29** — "endoflife.date" domaine nu, sans lien markdown ; figures invérifiables.
3. **L29** — la cellule « Images » casse la symétrie de la table ( cinq faits dans une cellule ) ; scinder en lignes.
4. **L35** — incohérence inline-code (« store » nu vs tout le reste en code).
5. **L27-40** — les deux paramètres de connexion non figés (DB_HOST=localhost, SSL=TRUE → sslmode=require, Makefile:91-92) sont précisément ceux qui peuvent casser l'invariant loopback (résolution `::1` , TLS jamais décidé).

### Audit vectors (structural)
- Missing Required Headers : 1 Blocker + 2 Major. Header hierarchy : aucun défaut (H1→H2→H3 propre). Formatting : table bien formée, aucun défaut. Link compliance : 1 Nit. Décision→Tâche : traceability complète, aucune décision orpheline ; le gap est l'absence de colonne artefact (F5). Motivation : satisfaite, caveat de précision F7.

---

## egm-style (style) — findings restants

### 🔴 1. Alternatives mandatory evaluation missing — voir Blocker partagé.

### 🟡 Majors
1. **L20-21 vs 29** — tags d'images contradictoires (Objectif debian vs décision alpine) ; **Action: CALIBRATE** — énoncer les tags une fois et y référer.
2. **L21, L29** — « parité dev = CI » jamais défini (l'axe réel : versions d'images uniquement — les ports diffèrent structurellement) ; et le claim « la CI n'a que deux tags à changer » est contredit par ci.yml:74/82 qui appellent des cibles make. **Action: CALIBRATE** — nommer l'axe + noter les invocations make côté CI.
3. **L31** — l'épinglage du superuser est attribué à la mauvaise composante : « épinglé par cmd/e2edb et la CI » contredit par L191-193 et le code — le pin vit dans le Makefile e2e (Makefile:122), e2edb/main.go:169-174 ne fait que les lire. **Action: CALIBRATE** — attribuer au Makefile.

### 🟢 Nits (style)
1. **L34** — le caveat « jamais en prod » duplique le Risque L257-260 ; **RELOCATE** (une fois, dans Risks).
2. **L34** — « vraies credentials » mélange FR/EN — « identifiants ».
3. **L8 vs L10-11** — incohérence interne sur 5173 (« Vite sur 5173 » présenté comme état courant alors que 5173 est listé parmi les ports libres) ; dire « port configuré de Vite, actuellement inoccupé ».
4. **L10-13, L29** — dates relatives jamais ancrées (« ce jour », « il y a 5 jours ») ; donner les dates absolues.
5. **L10-13** — 5433/6380 listés libres puis éliminés sans raison (« patron local » vague).
6. **L29** — cellule Images : ~50 mots, quatre faits (tags, source+date, alternative rejetée, conséquence CI) ; scinder.
7. **L34, L36** — « miroir exact de la CI » surestime : la CI pose trust+USER/PASSWORD mais pas POSTGRES_DB, et son healthcheck est `pg_isready -U postgres` sans `-d`.
8. **L37** — le plancher de version pour `--wait` manque dans la décision en range : Compose ≥ 2.17, pas seulement « v2 ». **CALIBRATE**.
9. **L40** — « non découpable proprement » cache la vraie raison : compose interpole des variables entières, pas de slicing. **CALIBRATE**.
10. **L5-8** — « ne peut plus démarrer proprement » masque le symptôme observable ; énoncer le fait vérifiable.
   
### Style vectors : pas de fluff marketing, pas de jugement subjectif, pas de buzzword stack ; 1 RELOCATE + 8 CALIBRATE ; aucun DROP.

### Notes hors range (non comptées comme findings)
- ci.yml:74 appelle make migrate_up → la ligne 129 (« la CI n'est pas affectée ») factuellement fausse.
- .env_rename_me:8 `DB_HOST=localhost` → L140 « inchangé » ne correspond pas au template.

---

## Synthèse orchestrator (Stage 7)

- **🔴 convergent unanime** : Absence d'évaluation d'alternatives (exigence standard).
- **Cross-stage déjà vu** : CI affectée via make (déjà Stage 5) ; parité à recalibrer (alpine/debian/objectif) ; duplication ports ; trust threat model.
- **Nouveaux Majors Stage 7** : schema absent du contrat des cibles (design #1) ; --wait infini + volume de major obsolète (design #3) ; exposition permanente restart policy (design #5) ; tags flottants vs pinning repo (design #7) ; Objectif vs Readiness divergents (structural #3) ; Goals & Non-Goals hors range (structural #4) ; table sans colonne artefact (structural #5) ; dates relatives non ancrées (structural #6).
