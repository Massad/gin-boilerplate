# Plan — Retirer l'accès public + rédiger la spec multi-tenant : `specs/multi-tenant.md`

## Contexte et décisions

- Décision 1 : la spec vit dans `specs/multi-tenant.md` (nouveau dossier racine, séparé de `docs/` réservé au Swagger auto-généré).
- Décision 2 (correction utilisateur) : le lookup public par slug est une **erreur** — « il n'y aura rien de public », l'accès aux outils du SaaS exige une identification. Le retrait intégral de l'accès public (lookup `/v1/public/:slug`, routes invoice démo, Swagger hors production) est porté par le plan dédié `.kilo/plans/1795284300000-remove-public-access.md`, prérequis de la présente spec.
- Le login/register/refresh (identification) et le shell statique SPA/assets restent les seuls non authentifiés — c'est une conséquence nécessaire : on ne peut pas s'identifier sans charger l'app ni sans routes d'authentification.

## Phase A — Prérequis : exécuter le plan de correction

Exécuter d'abord `.kilo/plans/1795284300000-remove-public-access.md` (fix intégral de l'accès public, périmètre et validation détaillés là-bas). La spec ci-dessous décrit le réel **d'après le résultat de ce fix**.

## Phase B — Rédaction de `specs/multi-tenant.md`

Règles : français, factuel, as-built (après phase A) ; invariants SQL recopiés exactement depuis `db/migrations/000002_multi_tenant.sql` ; chemins sources cités ; ce qui est posé mais non câblé est marqué comme tel. Ne pas éditer `internal/db/` (généré), ne pas régénérer Swagger.

Structure :

1. **Objet et périmètre** — spec as-built de l'isolation multi-tenant ; sources de vérité (migrations, pool, middleware, queries, CI).
2. **Décision retenue** — schéma partagé + colonne `tenant_id` + Row-Level Security PostgreSQL **échec fermé** ; alternatives écartées (base par tenant, schéma par tenant) ; défense en profondeur : RLS + double pool/rôles + paramètre `@tenant_id` explicite dans les requêtes sqlc.
3. **Terminologie** — `tenant` (racine d'isolation), `user` (utilisateur métier d'un tenant, rôles `owner`/`member`), `admin_user` (admin plateforme, hors isolation), pool plateforme (`AdminStore`, BYPASSRLS) vs pool tenant (`TenantStore`, RLS).
4. **Invariant d'accès — rien de public** — aucune route non authentifiée n'expose de données. Seules exceptions : login, register, refresh (identification) et le shell statique SPA/assets (`r.NoRoute` fallback, `/public` statique, `/assets` immutables). Tout le reste exige un token : `TokenAuth()` (JWT tenant signé `ACCESS_SECRET`) ou `AdminAuth()` (JWT admin signé `ADMIN_ACCESS_SECRET`) — 401 sinon. L'UI Swagger n'est servie qu'hors `ENV=PRODUCTION`. Le lookup public `GET /v1/public/:slug` et les routes démo invoice ont été retirés à cette fin (cf. plan de correction).
5. **Invariants d'isolation** (SQL exact de la migration 000002) :
   - Toute ligne métier porte `tenant_id uuid NOT NULL REFERENCES tenant(id) ON DELETE CASCADE`.
   - Tables protégées : `"user"`, `article` — `ENABLE` **et** `FORCE ROW LEVEL SECURITY`, policy `tenant_isolation` : `USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)` → échec fermé (pas de contexte ⇒ 0 ligne, même pour le propriétaire de la table grâce à FORCE).
   - Tables plateforme **sans** RLS : `tenant`, `admin_user`, `admin_audit`.
   - Intégrité intra-tenant : FK composite `article(tenant_id, user_id) → "user"(tenant_id, id)` (un `user_id` ne peut jamais référencer un user d'un autre tenant), index unique partiel `user_one_owner_idx` (1 owner par tenant), `CHECK (role IN ('owner','member'))`.
6. **Rôles et pools PostgreSQL** — `app_user` : login NON-superuser, 100 % du métier, contraint par RLS ; `app_admin` : `BYPASSRLS`, réservé aux opérations plateforme (login, register, refresh, backoffice admin, création/suspension tenant, audit). Justification œuf-poule : un SELECT user sans contexte tenant renverrait 0 ligne. Mots de passe dev dans la migration, prod via `ALTER ROLE ... PASSWORD`. Accesseurs `pgdb.TenantStore()` / `pgdb.AdminStore()` (`internal/db/pool.go`).
7. **Source de vérité du tenant et propagation** — le tenant vient **uniquement** du claim `tenant_id` du JWT signé, jamais d'un header/hôte/chemin client ; `pgdb.WithTenant(ctx, tenantID, fn)` = transaction + `SELECT set_config('app.tenant_id', $1, true)` (portée transaction, auto-reset au commit/rollback ; **interdit** : `SET` session-wide sur pool partagé) ; double défense : requêtes sqlc tenant-scoped avec `@tenant_id` explicite ; chaîne HTTP : `TokenAuth()` → `c.Set("tenantID")` (string) → `uuid.Parse` contrôleur → `WithTenant`.
8. **Sessions et impact multi-tenant** — access + refresh avec claims `tenant_id`/`role` ; Redis `user_sessions:{userID}` ; le refresh **relit** le user via le pool admin (rôle frais, existence, `tenant.status` — jamais les claims) ; tenant suspendu ⇒ 403 ; rejet croisé des tokens tenant/admin (secrets distincts, préfixe Redis `admin:`, pas de refresh admin, token ~8h).
9. **Backoffice plateforme et audit** — `admin_audit` insert-only ; `admin_audit.tenant_id` **sans** FK (l'audit survit à la suppression du tenant) ; suspension/réactivation tracée (`/admin/tenants/:id/status`).
10. **Cycle de vie du tenant** — création sur le pool admin (l'enregistrement n'est pas une opération tenant-scoped) ; slug : `Slugify` (NFD, sans accents, tirets) + pattern `^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$` (3–40 car.) + slugs réservés + suffixes `-2`, `-3`… jusqu'à 99 tentatives ; statut `1` = actif (défaut), `0` = suspendu. Le slug n'est plus résolvable publiquement depuis le retrait du lookup.
11. **RBAC minimal** — `owner`/`member`, `RequirePermission` (permissions `members:manage`, `tenant:manage`) : **posé mais non câblé** — le dire explicitement.
12. **Erreurs et sémantique HTTP** — `pgx.ErrNoRows` → 404 ; `23505` → 406/409 ; `23503` → 409 ; `42501` (RLS) → 500 ; jamais de texte SQL brut au client ; 401 = non authentifié, 403 = rôle insuffisant ou tenant suspendu.
13. **Prouver l'isolation** — tests et CI en rôles non-superuser `app_user`/`app_admin` : sans contexte ⇒ 0 ligne ; insert cross-tenant ⇒ 42501 — comportement **attendu**, ne jamais « corriger » en superuser ; CI : migrations en superuser puis tests en rôles applicatifs.
14. **Interdits (anti-patterns)** — `SET` session-wide ; concaténation SQL en Go (toute requête vit dans `db/queries/*.sql`) ; superuser pour l'app/les tests ; tenant déduit d'une entrée client ; métier tenant hors `WithTenant`/hors pool tenant ; éditer `internal/db/` généré ; nouvelle route publique exposant des données.
15. **Recette : ajouter une ressource tenant-scoped** — migration (`tenant_id NOT NULL` + FK + `ENABLE`/`FORCE` RLS + policy) → requêtes `db/queries/*.sql` avec `@tenant_id` → `make generate_sqlc` (généré commité) → accès `TenantStore().WithTenant` → tests RLS cross-tenant dans la suite `all`.
16. **Limitations connues** — `RequirePermission` non câblé (cf. §11) ; invitations/membres à venir.

Tâches phase B :
- Écrire `specs/multi-tenant.md` selon la structure ci-dessus, après validation de la phase A.
- Vérifier chaque fait technique contre les fichiers sources avant de l'écrire — n'inventer rien.
- Mise à jour légère du memory bank (convention « à chaque session ») : une ligne dans `memory-bank/activeContext.md` (récent) et `memory-bank/progress.md` mentionnant le retrait du lookup public et la création de la spec.

## Hors périmètre

- Câbler `RequirePermission`, invitations/membres, SPA frontend.
- Commit/git (aucun demandé).

## Risques

- Faible. Le principal risque de la phase A est une référence résiduelle oubliée → mitigé par `go vet` + build + grep final de `PublicController|GetTenantBySlug|OneBySlug|/v1/public`.
- Phase B : inexactitude factuelle → mitigée par recopie du SQL exact et vérification contre sources.
- `make test` nécessite PG + Redis actifs ; si indisponibles, le signaler explicitement.
