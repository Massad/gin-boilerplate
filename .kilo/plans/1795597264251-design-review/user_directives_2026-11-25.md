# Directives utilisateur — actées au gate Stage 5 (2026-11-25)

Réponse au gate Stage 5 : validation du rapport + passage Stage 6, conditionné à l'intégration de 4 directives architecturales AVANT toute génération de code. Les Nits sont explicitement exclus de cette passe (non bloquants).

## Directives et traduction en éditions

1. **Découplage inter-domaines (l.116)** — interdiction de passer `*pgdb.Queries` entre modules ; pattern Repository avec interfaces explicites (`UserRepository`, `TokenRepository`, …) ; transactions via un Unit of Work explicite.
   - Design : décision structurante nouvelle **6** ; §3.4 réécrit (Unit of Work `pgdb.WithinTx` + interface étroite `pgdb.InTx`).
   - Plan : P4 `TenantProvisioner.CreateInTx(ctx, tx pgdb.InTx, name)` (remplace `q *pgdb.Queries`) ; P5 orchestration par `pgdb.WithinTx`.
2. **Règle de dépendance (l.147 vs l.34)** — interfaces définies côté consommateur ; `TokenService`/`TenantProvisioner` déplacées vers les modules consommateurs ; distinction ports d'orchestration vs services techniques dans le wiring.
   - Design : décision 4 amendée + §4 réécrit (les deux familles, distinction au câblage `internal/app`).
   - Plan : P3/P5/P6 « Interfaces » réécrites (port `TokenService` déclaré par `user`/`admin`, implémentation d'auth injectée à l'app ; `TenantProvisioner`/`TenantReader` définis côté consommateur).
3. **Session/refresh token (l.152)** — soft-consume strict : `consumed_at` nullable, hash du token, TTL court 5-10 min, job de cleanup, pas de suppression immédiate ; rejeu pendant grâce → `401 invalid_grant` (ou re-émission selon politique, figée en exécution) ; traité comme événement support avec alternative motive.
   - Design : décision structurante nouvelle **7** ; §3.10 carve-out ; Questions ouvertes mise à jour (tranché).
   - Plan : tâche explicite en Phase 3 (+ carve-out contrat documenté en Risques transverses) ; ownership inchangé, tests dédiés nouveaux.
4. **Impact documenté (l.121)** — mutation globale `binding.Validator` élevée en tâche de backlog explicite, tracée et planifiée, non enfouie.
   - Design : §3.9 réserve (c)+ « Deux → Trois réserves ».
   - Plan : Phase 8 nouvelle puce « Sortie de dette tracée » (relocalisation hors de `NewRouter`, élimination requise à la sortie de migration).

Nits non édités (délégués aux étapes suivantes) : CORS hardcode, gabarits minces, `err.Error()` brut, références hors-doc, fenêtre transitoire en deux voix (partiellement couvert par la réécriture §4).
