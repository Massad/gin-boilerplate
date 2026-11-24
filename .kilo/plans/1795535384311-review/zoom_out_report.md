# zoom_out_report.md — Stage 5 : Initial Zoom Out

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md` (207 lignes)
**Périmètre :** fichier complet (1..207)
**Reviewers :** egm-zoom-out (Skeptical Architect), egm-inventive-catalyst (TRIZ) — contextes frais
**Standard appliqué :** Zoom-Out Standard (zoom_out + catalyst : aucun)

---

## A. egm-zoom-out — Skeptical Architect

**Verdict global : PASS conditionnel** — la conception architecturale est saine (problème racine traité, solution minimale, conventions du repo réutilisées). Les réserves portent sur la mécanique d'isolation elle-même et sur une dérive de problème racine non assumée côté PostgreSQL.

### F1 — **Major** — La protection `t.Setenv` expire à la fin du test qui l'a posée
- **Portée :** T5 (l. 115-122), T6 (l. 124-131), décision « Tests Go protégés à la source » (l. 30), Risques (l. 189-192).
- **Rationale :** `t.Setenv` restaure la valeur précédente à la fin du test. Dans le paquet `tests`, `REDIS_DB=2` n'existe que pendant `TestIntDB` ; l'isolation ne tient aujourd'hui que par accident de sequencing (le client global garde `DB=2` pour le reste du processus). Le jour où une suite appelle `db.InitRedis()` après `TestIntDB`, la variable est restaurée → résolution par défaut → **db 1 → écritures et purges dans les sessions dev, en silence** — exactement le mode de défaillance que le plan prétend rendre impossible (« Échec loud », l. 28). Remède le plus simple : **un `TestMain` par paquet** (`os.Setenv("REDIS_DB", "2")` pour toute la durée du processus), qui supprime aussi la prétendue contrainte d'ordre « `t.Setenv` avant `godotenv.Load` » (l. 189) — cette contrainte n'existe pas : `t.Setenv` posé après `godotenv.Load` donne le même résultat (`t.Setenv` écrase l'env du process, `godotenv.Load` n'écrase rien). Un risque documenté comme réel qui ne l'est pas coûte plus cher qu'il ne protège.

### F2 — **Major** — db 2 partagée entre `make test` et `make e2e` : la même classe de bug, un cran plus bas
- **Portée :** l. 26 (décisions), l. 44 (alternatives rejetées), l. 196 (risques).
- **Rationale :** l'argument qui rejette un 3e index est « les commandes sont séquentielles » — c'est exactement l'hypothèse qui a déjà failli (app dev et `make test` étaient eux aussi « jamais concurrents »). Un `make e2e` lancé pendant un `make test` fait un `FLUSHDB` en db 2 et efface les sessions des suites Go en cours → 401 aléatoires, échecs flaky difficiles à diagnostiquer. Coût de l'alternative : une valeur documentée de plus (`REDIS_DB=3` pour e2e). À défaut, l'invariant doit être énoncé comme règle du repo (« ne jamais lancer `make e2e` et `make test` en parallèle ») — ce que le plan ne fait pas.

### F3 — **Major** — Le plan résout Redis, mais la classe de bug « les tests utilisent l'environnement dev » reste entière côté PostgreSQL
- **Portée :** Contexte (l. 7-16), Hors scope (l. 203-207).
- **Rationale :** le mécanisme destructeur le plus brutal n'est pas Redis : `tests/article_test.go:909` exécute `DELETE FROM public.tenant` sur la base dev (`golang_gin_db` chargée depuis `.env`) — les tenants du développeur sont supprimés par `make test`. Séparer les index Redis sans mentionner ce pendant laisse croire au prochain lecteur que « l'environnement dev est protégé ». Sorties propres : (a) l'ajouter explicitement au « Hors scope » comme follow-up nommé ; (b) traiter la symétrie — base jetable dédiée (`artisan_test`) pour les suites Go. Au minimum, le plan doit le dire.

### F4 — **Nit** — Borne `[0,15]` codée en dur : constante magique adossée à une config serveur que l'app ne contrôle pas
- **Portée :** T1 (l. 56-66), décision l. 27.
- **Rationale :** `databases 16` est le défaut de Redis, pas un contrat de l'app ; un serveur `--databases 32` accepte l'index 16 que le client refuse, et Redis répond déjà `ERR DB index is out of range` au `SELECT`. Version minimale : parsing numérique + refus du négatif, laisser Redis refuser le hors-plage (l'échec reste « loud »). Si la borne reste, la constante mérite un commentaire sur son origine. Détail : les messages d'erreur du repo préfixent `error: ` (`config/secrets.go:20`) — la fonction proposée ne le fait pas.

### F5 — **Nit** — `cmd/e2edb` importe le package runtime `db` : l'outil d'infra perd son indépendance
- **Portée :** l. 29 (décisions), T4 (l. 102-113).
- **Rationale :** `cmd/e2edb` n'importait jusqu'ici aucun paquet applicatif (pgx + go-redis seulement) ; importer `db` le couple au paquet qui porte le client global `RedisClient`. La fonction de parsing trouverait une place plus neutre dans `config`, qui possède déjà la validation d'env avec la signature identique `getenv func(string) string` (`config.ValidateSigningSecrets`) et aucun état global. Le choix actuel reste défendable, mais l'argument « un seul parseur » est mieux servi par `config` que par `db`.

### F6 — **Nit** — Deux ancres de lignes fausses, dont une qui masque une dérive de documentation
- **Portée :** T7 l. 135 (« lignes ~121-122 »), T8 l. 162 (`activeContext.md` ligne 60).
- **Rationale :** l'appel e2edb est en **Makefile:155** (le fichier en compte 167) ; `~121-122` pointe sur la recette `migrate_up` — un implémenteur qui suit l'ancre littéralement modifie la mauvaise ligne. Pour `activeContext.md` : la ligne 60 parle de joignabilité WSL sans numéro de base ; l'affirmation qui devient **fausse** après le changement est en **activeContext.md:29** (« le `FLUSHDB` purge la base Redis 1 (sessions de l'app de dev) »). `activeContext.md` étant le fichier d'état courant, c'est la ligne 29 qu'il faut amender, pas la 60.

### F7 — **Nit** — La propriété centrale annoncée n'est pas dans le plan de validation
- **Portée :** T9 (l. 169-185), décision l. 30.
- **Rationale :** T9 valide `make test` (qui source `.env`), jamais l'invocation directe `go test -tags=all ./tests` — précisément le cas que la décision « protégés à la source » prétend couvrir. Un contrôle d'une ligne (`go test -tags=all -run TestIntDB ./tests` puis `redis-cli -n 1 keys '*'` inchangé) ferme la boucle.

### Vecteurs sans finding
- **Context Sufficiency :** rien — le Contexte énonce des faits vérifiables, l'objectif et la garantie de non-régression.
- **Root Problem Focus :** traité à la racine sur le périmètre Redis (séparation des index, pas un symptôme) ; la purge ciblée est correctement rejetée (clés identiques dev/test, `models/auth.go:119-126, 305, 359-361`). Seule réserve : F3 (pendant PostgreSQL).
- **YAGNI :** rien — une variable d'env, une fonction exportée, aucun abstraction spéculative ; suppression du variadic = bon nettoyage.
- **Anti-Weeds :** rien — altitude « quelle base, qui résout, qui épingle » ; l'énumération T8 est justifiée (répare des énoncés faux).
- **Réinvention :** rien — injection `getenv`, échec loud `log.Fatal` aligné `main.go:137-142`, tests table-driven : les précédents exacts (`internal/db/postgresDSN`, `config/secrets.go`) sont cités et suivis.

### Faits vérifiés (contrôle des affirmations du plan)
| Affirmation | Vérification |
|---|---|
| `cmd/e2edb/main.go:54` const, FLUSHDB l. 160 | exact |
| purges `article_test.go:914-928` / `login_test.go:267-279` | exact |
| 4 sites `db.InitRedis(1)` | exact |
| `-create-admin` retour anticipé sans `InitRedis` (main.go:112-117) | exact |
| `playwright.config.ts:57` env | exact |
| ancres AGENTS.md 9/33/59/68/83, README 61-64, HANDOFF 27, techContext 22/41 | exactes |
| compose redis:8-alpine sans `--databases` → 16 bases | exact |
| CI inchangée tenable | exact |
| « godotenv n'écrase pas une var déjà posée » | exact pour `Load` — mais la contrainte d'ordre n'en découle pas (F1) |

### Recommandation minimale (ordre coût/bénéfice)
1. Remplacer les `t.Setenv` par un `TestMain` par paquet (F1).
2. Trancher F2 explicitement : 3e index pour e2e ou invariant écrit.
3. Nommer le pendant PostgreSQL dans « Hors scope » (F3).
4. Corriger les deux ancres (Makefile:155, activeContext.md:29) et ajouter le contrôle « go test direct » à T9 (F6, F7).
5. Optionnel : borne → parsing seul, préfixe `error: ` (F4) ; helper déplacé vers `config` (F5).

---

## B. egm-inventive-catalyst — TRIZ

**Étape 0 :** des contradictions techniques réelles existent — le plan n'est pas « déjà équilibré ». Trois axes où il accepte des compromis présentés comme des lois.

### Ce qui n'est PAS un faux compromis (à ne pas rediscuter)
- L'axe d'isolation choisi (index Redis) plutôt qu'un conteneur séparé — préserve le `FLUSHDB` déterministe. Décision solide.
- Le refus d'exporter `REDIS_DB` depuis le Makefile pour `make test` — la protection doit vivre dans la source. Décision solide.
- Le refus du repli silencieux sur valeur invalide — cohérent avec `config.ValidateSigningSecrets`. Décision solide.
- Le refus de la purge ciblée comme solution principale — l'argument « mêmes clés, même db » est correct, à condition de l'appliquer aussi à la nouvelle paire tests/e2e (constat 2).

### Constat 1 — **Major** — `cmd/e2edb` reste destructif par défaut : l'objectif du plan n'est pas atteint sur le chemin d'invocation documenté de l'outil
- **Portée :** `cmd/e2edb/main.go:21` (usage documenté), `:43-58`, `:147-164` ; plan l. 27-29, l. 193-196.
- **Faux compromis :** le plan fusionne dans un seul couple (variable, défaut) deux besoins opposés : « quelle base l'app utilise » (défaut sûr = 1) et « quelle base l'outil purge » (défaut sûr = jetable). En imposant le défaut 1 aux deux consommateurs, l'invocation documentée dans l'en-tête de l'outil continue de purger les sessions dev. Le garde-fou « échec loud » ne couvre que les valeurs *invalides*, alors que la valeur par défaut de l'outil est précisément la valeur *dangereuse*.
- **Résolution proposée :** séparation des espaces de noms + précondition active, sur le modèle déjà présent dans le dépôt (`E2E_DB_NAME`, `E2E_DB_ADMIN_*` distincts de `DB_NAME`, `DB_*`) : `REDIS_DB` (défaut 1) pour l'app ; `E2E_REDIS_DB` (défaut 2) pour l'outil — parseur unique (`db.RedisDBIndex(getenv, name, fallback)`, un cœur, deux namespaces), constantes `db.DefaultRedisDB = 1` / `db.DisposableRedisDB = 2` ; **garde-fou actif** dans `flushRedis` sur le modèle du garde-fou anti-DROP `validateTarget` (`:104-118`) : refuser de purger l'index de dev sauf opt-in explicite. La recette `make e2e` épingle les deux valeurs via une seule variable make.
- **Responsabilités secondaires :** une variable de plus à documenter ; le garde-fou interdit une purge manuelle légitime sans opt-in (à documenter sinon contournement) ; l'accord tool/app devient une exigence de recette (une variable, deux affectations) à laisser visible dans le commentaire.

### Constat 2 — **Major** — La collision tests Go ↔ e2e est déplacée en db 2, pas supprimée : le plan s'applique à lui-même l'argument qu'il a rejeté
- **Portée :** `tests/article_test.go:914-928`, `tests/login_test.go:267-279`, `cmd/e2edb/main.go:147-164`, `app/e2e/login.spec.ts:10-11, 42-75`, `models/auth.go:119-127, 260, 305, 359-361`.
- **Faux compromis :** l'argument anti-purge-ciblée (l. 41) s'applique **verbatim** à la nouvelle paire : suites Go et app e2e partagent la db 2 et les mêmes formes de clés ; les nettoyages Go suppriment *toutes* les clés `admin:*` et `user_sessions:*` de la base (le commentaire « Flush this suite's Redis keys » est plus restrictif que le code). L'app e2e écrit dans cette base ; CUJ1 se connecte en superadmin → une clé `admin:<uuid>` que la suite Go supprime. Le plan fonde la sécurité sur une prémisse organisationnelle (« les commandes sont séquentielles », l. 26/44), pas sur un invariant technique.
- **Résolution :** (a) valeurs distinctes par chaîne — db 2 pour les suites Go, db 3 pour la chaîne e2e, la variable make alimentant l'appel e2edb ET l'env du webServer Playwright (la « duplication » refusée en l. 44 est exactement l'invariant qui remplace la prémisse ; résolution TRIZ canonique : séparation dans l'espace) ; (b) cumulable : purge scopée dans les suites Go (UUID déjà connus via `accessUUIDFromToken`, `login_test.go:295` → `Del` ciblés au lieu de `Keys("*")` + `Del` global ; bonus : disparition de `KEYS` O(N)).
- **Responsabilités secondaires :** (a) une constante et une ligne de doc de plus ; orphelins en db 3 purgés par chaque cycle ; (b) la purge scopée perd « base vierge » — orphelins de runs crashés subsistent (bornés par les TTL : accès 15 min, refresh 7 j, admin 8 h, `models/auth.go:74-78, 285-287`).

### Constat 3 — **Major** — « Une seule valeur partagée » est dupliquée en 5 sites, et l'invariant « même env » est faux à l'échelle du cycle e2e
- **Portée :** l. 26, 29, 31 ; T5/T6/T7 ; le littéral `"2"` apparaît dans 3 fichiers de test Go, la recette Makefile et la config Playwright.
- **Faux compromis :** la table revendique une valeur unique, refuse « deux index » au motif de la charge documentaire — puis s'impose cinq littéraux sans source de vérité. Et l'invariant « le même env » est faux **entre les deux processus du cycle e2e** : le tool tire son index d'un préfixe inline make, l'app du bloc `env` de `playwright.config.ts` — deux sources indépendantes qui coïncident par convention. Une divergence (Makefile à 3, config oublié) donne un FLUSHDB **silencieusement sans effet** sur les données de l'app : aucune erreur, symptôme de flakiness de rotation. « Explicite plutôt qu'hérité » (l. 31) et « une seule valeur » (l. 26) sont mutuellement exclusifs ici.
- **Résolution :** source unique Go (`const DisposableRedisDB = 2` dans `db/`, les tests posent `strconv.Itoa(db.DisposableRedisDB)`) ; chaîne e2e : une variable make `E2E_REDIS_DB` alimentant e2edb ET l'invocation Playwright, `playwright.config.ts` lisant `process.env.REDIS_DB` et **échouant bruyamment** si absent ; détecteur de dérive : un test Go relit `Makefile` et `app/playwright.config.ts` et vérifie l'égalité avec `db.DisposableRedisDB` (motif `check_ports` / `git diff --exit-code internal/db` déjà présent dans le dépôt).
- **Responsabilités secondaires :** le test détecteur couple la suite Go à l'arborescence (faux échecs au renommage) ; la config Playwright stricte casse le `playwright test` lancé à la main (comportement souhaité mais à documenter) ; l'indirection `$(E2E_REDIS_DB)` rend la recette moins lisible ; le détecteur ne vit que dans la suite taggée `all`.

### Constat 4 — **Nit** — Les opérations destructrices n'ont aucune précondition dans le code : l'invariant critique vit dans les commentaires
- **Portée :** `db/db.go` (client global), `cmd/e2edb/main.go:147-164`, `tests/article_test.go:914-928`, `tests/login_test.go:267-279`.
- **Faux compromis :** la suppression du variadic transforme un paramètre injectable en état ambiant global ; `REDIS_DB=1` est une valeur parfaitement valide au sens du parseur, et un `FLUSHDB` contre elle détruit les sessions dev sans signal. L'invariant « ne jamais purger l'index dev » n'existe que dans les commentaires.
- **Résolution :** précondition de code sur le modèle anti-DROP `validateTarget` : `db.RequireDisposableIndex(index)` appelé par chaque chemin destructeur, refusant l'index de dev sauf opt-in. Variante plus faible : `InitRedis` expose l'index résolu, chaque test destructeur exige `require.Equal(t, db.DisposableRedisDB, db.RedisDB())` **avant** de purger — le test devient lui-même la preuve de l'isolation.
- **Responsabilités secondaires :** garde adossée au singleton qu'elle contrôle (valeur ajoutée faible) ; préférer à terme un index passé en paramètre explicite aux fonctions de nettoyage (refactor hors périmètre).

### Constat 5 — **Nit** — La borne `[0,15]` fige une hypothèse de déploiement dans le parseur
- **Portée :** T1 ; `docker-compose.yml:25-36` ; `.github/workflows/ci.yml:33-41`.
- **Faux compromis :** la plage `databases` est une propriété de configuration **serveur**, pas du client. Le parseur peut refuser un index valide (serveur `databases 32`) **et** en accepter un invalide (serveur `databases 4`, index 8 → échec opaque au premier `SELECT`). L'« échec loud » échoue au mauvais moment, avec le mauvais message.
- **Résolution :** le parseur valide la **syntaxe** (entier ≥ 0), la plage reste au serveur ; traduire l'erreur « DB index is out of range » en message actionnable.
- **Responsabilités secondaires :** `CONFIG GET databases` exige une ACL (repoussée hors scope) ; un `SELECT` de contrôle au boot changerait le contrat opérationnel (démarrage tolérant → refus si Redis absent).

### Constat 6 — **Nit** — La protection est par site d'appel, pas par paquet : un futur test qui touche Redis repart silencieusement en db 1
- **Portée :** T5/T6 ; `tests/article_test.go:150-165` ; `tests/login_test.go:259-280` (nettoyage dans un **autre** test que celui qui a initialisé Redis) ; `tests/article_test.go:906-932`.
- **Faux compromis :** les purges destructrices s'appuient sur le client global initialisé par `TestIntDB` — couplage implicite entre tests ; `t.Setenv` étant restauré, `REDIS_DB` n'est **plus** posé pendant `TestCleanUp`. Un nouveau fichier de test qui appelle `db.InitRedis()` sans poser `REDIS_DB` repart en db 1 sans erreur. La fragilité repose en outre sur un comportement de bibliothèque subtil (un passage futur à `godotenv.Overload` casserait l'isolation silencieusement).
- **Résolution :** un `TestMain` par paquet taggé `all` posant `REDIS_DB` avant tout test + helper unique `db.InitTestRedis(t)` comme seule voie d'ouverture d'un client Redis dans un test. L'index devient une propriété du paquet ; l'ordre `t.Setenv`/`godotenv.Load` cesse d'être un invariant fragile.
- **Responsabilités secondaires :** restructure 4 paquets (diff plus large) ; le `TestMain` doit porter le tag `all` ; la convention « ne plus appeler `db.InitRedis` directement dans un test » à faire respecter par AGENTS.md.

### Constat 7 — **Nit** — Ancre de fichier erronée dans T7 et critère de validation figé
- **Portée :** T7 (« lignes ~121-122 ») ; T9 étape 6 (« 4/4 »).
- **Faux compromis :** l'appel `cmd/e2edb` est à **Makefile:155** (recette `e2e` à partir de la ligne 152) ; l'ancre du plan est fausse. Le critère « 4/4 » est correct en nombre (3 scénarios CUJ + projet setup) mais figé comme résultat plutôt qu'écrit comme attente vérifiable.
- **Résolution :** recaler l'ancre sur Makefile:155 ; reformuler T9.6 en « 3 scénarios + setup verts ». Les autres ancres vérifiées sont exactes.

### Ordre de traitement recommandé (catalyst)
1. Constat 1 (tool destructif par défaut) — garde-fou actif + namespace `E2E_REDIS_DB`.
2. Constat 2 (collision tests ↔ e2e en db 2) — décider explicitement la prémisse « jamais concurrents » ; si inacceptable, deux valeurs par chaîne (2 Go, 3 e2e).
3. Constat 3 (valeur dupliquée 5 fois, invariant « même env » faux entre processus) — constante Go unique + variable make unique + config Playwright fail-loud + détecteur de dérive.
4. Constats 4, 5, 6 — précondition de code sur les chemins destructeurs, séparation syntaxe/plage, `TestMain` par paquet.
5. Constat 7 — recaler l'ancre Makefile.
