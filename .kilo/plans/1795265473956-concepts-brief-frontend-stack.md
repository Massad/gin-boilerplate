# Stack Frontend — SPA tenant + console superadmin (go-artisan-services)

* **Author:** mistersly
* **Date:** 2026-11-21
* **Note** : en mode plan de ce harnes, les artefacts de session vivent dans `.kilo/plans/` — ce fichier tient lieu de `concepts_brief.md`.

## Overview

Outil SaaS multi-tenant entièrement authentifié (aucun enjeu SEO) : une SPA React servie en single-origin par le binaire Go existant, dans un nouveau répertoire `/app` à la racine du repo (monorepo, pas de repo séparé). Deux surfaces dans une seule app : l'espace tenant (usage mobile, desktop, tablette) et un sous-arbre `/admin` pour le superadmin plateforme. UI française en v1. Tests unitaires et e2e obligatoires. L'auteur est Lead Front-End, seul développeur.

Le backend est un plan multi-tenant RLS terminé (PostgreSQL fail-closed, Redis, JWT) : la stack front doit consommer l'API `/v1` existante sans la modifier en v1.

## Analysis

### Faits d'API vérifiés dans le code (cadrent le client)

- **Login** renvoie `{message, user, tenant, token}` (`controllers/user.go:65`). **Register** renvoie `user`+`tenant` **sans token** (`controllers/user.go:100`) : il faut se connecter après inscription.
- **Refresh** renvoie uniquement `{access_token, refresh_token}` (`controllers/auth.go:105-109`) — jamais de données user/tenant.
- **Aucun endpoint `GET /user/me`** : après un reload, l'identité de session doit être reconstruite en décodant les claims du JWT (`user_id`, `tenant_id`, `role` — `middleware/auth.go:51-53`) ; le profil complet n'est connu qu'au login. Gap backend documenté (candidat v1.5 backend).
- **Canal tenant** : access 15 min, refresh 7 jours (`models/auth.go:74-77`), rotation atomique Redis — le client porte une boucle 401 → refresh → retry.
- **Canal admin** : access **8 h** (`models/auth.go:281`), **aucun endpoint de refresh admin** — le client admin n'a PAS de boucle de refresh ; 401 → écran de re-login. (Corrige l'hypothèse initiale qui supposait un mécanisme centralisé aux deux canaux.)
- **Routes invoice** (`/v1/invoice`, `/v1/invoice/download`) : démo non authentifiée, `SampleInvoice()` codé en dur (`invoice/handler.go:30,60`) — aucune UI facture en v1.
- **Messages backend en anglais** (validation `forms.ArticleMessages` et messages 401/403 middleware) : l'UI FR doit posséder ses propres messages FR côté client ; on ne relaie jamais les chaînes backend.
- **RBAC backend** : `owner` = `members:manage` + `tenant:manage`, `member` = rien (`middleware/auth.go:13-22`) ; aucune route `RequirePermission` câblée ; aucune API invitations ni réglages tenant.
- **Console admin existante (API)** : login, logout, liste tenants, toggle status. Pas de route de lecture d'audit (la table `admin_audit` existe).

### Contradictions retenues (lentille inventive) et résolutions

1. **Sécurité des tokens vs simplicité** : RÉSOLUE en v1 (décision d'auteur du 2026-11-21, inversant le report initial) : refresh tenant (7 j) et access admin (8 h) en cookies `httpOnly` `SameSite=Strict` scopés par `Path` ; seul l'access tenant (15 min) reste en JS (mémoire, Bearer). Aucun token en localStorage ; CSRF couvert par `SameSite=Strict` + scoping, sans jeton CSRF.
2. **Monorepo simple vs CI hétérogène Go + Node** : résolu par segmentation — jobs CI séparés, conteneurs Node éphémères, aucun outillage Node ne fuite dans le job Go.
3. **Une SPA vs isolation du canal admin** : résolu par segmentation (chunk lazy `/admin`) + extraction (modules auth admin chargés à la demande). Rejet croisé backend déjà testé = filet de sécurité.
4. **UI anticipée vs APIs inexistantes** : résolu par contract-first — aucun scaffolding UI sans contrat d'API stabilisé ; feature flags au moment venu.

### Exigences issues des revues (obligatoires pour le draft)

- **Observabilité front** : Error Boundaries + error tracking (outil à décider), sourcemaps upload privé, version de release injectée au build, web-vitals. Absent de l'hypothèse initiale — ajouté.
- **Cache HTTP** : assets hashés `max-age=31536000, immutable`, `index.html` en `no-cache`.
- **CI** : job Node isolé du job Go, services PG/Redis existants réutilisés pour l'e2e, cache des navigateurs Playwright, le front casse la CI s'il échoue.
- **DX** : commande unique de dev (`make dev`) lançant Go + Vite (proxy `/v1` → `:8080`) ; automatiser build front + servage via Makefile.
- **Validation** : schemas zod alignés sur les tags `binding` backend (`required,min=3,max=100`…), messages FR possédés par le front ; duplication de règles assumée et documentée.
- **Client API** : deux instances (tenant / admin), boucle de refresh uniquement côté tenant (via cookie httpOnly), access tenant en mémoire, aucun token en localStorage.

### Points tranchés (revue acceptée par l'auteur, 2026-11-21)

1. **go:embed vs fichiers statiques** : **fichiers statiques** — `app/dist` servi par Gin via le mécanisme existant (fallback `NoRoute` + headers de cache). Déploiements front/back découplés, pas de rebuild binaire par release UI, CI Go indépendant du build Node. `go:embed` écarté jusqu'à un besoin réel de binaire unique.
2. **Zustand** : retiré de v1 — Context pour la session, TanStack Query pour l'état serveur ; ajout sur un besoin réel uniquement.
3. **Client généré depuis swagger** : client manuel typé en v1 (~15 endpoints) ; génération de types à réévaluer si la surface API croît.
4. **Cookie `httpOnly` en v1** : décision d'auteur du 2026-11-21 inversant le report initial — portée « hybride longue durée » : refresh tenant (7 j) et access admin (8 h) en cookies `httpOnly` `SameSite=Strict` scopés par `Path` ; access tenant (15 min) en mémoire Bearer ; pas de jetons CSRF (`SameSite=Strict` + `Path`) ; aucun endpoint `/me` requis en v1. Modifications backend limitées à la couche auth (login/refresh/logout, `AdminAuth`, tests).

## Goals vs. Non-Goals

* **In-Scope Goals (v1) :**
    * App tenant : login/register (+ login automatique après register), CRUD articles, session + gardes de rôle basées sur les claims (`owner`/`member`), logout.
    * Console superadmin (sous-arbre `/admin`) : login admin, liste tenants, toggle status.
    * Client API typé à deux instances, auth cookies httpOnly hybride (modifications backend couche auth incluses), boucle de refresh tenant via cookie.
    * Messages FR côté client, validation zod + react-hook-form alignée sur les règles backend.
    * Tests : unitaires (Vitest + Testing Library + MSW) et e2e (Playwright) sur backend réel.
    * Intégration monorepo : `/app`, CI Node séparé, servage Gin single-origin, `make dev`.
    * Observabilité front minimale : error boundary + reporting, version de release, sourcemaps privés.
* **Out-of-Scope Non-Goals (v1) :**
    * PWA / service workers / offline (Workbox) — différé, architecture non bloquante.
    * Impersonation superadmin dans un tenant — prérequis backend : API d'impersonation + traçabilité audit.
    * Invitations/membres + niveaux d'accès multiples — prérequis backend : API + câblage `RequirePermission`.
    * Réglages tenant — aucune API.
    * UI facture — endpoints démo non authentifiés.
    * i18n multi-langues — FR v1.
    * SEO/SSR — produit intégralement authentifié.

## Alternatives Weighed

* **Next.js / Remix (SSR)** — Rejected by author : aucun SEO (outil authentifié), single-origin servi par le binaire Go, un runtime Node en prod n'apporte rien.
* **HTMX/Alpine + templating Go** — Rejected by author : l'écosystème React est une décision d'auteur (Lead FE), et l'architecture SPA est requise pour la PWA future.
* **TanStack Router** — Rejected by author : bénéfice de typage des routes marginal face à la maturité de l'écosystème React Router ; le lazy `/admin` fonctionne avec les deux.
* **Redux Toolkit** — Rejected by author : surdimensionné pour un solo lead ; Zustand lui-même est reporté (Context en v1).
* **MUI / Ant Design** — Rejected by author : verrouillage design vs modèle d'appartenance shadcn (le fardeau de maintenance des composants copiés est assumé).
* **Client API généré depuis la doc swag** — Rejected for v1 : schémas swag bruités pour ~15 endpoints, un client manuel typé coûte moins ; génération à réévaluer si la surface croît.
* **Biome** — Rejected : écosystème ESLint plus profond (typescript-eslint, react-hooks), malgré un outil unique plus léger.
* **go:embed du build dans le binaire** — Rejected by author (après revue Goldfish) : couplage des releases front/back, taille du binaire, CI Go dépendant du build Node ; le servage en fichiers statiques via le fallback `NoRoute` existant suffit et découple les déploiements.
* **Repo séparé** (plan initial du memory bank) — Rejected by author : monorepo `/app` décidé, memory bank à mettre à jour.

## Key Decisions & Staging for Next Steps

* **Stack v1** : TypeScript strict, React 19, Vite 7, React Router v7 (lazy routes), TanStack Query v5, zod + react-hook-form, Tailwind CSS v4 + shadcn/ui, Vitest + Testing Library + MSW, Playwright, ESLint flat + Prettier, pnpm. **Zustand retiré de v1** (Context pour la session).
* **Client API** : deux instances — tenant (Bearer access 15 min en mémoire ; 401 → refresh via cookie httpOnly single-flight → retry ×1) et admin (cookie httpOnly `Path=/v1/admin`, pas de refresh, 401 → re-login). Aucun token en localStorage.
* **Bootstrap de session** : reload → refresh silencieux via cookie → access en mémoire → claims décodés + profil caché en `sessionStorage` ; bootstrap admin par sonde (liste tenants, 401 → login). Gap `GET /user/me` documenté pour v1.5 (rafraîchissement du profil complet).
* **Messages** : FR possédés par le front, alignés sur les règles `binding` backend ; jamais de relais des chaînes backend anglaises.
* **Structure `/app`** : une seule SPA, sources admin séparées (`src/admin/`), chunk lazy, login séparé.
* **Servage prod** : build Vite vers `app/dist`, servi par Gin en fichiers statiques (fallback `NoRoute` + headers cache : hashés `immutable`, `index.html` `no-cache`). Pas de `go:embed`.
* **Staging** :
    * v1 : auth (cookies httpOnly hybride, modifications backend couche auth) + articles + console tenants + observabilité minimale + CI/dx.
    * v1.5 (prérequis backend) : `GET /user/me`, câblage `RequirePermission`, API invitations.
    * v2 : PWA/Workbox, impersonation superadmin (API dédiée), réglages tenant, UI audit.
* **Décision ouverte restante pour egm-draft** : outil d'error tracking (SaaS type Sentry vs self-hosted vs différé en v1) — tout le reste est tranché.
