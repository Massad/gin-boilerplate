# productContext.md — go-artisan-services

## Pourquoi

Héberger plusieurs entreprises (tenants) sur une seule instance API + une seule base + un seul domaine, sans multiplier le code ni le coût opérationnel, tout en rendant la fuite de données **techniquement impossible** côté stockage.

## Modèle produit

- L'utilisateur s'inscrit avec sa société : `company` → slug unique (`TenantModel.Slugify`, liste noire de mots réservés, suffixes `-2`, `-3`…, 406 au-delà).
- Chaque user appartient à **un** tenant avec un rôle (`owner` créé à l'inscription, `member` par défaut) ; un seul owner par tenant (index partiel unique en base).
- Le superadmin (table `admin_user` distincte) liste les tenants, suspend/réactive un tenant (audit dans `admin_audit`, sans FK pour survivre à une suppression de tenant). Tenant suspendu : login/refresh refusés (403), accès résiduel ≤ TTL access (15 min) — trade-off assumé.
- Routes publiques : `GET /v1/public/:slug` expose id/slug/name/status d'un tenant actif uniquement.
- La SPA (à venir, repo séparé) est servie par le backend : fallback `r.NoRoute` 404 sur `/v1*` et non-GET, sinon `index.html` (PWA single-origin).

## Garanties vécues par le client

- 401 = non authentifié (token expiré/absent/mauvais plan) ; 403 = connecté mais rôle insuffisant ou tenant suspendu.
- Un token admin ne passe jamais les routes tenant, un token tenant ne passe jamais les routes admin (claims secrets distincts, rejet croisé testé).
- Les données d'un tenant sont invisibles aux autres tenants même en cas de bug applicatif : requête sans contexte tenant = 0 ligne, insert cross-tenant = 42501.
