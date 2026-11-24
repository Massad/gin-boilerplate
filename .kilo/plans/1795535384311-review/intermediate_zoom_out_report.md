# intermediate_zoom_out_report.md — Stage 8 : Intermediate Zoom Out

**Document cible :** `.kilo/plans/1795535384311-redis-db-isolation.md` (207 lignes, relu intégralement)
**Reviewer :** egm-zoom-out — Skeptical Architect (contexte frais), avec les rapports S5/S6/S7 en entrée
**Fait vérifié par l'orchestrateur via Context7 (docs officielles, 24 nov. 2026) :**
- go-redis : `Client.Options() *Options` retourne les options passées à `NewClient` (dont le champ `DB`) — l'assertion `db.GetRedis().Options().DB` est une API valide ;
- godotenv : `Load` ne surcharge jamais une variable d'env déjà posée (seul `Overload` le fait) — la contrainte d'ordre `t.Setenv`/`godotenv.Load` est un non-invariant ;
- Redis : les db logiques sont du *namespacing* (doc officielle : « Use Redis databases to separate keys within the same application. Don't use them to run multiple unrelated applications in a single Redis instance ») ; `SELECT` hors plage → `ERR DB index is out of range`.

---

## 0. Constat préalable — le document n'a pas bougé

Le fichier relu est identique à celui des étapes 5-7. Aucune correction issue des rapports précédents n'a été réintégrée : la question « les problèmes signalés sont-ils adressés par le document ? » a une réponse uniforme : non, sauf là où le plan était déjà complet.

Ce que le plan résout **déjà bien** (à ne pas remanier) : refus du repli silencieux, refus de l'export `REDIS_DB` par le Makefile pour `make test`, rejet justifié de la purge ciblée, suppression du variadic `selectDB ...int`, non-touchement de `.env`/CI/Swagger, refus d'un conteneur Redis séparé. Le diagnostic d'ouverture est exact et prouvé (4 sites `InitRedis(1)`, `redisDBIndex = 1`, purges globales `user_sessions:*`/`admin:*`).

## 1. Tableau de mitigation des constats antérieurs

| Constat antérieur | Origine | État | Nature |
|---|---|---|---|
| `t.Setenv` expire / protection « inconditionnelle » fausse | S5-F1, S6-#4, S7-style-1, S7-F-02 | **Non mitigé** | calibrage + preuve |
| db 2 partagée tests/e2e, prémisse « jamais concurrents » | S5-F2, S6-#3, S7-F-01 | **Non mitigé** | architecture |
| Pendant PostgreSQL non nommé (`DELETE FROM public.tenant` sur la base dev) | S5-F3 | **Non mitigé** (absent du « Hors scope ») | architecture |
| Borne [0,15] codée en dur | S5-F4, S7-F-07, S6-#8 | **Non mitigé** | nit |
| Parseur dans `db` vs `config` | S5-F5, S7-F-06 | **Non mitigé — le refus est confirmé** (voir R2) | sur-ingénierie évitée |
| Ancres fausses (Makefile, activeContext) | S5-F6, S7-style-annexe | **Non mitigé** | mécanique |
| Garantie centrale non prouvée par un test | S7-F-03 | **Non mitigé** | preuve |
| Nature de l'isolation non énoncée | S7-F-04 | **Non mitigé** | calibrage |
| Outil destructif par défaut | S5-catalyst-1, S7-F-05 | **Non mitigé** | architecture |
| Valeur dupliquée 5 sites / invariant « même env » faux entre processus | S5-catalyst-3 | **Non mitigé — le remède proposé est refusé** (voir R3) | sur-ingénierie évitée |
| `redis-cli` non reproductible en T9 | S6-#2 | **Non mitigé** | validation |

## 2. Constats non mitigés — architecturaux

### G1 — **Major** — L'objectif n'est pas appliqué là où habite le pouvoir destructeur
**Portée :** l. 26-29, l. 45, T4 (l. 102-113), Risques l. 193-196 ; `cmd/e2edb/main.go:12, 21, 52-54, 147-164`.
Après ce changement, **la valeur par défaut du parseur (1) est précisément l'index que le plan veut réserver au dev**. L'outil `cmd/e2edb` a un contrat explicite — « The guard rails come first » (`main.go:12`), `validateTarget` refuse déjà 4 classes de cibles PostgreSQL — et ne possède **aucun** garde-fou côté Redis. Sa commande d'usage documentée (`main.go:21`) vide l'index 1, donc les sessions dev. Le plan l'assume (l. 196) sans évaluer d'alternative.
**Remède proportionné** (pas une nouvelle variable) : une garde dans `flushRedis`, sur le modèle exact de `validateTarget` :
```go
if index == db.DefaultRedisDB && strings.TrimSpace(getenv("REDIS_DB")) == "" {
    return errors.New("REDIS_DB is unset: refusing to flush the dev Redis database (index 1)")
}
```
≈6 lignes, zéro variable nouvelle, un const exporté (`db.DefaultRedisDB = 1`) réutilisé par le parseur et la garde. Effet : le risque l. 193-196 **disparaît** au lieu d'être documenté, et l'objectif « db 1 = dev exclusivement » devient une garantie du code. Le plan ne peut pas à la fois affirmer « db 1 = dev exclusivement » et laisser l'outil vider la db 1 par défaut.

### G2 — **Major** — Le plan s'applique à lui-même son propre critère de rejet
**Portée :** l. 26, l. 41, l. 44, l. 196 ; `tests/article_test.go:914-928`, `tests/login_test.go:267-279`, `models/auth.go:119-127, 260, 360`.
L'argument qui rejette la purge ciblée — « **mêmes clés, même db** » (l. 41) — s'applique **verbatim** à la paire que le plan crée : suites Go et cycle e2e partagent l'index 2 et les mêmes formes de clés. Deux acteurs destructeurs coexistent sur un index : le `FLUSHDB` e2e supprime les sessions des suites Go en cours ; le nettoyage global de `tests/` supprime la session du superadmin e2e → 401 au login Playwright. Le plan fonde la sécurité sur « les commandes sont séquentielles », prémisse écrite nulle part — alors qu'elle a déjà failli (l'app dev et `make test` étaient eux aussi « jamais concurrents »). Coût de l'alternative : **un seul entier** (`REDIS_DB=3` sur la chaîne e2e : Makefile + playwright.config.ts + 2 lignes de doc).
**Recommandation :** index distinct pour la chaîne e2e (le repo est en doctrine « échec fermé / garde-fous d'abord »). À défaut, **l'invariant doit être écrit comme règle du repo** dans `AGENTS.md` et l'en-tête d'e2edb — en l'état il n'existe nulle part.

### G3 — **Major** — La classe de bug « les tests utilisent l'environnement dev » reste entière côté PostgreSQL, et le plan ne le dit pas
**Portée :** Contexte l. 7-20, Objectif l. 18-20, « Hors scope » l. 203-207 ; `tests/article_test.go:153` (`godotenv.Load("../.env")`), `:909` (`DELETE FROM public.tenant`).
Les suites Go chargent `.env` → base **dev** `golang_gin_db`, et `TestCleanUp` y exécute `DELETE FROM public.tenant` (cascade sur `user`/`article`). `make test` supprime donc les tenants du développeur — impact **plus grave** que la purge Redis (données, pas seulement re-login). Le plan, dont l'objectif est formulé en termes de sessions, ne mentionne ce pendant ni dans le Contexte, ni dans « Hors scope », ni dans les Risques.
**Remède : une ligne** dans « Hors scope » nommant le pendant PostgreSQL comme follow-up (base jetable `artisan_test`, symétrique d'`artisan_e2e`). C'est la seule façon d'éviter que le prochain lecteur ne croie le problème réglé.

## 3. Constats non mitigés — preuve, calibrage, garanties

### G4 — **Major** — La garantie centrale n'est prouvée par aucun test ; le seul test livré teste le parseur, pas le câblage
**Portée :** l. 30, T2 (l. 88-92), T9 (l. 169-185).
T2 teste `RedisDBIndex` — la fonction pure. Rien ne vérifie que **l'index effectif du client** est 2 dans chaque suite. La validation T9 est manuelle (`redis-cli`), non rejouée par la CI. Une régression banale (suppression d'un `t.Setenv`, préfixe Makefile retiré, `REDIS_DB` glissé dans un `.env`) restaure le bug sans signal — alors qu'`AGENTS.md` impose « toute nouvelle feature livrée avec ses tests ».
**Deux remèdes proportionnés** (calqués sur les garde-fous CI existants `ci.yml:88` legacy drivers, `ci.yml:94` isolation helpers) :
1. assertion permanente par suite : `require.Equal(t, 2, db.GetRedis().Options().DB)` juste après chaque `InitRedis` — API vérifiée contre les docs go-redis (`Options()` retourne les options passées à `NewClient`, champ `DB`) ;
2. une étape CI : `! git grep -n "InitRedis(1)" -- '*.go'`.

### G5 — **Major** — « Inconditionnel » est faux, et la contrainte d'ordre documentée est un non-invariant
**Portée :** l. 30, l. 43, T5 (l. 115-122), Risques l. 188-192.
`t.Setenv` **restaure** la valeur à la fin du test ; dans le paquet `tests`, `REDIS_DB=2` n'existe que pendant `TestIntDB`. Un `db.InitRedis()` appelé plus tard résout 1 → dérive silencieuse vers la db dev, exactement ce que l. 28 déclare impossible. Symétriquement, la contrainte « `t.Setenv` **avant** `godotenv.Load` » est **sans effet** — confirmé par les docs godotenv : `Load` n'écrase jamais une variable déjà posée, l'ordre est indifférent. Un risque documenté comme réel qui ne l'est pas coûte plus cher qu'il ne protège ; une garantie qualifiée d'« inconditionnelle » qui ne l'est pas induit en erreur.
**Choix (un seul) :** soit un `TestMain` par paquet taggé `all` posant `REDIS_DB=2` au niveau processus (5 lignes par paquet), soit l'assertion d'index de G4 posée **à chaque création de client** + une ligne de convention AGENTS.md. Le `TestMain` devient inutile si l'index effectif est asserté là où le client est créé.

### G9 — **Nit** — La nature de l'isolation n'est pas énoncée : `REDIS_DB` n'est pas un mécanisme de sécurité
**Portée :** l. 18-20, l. 27-28.
L'isolation est un **index logique sur une instance unique, mêmes identifiants, sans ACL** — la doc officielle Redis le dit elle-même : les db logiques sont du *namespacing*, « Don't use them to run multiple unrelated applications in a single Redis instance ». `REDIS_DB=0` est une valeur **valide** non refusée : l'échec explicite ne couvre que les valeurs *mal formées*, pas les valeurs *mal choisies*. **Deux lignes** (nature de l'isolation + contrainte de déploiement), le plan en zéro aujourd'hui.

## 4. Constats non mitigés — mécaniques / éditoriaux

### G6 — **Major (validation)** — Le critère d'acceptation central n'est pas reproductible tel qu'écrit
**Portée :** T9 étapes 3-6 (l. 173-183).
L'étape 3 exécute `redis-cli` dans le conteneur (`docker compose exec`), les étapes 4-6 exécutent `redis-cli` **sans hôte ni port** : le Redis de dev est publié sur `127.0.0.1:16379`, pas 6379 → commande sur le mauvais port, réponse vide, et le lecteur conclut **à tort que l'isolation a échoué**. L'étape 5 (`POST /v1/login`) ne donne ni URL, ni corps, ni commande exécutable. **Remède :** contexte d'exécution + port par étape, ou une forme unique (`redis-cli -h 127.0.0.1 -p 16379 -n N …`).

### G7 — **Major (mécanique)** — Ancres fausses sur des fichiers que le plan va modifier
**Portée :** T7 (l. 135), T8 (l. 162-164).
L'appel `cmd/e2edb` est en **`Makefile:155`** (recette `e2e` dès l. 152) ; « ~121-122 » pointe sur `migrate_down`. Même famille : T8 vise `activeContext.md` **ligne 60**, mais l'affirmation qui devient **fausse** après le changement est en **ligne 29** (« le `FLUSHDB` purge la base Redis 1 ») ; la ligne 60 ne parle que de joignabilité WSL. **Remède :** remplacer les numéros de ligne par des **noms de recette/section**.

### G8 — **Nit** — La borne [0,15] fige une hypothèse de serveur ; l'échec explicite ne couvre que la syntaxe
**Portée :** l. 27, T1, T2.
`databases 16` est le défaut du conteneur dev, pas une propriété du client. Sur un Redis managé à `databases 4`, `REDIS_DB=2` est accepté puis échoue au premier usage par `ERR DB index is out of range` (confirmé, docs Redis) ; comme `InitRedis` ne fait pas de ping (choix assumé), l'app démarre et casse à la première opération. **Proportionné :** clause dans le message (« 0-15, plage du stack dev ») + préfixe `error: ` du repo (`config/secrets.go:20`). **Disproportionné :** supprimer la borne — l'échec serveur est plus tardif et moins actionnable.

### G10 — **Nit** — Rupture d'API non mentionnée ; vocabulaire « base »/« db »/« index »
**Portée :** l. 28 vs l. 18-20 ; l. 7-16, l. 26-27, T8.
`InitRedis(selectDB ...int)` → `InitRedis() error` est une **rupture de signature** d'un paquet exporté d'un module publié : un importateur externe doit s'adapter. Une phrase suffit. Vocabulaire : « index Redis N » comme terme juste (« db N » seulement dans les commandes) ; T8 propage l'ambiguïté dans trois fichiers de doc.

### G11 — **Nit** — Terme d'équipe non défini ; symptôme utilisateur et critère de succès jamais énoncés
**Portée :** l. 5, l. 7-20.
« validation fenêtre 5 » renvoie à un artefact introuvable ; les références `fichier:ligne` sont auto-vérifiables, la parenthèse n'apporte rien (supprimer). Le Contexte n'énonce jamais ce que le développeur **voit** (session invalidée → re-login requis) ni le critère de succès (le « marqueur » de T9) — c'est pourtant tout le bénéfice de l'objectif, et le plan le démontre en T9 sans l'annoncer.

## 5. Remédiations disproportionnées — à ne PAS appliquer

Le lens YAGNI coupe dans les deux sens : plusieurs remèdes proposés aux étapes 5-7 introduiraient plus de complexité que le problème n'en justifie.

- **R1 — Rejeté : `E2E_REDIS_DB` + parseur à deux namespaces + constantes `DefaultRedisDB`/`DisposableRedisDB` complètes (S5-catalyst-1).** Une variable de plus et deux noms à synchroniser — exactement la charge que le plan refuse pour « deux index ». Le défaut 1 reste le défaut. Le remède proportionné est **G1** (garde dans l'outil, zéro variable).
- **R2 — Rejeté : déporter le parseur dans `config/` (S5-F5, S7-F-06).** L'argument « importer `db` lie le paquet applicatif complet » est **factuellement inexact** : `db/db.go` fait 32 lignes et n'importe que `os` + go-redis, **déjà importé directement** par `cmd/e2edb` (même module, même version). Coût de transitivité : **zéro**. `config/` ne contient qu'un fichier (validation des secrets) ; y loger la résolution d'un index Redis mélange deux responsabilités sans gain. **Le plan a raison de garder le parseur dans `db`.** Seul ajout utile : une ligne dans le tableau des alternatives (« parseur dans `config` — rejeté : `db` est le point d'accès unique au client Redis, import déjà présent »).
- **R3 — Rejeté : le « détecteur de dérive » qui relit `Makefile` et `app/playwright.config.ts` depuis un test Go (S5-catalyst-3).** Il couple la suite Go à l'arborescence (faux échecs au renommage) et ne remplace pas une relecture : vérifier qu'un entier est identique dans trois fichiers n'est pas de l'ingénierie. La dérive se voit en une ligne de `redis-cli` dans T9.
- **R4 — Rejeté (hors périmètre) : re-filtrer le nettoyage Redis des suites Go sur leurs seules clés (S5-catalyst-2b).** C'est le vrai correctif de la course inter-paquets (préexistante) et il supprime le `KEYS` O(N) — mais il sort du périmètre de ce plan et ne couvre pas la moitié e2e ↔ suites Go. Follow-up, pas ici.
- **R5 — À proscrire :** un 4e index « au cas où » ; un nouveau paquet `internal/redisconf` ; un `PING` Redis au boot (changerait le contrat opérationnel pour un gain nul).

## 6. Remédiations proportionnées — ordre coût/bénéfice

1. **G1** — garde « ne pas vider l'index 1 » dans `flushRedis` (≈6 lignes + 3 cas dans la table existante de `cmd/e2edb/main_test.go`) ; mettre à jour la ligne `Usage:` de l'en-tête.
2. **G4** — `require.Equal(t, 2, db.GetRedis().Options().DB)` par suite + garde CI `! git grep -n "InitRedis(1)"`.
3. **G2** — trancher : `REDIS_DB=3` pour la chaîne e2e **ou** invariant écrit « ne jamais lancer `make test` et `make e2e` simultanément ».
4. **G3** — une ligne dans « Hors scope » nommant le pendant PostgreSQL.
5. **G5** — réécrire l. 30/43/188-192 sans « inconditionnel », supprimer la contrainte d'ordre `t.Setenv`/`godotenv` (docs godotenv) ; choisir `TestMain` **ou** l'assertion d'index à chaque `InitRedis`.
6. **G9, G10, G11** — deux lignes sur la nature de l'isolation (citer la doc Redis : *namespacing*) ; une phrase sur la rupture d'API ; supprimer « validation fenêtre 5 » ; énoncer le symptôme et le critère de succès.
7. **G6, G7, G8** — reprendre les commandes de validation (contexte + port), recaler les ancres par nom de recette/fichier, borner `[0,15]` avec sa justification.

## 7. Verdict d'altitude

**L'architecture est saine.** La racine est traitée au bon endroit : séparation des index plutôt que purge ciblée, un seul parseur partagé, variadic supprimé, échec explicite plutôt que repli silencieux, non-touchage de `.env`/CI/Swagger. Le plan fait ~85 % de son travail ; les 15 % restants sont **quatre décisions et une dizaine de phrases**, pas une refonte. **Aucun Blocker.**

- **Findings architecturaux (à trancher avant exécution) :** G1 (l'objectif n'est appliqué ni au seul point où le pouvoir destructeur existe, ni par le code), G2 (le plan applique verbatim son propre critère de rejet à la paire tests/e2e qu'il crée), G3 (le pendant PostgreSQL n'est nommé nulle part — le silence fait croire à une protection plus large que celle livrée). G1 et G2 se traitent avec **une ligne de garde** et **un entier** — ce ne sont pas des chantiers.
- **Findings de calibrage et de preuve (coût quasi nul) :** G4 (l'index effectif n'est asserté nulle part), G5 (« inconditionnel » faux, ordre `t.Setenv`/`godotenv` non-invariant), G9 (nature de l'isolation).
- **Findings mécaniques/éditoriaux :** G6, G7, G8, G10, G11 — une passe, aucun n'affecte l'architecture.
- **Point de vigilance :** G2 et G3 sont les deux endroits où le plan promet une garantie plus large que ce qu'il livre. Ces deux lignes à écrire, plus la garde de G1, suffisent à le rendre exact.
- **Ce qu'il ne faut surtout pas faire :** absorber les remèdes R1-R5 — plus de complexité que le problème, pour un bénéfice nul ou négatif. Le plan a raison d'être minimal ; il doit le rester en corrigeant, pas en s'alourdissant.
