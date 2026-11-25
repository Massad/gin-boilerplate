# Stage 5 — Initial Zoom Out Review

Document audité : `.kilo/plans/1795597264251-architecture-modulaire-par-domaine.md` (fichier complet).
Contexte Goldfish : aucun. Périmètre : intégralité. Découpage sémantique acté par l'utilisateur : FeatureRange 1..174, DetailedRange 175..End.

---

## Reviewer 1 — egm-zoom-out (Skeptical Architect)

# Rapport d'audit Zoom-Out — « Migration incrémentale vers une architecture modulaire par domaine »

## Vérifications factuelles effectuées (recherche code)

La cartographie du document est exacte et vérifiée dans le dépôt : `SetupRouter` à `tests/article_test.go:34`, helper Refresh à `controllers/auth.go:84+`, accesseurs `getUserID/getTenantID/getRole` à `controllers/user.go:19-33`, casse mixte `"Message"`/`"message"` dans `controllers/article.go`, `RequirePermission` non câblé (usages limités à `middleware/auth.go` et son test), dérive smoke CI réelle (`.github/workflows/ci.yml:139,157,166-168` sonde `/v1/invoice` supprimé de `main.go:148-182`), backlog à `memory-bank/progress.md:56`. Aucune réinvention de roue détectée : il n'existe ni helper de réponse JSON central, ni mapping pgx→HTTP, ni parseur de param UUID ailleurs dans le code — `httpx` est une création nouvelle, pas un duplicata.

## Constats

**1. [Major] YAGNI — `internal/platform/httpx` créé sans aucun appelant**
- Scope : P1.6 (§6, l.227-229), layout §4 (l.104), Phase 2 risque (l.132).
- Le paquet est construit en Phase 1 avec « aucun appelant prod dans cette phase », adoption différée aux Phases 2-7. C'est du code spéculatif maintenu mort pendant toute la migration, et la Phase 2 exige que les handlers restent identiques au byte près — donc `httpx` ne sert aucun filet de sécurité réel tant qu'il n'est pas adopté. `MapPgxError` (23503→409, 42501→500, etc.) industrialise un mapping qu'aucun handler actuel n'utilise : c'est une nouvelle convention introduite au nom d'AGENTS.md, pas un besoin démontré.
- Raison de simplicité : créer le paquet au moment de la première adoption (Phase 2, avec les vrais signatures appelants réels) évite une conception à l'aveugle et une fenêtre de code mort — même principe que le reste du plan, qui adopte chaque déplacement immédiatement.

**2. [Major] Impasse architecturale — `httpx/identity.go` dans le platform lit le contrat de contexte du module auth**
- Scope : P1.6 (l.228 : « recopiés de controllers/user.go:19-33 »), Phase 2 (l.132), règle de dépendance §4 (l.116).
- `getUserID/getTenantID/getRole` lisent les clés de contexte (`userID`, `tenantID`, `role`) posées par `TokenAuth`, qui est déplacé dans le **module** `auth` (décision 3). Placer les accesseurs dans `internal/platform/httpx` fait dépendre le platform d'un savoir détenu par un module que le platform « n'importe jamais » : la règle est respectée en Go (aucun import), violée en sémantique (la source de vérité des clés est dupliquée hors du module qui la définit).
- Alternative plus simple : mettre ces accesseurs dans le module `auth` (propriétaire du middleware qui pose les clés), ou les laisser avec les handlers qui les consomment — ce qui ne crée aucun cycle puisque les handlers sont dans les modules.

**3. [Nit] Montée en charge des couches par rapport à la taille réelle des domaines**
- Scope : §4 (l.107-112), Phase 2 (l.129-130), décision 4.
- Cinq fichiers par module (handler/service/repository/dto/model) pour des domaines d'une seule table (article : 5 requêtes sqlc) ; le module `tenant` est réduit à un slug et trois opérations. Chaque repository se voit une interface à implémentation unique dont l'unique bénéfice mesuré est un fake de test. C'est défendable (idiome Go « interface côté consommateur », validation utilisateur actée), mais le risque à surveiller est la couche vide : si `dto.go` et `model.go` ou `service.go`/`repository.go` finissent par déformer le code une seule fois, le coût de la cible dépassera son bénéfice. Le plan ne fixe aucun critère de fusion de couches.

**4. [Nit] Extension de couverture de tests mêlée à un refactoring à comportement inchangé**
- Scope : §7 (l.255), Phases 2/5/6.
- Chaque phase module doit ajouter une suite `service_test.go` avec fake. Or les cas exigés (0 ligne → 404, email dupliqué, audit dans la Tx) sont déjà épingle par `tests/article_test.go` et `login_test.go`, que la migration s'interdit de modifier. Ajouter des tests unitaires est utile mais c'est un nouveau périmètre pendant un chantier dont la promesse est le zéro changement de comportement ; le risque de calendrier (chaque phase grossit sans filet proportionnel) n'est pas discuté. Un critère « nouveau test seulement si comportement non couvert par la suite HTTP » suffirait.

**5. [Nit] Weeds — précision au niveau des lignes d'imports**
- Scope : §6 P1.2/P1.3 (l.207, 213 : listes de 7 puis 12 fichiers avec numéros de ligne).
- Les numéros de ligne (`main.go:16`, `models/auth.go:11`, etc.) seront obsolètes dès la première étape validée ; le détail utile est la liste des fichiers, pas leurs offsets. Tenir ce niveau de granularité fragilise la Phase 1 sans la rendre plus sûre (la compilation verte après chaque git mv est le vrai contrôle).

**6. [Nit] Bénéfice attendu de la migration non énoncé en tête**
- Scope : Objectif (l.3) vs §3 (l.82-91).
- Le document ouvre sur la cible (layout) et non sur le problème racine (état global non injecté, logique métier dans les handlers, câblage dupliqué — qui n'apparaissent qu'en §3 comme « couplages et points de friction »). Un lecteur naïf doit inverser le document pour comprendre le *pourquoi*. Une phrase de motivation en-tête suffirait ; rien à restructurer.

## Bilan

Le document respecte le Zoom-Out Standard sur l'essentiel : contexte suffisant, cartographie vérifiée exacte contre le code, phasage avec critères de sortie vérifiables, invariants RLS/sécurité explicitement préservés, speculative features explicitement repoussées hors périmètre (RBAC, CORS, httpOnly, montées de versions). Les phases correspondent au layout réel du dépôt. Les deux constats Majors portent sur `internal/platform/httpx` : un paquet créé trop tôt sans appelant, et un placement d'accesseurs d'identité qui duplique la source de vérité du module `auth` dans le platform.

---

## Reviewer 2 — egm-inventive-catalyst (TRIZ, Mode REVIEW)

# Rapport d'audit stratégique — Migration modulaire par domaine (Inventive Catalyst, Mode REVIEW)

## Étape 0 — Existence de contradictions

Le document ne présente aucun compromis structurel majeur comme inévitable : les ports sont définis côté consommateur, la plateforme n'importe pas les modules, la racine de composition `internal/app` centralise le câblage, la transaction d'enregistrement est reconnue comme couplage inter-domaines et traitée par orchestration explicite, et les compromis de sécurité (bcrypt à coût constant, pas de double sondage d'annuaire) sont déjà résolus. Deux résidus subsistent : un potentiel d'échec non géré et un couplage manuel non examiné. Ce sont des avis, pas des blocages.

---

**Sévérité : Major** (avis)
**Périmètre concerné :** Phase 7, module `auth` — flux Refresh (`ConsumeRefreshAuth` puis relecture utilisateur puis ré-émission, cf. §5 Phase 7 et §3.10).
**Compromis implicite :** le plan fige l'ordre « consommation du refresh d'abord, validation ensuite » comme contrat testé (`TestRefreshOnlyConsumesMatchingOwner`) sans examiner le chemin d'échec après consommation : si la relecture utilisateur échoue ou que le compte est supprimé/suspendu entre `ConsumeRefreshAuth` et l'émission, le jeton de rafraîchissement est détruit sans contrepartie — le client perd sa session de manière définitive. Ce mode d'échec existant n'est ni signalé comme dette ni couvert par un test d'échec ; l'audit le considère comme une loi du système alors qu'il est un choix d'implémentation.
**Résolution proposée :** ne pas changer le comportement pendant la migration (le contrat est épinglé), mais ajouter au backlog de Phase 7 le motif *Beforehand Cushioning* : consommation atomique conditionnelle dans le script Lua (consommation conditionnelle au résultat de validation n'est pas réalisable en un script, mais un motif *Blessing in Disguise* est applicable — enregistrer l'échec post-consommation comme événement de session (clé mortelle `user_sessions:*` ou file d'échecs) plutôt que comme simple retour 401, pour rendre l'incident observable et réversible côté support.
**Liabilités secondaires :** logique d'orchestration supplémentaire dans le module `auth`, clé Redis additionnelle à purger dans les environnements jetables (index 2/3), risque de fausse alerte si la détection d'échec est approximative.

---

**Sévérité : Nit** (avis)
**Périmètre concerné :** Phase 4, module `tenant` — liste `reservedSlugs` et boucle de collision slug (§5 Phase 4, Risques).
**Compromis implicite :** deux traitements manuels présentés comme acceptables : (a) la liste `reservedSlugs` contenant « des racines de routes » doit être maintenue à la main en synchronisation avec la table de routes — un couplage étroit entre deux sources de vérité sans mécanisme de dérivation ; (b) l'épuisement des 99 tentatives de collision de slug n'a pas de comportement documenté (erreur renvoyée ? forme ? code HTTP ?) ni de test le couvrant — potentiel d'échec traité comme détail d'implémentation.
**Résolution proposée :** (a) motif *Intermediary* : dériver les slugs réservés de la table de routes montée par `internal/app` (le router devient l'intermédiaire unique, la liste statique ne sert plus qu'au seed initial), ou à défaut un test contrat qui échoue si une nouvelle racine de route n'est pas dans `reservedSlugs` — réutilisant la mécanique `TestAPISurface` (P0.3) ; (b) spécifier et tester l'échec d'épuisement (erreur discriminée, code HTTP attendu) dans le `service_test.go` pur exigé par §7 pour la Phase 4.
**Liabilités secondaires :** la dérivation depuis le routeur accroît le couplage du module `tenant` à `internal/app` (à casser par une interface fournie en dépendance `ReservedSlugs`), coût de coordination supplémentaire ; le test contrat sur les slugs est un artefact de maintenance de plus.
