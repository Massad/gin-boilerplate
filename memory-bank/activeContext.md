# activeContext.md — go-artisan-services

> File de contexte courant — mettre à jour à CHAQUE session (pattern memory bank).

## État actuel (2026-11-21)

**Plan multi-tenant RLS conclu à 100 %** (12/12 étapes, parties I + II du plan `1795250166405-multi-tenant-postgres-rls.md`). Aucun travail en cours, aucune tâche bloquée. Le bilan d'exécution détaillé et les pièges d'environnement sont dans `.kilo/HANDOFF.md`.

## Récent (dernière session)

- Étape 12 (T18/E7) : retrait sqlx/lib/pq vérifié no-op, garde-fou CI « Check removed legacy drivers » (pattern sur imports quotés — le commentaire `lib/pq` de `internal/db/pool.go:70-72` est volontairement non matché), AGENTS.md + README.md mis à jour, docs Swagger non régénérées (aucune annotation touchée).
- Nettoyage : 40 `.kilo/*.sh` one-offs supprimés (+ secrets embarqués), 6 `*_report.md` de la phase design supprimés, `.cache/` (137 Mo) supprimé, binaire `gin-boilerplate` supprimé + pattern `.gitignore` corrigé (`/gin-boilerplate`).
- Revalidation complète : build, vet, unitaires purs (config, models), suite tag `all` sur 7 packages **y compris `./` (main_test.go, maillon manquant de l'étape 12)** — tout vert en `-race` (RLS prouvée en `app_user`/`app_admin`).
- Création de ce memory bank.

## Candidats pour la prochaine session (à arbitrer avec l'utilisateur, hors périmètre du plan actuel)

- Câbler `RequirePermission` (routes owner-only) + invitations/membres (colonne et map déjà posées).
- SPA PWA (repo séparé) — le fallback backend est déjà en place.
- Conséquences du nettoyage : `.cache/` sera retéléchargé au prochain `make install_*` ; tout le travail du plan reste non commité (aucun commit demandé jusque-là) — à committer quand l'utilisateur le décide.
