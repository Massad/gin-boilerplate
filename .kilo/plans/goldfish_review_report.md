# Goldfish Annotated Review Report — Résolution

**Target :** `.kilo/plans/1795597264251-architecture-modulaire-par-domaine.md` (fichier complet, depuis scindé) — correctifs appliqués sur `1795597264251-architecture-modulaire-par-domaine-design.md` + `1795597264251-architecture-modulaire-par-domaine-plan-phases.md` (ancrages remappés via le mapping du stub)
**FindingsList :** `.kilo/plans/1795597264251-consolidated-review-report.md` (A1-A2 Blockers, B1-B10 Majors, C1-C4 mineurs)
**Status :** resolved — 16/16 correctifs PROPOSE FIX appliqués (partie 2) ; décisions validées : A2 → Option 1 « Tolérance transitoire explicite », B10 → OUI (déplacement `invoice/` acté). Bilan : 16 acceptés / 0 contestés / 0 escaladés / 0 déclinés (session du 2026-11-25).

## 1. Bilan

16 findings évalués : **16 PROPOSE FIX — 0 CONTEST — 0 ESCALATE — 0 DECLINE.**
Ancrages vérifiés contre le dépôt et les standards du plugin (constat confirmé, non répété ici). Aucun finding contestable : tous les correctifs sont des révisions du **document** ; aucune retouche de code requise avant exécution. Descriptions/rationales complets : cf. rapport consolidé (matrice de traçabilité incluse).

| ID | Sévérité | Suggestion | Correctif (résumé) |
|---|---|---|---|
| A1 | Blocker | PROPOSE FIX | Insérer « Alternatives considérées » (4 arbitrages + coût du rejet) après les décisions structurantes |
| A2 | Blocker | PROPOSE FIX | **Décision §2** : fenêtre transitoire racine↔modules Phases 3-6 + critère §7 reformulé + métrique réelle |
| B1 | Major | PROPOSE FIX | Préambule Problème / Objectifs / Non-objectifs consolidés |
| B2 | Major | PROPOSE FIX | « Risques transverses et rollback » en fin de §5 (ligne de revert par phase) |
| B3 | Major | PROPOSE FIX | `httpx` différé en Phase 2 — P1.6 devient une note de renvoi ; compteurs Phase 1 ajustés |
| B4 | Major | PROPOSE FIX | `identity.go` retiré de `httpx` ; helpers locaux aux consommateurs ; contrat des clés documenté dans `auth` |
| B5 | Major | PROPOSE FIX | Ports `UserAuthenticator` + `UserReader.OneWithTenant` + `AdminAuthenticator` définis en Phase 3, implémentés Phases 5-6 |
| B6 | Major | PROPOSE FIX | `Deps` figé `{TemplateGlob, Env}` ; plage de recopie 127-134 + 148-212 ; énumération exhaustive (TrustedProxies, gzip) ; test d'équivalence `Routes()` ; dettes documentées |
| B7 | Major | PROPOSE FIX | Inventaire P1.5 corrigé : site cassant réel `factory_test.go:86`, retrait de `factory.go`, import `main.go:20` supprimé, lignes « Vert » ajoutées |
| B8 | Major | PROPOSE FIX | Sémantique Lua (2 phrases), valeurs `kind` (`tenant_user`/`superadmin`), `DummyHashComparisons` en risque Phase 7, dette refresh en backlog |
| B9 | Major | PROPOSE FIX | Calibrations factuelles : cadence e2e (l.90), smoke « rouge » → conflit (l.88), P0.3 services (l.192), « prouve » → « vérifie » (l.241), « sécurise » (l.91), engagement futur contrat SPA (l.90) |
| B10 | Major | PROPOSE FIX | **Décision §3** : `invoice/` — déplacement Phase 8 tranché |
| C1 | Nit | PROPOSE FIX | Hygiène éditoriale : platforme→plateforme (l.9, 64, 68), RésOLU (l.87), sécoupe (l.83), épinglé(es) (l.45, 75, 87, 91), Bougeera (l.22), « ancienne monde » (l.137), métaphores (l.23, 38, 83, 160), hedge (l.72) |
| C2 | Nit | PROPOSE FIX | Préambule titré ; titre §6 reformulé ; « (l'étape structurante) » supprimé ; comptages 11 fichiers / 6 commits ; renvois §1↔§3 et §7 source unique ; P0.2 décompressé ; `.kilocode/rules/` reformulé (l.170) ; « répertoires sondés » (l.164) ; alignement ASCII |
| C3 | Nit | PROPOSE FIX | `-mod=readonly` sur toutes les lignes Vert (l.203, 208, 216) ; renommages `db.*`→`redisdb.*` + alias `main.go:18` ; commentaire `cmd/e2edb/main.go:200` ; Vert P1.4 ; prérequis `make migrate_up` en P0.2 ; normalisation P0.3 (`{id}`, préfixe `/v1`, égalité bidirectionnelle) ; note lignes précaires (Z5) ; critères Z3/Z4 |
| C4 | Nit | PROPOSE FIX | `reservedSlugs` : test contrat (mécanique P0.3) + échec d'épuisement des 99 tentatives spécifié (contrat actuel épinglé, erreur discriminée post-migration) |

## 2. Décision A2 — Mécanisme de transition racine↔modules (Phases 3-7)

**Constat (vérifié).** Le critère §7 (l.258, formulation ambiguë « en sens inverse ») contredit la Phase 3 (l.137 : injection du `*auth.Service` dans `models/user.go`, `models/login.go`, `controllers/user.go`, `controllers/admin.go`). Les Phases 3-6 nécessitent des imports racine→module **et** module→racine : `auth/handler.go` (admin Login/Logout, user Logout migrés dès la Phase 3, l.136) consomme `models.AdminModel`/`UserModel` tant que les Phases 5-7 ne les ont pas déplacés. La Phase 7 supprime les répertoires racine (l.163) : la fenêtre se referme naturellement.

### Option 1 — Tolérance transitoire explicite (recommandée)

- **Principe** : fenêtre transitoire actée Phases 3-6 — imports racine↔modules autorisés **dans les deux sens**, uniquement pour le code en cours de migration ; interdiction vérifiée en sortie de Phase 7 (répertoires supprimés → décompte 0).
- **Modifications documentaires** : l.137 (encadré « Tolérance transitoire » en Phase 3) ; l.258 (critère remplacé : décompte réel des fichiers sous `controllers/`, `models/`, `forms/`, `middleware/` — décroissant à chaque phase, 0 en sortie de Phase 7, vérifié par `git ls-files` ; invariant absolu conservé : `internal/platform/*` et `cmd/e2edb` n'importent jamais `internal/modules/*`).
- **Impacts fichiers/CI** : aucun — pure clarification du plan ; les phases restent buildables telles quelles, aucun code supplémentaire.
- **Risque résiduel** : la fenêtre dure plus longtemps que prévu si les phases glissent — couvert par la métrique publiée à chaque sortie de phase + vérification finale.

### Option 2 — Signatures paramétrées (interfaces consommateur)

- **Principe** : éliminer l'import racine→module — chaque consommateur racine reçoit le service en paramètre (`UnifiedLogin(form, ts TokenIssuer)`), interface définie côté consommateur (l.116 l'autorise), injectée depuis `internal/app`.
- **Impacts** : interfaces transitoires créées dans `models/`/`controllers/` puis supprimées en Phase 7 ; retouches de signatures en pleine migration (éditions de contenu mêlées aux déplacements — tension avec la discipline §6 « jamais mélanger `git mv` et édition dans un même commit ») ; charge de test accrue pour du code jetable ; **le sens module→racine subsiste quand même** (les handlers `auth` migrent en Phase 3 avant les models `admin`/`user` des Phases 5-7) : l'option ne ferme qu'un des deux sens.
- **Rejet** : coût réel, bénéfice partiel, code transitoire supprimé de toute façon en Phase 7.

**Recommandation : Option 1** — c'est le mécanisme que le plan applique déjà de fait (l.137) ; la Zero-Decision Rule exige de l'écrire noir sur blanc : fenêtre explicite, invariants absolus, métrique vérifiable. Convergent avec le Stage 8 (« transition racine→module explicitée »).

## 3. Décision B10 — Déplacement `invoice/` (Phase 8)

**Constat.** l.263 laisse le déplacement « recommandé mais optionnel » en Question ouverte ; le layout cible §4 (l.112) affiche déjà `internal/modules/invoice/` ; l'exécutateur de la Phase 8 devrait décider → viole la Zero-Decision Rule.

### Option OUI — déplacer (recommandée)

- **Mouvement** : `git mv invoice internal/modules/invoice` — paquet pur (lib maroto + `invoice_test.go` poppler), sans route ni DB, aucun appelant applicatif.
- **Impacts fichiers** : le paquet lui-même ; `AGENTS.md` (structure + commande tests PDF `./invoice` → `./internal/modules/invoice`) et `README.md`/`memory-bank/` — déjà dans le périmètre Phase 8 (l.169) : aucun fichier hors périmètre.
- **Impacts CI/Makefile** : aucun workflow ne référence le chemin du paquet (les seules mentions `invoice` de `ci.yml` sont les routes `/v1/invoice` supprimées par P0.1) ; la commande documentée devient `go test -mod=readonly -count=1 -race -v ./internal/modules/invoice` (poppler requis, inchangé).
- **Alignements** : layout §4 exact ; l.263 retirée des Questions ouvertes ; cohérent avec « domaine futur billing » (§2, l.58).

### Option NON — rester à la racine

- **Impacts** : retirer `internal/modules/invoice/` du layout §4 (l.112) ; réécrire l.263 en décision actée (« reste racine ») ; `AGENTS.md` inchangé sur ce point.
- **Inconvénient** : un répertoire métier hors du layout modulaire après une migration dont c'est l'objet ; le futur module billing repartirait d'un second déménagement.

### Option déclencheur conditionnel

- Rejetée : repousse la décision à un événement non daté — viole de nouveau la Zero-Decision Rule au point d'entrée Phase 8.

**Recommandation : OUI, acté** — un commit mécanique de plus en Phase 8, tests inchangés, dernière décision ouverte du plan supprimée.

## 4. Plan de résolution — partie 2 (appliqué)

Modifications détaillées (localisation ligne/section → texte exact de remplacement) générées dans cet ordre sur le document cible :

1. **Préambule** : B1 (Problème/Objectifs/Non-objectifs), A1 (Alternatives ×4), C2 (titre du préambule).
2. **§3 Couplages** : B8(1) sémantique Lua, B9 (l.88, l.90, l.91), C1/C2 (épinglé(es), renvoi singletons l.82).
3. **§4 Layout** : B3 (httpx « créé en Phase 2 », l.104), C2 (alignement ASCII).
4. **§5 Feuille de route** : A2 (l.137 fenêtre transitoire), B3/B4 (étape httpx + helpers locaux en Phase 2), B5 (ports Phase 3 + réorientations l.150/156/161-162), B8(2)(3) (`kind`, `DummyHashComparisons`, l.164), B2 (Risques transverses et rollback), B10 (l.168 acté ; l.263 supprimée), C4 (l.145-146), C1 (registre l.23, 38, 83, 137, 160), C2 (rubriques Phases 0/1).
5. **§6 Phases 0-1** : C3 (prérequis `make migrate_up` P0.2 ; normalisation P0.3 ; `-mod=readonly` l.203/208/216 ; renommages `redisdb.*` + alias + commentaire e2edb ; Vert P1.4 ; note Z5), B7 (inventaire P1.5 + Vert ; l.213 « 11 fichiers »), B3 (l.126 compteurs ; P1.6 différé ; l.199 « 6 commits »), B6 (l.231 titre ; l.232 `Deps` figé + plage 127-134/148-212 + énumération ; l.234 appel sans « … » ; test d'équivalence `Routes()` ; dettes Δ surface + `binding.Validator`), C2 (titre §6 l.175 ; l.170 ; P0.2 décompressé), C1 (l.22, 45, 75, 87).
6. **§7 Critères** : A2 (l.258 métrique réelle + tolérance + invariants), C3 (Z3, Z4), C2 (§7 source unique).
7. **Questions ouvertes** : B8(4) dette refresh (jeton consommé puis relecture échouée = session perdue sans contrepartie), B10 (retrait de la ligne).
8. **Passe finale** : relecture C1/C2 — cohérence des renvois, titres, hedges restants.

Application : exécutée en une passe le 2026-11-25 (décisions A2/B10 validées en gates interactifs) ; passe finale C1/C2 rejouée — P0.1 du plan-phases harmonisé avec la calibration B9, clause évaluative résiduelle de §3.2 retirée.
