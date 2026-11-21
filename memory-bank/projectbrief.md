# projectbrief.md — go-artisan-services

## But

API REST Go (Gin) multi-tenant sur un seul domaine, fondation SaaS pour une PWA servie par le même backend. Isolation des données par tenant **garantie en base** (RLS PostgreSQL échec fermé), pas seulement en code applicatif.

## Périmètre livré (plan multi-tenant RLS, terminé 2026-11-21)

- Inscription/login simples (`email + password`, email globalement unique) ; le tenant est créé à l'inscription (`company` → slug).
- Résolution du tenant par le **claim JWT signé uniquement** (jamais header/hôte client, pas de sous-domaines — incompatible PWA).
- Fondation backoffice admin (login superadmin, liste tenants, toggle status, audit) + CLI `-create-admin`.
- RBAC minimal `owner`/`member` (colonne + map middleware, une seule route pas encore câblée — YAGNI assumé).
- Migration layer données : pgx/v5 (pgxpool) + sqlc + goose, plus aucun sqlx/lib/pq.

## Hors périmètre (déferré, §8 du plan)

SPA PWA (repo séparé), invitations/gestion membres (routes owner-only), dashboard admin complet, domaines custom, refresh cookie `httpOnly`, multi-appartenance, facturation.

## Contraintes fortes

- Jamais de `SET` session-wide sur un pool partagé (fuite inter-tenants) → `set_config(..., true)` en transaction uniquement.
- L'app et les tests se connectent en non-superuser (`app_user`/`app_admin`) ; superuser uniquement pour `make migrate_up`.
- Toute requête SQL vit dans `db/queries/*.sql` (sqlc) — zéro SQL concaténé en Go.
- Toujours `-mod=readonly` sur les commandes go.
