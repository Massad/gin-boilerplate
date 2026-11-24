# Fenêtres de traitement — points 6-9 non verts du plan REDIS_DB

**Objet :** rendre verts les quatre points restants du bilan de la fenêtre de
résolution. Deux fenêtres auto-contenues, exécutées chacune par un agent
**neuf** (aucun historique), dans l'ordre **A puis B**. Chaque agent lit ce
brief + les fichiers listés, exécute **strictement** son périmètre, prouve ses
critères de sortie.

| # | Point non vert | Fenêtre |
|---|---|---|
| 6 | Majors du 2e re-audit corrigés après le cap, sans re-vérification | A |
| 7 | Nits résiduels (« fail loudly », garde CI filet partiel, `REDIS_DB=0`, pas de ping) | A |
| 8 | Implémentation T1-T9 non commencée | B |
| 9 | Validation T9 non exécutée | B |

Deux fenêtres, pas quatre : les points 6-7 sont deux passes sur le même
fichier avec la même règle de périmètre ; les points 8-9 sont un seul flux
d'exécution (la validation T9 est imbriquée dans les phases). **B ne démarre
qu'après le message final de A** (points 6-7 verts).

**Mode d'emploi :** ouvrir une nouvelle fenêtre de contexte par fenêtre, avec
le prompt d'amorçage donné en tête de chaque section. Si les conditions
d'entrée d'une fenêtre ne sont pas réunies (déjà jouée, ou état divergé), le
signaler au lieu d'exécuter.

**Ground truth (docs officielles via Context7 — accepter telles quelles, ne
jamais relire le module cache Go) :** go-redis `Client.Options() *Options`
retourne les options passées à `NewClient` (champ `DB`) ; godotenv `Load` ne
surcharge jamais une variable déjà posée (seul `Overload`) ; les db logiques
Redis sont du *namespacing* (« Don't use them to run multiple unrelated
applications in a single Redis instance »), `SELECT` hors plage →
`ERR DB index is out of range`.

**Interdits communs :** ne pas relire le module cache Go ; ne pas modifier
les rapports du dossier `.kilo/plans/1795535384311-review/` ni
`resolution_fenetre.md` ; un commit n'intervient que sur demande explicite
de l'utilisateur.

---

## Fenêtre A — passe documentaire sur le plan (points 6-7)

**Prompt d'amorçage (copier tel quel) :**

> Lis `.kilo/plans/1795541661880-redis-db-fenetres-points-6-9.md` et exécute
> strictement la **Fenêtre A** ; prouve les critères de sortie.

**Objet :** re-vérifier ciblément les trois passages corrigés après le second
re-audit (point 6), résoudre le nit résiduel de T2 et confirmer les
arbitrages assumés (point 7), et inscrire dans le plan le chemin d'exécution
(complément documentaire des points 8-9).

**Périmètre :** UNIQUEMENT le fichier plan
`.kilo/plans/1795535384311-redis-db-isolation.md` (écriture). Le dépôt est
lisible en lecture seule pour la vérification. Aucun fichier source, aucune
doc, aucun rapport n'est modifié. Cette passe ne rouvre **pas** la revue
Goldfish complète (le cap sur les re-audits complets tient) : c'est une
vérification ciblée des passages listés, demandée par l'utilisateur.

**Entrée (rejouer avant de commencer) :**
1. Le plan existe : 493 lignes, titre « Isolation REDIS_DB : dev (index 1),
   suites Go (2), cycle e2e (3) ».
2. La chaîne `must fail loudly` figure en T2 (~l. 174) ; aucun bloc
   « **Statut** » sous le titre ; aucune cartographie en tête de T9.
3. Rien n'est implémenté : `db/db_test.go`, `tests/main_test.go`,
   `middleware/main_test.go`, `controllers/main_test.go` absents du dépôt.

**Lecture obligatoire avant d'éditer :** le plan en entier ; pour A1, les
fichiers du dépôt cités dans les vérifications.

### A1. Re-vérification ciblée des trois passages corrigés post-cap (point 6)

Vérifier contre le dépôt, en lecture seule, indépendamment ; tout écart
relevé est corrigé dans le plan avant de continuer :

1. **Garde D3 (T9 phase 1, jalon 4)** : le message attendu « REDIS_DB is
   unset: refusing to flush the dev Redis index (1) » est identique, au mot
   près, à celui du snippet `redisFlushIndex` (T4). Causalité : la garde
   existe dès T4 (phase 1), le préfixe `REDIS_DB=3` du Makefile n'est posé
   qu'à T7 (phase 2) ; l'échec survient à l'étape e2edb — après `services_up`,
   **avant toute action destructive** (aucun DROP : la résolution précède
   `recreateDatabase`, aucun FLUSH).
2. **Preuve D5 recadrée (T9 phase 2, étape 5)** : le contrôle `dbsize` est
   exécutable (`docker compose exec -T redis redis-cli -n 2 dbsize`) et
   décidable — capture juste avant `make e2e` (donc après le `make test` de
   l'étape 3 et ses résidus), comparaison juste après ; la paire
   {index 3 non vide, dbsize(2) inchangé} prouve bien l'accord
   Makefile/Playwright, aucune clé « e2e » n'étant distinguable par nom.
3. **Cycle de vie de l'app dev (T9 phase 2, étapes 4b/4c/4f)** : `main.go:97`
   charge bien `.env` via godotenv ; le patron d'attente référencé existe
   (`ci.yml:118-129`) ; les commandes sont du POSIX sh valide (`&`, `$!`,
   `kill -0`, `seq`) ; le port dev est 9001 ; `POST /v1/login` répond 200
   avec `"kind":"superadmin"` à la racine du JSON pour un compte superadmin
   (`models/login.go`).

### A2. Nit « fail loudly » (point 7)

Dans T2, remplacer la ligne :

```go
require.Error(t, InitRedis(), "an out-of-range REDIS_DB must fail loudly")
```

par :

```go
require.Error(t, InitRedis(), "an out-of-range REDIS_DB must be rejected explicitly")
```

Motif : la règle « échec loud → échec explicite » partout (P11 du brief de
résolution) ; cette chaîne était le dernier reste. La chaîne du chemin
nominal (« the parsed index must reach the client ») est inchangée.

### A3. Arbitrages assumés des trois autres nits (point 7)

Vérifier la présence explicite de chaque marqueur (ne rien inventer si déjà
présent ; ajouter le marqueur « assumé » sinon) : pas de ping (note T1 +
Risques « Serveur à `databases` réduit ») ; `REDIS_DB=0` accepté (Risques
« `REDIS_DB` mal choisi ») ; garde CI (T7 « Filet partiel assumé »).

### A4. Bloc Statut (complément documentaire des points 8-9)

Insérer immédiatement sous la ligne de titre, avant « ## Contexte » :

> **Statut** : document prêt pour exécution — l'étape suivante est
> l'implémentation T1-T9 (3 phases, § Phasage), du ressort d'un agent
> d'implémentation ; chaque phase se ferme sur son jalon T9 ; commit
> uniquement sur demande explicite.

### A5. Cartographie passages ↔ jalons (point 6, ancrage documentaire)

Insérer entre « ### T9. Validation (jalons par phase) » et
« **Phase 1 (T1-T6)** — jalon : » :

> Les jalons ci-dessous vérifient en particulier les passages sensibles
> introduits en dernière passe de correction — toute erreur s'y manifeste
> en jalon rouge, pas en faux vert :
>
> - message exact de la garde D3 → phase 1, jalon 4 : exit 1 avec « REDIS_DB
>   is unset: refusing to flush the dev Redis index (1) » ;
> - preuve D5 (une clé « e2e » n'est pas distinguable par nom) → phase 2,
>   étape 5 : index 3 non vide **et** `dbsize` de l'index 2 identique avant
>   et après `make e2e` ;
> - cycle de vie de l'app dev (binaire, attente bornée, arrêt vérifié) →
>   phase 2, étapes 4b/4c/4f : le PID tué est celui du serveur, le port
>   9001 est confirmé libéré.

### A6. Littéraux d'exécution des étapes 4b/4c (point 6, raffinement zéro-décision)

- **4b** : remplacer « puis lancer **le binaire** en arrière-plan depuis la
  racine (il charge `.env` lui-même, main.go:97) » par « puis
  `/tmp/artisan-dev & pid=$!` depuis la racine (le binaire charge `.env`
  lui-même, main.go:97) » — le PID devient littéral, réutilisé en 4c/4f.
- **4c** : remplacer la commande de la boucle par celle-ci (sonde **sans**
  `--fail` : toute réponse HTTP = prêt — l'app dev ne sert le SPA que si
  `app/dist` existe, un 404 resterait une réponse valide ; refus de connexion
  = pas prêt ; branche d'échec explicite post-boucle) :

  ```
  for i in $(seq 1 50); do curl -s -o /dev/null --max-time 2 http://127.0.0.1:9001/ && break; kill -0 "$pid" || { echo "le serveur est mort"; exit 1; }; sleep 0.2; done; curl -s -o /dev/null --max-time 2 http://127.0.0.1:9001/ || { echo "app non disponible après 50 tentatives"; exit 1; }
  ```

**Sortie (critères) :**
1. A1 : les trois passages vérifiés — constats rapportés (0 écart, ou écarts
   corrigés dans le plan).
2. A2-A6 appliqués ; relecture du plan modifié (cohérence, aucune référence
   cassée, numérotation T9 intacte).
3. Message final : constats A1, modifications appliquées, statut des points
   6-7 (vert), confirmation que le plan est prêt pour la Fenêtre B.

---

## Fenêtre B — exécution du plan REDIS_DB (points 8-9)

**Prompt d'amorçage (copier tel quel) :**

> Lis `.kilo/plans/1795541661880-redis-db-fenetres-points-6-9.md` et exécute
> strictement la **Fenêtre B** : implémenter le plan
> `.kilo/plans/1795535384311-redis-db-isolation.md` (phases 1 à 3, jalons
> T9) ; prouve les critères de sortie.

**Objet :** implémenter T1-T9 du plan REDIS_DB (point 8) et faire passer
tous les jalons de validation T9 (point 9). Fenêtre d'implémentation : l'agent
doit pouvoir écrire le code source.

**Périmètre :** strictement l'allowlist du plan (§ Allowlist — listes
Créés / Modifiés / Interdits ; le plan est la source de vérité, elle n'est
pas recopiée ici). Aucune écriture hors allowlist.

**Entrée (rejouer avant de commencer) :**
1. Fenêtre A terminée : le plan contient le bloc « **Statut** » et la
   cartographie en tête de T9 ; la chaîne « fail loudly » n'y figure plus.
2. `.env` présent à la racine (copié depuis `.env_rename_me`), jamais commité.
3. Stack dev opérationnelle : `make services_up` vert (PostgreSQL sur
   127.0.0.1:15432, Redis sur 127.0.0.1:16379).
4. Prérequis tests : poppler réel dans le PATH (`pdftotext`) pour les tests
   PDF ; prérequis e2e : `pnpm -C app exec playwright install chromium` déjà
   exécuté une fois.
5. `git status` relu avant de commencer (état de départ connu ; travail sans
   commit — convention : commit sur demande explicite).

**Lecture obligatoire avant d'éditer :** le plan en entier (spécification
exécutable : snippets verbatim, ancres fichier:ligne, messages exacts) ;
`AGENTS.md` (conventions : `-mod=readonly` systématique, requêtes SQL
uniquement dans `db/queries/`, mapping d'erreurs pgx, validation via
`forms.Translate`, tests obligatoires pour toute feature).

**Programme (le plan fait foi ; cet ordre est impératif) :**
1. **Phase 1 (T1-T6)** — livrer d'un seul bloc (la rupture de signature
   `InitRedis` rend `main.go` et les paquets de test non compilables entre
   T1 et T3/T5/T6). Jalon T9 phase 1 : `gofmt -l .` vide ;
   `go vet -mod=readonly -tags=all ./...` vert ;
   `go test -mod=readonly -count=1 -race ./db` vert ; `make test` complet
   vert ; **`make e2e` en échec attendu** (exit 1, message exact de la
   garde D3 — état voulu documenté au § Phasage, ne pas le « réparer »).
2. **Phase 2 (T7)** — préfixe `Makefile:155`, `webServer` Playwright, garde
   CI. Jalon T9 phase 2 : les 6 étapes du plan, dans l'ordre — marqueur
   `admin:sentinel` ; contrôle `go test` direct ; `make test` + marqueur
   intact ; preuve app dev en 1 (démarrage binaire, attente bornée, arrêt
   vérifié) ; `make e2e` vert + preuves d'index (3 non vide,
   `dbsize(2)` inchangé) ; nettoyage du marqueur.
3. **Phase 3 (T8)** — documentation + mémoire (AGENTS, README,
   techContext, systemPatterns, activeContext annoté daté, progress entrée
   23, HANDOFF). Jalon T9 phase 3 : `git status` + `git diff` relus face à
   l'allowlist ; pas de commit.
4. Tout au long : obligations AGENTS — `go vet -mod=readonly ./...` et tests
   du/des paquets touchés après chaque modification ; jamais terminer avec
   des tests en échec ou sautés par erreur ; si l'environnement manque
   (PostgreSQL/Redis/poppler indisponibles), le signaler explicitement.

**Interdits :** aucune écriture hors allowlist ; ne jamais relâcher la garde
D3 (le `make e2e` rouge de la phase 1 est l'état voulu) ; ne jamais passer en
superuser pour « corriger » un comportement RLS (0 ligne / 42501 attendus
sont des comportements prouvés) ; ne pas relire le module cache Go ;
commit uniquement sur demande explicite de l'utilisateur.

**Sortie (critères) :**
1. Jalons T9 des trois phases verts, dans l'ordre — dont : marqueur
   `admin:sentinel` intact après `make test` **et** après `make e2e` ;
   sessions e2e visibles sur l'index 3 ; `dbsize` de l'index 2 inchangé
   avant/après `make e2e` ; session dev vivant en 1 après le cycle.
2. `git status` + `git diff` relus face à l'allowlist (aucun fichier hors
   liste) ; pas de commit.
3. Message final : tableau des jalons (phase → attendu → constat), fichiers
   créés/modifiés face à l'allowlist, résidus éventuels (ex. orphelins sans
   TTL sur les index 2/3 — attendus), et bilan des points 6-9 (tous verts,
   ou restants avec motif).
