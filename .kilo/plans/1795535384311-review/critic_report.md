# critic_report.md — Stage 7 : Critic Review (FeatureRange 1..47)

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md`
**Périmètre :** FeatureRange 1..47 (Contexte, Décisions de design, Alternatives rejetées)
**Reviewers :** egm-design (Hostile Design Critic), egm-structural (Layout Validator), egm-style (Technical Copy Editor) — contextes frais
**Standards appliqués :** Feature Specification Standard (design, structural) ; aucun (style)

---

## A. egm-design — Hostile Design Critic

**Verdict global :** le diagnostic de départ est exact et l'option retenue est la bonne direction (déplacement du FLUSHDB vers un index dédié). En revanche la FeatureRange repose sur un invariant non appliqué (« jamais concurrents »), sur une protection de test qui n'est pas inconditionnelle, et ne comporte ni assertion automatisée de la garantie centrale, ni implication de sécurité explicite, ni garde-fou sur l'outil destructif. **6 Major, 4 Nit — spec non conforme en l'état** (Alternatives Rule, Risks & Limits, tests).

### F-01 — **Major** — La mutualité « jamais concurrents » n'est appliquée nulle part (l. 26, l. 44)
Chemins vérifiés : `Makefile:67` et `ci.yml:100` sérialisent bien avec `-p 1` ✔ ; mais `AGENTS.md` (Commandes) documente `go test ... ./...` **sans `-p 1`** — ce `go test ./...` exécute les paquets en parallèle, or les purges de `tests/` suppriment toutes les clés `user_sessions:*`/`admin:*` → interférence inter-paquets possible (préexistant, mais le plan le rationalise au lieu de le traiter). Second cas, plus probable : `make e2e` dans un terminal pendant `make test` dans l'autre → le `FLUSHDB` e2e sur db 2 **efface les sessions des suites Go en cours** (401 flaky). Le rejet de « db 3 pour e2e » repose sur une prémisse non garantie ; son coût réel est une constante + une ligne de doc.
**Recommandation :** soit restaurer l'index distinct (e2e = 3), soit assumer l'invariant comme règle dure et l'appliquer : (1) documenter dans AGENTS.md que `make e2e` ne doit jamais tourner en parallèle de `make test` et que toute suite complète exige `-p 1` (corriger la commande documentée qui ne l'a pas) ; (2) rendre le nettoyage du paquet `tests` non global (supprimer uniquement les clés créées par la suite) — seule façon de rendre les paquets parallèles sûrs ; (3) le cas échéant, faire purger l'e2e un index ne contenant que des clés e2e.

### F-02 — **Major** — La protection « à la source » n'est pas inconditionnelle (l. 26, l. 30)
`t.Setenv` restaure la valeur précédente à la fin du test. Dans le paquet `tests`, seul `TestIntDB` appelle `db.InitRedis` ; tous les tests suivants utilisent le client global laissé par `TestIntDB`. La variable `REDIS_DB` n'est donc **plus positionnée** pour la suite : tout appel `db.InitRedis()` supplémentaire (futur ajout, refactor) retomberait silencieusement sur db 1 — exactement le bug que ce plan corrige, sans erreur. `go test ./tests -run TestCleanUp` seul laisse le client `nil` → panic (préexistant). La validité repose en outre sur l'ordre de définition des fichiers (alphabétique) — invariant fragile non documenté. Dans `middleware`/`controllers`, le `t.Setenv` et le `InitRedis` sont dans la même fonction de setup ✔.
**Recommandation :** ne pas s'appuyer sur la portée d'un `t.Setenv` pour une propriété de processus. Cumulable : (1) helper unique `initTestRedis(t)` par paquet (`t.Setenv` + `db.InitRedis()` + `require` sur l'index) ; (2) ou `TestMain` par paquet (`os.Setenv("REDIS_DB", "2")` au niveau processus). Retirer la formulation « inconditionnel » de l. 30 et la remplacer par l'énoncé exact du mécanisme.

### F-03 — **Major** — La garantie centrale n'est prouvée par aucun test (l. 27, l. 28, l. 30)
Le plan livre un test du **parseur** (T2) mais aucun test du **câblage** — de la propriété annoncée comme objectif. Une régression banale (suppression du `t.Setenv`, préfixe Makefile retiré, `REDIS_DB` exporté par un `.env`) restaure le bug sans signal. La validation T9 est **manuelle** (`redis-cli`), non rejouable par la CI. AGENTS.md pose pourtant « toute nouvelle feature doit être livrée avec ses tests » — la feature étant de l'infrastructure de test, elle mérite un contrôle automatisé.
**Recommandation :** (1) dans chaque suite d'intégration, après `InitRedis` : `require.Equal(t, 2, db.GetRedis().Options().DB)` — l'index effectif est lisible via `Options()` du client go-redis, preuve permanente, une ligne par paquet ; (2) garde-fou CI sur le modèle « legacy drivers » (`ci.yml:88`) : `! git grep -n "InitRedis(1)" -- '*.go'` ; (3) pour l'e2e : assertion de bout en bout (la session du superadmin doit exister en db 2 et pas en db 1).

### F-04 — **Major** — Aucune implication de sécurité énoncée (l. 18-20, l. 27, l. 28, l. 34)
Faits à énoncer : l'isolation obtenue est une isolation de **namespace Redis** sur **une instance unique** avec les **mêmes identifiants** — radicalement plus faible que la posture PostgreSQL du repo (rôles dédiés, RLS FORCE, échec fermé) : un test, un bug ou `e2edb` dispose du même mot de passe sur tous les index, y compris db 1. `REDIS_DB` devient un aiguillage de sécurité pour le magasin de sessions : une valeur valide mais erronée (`REDIS_DB=0`) n'est **pas** refusée — l'« échec explicite » ne couvre que les valeurs mal formées, pas les valeurs mal choisies. Une instance Redis mutualisée (managé, `databases` réduit, autre service) rend l'index 1/2 partagé avec un tiers.
**Recommandation :** ajouter deux lignes dans la FeatureRange : (1) *Nature de l'isolation* — « logique (index Redis), pas un mécanisme de sécurité : même instance, mêmes identifiants, aucun ACL ; protège contre les purges accidentelles, pas contre un accès malveillant » ; (2) *Contraintes de déploiement* — « `REDIS_DB` ne doit pas pointer vers un index partagé avec une autre application ; en production la variable reste absente (défaut 1) ».

### F-05 — **Major** — L'outil destructif conserve par défaut le pouvoir de purger la db dev, sans garde-fou (l. 29, l. 45)
`cmd/e2edb` a été écrit avec une doctrine de garde-fous explicite (« The guard rails come first », `main.go:12`) côté PostgreSQL ; côté Redis cette doctrine est **absente** : le tool fera un `FLUSHDB` de l'index résolu, dont le défaut est **1** — la db des sessions dev. La commande d'usage documentée dans l'en-tête (`main.go:21` : `E2E_DB_NAME=artisan_e2e go run ./cmd/e2edb`) détruit les sessions dev. Le plan l'assume mais n'évalue **aucune alternative**.
Alternatives non évaluées : (a) **défaut de l'outil = 2** (une constante + un test : l'outil ne peut plus purger la db dev par accident) ; (b) **refus explicite de purger l'index 1** sauf opt-in, sur le modèle de `validateTarget` (une garde + un test unitaire : l'erreur devient explicite au lieu d'un comportement documenté) ; (c) **`E2E_REDIS_DB` dédié** à l'outil (une variable de plus : supprime l'ambiguïté « même env »).
**Recommandation :** ajouter ces trois lignes au tableau des rejets (ou en retenir une) et mettre à jour la ligne `Usage:` de l'en-tête (`E2E_DB_NAME=artisan_e2e REDIS_DB=2 go run ./cmd/e2edb`) — aujourd'hui le seul mode d'emploi découvrable devient sinon faux.

### F-06 — **Major** — Alternative viable non évaluée : le parseur dans un package feuille (l. 29, l. 46)
Le tableau des rejets évalue « dupliquer » vs « un seul parseur via `db` » mais ignore la troisième option qui cumule les avantages : extraire `RedisDBIndex` dans un paquet feuille (`config/` — qui contient déjà `ValidateSigningSecrets`, le motif « validation d'env partagée » existe — ou `internal/redisconf`). Résultats : source unique ✔ ; l'outil d'infra **ne dépend plus du paquet runtime de l'app** (importer `db` lie le client global `RedisClient` et le paquet applicatif complet — inversion de dépendance à éviter, le repo ayant déjà la convention `internal/*`) ; pas de nouvelle dépendance ✔. Une justification de refus serait recevable (« le paquet `db` est déjà le point d'accès unique au client Redis ; un paquet de plus serait de la cérémonie pour 10 lignes ») mais elle est absente.
**Recommandation :** ajouter la ligne au tableau des rejets avec ce raisonnement explicite.

### F-07 — **Nit** — La borne `[0,15]` est une heuristique de conteneur, pas du serveur (l. 27)
Une instance managée configurée avec `databases 1` ou `4` accepte `REDIS_DB=2` comme « valide » puis échoue au premier usage par `ERR DB index is out of range`. Comme `InitRedis` ne fait pas de `Ping` (voulu : « comportement inchangé »), l'app **démarre** et ne casse qu'à la première opération — l'« échec explicite » couvre le parsing, pas la faisabilité.
**Recommandation :** corriger le message (« must be an integer between 0 and 15 *and below the server's `databases` setting* »), mentionner la contrainte dans AGENTS/README ; si le choix est de ne pas pinger, le dire explicitement.

### F-08 — **Nit** — La FeatureRange ne renvoie vers aucun mode de défaillance (l. 1-47)
Lus en isolation, les L1-47 ne comportent ni risques, ni non-objectifs, ni renvoi vers « Risques / cas d'échec » (l. 187) et « Hors scope » (l. 203).
**Recommandation :** une ligne de renvoi en fin de FeatureRange — le standard « Risks & Limits » serait satisfait dans le périmètre.

### F-09 — **Nit** — `REDIS_DB` documentée dans AGENTS.md mais absente du template `.env` (l. 34)
Un développeur qui customize son `.env` (ou un shell/CI qui exporte `REDIS_DB`) bascule le **dev** hors db 1 sans signal — d'autant que le Makefile fait `-include .env` + `export` : une valeur posée dans `.env` s'applique à `make run`. `make test` (t.Setenv) et `make e2e` (préfixe inline) la surchargent ✔, mais `make run` la suit.
**Recommandation :** soit une ligne commentée `# REDIS_DB=1  # index Redis : 1 = dev (défaut), 2 = tests/e2e` dans `.env_rename_me`, soit assumer explicitement dans AGENTS.md que la variable est volontairement absente du template.

### F-10 — **Nit** — Rupture d'API publique de `db.InitRedis` non mentionnée (l. 28 vs l. 18-20)
`InitRedis(selectDB ...int)` → `InitRedis() error` est une **rupture de signature** d'un paquet exporté, pas un simple changement de comportement. Les 5 appelants internes sont couverts ✔ (« plus aucun consommateur » exact dans le dépôt), mais un consommateur externe du module devra s'adapter.
**Recommandation :** une phrase dans l'Objectif : « l'API `db.InitRedis` change de signature — rupture de compatibilité pour les importateurs externes du module ».

### Vérifications factuelles (affirmations du plan confrontées au code)
| Affirmation | Verdict |
|---|---|
| `cmd/e2edb` FLUSHDB db 1, const `main.go:54` | Confirmé (`:54`, `:155`, `:160`) |
| `make test` détruit les sessions dev (`user_sessions:*`/`admin:*`) | Confirmé (`article_test.go:915-927`, `login_test.go:269-279` ; formes des clés `models/auth.go:260, 359-360`) |
| middleware/controllers ne suppriment que leurs propres clés | Confirmé (`middleware/auth_test.go:43,83,109,127,144` ; `controllers/auth_test.go:105`) |
| 5 appelants `InitRedis`, tous listés | Confirmé (grep exhaustif) |
| « purge ciblée ne protégerait rien » | Confirmé (aucun préfixe distinctif dev/test dans `models/auth.go`) |
| godotenv n'écrase pas | Confirmé (`Load` ne surcharge pas ; seul `Overload` le fait) |
| CI inchangée tenable | Confirmé (`ci.yml:100` `-p 1`, service redis:8-alpine → databases 16 ; `ci.yml:53` pas de REDIS_DB → smoke binaire en db 1) |
| `-create-admin` n'appelle pas InitRedis | Confirmé (`main.go:112-117` avant `:144`) |
| `make e2e` n'exporte pas REDIS_DB aujourd'hui | Confirmé (`Makefile:154-165` — le préfixe inline proposé est nécessaire et suffisant) |
| webServer.env explicite = bonne décision | Confirmé (`playwright.config.ts:57` ; le binaire relit `.env` via godotenv) |
| Aucune autre source Redis que `models/auth.go` | Confirmé (grep exhaustif — le périmètre « seul l'index varie » est exhaustif) |

### Points conformes
Diagnostic exact et prouvé ; mécanisme (index dédié) = solution la plus simple ; refus du repli silencieux cohérent avec le repo ; non-régression fonctionnelle exacte (script Lua `consumeRefreshAuth` et `SAdd` inchangés) ; emplacements de doc correctement identifiés ; « pas de generate_docs » correct.

### Priorités
1. F-02 + F-03 : protection effective et **prouvée** (helper/TestMain + `Options().DB == 2` + garde CI anti-`InitRedis(1)`).
2. F-01 : trancher index unique vs distinct ; si unique, appliquer `-p 1` partout et documenter l'interdiction de concurrence.
3. F-05 + F-06 : deux lignes au tableau des alternatives ; au minimum, mettre à jour la ligne `Usage:` d'e2edb.
4. F-04 : deux lignes « nature de l'isolation » / « contraintes de déploiement ».
5. F-08 : une ligne de renvoi.

---

## B. egm-structural — Layout Validator

**Constats : 3 Major, 6 Nit. Aucun Blocker** : hiérarchie propre (`H1 → H2 ×3`, pas de saut), deux tableaux bien formés (colonnes constantes, séparateurs conformes), liste à puces cohérente, aucune ancre cassée — **toutes les ancres `fichier:ligne` de la portée vérifiées exactes**, énumération des appelants `InitRedis` complète (4/4).

### F1 — **Major** — Aucune zone « Objectifs / hors-objectifs » dans la portée ; l'objectif est un paragraphe sans titre (l. 3-20, l. 22-35)
L'objectif n'existe que comme paragraphe en gras « **Objectif** : » (l. 18) — non repérable dans une table des matières. Les non-objectifs n'ont aucune place dans la portée : quatre lignes « Inchangée / Aucune… » noyées dans la table des décisions (l. 32-35), et la section formelle « Hors scope » (l. 203) hors portée, à 185 lignes de distance. **Remédiation :** insérer un `## Objectifs et hors-objectifs` entre Contexte et Décisions ; y remonter les non-objectifs distributionnels (make test, CI, `.env`, Swagger) et la clause « aucun changement fonctionnel ».

### F2 — **Major** — Aucun bloc « Risques / mitigations » dans la portée et aucun renvoi vers la section existante (l. 22-46 ; section à l. 187)
La section existe mais hors portée, et **aucune section de la portée n'y renvoie**, alors que la portée porte déjà des affirmations à charge de risque : l. 26 (« les commandes sont séquentielles » — hypothèse non justifiée, jamais déplacée vers les Risques), l. 28 (risque cité comme justification, jamais relié à sa mitigation), l. 30 (garantie technique posée sans limites). Écart d'ordre canonique en sus : le standard place Risks après Alternatives ; le document suit `Alternatives (l. 37) → Tâches (l. 48-185) → Risques (l. 187)`, enterrant les risques sous 140 lignes de détail. **Remédiation :** remonter `## Risques / cas d'échec` avant `## Tâches`, ou à minima un renvoi explicite dans la portée.

### F3 — **Major** — La table « Décisions de design » mélange trois classes de décisions (l. 24-35)
Sans catégorie ni ordre, les 10 lignes relèvent de trois natures : architecture/runtime (l. 26-29), câblage par consommateur (l. 30-31), **constats d'inaction** (l. 32-35 : Makefile test, CI, `.env`, Swagger « inchangés »). La troisième famille n'est pas une décision : c'est le contenu de la zone « Non-Goals », déposé dans la table pour qu'il passe inaperçu — la zone « Proposed Architecture » n'a pas de lecture à plat. **Remédiation :** scinder en `## Décisions de design (fixées)` (l. 26-31) et `## Ce qui ne change pas` (l. 32-35), cohérent avec le bloc Objectifs/hors-objectifs.

### F4 — **Nit** — Référence interne non résoluble : « validation fenêtre 5 » (l. 5)
Occurrence unique, aucun artefact identifiant. La provenance des « faits établis » — justification de tout le bloc Problème — est inatteignable. Nommer la source ou supprimer la parenthèse.

### F5 — **Nit** — La table des alternatives ne pose aucun critère d'évaluation, ni option retenue en ligne de base (l. 37-46)
Six rejets argumentés cas par cas, sans grille commune (d'où des justifications plus difficiles à rapprocher), et l'option retenue (db 1 = dev / db 2 = tests + e2e) ne figure pas comme ligne de comparaison dans la table. **Remédiation :** un chapeau « Critères d'évaluation : … » + une ligne « RETENU : … » en tête de table.

### F6 — **Nit** — Doublon de l'argument central entre le Contexte (l. 16) et la table des alternatives (l. 41)
« mêmes clés, même db » énoncé deux fois sans renvoi. Garder l'argument dans la table des alternatives, ne laisser au Contexte que la formulation du problème (avec renvoi).

### F7 — **Nit** — Incohérence terminologique : « index », « db » et « base » désignent le même objet (l. 7-16, l. 18, l. 26-27, l. 45)
Retenir « index » partout ; réserver « db N » aux commandes (`redis-cli -n N`). La confusion se propage dans T8 vers trois fichiers de documentation.

### F8 — **Nit** — Une contrainte d'environnement posée comme un fait, sans renvoi vers la zone des limites (l. 27)
La borne 0-15 est vraie pour le compose de dev, pas pour un Redis managé. Formuler « supposé par le stack dev (docker-compose) ; à vérifier sur un Redis managé » et pointer vers les limites.

### F9 — **Nit** — Le Problème reste au niveau du mécanisme : ni symptôme utilisateur, ni critère de succès dans la portée (l. 7-20)
Le Contexte n'énonce jamais ce que le développeur constate (session invalidée, re-login requis) ; le critère de succès (le « marqueur » Redis de T9) n'est pas annoncé dans la zone Objectif. **Remédiation :** une phrase de bénéfice et de critère de succès dans le bloc Objectif, avec renvoi vers T9.

### Vérifications de conformité (sans constat)
Hiérarchie `H1 → H2` propre ; `### T1..T9` correctement imbriqués sous `## Tâches` ; tableaux syntaxiquement bien formés ; liste à un seul niveau, continuations indentées correctement ; blocs de code fermés ; backticks appariés ; toutes les ancres de la portée exactes (`main.go:54` ✓, `article_test.go:914-928` ✓, `login_test.go:267-279` ✓, `main.go:144` ✓, `main.go:137-142` ✓, `middleware/auth_test.go:30` ✓, `controllers/auth_test.go:34` ✓, `article_test.go:159` ✓, `config/secrets.go:16` ✓) ; énumération des appelants InitRedis **complète et vérifiée** (4/4) ; prémisses des décisions vérifiables (playwright `env` explicite réalisable ✓, préfixe inline l'emporte sur l'env ✓, aucune ligne REDIS_DB dans le dépôt ✓, formulations AGENTS.md à corriger confirmées ✓).

---

## C. egm-style — Technical Copy Editor

**Constats : 3 Major, 14 Nit.** Actions dominantes : CALIBRATE (garanties plus larges que le mécanisme, hypothèses posées comme faits).

### Major
1. **l. 30** — « la protection tient même pour `go test … ./tests` direct » décrit un état permanent ; `t.Setenv` expire à la fin du test ; dans `tests`, la protection ne vit que pendant `TestIntDB`. La contrainte « avant `godotenv.Load` » est présentée comme un invariant alors qu'elle est **sans effet** (godotenv `Load` n'écrase jamais). **CALIBRATE** : « `REDIS_DB=2` doit être présent dans l'environnement du processus au moment où le client Redis est créé ».
2. **l. 26, l. 44** — « les commandes sont séquentielles » / « (jamais concurrents) » : hypothèses sur le comportement de l'opérateur présentées comme faits ; conséquence non couverte (`make e2e` pendant `make test` → FLUSHDB db 2 → 401 flaky). **CALIBRATE** : énoncer la règle une fois, comme invariant du dépôt ; supprimer la parenthèse de l. 44.
3. **l. 18-20** — « Aucun comportement fonctionnel ne change » est faux pour l'observateur : ce que le plan change est précisément le comportement visible de `make test`/`make e2e` vis-à-vis des sessions dev. **CALIBRATE** : « le comportement des sessions (rotation atomique, préfixes, TTL) est inchangé ; seul l'index varie. Effet visible attendu : une session dev survit à `make test` et à `make e2e` ».

### Nit
4. **l. 27** — borne 0-15 : justifiée par le conteneur de dev seulement. **CALIBRATE** (« borne alignée sur la configuration par défaut du conteneur de dev »).
5. **l. 16 vs l. 41** — « mêmes clés, même db » deux fois. **DROP** l. 16 (ou renvoi), **KEEP** l. 41.
6. **l. 28 vs l. 45** — même argument dérive silencieuse, deux fois. **KEEP** l. 45 (où il justifie le rejet) ; la l. 28 doit énoncer la décision, pas la réargumenter.
7. **l. 29** — « l'app qu'il sert » factuellement faux : `cmd/e2edb` ne sert rien, il prépare l'environnement ; l'app du cycle est servie par le webServer Playwright. **CALIBRATE** : « l'outil qui purge et l'app du cycle e2e lisent le même index issu du même env ».
8. **l. 14** — le titre de puce (« L'app écrit en db 1 partout ») mélange app et tests (3 des 4 références sont des fichiers de test). **CALIBRATE** : « Tous les clients Redis sont créés en db 1 : l'application (`main.go:144`) et trois suites de test ».
9. **l. 31** — « `.env` ne peut pas l'écraser » : mécanisme inexact — `cmd/e2edb` ne charge jamais `.env` ; c'est le préfixe inline qui écrase la valeur héritée. **CALIBRATE** : « l'outil ne charge pas `.env` ; le préfixe inline écrase toute valeur héritée ».
10. **l. 34** — « la variable n'existe que posée par les tests/e2e » contredit T8 (variable documentée applicative) ; « EST » = capitalisation d'emphase. **CALIBRATE** : « aucune ligne REDIS_DB dans `.env`/`.env_rename_me` : le défaut 1 est la configuration dev ; `REDIS_DB` n'est posée que par les tests et le cycle e2e ».
11. **l. 42** — « Surdimensionné » = jugement ; la justification factuelle qui suit suffit. **CALIBRATE** : « un index Redis suffit : pas de port, de volume ni de cycle de vie supplémentaire ».
12. **l. 45** — « l'erreur loud est le pattern du repo » élargit la preuve (une fonction citée). **CALIBRATE** : « erreur explicite avant démarrage (même forme que `config.ValidateSigningSecrets`) ».
13. **l. 46** — « exactement la classe de bug corrigée » = emphase. **CALIBRATE** : « deux parseurs à resynchroniser = la dérive que ce plan supprime ».
14. **l. 5, 30, 34** — « validation fenêtre 5 » : **DROP** la parenthèse (les références fichier:ligne sont auto-vérifiables) ; « une var déjà posée » → « variable » ; « EST » → gras.
15. **l. 7, 18, 26, 27, 34** — « db » / « base » / « index » alternent ; « base Redis 1 » ambigu (une « base » Redis EST un index numéroté). **CALIBRATE** : « index Redis N » comme terme précis, « db N » en forme courte, « base » supprimé pour un index.
16. **l. 28, 45** — « Échec loud » : anglicisme non défini ; le vocabulaire du repo est « échec fermé » / « échec explicite ». **CALIBRATE** : « Échec explicite (erreur remontée) ».
17. **l. 33** — « ses tests hériteront de `t.Setenv` » : le verbe est faux (chaque suite pose la variable ; pas d'héritage). **CALIBRATE** : « ses tests poseront eux-mêmes `REDIS_DB=2` (db 2 du Redis éphémère du job) ; le smoke du binaire reste en db 1 par défaut ».

### Éléments vérifiés — KEEP (aucun constat)
Références l. 7-8 (`main.go:54`, `flushRedis` `:147-163`) exactes ; l. 9-13 corroborées par le code ET par l'observation live du 24 nov. 2026 ; l. 14-15 exactes ; l. 28 « plus aucun consommateur » exact (recherche exhaustive) ; l. 35 Swagger justifié ; l. 33 CI inchangée confirmée (`ci.yml` sans REDIS_DB, Redis en service container).

### Annexe — références hors périmètre (l. 48-207), signalées au passage
- **l. 135** : « lignes ~121-122 » — l'appel e2edb est en réalité à `Makefile:155` ; la référence courante induit en erreur sur un fichier que la tâche va précisément modifier.
- **l. 129** : `TestRefreshOnlyConsumesMatchingOwner` (`controllers/auth_test.go:27-38`) — la fonction s'étend de l. 27 à ~144 ; seules les l. 31-34 sont concernées par l'insertion.
