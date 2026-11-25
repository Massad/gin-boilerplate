# Stage 7 — Critic Review

Périmètre audité : FeatureRange 1..174 (design). Contexte Goldfish : aucun.
Délégation directe (mode uninterrompu) : egm-design + egm-structural + egm-style.

---

## Reviewer 1 — egm-design (Design Critic, standard feature-spec-standard)

## Rapport d'audit — Design Critic Goldfish

**Mode** : REVIEW (Document Audit). **Périmètre** : lignes 1-174 du document cible (design : décisions, cartographie, inventaire, couplages, layout, phases 0-8).

**Vérification factuelle préalable** : les affirmations contrôlables l'ont été. `main.go` (~231 l., routes `/v1` 148-182, `NoRoute` 404 sur `/v1*` en l.207) concorde ; les routes invoice sont absentes de `main.go` et `ci.yml:139,157,166-168` sonde encore `/v1/invoice` avec attendus 200/PDF — la dérive smoke CI décrite au §3.7 est réelle et le smoke est rouge sur master à ce jour. `models/login.go` (hash factice l.30, orchestration l.80-120) et `controllers/auth.go` (Refresh l.84-162, relecture user l.134-144) concordent avec le §3. La cartographie est fiable.

### Conformité au standard (feature-spec-standard.md)

- **[Major] — Périmètre : document entier (l.1-274), vs standard §« Alternatives Considered »** — Le standard rend obligatoire une évaluation des alternatives rejetées avec rationaux concrets. Le document ne contient aucune section « alternatives » : ni microservices vs modulaire-monolithe, ni découpage par couche dans les modules (handler/service/repository imposé, jamais justifié face à un handler/service simple), ni alternative « ne pas migrer / migrer seulement controllers », ni alternative sur le placement des middlewares (module auth vs platform). Les décisions structurantes (§ actes) sont présentées comme validées par l'utilisateur, sans trade-offs documentés. **Recommandation** : ajouter une section « Alternatives considérées » couvrant au minimum : (1) découpage en modules vs réorganisation légère des paquets racine, (2) hexagonal complet (ports/repos) vs service simple sur Queries, (3) middlewares dans platform vs module auth — avec le coût/complexité qui a motivé chaque rejet.

- **[Major] — Périmètre : l.1-3, §5 — vs standard §« Problem Statement / Goals & Non-Goals »** — Le problème n'est jamais énoncé comme tel. Les douleurs constatées (état global, logique métier dans les handlers, duplication `SetupRouter`) sont reléguées dans §3 comme observations, sans lien explicite vers l'objectif ni mesure du problème (pourquoi migrer maintenant ? quel coût de ne pas migrer ?). Les objectifs sont implicites (l'« Objectif » l.3 décrit la cible, pas le bénéfice). Les non-goals existent mais dispersés (§7 questions ouvertes, l.90, l.78). **Recommandation** : ouvrir le document par « Problème / Objectifs / Non-objectifs » explicites : bénéfices attendus (testabilité, onboarding, extensibilité RBAC/members), et consolider les non-goals (RBAC câblé, CORS configurable, refresh httpOnly, homogénéisation casse, montées de versions) en une seule liste avec leur statut.

### Risques architecturaux du plan

- **[Major] — Périmètre : §5 Phase 3 (l.135-140), Phase 7 (l.160-165), règle l.116/l.258** — La Phase 3 déplace `models/auth.go` (AuthModel : CreateToken/CreateAuth/ConsumeRefreshAuth) vers le module `auth`, alors que `models/login.go` (unified login, déplacé en Phase 7) et `models/user.go` (Login, Phase 5) et `controllers/auth.go` (Login) continuent de l'utiliser via la var globale `authModel`. Le mécanisme de transition entre Phases 3 et 7 n'est pas spécifié : soit les paquets racine importent `internal/modules/auth` (interdit ? la règle l.258 « n'importe … en sens inverse » est ambiguë sur ce point), soit des signatures changent par injection — mais le doc dit injecter « via internal/app » vers les consommateurs (controllers/models à l'ancienne monde) sans préciser comment `UnifiedLogin(form)` reçoit le service sans casser sa signature ni créer d'import racine→module. C'est la zone la plus à risque de la migration et elle est la moins décrite. **Recommandation** : spécifier explicitement la transition Phase 3-6 (schéma d'import autorisé : racine→module toléré en transitionnel, ou signature `UnifiedLogin(form, ts)` passée depuis le controller recevant le service), et clarifier la formulation de l.258.

- **[Major] — Périmètre : §5 Phase 7 (l.161-162)** — Le port `UserAuthenticator.Login(form)` (l.162) ne couvre pas la relecture utilisateur requise par Refresh (relecture `user` par ID avec tenant/role/status — `controllers/auth.go:136`, invariant « rôle re-fraîchi depuis la base »). Le texte dit « relecture user via port » sans définir ce port ni sa méthode. Sans port défini, l'exécuteur improvisera un import user→auth ou une fuite de contrat. **Recommandation** : définir un second port dans `auth` (ex. `UserRefresher` / `UserReader { GetWithTenant(ctx, userID) }`), implémenté par `user`, injecté aux côtés de `UserAuthenticator`, et l'ajouter à la liste des ports de la Phase 7.

- **[Major] — Périmètre : §5 P1.7 (l.232), §3.9 (l.90)** — La liste énumérée du câblage à recopier dans `internal/app/router.go` omet `r.SetTrustedProxies(nil)` (main.go:128) et le middleware `gzip.Gzip` (main.go:134). La plage « main.go:127-212 » les couvre en théorie, mais un exécuteur qui suit la seule liste produira un routeur sans compression ni sécurisation proxy — détecté seulement a posteriori par le smoke CI (assertions gzip/identity), c'est-à-dire une phase d'investigation inutile. **Recommandation** : compléter la liste de P1.7 avec `SetTrustedProxies` et `gzip`, et exiger une diff structurale (gin.Routes() + middlewares) entre l'ancien routeur et `app.NewRouter` à la fin de P1.7.

- **[Nit] — Périmètre : §6 P0.3 (l.191-194), §7 (l.254)** — Le test contrat compare `docs/swagger.json` à une liste figée, pas aux routes réellement câblées. Il ne détecte donc pas une déconnexion routeur ↔ annotations tant que les annotations restent. L'affirmation « échoue si une route disparaît » n'est vraie que pour la couche annotations. **Recommandation** : après la Phase 1 (quand `app.NewRouter` devient importable), ajouter un test d'équivalence `app.NewRouter(deps).Routes()` ↔ liste figée pour épingler le câblage réel, pas seulement la doc.

- **[Nit] — Périmètre : §6 P0.2 (l.185-189)** — `make test` exige, selon la convention du dépôt, une première `make migrate_up` sur la base dev (à cause de `internal/db/pool_test.go`) en plus de `services_up`. Le prérequis listé (`make services_up`) est incomplet ; sur un checkout frais la baseline P0.2 échouera sur un prérequis non documenté. **Recommandation** : ajouter `make migrate_up` aux prérequis de P0.2.

- **[Nit] — Périmètre : §5 Phase 7 (l.161), §3.10 (l.91)** — Le hook d'observabilité `DummyHashComparisons` est un état global package `models` (models/login.go:34-38) consommé par `tests/login_test.go`. Après Phase 7, il déménage dans le module `auth` ; le §7 n'est pas explicite sur le déplacement mécanique de cet import (c'est couvert par « déplacement mécanique de helpers » mais sans l'identifier). **Recommandation** : citer explicitement le déplacement de ce hook dans les risques Phase 7.

- **[Nit] — Périmètre : §5 P1.7 (l.232-234)** — `app.NewRouter` recopie l'affectation de l'état global `binding.Validator` (main.go:130) : un constructeur qui mute l'état global est un effet de bord caché, incohérent avec l'objectif de DI explicite affiché en l.3. **Recommandation** : acceptable en transitionnel, mais le consigner comme dette (la validation globale devrait finir dans l'entrypoint / platform validate, pas dans `NewRouter`).

- **[Nit] — Périmètre : l.116 vs l.258** — Les deux énoncés de la règle de dépendance se contredisent partiellement : l.116 « platform n'importe jamais modules » (silencieux sur racine→modules en transition) ; l.258 « aucun paquet racine n'importe un paquet internal/modules en sens inverse » — direction « inverse » mal définie. Voir finding Phase 3 ci-dessus.

- **[Nit] — Périmètre : §5 (l.118-172)** — Pas de stratégie de rollback explicite : la discipline « un commit par étape, branche par phase, merge si CI verte » l'implicite (revert), mais en cas d'échec en cours de phase (build rouge entre deux commits), la reprise n'est pas décrite. **Recommandation** : une ligne par phase : « en cas d'échec, revert jusqu'au dernier commit vert, re-diagnostic ».

### Points vérifiés et validés (sans finding)

- L'inventaire des domaines (§2) et la séparation super-admin/tenant (l.60-78) sont exacts et correctement sous-notés (« à ne pas surévaluer ») — le RBAC non câblé est confirmé.
- Les invariants (§ l.40-47) couvrent les contraintes validées : RLS fail-closed, pools, coût bcrypt constant, contrat HTTP au byte près (y compris la casse `"Message"`/`"message"` — risque rare mais correctement repéré), env de test.
- Le pilotage par article en Phase 2, le placement des middlewares dans `auth` sans cycle (ports côté consommateur), la suppression de la duplication `SetupRouter` dès la Phase 1, et le séquencement 7 « en dernier » pour le collage unified login sont des choix sains.
- Le plan ne modifie rien au comportement ; les risques opérationnels résiduels (smoke CI déjà rouge, prérequis env) sont correctement traités en Phase 0.

---

## Reviewer 2 — egm-structural (Layout Validator, standard feature-spec-standard)

# Rapport d'audit structurel — FeatureRange (lignes 1-174)

## Constats

**1. [Blocker] — Portée : document entier (lignes 1-174)**
Absence totale de section « Alternatives envisagées » / « Alternatives rejetées ». Le standard « Feature Specification » impose une évaluation obligatoire des alternatives avec raisons de rejet documentées (« The Alternatives Rule » : le document échoue s'il propose un chemin unique sans preuve d'alternatives considérées et rejetées). La section « Décisions structurantes actées » (lignes 5-11) énonce des choix (déplacement de `internal/db` en Phase 1, découpage en 4 modules, `internal/app`…) sans aucune alternative adverse (ex. gros mouvement unique vs migration incrémentale, maintien de `controllers/` avec DI partielle, placement des middlewares en platform plutôt qu'en module `auth`, autre granularité de découpage) ni leur rejet motivé. Aucun arbitrage n'est traçable.

**2. [Major] — Portée : document entier (lignes 1-174)**
Absence de section explicite « Objectifs et non-objectifs » (Goals & Non-Goals). Une notion d'objectif apparaît dans l'alinéa gras « **Objectif** » (ligne 3) et des non-objectifs sont dispersés inline (ligne 3 « aucune montée de version de dépendance », ligne 90 refresh httpOnly hors périmètre, ligne 168 migration `invoice/` « optionnel ») mais aucun n'est structuré en section dédiée avec les deux volets symétriques attendus par le standard.

**3. [Major] — Portée : document entier (lignes 1-174)**
Absence de section consolidée « Risques et mitigations ». Les risques existent uniquement par phase (puces **Risques** des §Phase 2-7, lignes 132, 139, 145, 151, 157, 164) et sont partiellement couplés à l'atténuation via les puces **Validation** ; les Phases 0, 1 et 8 n'ont aucune mention de risque. Le standard exige l'appel explicite des modes de défaillance (sécurité, scalabilité, limites) et de leurs mitigations ; les invariants ligne 42-47 couvrent RLS mais rien ne consolide les risques transverses à la migration (régression du contrat SPA, fenêtres de câblage cassé entre phases, rollback).

**4. [Nit] — Portée : « Problème » (lignes 3, 80-91)**
Pas de section « Problème à résoudre » (Problem Statement) dédiée. Le problème est dérivable de l'alinéa Objectif et de la section « 3. Couplages et points de friction », mais il n'est jamais formulé comme motivation métier/technique autonome, ce que le standard liste comme aire requise.

**5. [Nit] — Portée : §5 « Feuille de route par phases » (lignes 118-171), symétrie des sections parallèles**
Les neuf sous-sections de phase (Phase 0 à Phase 8) ne sont pas symétriques : les Phases 2-7 suivent un gabarit uniforme en puces (**Sources / Interfaces / Risques / Validation**, avec **Tests** en Phase 2), tandis que Phases 0 et 1 sont des coquilles de deux lignes déléguant à « §6 » (lignes 121, 125) sans gabarit, et la Phase 8 use d'un gabarit libre sans **Risques/Validation**. La délégation à §6 est légitime (hors périmètre d'audit), mais la rupture de gabarit au sein d'un même niveau H3 affaiblit la lisibilité comparée phase à phase ; au minimum un résumé structuré (livrables/risques/validation) par phase resterait attendu dans la partie design.

**6. [Nit] — Portée : titres H2 (lignes 5 et 15-118)**
Convention de numérotation incohérente : « Décisions structurantes actées » (ligne 5) est un H2 non numéré qui précède des H2 numérotés 1 à 5. Soit numéroter tous les H2, soit réserver la numérotation aux sections numérotées et expliciter le statut du bloc de décisions (préambule).

**7. [Nit] — Portée : §4 « Layout cible » (lignes 95-114)**
Alignement incohérent dans le bloc de code ASCII : les commentaires `# ←` ne sont pas alignés sur une colonne commune (espaces supplémentaires ligne 100 `internal/platform/redisdb/` et lignes 108 `internal/modules/auth/`, 111 `internal/modules/admin/`). Le bloc est syntaxiquement fermé et intact, mais sa géométrie rompt la symétrie interne de l'arborescence affichée.

**8. [Nit] — Portée : texte courant (lignes 9, 68, 87, 83, 137, 45)**
Coquilles récurrentes : « platforme » (lignes 9, 68 — deux variantes coexistent avec « plateforme » ligne 43), « **RésOLU** » (ligne 87, casse incohérente), « la sécoupe » (ligne 83), « restés à l'ancienne monde » (ligne 137), accords « les tests les épingle » (ligne 45) / « épinglees » (ligne 91). Aucun impact sémantique, mais la fiabilité typographique d'un document de référence en souffre.

## Structure conforme (constat positif)
La hiérarchie H1→H2→H3 est séquentielle et sans saut de niveau (aucun H3 sous H1, aucun H4 isolé). Les trois tables (lignes 19-38, 51-58, 64-70) sont bien formées : bordures d'en-tête cohérentes, nombre de colonnes homogène par table, aucune ligne cassée. L'unique bloc de code (lignes 95-114) est correctement ouvert et fermé. Les listes à puces sont homogènes (tirets) et le séquencement logique des cinq sections numérotées (cartographie → domaines → frictions → cible → roadmap) est sain.

---

## Reviewer 3 — egm-style (Technical Copy Editor)

# Rapport Style Goldfish — Audit (Mode REVIEW, FeatureRange lignes 1-174)

**Lecture** : document lu en intégralité ; audit limité à la partie design (lignes 1-174). Le ton général est factuel, dense et correctement ancré (références de code à l'appui) ; les findings ci-dessous sont des corrections de calibration et de formulation, pas de remise en cause du fond.

---

**F1 — Major — §3.9, ligne 90 — Claim non calibré sur la cadence e2e**
Style Rationale : « `make e2e` est rejoué aux Phases 1 et 7 précisément parce que ce sont les deux seules phases qui touchent au câblage servi à la SPA » affirme « les deux seules » de façon absolute. Contre-preuve dans le document lui-même : la Phase 2 câble un constructeur dans `internal/app` (ligne 130), la Phase 3 paramètre les middlewares et migre Logout/Login (lignes 136-139), les Phases 5-6 y injectent les services (lignes 149, 156). Le câblage servi est donc touché par les Phases 2 à 7. **CALIBRATE** : reformuler en « les deux seules phases qui changent l'entrypoint (`main.go` → `cmd/api`) et la structure de `router.go` », ou élargir la cadence e2e aux phases modifiant `internal/app`. En l'état, la phrase justifie l'absence de gate e2e sur des phases qui modifient quand même le câblage.

**F2 — Major — §3.3, ligne 88 — Claim d'état CI non adossé à une observation**
Style Rationale : « le smoke step est rouge sur master » est énoncé comme un fait présent observé, mais l'évidence décrite est une lecture croisée (routes supprimées de `main.go`, smoke attend 200/PDF) — une inférence, pas un run observé (pas de lien de run CI, pas de capture). **CALIBRATE** : « le smoke step entre en conflit avec `main.go` actuel — échec au prochain run » (vérifiable par la lecture croisée déjà décrite) ; si le rouge a été constaté sur un run réel, renvoyer explicitement à ce run.

**F3 — Nit — ligne 23, §3.2 (ligne 83), ligne 160 — Jugements subjectifs et métaphores**
Style Rationale : « Refresh **gorgé de logique métier** » (métaphore évaluative), « CRUD **propre** » (éloge subjectif), « `models/` n'est pas plus « métier » que les controllers : la découpe actuelle est **cosmétique** » (jugement de valeur traduisible en faits : la logique parse/consommation/ré-émission est dans le handler, lignes citées à l'appui), « **le plus critique**, en dernier » (intensif subjectif ; « le plus complexe » serait mesurable par le nombre de responsabilités listées). Remplacer par les descriptions factuelles déjà disponibles dans le document ; les métaphores n'ajoutent aucune décision.

**F4 — Nit — lignes 23, 121, 125 — Registre informel et phrasé marketing**
Style Rationale : « **Bougeera** vers `cmd/api/main_test.go` » (registre familial ; écrire « sera déplacé ») ; « §6 (**proposition prête à l'emploi**) » répété deux fois — formule de brochure sans valeur technique ; remplacer par « (description étape par étape en §6) ».

**F5 — Nit — ligne 38 — Étiquette subjective**
Style Rationale : « artefacts **parasites** » : jugement chargé ; « artefacts racine à supprimer (binaire de build, binaires Windows) » suffit. La décision (suppr. Phase 8) est identique.

**F6 — Nit — ligne 72, titre « Niveaux de droits » — Hedge non opérant dans le titre**
Style Rationale : « état réel **(à ne pas surévaluer)** » est une consigne d'interprétation floue en parenthèse. Les quatre puces suivantes énoncent déjà les faits calibrés (non câblé, grep, `member` non créable). **CALIBRATE** : supprimer le parenthétique et laisser le titre factuel (« état réel : RBAC minimal non câblé ») ; l'instruction au lecteur découle des puces sans injonction.

**F7 — Nit — §3.10, ligne 91 — Claim de couverture légèrement survendu**
Style Rationale : « **sécurise la migration** » : la couverture protège les invariants listés, elle ne sécurise pas la migration entière (la ligne elle-même admet des « zones à risque » restantes). **CALIBRATE** : « couvre les invariants listés ; zones restantes : câblage main, formes de réponses, script Lua ».

**F8 — Nit — §1 (lignes 19-38) vs §3 (lignes 82-83) — Redondance d'énumération**
Style F8 Rationale : l'inventaire des singletons/vars de package (`TenantStore`/`AdminStore`, `GetRedis`, `authModel`/`userModel`/`articleModel`/`adminModel`, références pool.go:92-121) et le détail du Refresh (controllers/auth.go:84-162) sont énoncés intégralement deux fois : cartographie puis friction §3.1/§3.2. La friction peut se contenter de renvoyer vers §1 (« mêmes singletons que cartographiés en §1 ; conséquence : tests peu isolables ») sans dupliquer les listes et numéros de ligne.

**F9 — Nit — orthographe, typographie et accords (lignes 9, 23, 45, 64, 68, 83, 87, 91)**
Style Rationale : erreurs qui dégradent la lecture d'un document d'ingénierie : « platforme » (lignes 9, 64, 68 — trois occurrences), « sécoupe » (ligne 83), « RésOLU » (casse aberrante, ligne 87), « Bougeera » (ligne 23), « les tests les épingle » (ligne 45, accord : « épinglent »), « épinglees » (ligne 91, accent : « épinglées »). Correction mécanique, aucune incidence décisionnelle.

**F10 — Nit — ligne 90 — Présent au lieu du futur pour un engagement de design**
Style Rationale : « Contrat SPA single-origin — **préservé** au byte près » : au stade plan, le contrat *sera* préservé par la copie verbatim décrite. Un titre au présent lit comme un constat déjà vérifié. Reformuler « préservé tel quel (copie verbatim de `main.go:127-212`) — vérifié par smoke + e2e ». Les deux caveats internes (CORS codé en dur, tokens-in-body) nomment chacun une décision conditionnelle concrète (« dès qu'une origine croisée existera », backlog explicite) : **KEEP**, non signalés comme défaut de calibration.

---

**Synthèse** : 2 Major (F1 calibration de la cadence e2e, F2 calibration de l'état CI), 8 Nit. Aucun Blocker. Aucune finding de type DROP : les caveats du périmètre nomment tous des décisions atteignables. Le fichier n'a pas été modifié (lecture stricte).
