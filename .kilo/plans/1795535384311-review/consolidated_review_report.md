# consolidated_review_report.md — Goldfish Review consolidée

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md` (207 lignes)
**Semantic split :** FeatureRange 1..47 · DetailedRange 48..207
**Preset :** Consolidated report — 5 stages sans checkpoints
**Reviewers :** egm-zoom-out ×2 (S5, S8), egm-inventive-catalyst (S5), egm-comprehension (S6), egm-design (S7), egm-structural (S7), egm-style ×2 (S7, S9), egm-implementation-readiness (S9), egm-codebase-details (S9) — **10 dispatches en contextes frais, 0 not-run**
**Rapports bruts :** `zoom_out_report.md`, `comprehension_report.md`, `critic_report.md`, `intermediate_zoom_out_report.md`, `implementation_readiness_report.md` (même dossier)
**Note méthodo :** comportements de bibliothèques vérifiés contre les docs officielles via Context7 (go-redis `Client.Options()`/`DB` ; godotenv `Load` n'écrase jamais ; Redis db logiques = *namespacing*, `SELECT` hors plage → `ERR DB index is out of range`) — injection en ground truth aux reviewers, lecture du module cache proscrite.

---

## Verdict global

**FAIL contre les standards appliqués — mais un FAIL sain.** L'architecture est jugée saine par les deux passes zoom-out (index dédié, un seul parseur, variadic supprimé, échec explicite, non-touchage `.env`/CI/Swagger) ; le diagnostic d'ouverture est exact et **exhaustivement vérifié** : toutes les ancres `fichier:ligne` re-vérifiées par 3 auditeurs indépendants (une seule paire fausse, corrigée ci-dessous), 4/4 appelants `InitRedis` confirmés, **0 « contradicts the code », 0 Blocker**. Ce qui échoue au standard Implementation Readiness : **deux ancres fausses sur des opérations d'écriture**, des **commandes de preuve non exécutables** (le plan échouerait sa propre validation), un **invariant faux enseigné comme vrai** (ordre `t.Setenv`/`godotenv`), une **garantie centrale prouvée par aucun test**, et **6 décisions laissées ouvertes** (Zero-Decision). Rien de tout cela ne remet en cause l'architecture : chaque correctif est une ligne, un entier ou une phrase — le plan est à ~85 % de son objectif (estimation du Stage 8).

## Readiness par task (Stage 9)

| Task | Verdict | Bloquant |
|---|---|---|
| T1 `db/db.go` | NOT READY | D1 (garde `DefaultRedisDB`), calibrage du message, injection `getenv` |
| T2 `db/db_test.go` | NOT READY | cas `15`, formes `Atoi`, package, test du chemin d'erreur `InitRedis` |
| T3 `main.go` | READY | — |
| T4 `cmd/e2edb` | NOT READY | D3 (garde index dev), `Usage:`, liste des vars, résolution avant le DROP |
| T5 `tests/` | NOT READY | D1 (mécanisme), invariant réel, séquencement |
| T6 `middleware`/`controllers` | NOT READY | D1 (mécanisme) |
| T7 chaîne e2e | NOT READY | ancre `Makefile:155`, D2 (index 2 vs 3) |
| T8 documentation | NOT READY | D6 (ancres/sections historiques), terminologie, bullet `db/` AGENTS |
| T9 validation | NOT READY | commandes exécutables, contrôle `go test` direct, étape login |

---

## Décisions à trancher (Zero-Decision Rule) — le cœur du correctif

- **D1 — Mécanisme de protection des suites.** `t.Setenv` expire à la fin du test qui le pose ; dans `tests`, l'isolation ne tient que pendant `TestIntDB`. **Reco : `TestMain` par paquet taggé `all`** (`tests`, `middleware`, `controllers`) posant `REDIS_DB=2` au niveau processus, **plus** l'assertion d'index de M-04 (preuve permanente ; le Stage 8 juge l'un des deux suffisant, les deux coûtent ~6 lignes au total). Supprimer toute mention d'ordre `t.Setenv`/`godotenv.Load` et écrire l'invariant réel : « le client Redis doit être créé tant que `REDIS_DB` est présent dans l'environnement du processus ».
- **D2 — Index e2e partagé ou distinct.** La prémisse « jamais concurrents » a déjà failli une fois dans ce dépôt, et le plan applique à la paire tests/e2e qu'il crée son propre argument de rejet (« mêmes clés, même db »). **Reco : `REDIS_DB=3` pour la chaîne e2e** (Makefile + `playwright.config.ts`), 2 reste aux suites Go — un entier supprime la classe entière d'interférence (FLUSHDB e2e ↔ sessions des suites, `TestCleanUp` ↔ session superadmin e2e). À défaut : index partagé **assumé par un invariant écrit** dans AGENTS.md (« ne jamais lancer `make test` et `make e2e` simultanément ») — en l'état, il n'existe nulle part.
- **D3 — Garde active dans `cmd/e2edb`.** La commande d'usage documentée purge l'index dev par défaut. **Reco : oui** — refuser de purger l'index de dev (`db.DefaultRedisDB`) quand `REDIS_DB` est absent, sur le modèle testé de `validateTarget` (~6 lignes + cas dans `main_test.go`), mettre à jour `Usage:` (`E2E_DB_NAME=artisan_e2e REDIS_DB=2 go run ./cmd/e2edb`), ajouter `REDIS_DB` à l'énumération des variables lues, et **résoudre l'index dans `run()` avant `recreateDatabase`** (aujourd'hui le FLUSH vient après le DROP : une erreur de parsing remonterait après la destruction de la base).
- **D4 — Garde CI anti-régression.** `! git grep -n "InitRedis(1)" -- '*.go'` (modèle « legacy drivers », `ci.yml:88`). Contredit « Hors scope : CI inchangée » → **soit l'ajouter en amendant le Hors scope, soit la rejeter explicitement**. **Reco : l'ajouter** (une ligne, la preuve devient permanente).
- **D5 — Littéraux d'index.** La valeur est dupliquée (3 fichiers de test + Makefile + playwright après D2 : deux valeurs). **Reco : garder les littéraux nus mais les reconnaître dans le plan** (une phrase : « l'accord entre ces fichiers est prouvé par la sentinel T9 ») ; rejeter le détecteur de dérive qui relirait Makefile/playwright depuis un test Go (R3, disproportionné).
- **D6 — Sections historiques (`activeContext.md:26/29`, `HANDOFF.md:27`).** `activeContext.md` est le fichier d'état courant : y laisser « le FLUSHDB purge la base Redis 1 » propage un énoncé devenu faux. **Reco : annoter daté** (« à la date du 2026-11-23 ; l'isolation db 1/db 2 est en vigueur depuis le plan REDIS_DB ») et **assumer explicitement l'édition de HANDOFF** (section encore utilisée comme référence) — le plan applique aujourd'hui deux traitements opposés à deux documents de même nature.

---

## 🟡 Majors consolidés (12)

### M-01 — Ancres de lignes fausses sur des opérations d'écriture
- **Trace :** S5-F6 · S6-#1 · S7-style-annexe · S8-G7 · S9-IR-M1/M2 · S9-CB-M1/M2 · S9-style-F-01/F-02/F-09
- **Fait :** l'appel e2edb est en `Makefile:154-155` (recette `e2e` l. 152-165), pas « ~121-122 » (lignes DSN/`install_goose`) ; `activeContext.md:60` (joignabilité WSL) est sans rapport — les énoncés qui deviennent faux sont `:26` et `:29`, dans la section historique que T8 interdit de retoucher (contradiction interne).
- **Correctif :** recaler par **noms de cible** (recette `e2e`, « bloc de setup », « § Leçons durables ») plutôt que par numéros volatils ; trancher D6.

### M-02 — Les commandes de preuve de T9 ne sont pas exécutables telles qu'écrites
- **Trace :** S6-#2/#10 · S8-G6 · S9-IR-M8 · S9-CB-N7 · S9-style-F-04 · S5-F7 (contrôle `go test` direct absent)
- **Fait :** étape 3 exécute `redis-cli` dans le conteneur, étapes 4-6 le lancent nu — le Redis de dev est publié sur `127.0.0.1:16379` (6379 appartient à une autre stack) : la commande interrogerait une autre instance → **faux négatif sur la preuve d'isolation**. L'étape 5 (login) ne donne ni URL, ni corps, ni commande. La propriété « la protection tient pour `go test` direct hors make » (l. 30) n'est jamais validée.
- **Correctif :** uniformiser sur `docker compose exec -T redis redis-cli …` ; préciser l'étape 5 (URL 9001, JSON, arrière-plan) ; ajouter le contrôle direct `go test -tags=all -run TestIntDB ./tests` + sentinel intacte.

### M-03 — Invariant faux enseigné comme vrai : l'ordre `t.Setenv`/`godotenv.Load`
- **Trace :** S5-F1 · S6-#4 · S7-design-F-02 · S7-style-1 · S8-G5 · S9-IR-M3 · S9-CB-M3 · S9-style-F-03
- **Fait :** godotenv `Load` n'écrase jamais une variable déjà posée (docs officielles) — l'ordre est indifférent ; le plan en fait un point critique, exige un commentaire de code et l'inscrit aux Risques. L'invariant réel (client créé tant que `REDIS_DB` est présent) n'est pas énoncé, et la portée d'un `t.Setenv` (restauré à la fin du test) n'est pas documentée.
- **Correctif :** **DROP** la contrainte d'ordre partout ; écrire l'invariant réel ; D1 pour le mécanisme.

### M-04 — La garantie centrale n'est prouvée par aucun test ; le mode d'échec « loud » non plus
- **Trace :** S7-design-F-03 · S8-G4 · S9-IR-M5/M6 · S9-CB-M4
- **Fait :** T2 teste le parseur, pas le câblage — rien ne vérifie l'index effectif du client ; le chemin « `InitRedis()` retourne l'erreur sur valeur invalide » et la propagation côté e2edb ne sont pas testés. Une régression banale restaure le bug sans signal ; la validation T9 est manuelle, non rejouable par la CI. AGENTS.md impose pourtant les tests pour toute feature.
- **Correctif :** `require.Equal(t, 2, db.GetRedis().Options().DB, …)` après chaque `InitRedis` de suite (API confirmée par les docs go-redis) ; cas sans service `t.Setenv("REDIS_DB","16")` + `require.Error(t, db.InitRedis())` ; D4 pour la garde CI.

### M-05 — La paire tests/e2e partage l'index 2 sur une prémisse non garantie
- **Trace :** S5-F2 · S6-#3 · S7-design-F-01 · S8-G2 · S9-IR-M7 · S9-CB-M8
- **Fait :** deux acteurs destructeurs et globaux, mêmes formes de clés (`user_sessions:*`, `admin:<uuid>`) ; « les commandes sont séquentielles » n'est écrit nulle part comme règle, et AGENTS.md documente par ailleurs un `go test ./...` **sans `-p 1`** (interférence inter-paquets possible en sus du cas inter-commandes). Le risque n'apparaît pas dans la section Risques.
- **Correctif :** D2 (reco : index 3 pour e2e) ; à défaut invariant écrit + corriger la commande documentée dans AGENTS.md.

### M-06 — `cmd/e2edb` reste destructif pour l'index dev par défaut, sans garde
- **Trace :** S5-catalyst-1 · S7-design-F-05 · S8-G1 · S9-IR-M9 · S9-CB-M9/M12 · S9-style-F-07/F-08
- **Fait :** la doctrine de l'outil (« The guard rails come first ») s'arrête aux portes de Redis : `REDIS_DB` absent → index 1 → FLUSHDB des sessions dev, silencieusement ; la ligne `Usage:` (seul mode d'emploi découvrable) reste celle qui purge la db dev ; l'énumération des variables lues deviendrait fausse par omission ; la résolution se ferait après le DROP.
- **Correctif :** D3 (garde + `Usage:` + liste des vars + résolution dans `run()` avant `recreateDatabase`).

### M-07 — Le pendant PostgreSQL de la même classe de bug n'est pas nommé
- **Trace :** S5-F3 · S8-G3
- **Fait :** `make test` exécute `DELETE FROM public.tenant` sur la base **dev** (`golang_gin_db` chargée depuis `.env`) — impact plus grave que la purge Redis (données, pas seulement re-login) ; ni Contexte, ni Risques, ni Hors scope ne le mentionnent : le lecteur croira « l'environnement dev est protégé ».
- **Correctif :** une ligne dans « Hors scope » nommant le follow-up (base jetable `artisan_test`, symétrique d'`artisan_e2e`).

### M-08 — Phasage cassé (T1→T6 non compilable) et allowlist absente
- **Trace :** S9-IR-M10 · S9-CB-M7
- **Fait :** après T1, `main.go:144` et les trois paquets de test ne compilent plus jusqu'à T3/T5/T6 — un implémenteur qui lance `go build` entre deux tâches croira à une régression ; aucune phase ni allowlist consolidée (le standard l'exige).
- **Correctif :** 3 phases avec jalons (cœur+appelants → chaîne e2e → docs) ; allowlist créés/modifiés/interdits (proposition détaillée dans `implementation_readiness_report.md` §A-M10).

### M-09 — La nature de l'isolation n'est pas énoncée (sécurité)
- **Trace :** S7-design-F-04 · S8-G9
- **Fait :** l'isolation est un index logique sur une instance unique, mêmes identifiants, sans ACL — la doc officielle Redis le dit elle-même (*namespacing* ; « Don't use them to run multiple unrelated applications in a single Redis instance ») ; `REDIS_DB=0` est une valeur valide non refusée (l'échec explicite ne couvre que le mal formé, pas le mal choisi).
- **Correctif :** deux lignes dans la FeatureRange (nature de l'isolation + contrainte de déploiement), citation de la doc Redis.

### M-10 — Structure de la FeatureRange non conforme au standard
- **Trace :** S7-structural F1/F2/F3 · S6-#11
- **Fait :** pas de zone Objectifs/hors-objectifs (l'objectif est un paragraphe en gras ; les non-objectifs sont noyés dans la table des décisions) ; aucun renvoi depuis la portée vers « Risques » (l. 187) et « Hors scope » ; ordre canonique Alternatives → Tâches → Risques qui enterre les risques sous 140 lignes.
- **Correctif :** insérer `## Objectifs et hors-objectifs`, scinder la table (décisions / ce qui ne change pas), un renvoi explicite vers les Risques.

### M-11 — Mode d'échec « serveur à `databases` réduit » non documenté
- **Trace :** S5-F4 · S6-#8 · S7-design-F-07 · S8-G8 · S9-CB-M5 · S9-style-F-05
- **Fait :** la borne 0-15 est vraie pour les stacks du dépôt (`redis:8-alpine` dev + CI), pas pour un Redis managé réduit : `REDIS_DB=2` passe le parseur puis échoue à la première commande (`ERR DB index is out of range`), sans ping (choix assumé), avec une erreur qui ne mentionne pas `REDIS_DB`.
- **Correctif :** calibrer le message (« 0-15, plage du stack dev ; le `databases` effectif du serveur reste la référence ») + une ligne dans Risques. Préfixe `error: ` attendu par la convention du repo.

### M-12 — « Une seule valeur partagée » dupliquée en littéraux non reconnus
- **Trace :** S5-catalyst-3 · S9-CB-M6
- **Fait :** la valeur est écrite en dur dans 3 fichiers de test + 2 emplacements e2e sans source de vérité ; le plan revendique simultanément « une seule valeur » et « explicite plutôt qu'hérité ».
- **Correctif :** D5 (reconnaître les littéraux, la sentinel T9 prouve l'accord) ; rejeter le détecteur-by-test (R3).

---

## 🟢 Nits (une passe de relecture)

Terminologie unique « index Redis N » (« db N » réservé aux commandes, « base » proscrit — propagée dans T8) · « validation fenêtre 5 » : DROP · « jamais concurrents » : CALIBRATE si D2 rejeté · « ses tests hériteront » → « poseront eux-mêmes » · « l'app qu'il sert » → « l'app du cycle e2e » · « inconditionnel » : supprimer · doublons d'arguments (godotenv ×2, « dérive silencieuse » ×2 — garder l'argument dans la table des alternatives) · « Échec loud » → « échec explicite » · formes `Atoi` (`+2`/`02`/`-0`) à trancher + cas `15` · package de `db_test.go` (`db`) · message d'assertion « resolving the Redis database index » · rupture d'API externe de `InitRedis` à mentionner · bullet AGENTS.md `db/` à mettre à jour (`RedisDBIndex`, nouvelle signature) · avertissement AGENTS « ne pas poser `REDIS_DB` dans `.env` » · bénéfice utilisateur à énoncer (« une session dev survit à `make test` et `make e2e` ») · destination `progress.md` (entrée 23 + liste « Ce qui fonctionne ») · `-race` sur T9.2 (ou signaler l'omission) · `F-09` plage 27-38 = bloc de setup · `F-11` Hors scope à scinder · `F-14` « pas d'héritage implicite » à préciser · `F-16` redondances T5/Risques/T8 · injection `getenv` dans `InitRedis` : aligner sur la convention du repo ou assumer en une phrase.

---

## Remédiations disproportionnées — à NE PAS appliquer

Décisions du Stage 8 (zoom-out), confirmées par la consolidation :
1. **R1** Namespace `E2E_REDIS_DB` complet + constantes jumelles `DefaultRedisDB`/`DisposableRedisDB` — deux noms à synchroniser ; la garde D3 fait mieux pour zéro variable.
2. **R2** Parseur déporté dans `config/` ou `internal/redisconf` — coût de transitivité **zéro** (`db` = 32 lignes, go-redis déjà importé par l'outil) ; ajouter simplement la ligne au tableau des alternatives rejetées.
3. **R3** Détecteur de dérive relisant `Makefile`/`playwright.config.ts` depuis un test Go — couple la suite à l'arborescence ; la sentinel T9 suffit.
4. **R4** Purge scopée des suites Go sur leurs seules clés — le vrai correctif de la course inter-paquets mais hors périmètre, et ne couvre pas e2e ↔ suites : follow-up.
5. **R5** 4e index « au cas où », `PING` au boot (changerait le contrat opérationnel pour un gain nul).

---

## Plan d'action recommandé (ordre)

1. Trancher **D1-D6** (décisions ci-dessus).
2. Corriger **M-01** (ancres) et **M-02** (commandes T9 exécutables + contrôle `go test` direct).
3. Réécrire l'invariant (**M-03**) et le mécanisme (**D1**) ; ajouter la preuve (**M-04**) et le test du mode d'échec.
4. Garde `cmd/e2edb` (**M-06/D3**) + `Usage:` + en-tête + ordre de résolution.
5. Lignes sécurité (**M-09**), Hors scope PostgreSQL (**M-07**), restructure FeatureRange (**M-10**), calibrage borne (**M-11**), reconnaissance des littéraux (**M-12/D5**).
6. Phasage 3 phases + allowlist (**M-08**).
7. Passe nits.
8. Rejouer l'audit Stage 9 sur le plan corrigé (1 re-audit max, cap du protocole).

## Matrice de traçabilité (condensée)

| Consolidé | S5 | S6 | S7 | S8 | S9 |
|---|---|---|---|---|---|
| M-01 ancres | F6 | #1 | style-annexe | G7 | IR-M1/M2, CB-M1/M2, style-F01/02/09 |
| M-02 preuves T9 | F7 | #2, #10 | — | G6 | IR-M8, CB-N7, style-F04 |
| M-03 invariant faux | F1 | #4 | design-F02, style-1 | G5 | IR-M3/M4, CB-M3, style-F03 |
| M-04 preuve par test | — | — | design-F03 | G4 | IR-M5/M6, CB-M4 |
| M-05 index partagé | F2 | #3 | design-F01 | G2 | IR-M7, CB-M8 |
| M-06 garde e2edb | cat-1 | #7 | design-F05 | G1 | IR-M9, CB-M9/M12, style-F07/08 |
| M-07 pendant PG | F3 | — | — | G3 | — |
| M-08 phasage | — | — | — | — | IR-M10, CB-M7 |
| M-09 nature isolation | — | — | design-F04 | G9 | — |
| M-10 structure FR | — | #11 | structural-F1/2/3 | — | — |
| M-11 borne serveur | F4 | #8 | design-F07 | G8 | CB-M5, style-F05 |
| M-12 littéraux | cat-3 | — | — | (R3) | CB-M6 |

**Faits vérifiés sans réserve (tous stages, 3 passes indépendantes) :** diagnostic exact ; 4/4 appelants `InitRedis` ; `main_test.go` non impacté ; mécanisme e2e cohérent (préfixe inline > `.env` ; `webServer.env` fusionné ; `-create-admin` sans Redis) ; « 4/4 » e2e correct ; CI tenable sans modification ; Swagger correct ; toutes les ancres exactes sauf M-01 ; 0 Blocker ; 0 « contradicts the code ».
