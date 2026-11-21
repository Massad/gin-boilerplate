# Stack Frontend — SPA tenant + console superadmin

* **Author:** mistersly
* **Status:** Proposed (sections approuvées gate par gate ; prêt pour `/egm-review`)
* **Date:** 2026-11-21

> Artefact `egm-draft`. En mode plan de ce harness, les artefacts de session vivent dans `.kilo/plans/`. Source ingérée (Priority 1) : `concepts_brief.md` → `.kilo/plans/1795265473956-concepts-brief-frontend-stack.md`.

--------------------------------------------------------------------------------

## Overview

Une SPA React (TypeScript strict) dans un nouveau répertoire `/app` à la racine du monorepo, servie en single-origin par le binaire Gin existant. L'outil SaaS est intégralement authentifié : aucun SEO, aucun SSR, pas de runtime Node en production. Deux surfaces dans une seule app : l'espace tenant (responsive mobile/tablette/desktop, UI française) et un sous-arbre `/admin` (console superadmin plateforme) en chunk chargé à la demande (lazy) avec canal de token isolé. L'auteur est Lead Front-End, seul développeur : la stack est alignée sur ses compétences et sur l'API `/v1` existante ; la couche backend visée par le pivot cookie est **déjà livrée** (couche auth cookies — vérification uniquement, voir plus bas) ; aucune autre modification backend n'est requise en v1.

## The Problem

Le backend multi-tenant (Go/Gin, PostgreSQL RLS échec fermé, Redis, JWT) est terminé mais n'a aucun frontend : les seules pages HTML existantes sont des gabarits de démo (`public/html/`). Les utilisateurs ne peuvent ni s'inscrire ni utiliser le produit au-delà d'appels API bruts, et le superadmin n'a que 4 routes API nues.

La solution : une SPA React qui consomme l'API `/v1` telle quelle, en traitant en first-class les contraintes réelles du backend : deux canaux de tokens étanches (tenant : access 15 min + refresh 7 j avec rotation atomique ; admin : access 8 h **sans** endpoint de refresh), absence d'endpoint `GET /user/me`, messages backend en anglais que l'UI française ne relaie jamais, et endpoints facture qui sont des démos non authentifiées. Le servage est assuré par le mécanisme déjà en place (fallback `r.NoRoute` → `index.html`) en **fichiers statiques** — pas de `go:embed` (décision d'auteur après revue).

## Goals

* Un visiteur peut s'inscrire avec sa société, se connecter et gérer les articles de son tenant (CRUD complet) — sur mobile, tablette et desktop.
* Le créateur du tenant (rôle `owner` dans le JWT) voit en v1 les mêmes écrans que les membres ; les gardes de rôle (basées claims) sont posées, mais aucune feature owner-only n'est exposée tant que `RequirePermission` n'est câblé sur aucune route (état réel du code).
* Le superadmin se connecte sur `/admin` (login séparé), liste les tenants, bascule leur statut.
* Aucun token de durée longue n'est accessible au JS (détail de la décision : Open Questions ; norme : Non-Functional Requirements). La session survit au reload sans re-login (refresh silencieux via cookie).
* Les messages d'erreur et de validation sont en français, possédés par le front, alignés sur les règles `binding` backend.
* Tests : unitaires (Vitest + Testing Library + MSW) sur composants et client API ; e2e (Playwright) sur backend réel.
* CI : un job Node séparé du job Go, réutilisant les services PG/Redis existants pour l'e2e ; le front casse la CI s'il échoue.
* Observabilité minimale : Error Boundaries + error reporting, sourcemaps privés, version de release injectée au build.

## Non-Goals

* PWA / service workers / offline (Workbox) — différé, architecture non bloquante.
* Impersonation superadmin dans un tenant — prérequis backend (API dédiée + audit).
* Invitations/membres et niveaux d'accès multiples — prérequis backend (API + câblage `RequirePermission`).
* Réglages tenant — aucune API n'existe.
* UI facture — les endpoints `/v1/invoice*` sont des démos non authentifiées (`invoice/handler.go:30,60`).
* i18n multi-langues (FR v1) ; SEO/SSR.
* Toute modification de l'API métier en v1 : la **couche auth cookies** (`login`/`refresh`/`logout`, `AdminAuth`, tests) est livrée et testée ; les autres gaps restent documentés pour v1.5 (`GET /user/me`, câblage RBAC, API invitations).

## Open Questions

Décisions résolues pendant la rédaction (conservées pour mémoire ; à retirer à la finalisation) :

* Error tracking : **SaaS type Sentry** (DSN via `VITE_SENTRY_DSN`, sourcemaps uploadés en CI, release = SHA git).
* Portée httpOnly : **hybride longue durée** — refresh tenant (7 j) et access admin (8 h) en cookies `httpOnly` `SameSite=Strict` scoper par `Path` ; access tenant (15 min) en mémoire Bearer ; aucun endpoint `/me` requis en v1.

## Critical User Journeys (CUJs)

### CUJ 1 : Inscription + première connexion
* **Given** : un visiteur sans compte, formulaire register (email, mot de passe, société).
* **When** : il s'inscrit ; le backend crée user `owner` + tenant (slug) et renvoie user+tenant **sans token** (handler `Register`, `controllers/user.go`).
* **Then** : le front chaîne automatiquement l'appel login avec les identifiants déjà en mémoire ; le backend pose le cookie refresh (httpOnly, `SameSite=Strict`, `Path=/v1/token`) et renvoie l'access token en JSON (mémoire uniquement) ; session établie ; redirection vers l'app.

### CUJ 2 : CRUD articles en session
* **Given** : membre connecté, token access valide.
* **When** : il crée/modifie/supprime un article ; le token access en mémoire expire (401).
* **Then** : le client tenant déclenche le refresh **via le cookie httpOnly** (rotation atomique single-flight : N requêtes 401 concurrentes déclenchent un seul refresh ; nouvel access en mémoire, cookie ré-émis) puis rejoue la requête une fois, de façon invisible ; si le refresh échoue (révocation), redirection vers le login.

### CUJ 3 : Tenant suspendu
* **Given** : utilisateur connecté ; le superadmin suspend son tenant.
* **When** : l'accès résiduel expire (≤ 15 min, trade-off backend assumé) et le front tente un refresh.
* **Then** : l'API refuse le refresh (401 — tenant suspendu rejeté dans `Refresh`, `controllers/auth.go`) et le login (403 « Tenant suspended », `Login`, `controllers/user.go`) ; le front purge la session et affiche l'état « tenant suspendu » ; plus aucun appel ne passe.

### CUJ 4 : Console superadmin
* **Given** : superadmin sur la route `/admin` (chunk lazy, login séparé) ; son access token 8 h vit dans un cookie httpOnly `Path=/v1/admin` (aucun refresh n'existe, le JS ne le voit pas).
* **When** : il liste les tenants et bascule un statut.
* **Then** : liste via `GET /v1/admin/tenants`, toggle via `POST /v1/admin/tenants/:id/status` ; sur 401 (expiration à 8 h), écran de re-login — jamais de boucle de refresh ; après reload, la session est revalidée par un appel admin (401 → login).

### CUJ 5 : Reload de page
* **Given** : utilisateur connecté qui recharge la page ; l'access token était en mémoire (perdu), aucun endpoint `GET /user/me` n'existe.
* **When** : la SPA redémarre.
* **Then** : le client déclenche le refresh silencieux via le cookie httpOnly ; le nouvel access token (claims `user_id`/`tenant_id`/`role` décodables côté client) restaure la session sans re-login ; le profil d'affichage (user/tenant du login) est repris du cache `sessionStorage` — rafraîchi au prochain login.

--------------------------------------------------------------------------------

## Technical Plan (Architecture)

### Vue d'ensemble

```text
┌─────────────────────────── Navigateur ───────────────────────────┐
│  SPA React 19 + TS strict (/app)                                  │
│  ├── Espace tenant : SessionContext + TanStack Query              │
│  │     └── client API tenant (fetch /v1, 401 → refresh → retry ×1)│
│  ├── Sous-arbre /admin (chunk lazy, login séparé)                 │
│  │     └── client API admin (fetch /v1/admin, 401 → re-login)     │
│  ├── Adaptateur TokenStore (localStorage, 2 espaces isolés)       │
│  └── Sentry (Error Boundaries, web-vitals, release SHA)           │
└───────────────────────────┬───────────────────────────────────────┘
                            │ same-origin : /v1 (API) + /assets (hashés)
┌───────────────────────────┴───────────────────────────────────────┐
│  Gin (main.go) — API /v1 inchangée (RLS, JWT, Redis)               │
│  ├── /assets/* → ./app/dist/assets (Cache-Control: immutable)      │
│  ├── GET /* hors /v1 → ./app/dist/index.html (Cache-Control:      │
│  │   no-cache) — racine + NoRoute modifiés                         │
│  └── conservés : LoadHTMLGlob (démo invoice), r.Static /public    │
└─────────────────────────────────────────────────────────────────────┘
```

La SPA consomme les routes `/v1` existantes (articles, public). Les contrats auth ont été modifiés par la couche cookies **livrée au 2026-11-21** (login/refresh/logout, admin login/logout, `AdminAuth` par cookie — voir « Couche auth cookies backend — livrée »). Les manques backend restent documentés en v1.5 (`GET /user/me`, câblage RBAC, invitations).

### Décision tranchée : error tracking = SaaS type Sentry

`@sentry/react` initialisé via `VITE_SENTRY_DSN`, Error Boundary global + par route, release = SHA git injecté au build, sourcemaps générés par Vite et uploadés en CI (sentry-cli, secret `SENTRY_AUTH_TOKEN`), web-vitals reportés. Les options self-hosted (infrastructure à opérer) et différée (métier sans observabilité au lancement) ont été écartées pour v1.

### Couche auth cookies backend — livrée (constat d'existant, vérification uniquement)

* (LIVRÉ) `POST /v1/user/login` : pose le cookie refresh (`httpOnly`, `Secure` si `ENV=PRODUCTION` ou `SSL=TRUE`, `SameSite=Strict`, `Path=/v1/token`) ; réponse `{message, user, tenant, token: {access_token}}` — `refresh_token` absent du JSON.
* (LIVRÉ) `POST /v1/token/refresh` : lit le refresh **dans le cookie** (plus de token dans le body) ; rotation atomique Redis conservée ; réponse = `{access_token}` + cookie ré-émis (nouveau refresh).
* (LIVRÉ) `GET /v1/user/logout` : retire le cookie refresh (Max-Age=0) en plus de la révocation Redis existante.
* (LIVRÉ) `POST /v1/admin/login` / `GET /v1/admin/logout` : posent/retirent le cookie admin access 8 h (`httpOnly`, `SameSite=Strict`, `Path=/v1/admin`) ; admin login répond `{message}` ; `AdminAuth` lit le token **depuis le cookie** (fallback header Authorization).
* CSRF : `SameSite=Strict` + scoping `Path` (le cookie refresh n'est envoyé qu'à `/v1/token/refresh`, le cookie admin qu'aux routes `/v1/admin`) ; aucun jeton CSRF requis ; le middleware CORS existant reste en same-origin. Résiduel connu et assumé en v1 : dégradation navigateur possible en Lax, et un XSS peut émettre des requêtes avec le cookie attaché (sans vol du token) — fenêtre bornée par expiration (7 j / 8 h) ; Alternative 13 écarte le jeton double-submit ; une vérification d'`Origin` sur les mutateurs reste une option v1.5.
* Le cookie refresh ne peut être retiré que par le serveur (logout réussi) ; en cas de purge client (refresh refusé — révocation/suspension), le cookie persiste jusqu'à expiration mais est inutilisable tant que la session Redis est morte.
* (LIVRÉ) Tests backend : `tests/article_test.go`, `controllers/auth_test.go`, `middleware/auth_test.go` — assertions `Set-Cookie`, envoi du header `Cookie` au refresh, absence de `refresh_token` dans les bodies déjà en place.
* Inchangés : `TokenAuth` (Bearer, access 15 min), toutes les routes métier, RLS, Redis (métadonnées de session identiques).

### Composants de la SPA (/app)

* **Client API (`src/lib/api/`)** : un wrapper fetch de base (JSON, erreurs typées, abort) + deux instances :
    * *tenant* : `Authorization: Bearer <access>` depuis le magasin mémoire ; sur 401 → refresh **via le cookie httpOnly** (POST `/v1/token/refresh` sans token dans le body, rotation single-flight) → retry ×1 ; échec de refresh → purge session + redirection login.
    * *admin* : aucun header `Authorization` — le cookie httpOnly `Path=/v1/admin` est attaché automatiquement par le navigateur ; aucun endpoint de refresh n'existe ; 401 → purge + re-login admin, jamais de boucle.
    * Endpoints typés : register, login, logout, refresh, articles CRUD, `public/:slug`, admin login/logout/tenants/status.
* **Magasin de tokens (`src/lib/auth/`)** : l'access tenant vit **en mémoire uniquement** (jamais localStorage) — seul token lisible par le JS, durée 15 min ; les cookies refresh (7 j) et admin (8 h) sont `httpOnly`, invisibles au JS, gérés par le navigateur. Le profil d'affichage (user/tenant du login) est caché en `sessionStorage` — donnée non sensible, jamais de token dedans.
* **Session (`SessionContext`)** : bootstrap = refresh silencieux via cookie → access en mémoire → claims décodés (`user_id`, `tenant_id`, `role` — le serveur reste seul juge de validité) + profil du cache `sessionStorage`. Bootstrap admin = appel admin léger (la liste tenants sert de sonde de session) ; 401 → écran login. Gardes de routes : `ProtectedRoute` (tenant) et `ProtectedAdminRoute`.
* **Routage** : React Router v7 (`createBrowserRouter`), `lazy()` par route ; `/admin/*` = un chunk lazy distinct ; la navigation tenant ne charge jamais le code admin.
* **Données** : TanStack Query v5 — queryKeys normalisés (`['tenant','articles']`, `['admin','tenants']`), mutations avec invalidation ciblée, pas de retry automatique sur 401/403 (le client gère).
* **Formulaires** : react-hook-form + zod, schemas alignés sur les tags `binding` des `forms/*.go`, messages FR possédés par le front ; les réponses de validation backend ne sont jamais affichées telles quelles.
* **UI** : Tailwind v4 + shadcn/ui, responsive 3 facteurs ; mapping sémantique des erreurs : 401 → session, 403 → suspendu/droits, autres 4xx → validation, 5xx → générique.
* **Observabilité** : Sentry + Error Boundaries + version de release + web-vitals.

### Servage — en place dans `main.go` (lignes 76-214)

* (LIVRÉ) Racine (`r.GET("/")`) et fallback `NoRoute` servent `./app/dist/index.html` s'il existe (`os.Stat`), avec `Cache-Control: no-cache`, sinon fallback template démo. Le comportement API est préservé : `/v1*` ou non-GET → JSON 404.
* (LIVRÉ) Servage des assets hashés (`/assets` → `./app/dist/assets`) avec `Cache-Control: max-age=31536000, immutable`.
* Conservés en l'état : `r.LoadHTMLGlob("./public/html/*")` (le template `invoice.html` reste requis par la démo invoice), `r.Static("/public", "./public")`, toutes les routes `/v1`.

### Dev & CI

* **Dev** : `make dev` lance en parallèle le backend Go (PORT par défaut) et `pnpm --dir app dev` (Vite :5173, proxy serve-side `/v1` → backend) — pas de CORS nécessaire ni en dev (proxy) ni en prod (same-origin). Toutes les URL API sont relatives (`/v1/...`). Les `Set-Cookie` traversent le proxy (dev same-origin) ; le flag `Secure` des cookies suit `ENV`/`SSL` pour rester posable sur localhost en dev.
* **Build** : `make build_spa` → `pnpm --dir app build` → `app/dist` (gitignoré).
* **CI (`.github/workflows/ci.yml`)** : nouveau job `frontend` isolé du job Go : pnpm install (fichier-lock commité) → lint ESLint + Prettier → `tsc --noEmit` → Vitest (MSW) → build → Playwright e2e sur backend réel (réutilise les services PostgreSQL/Redis et le `.env` applicatif existants, rôles `app_user`/`app_admin`) avec cache des navigateurs ; e2e isolés par run : tenant dédié au run (suffixe unique) purgé en teardown (succès comme échec) ; upload sourcemaps Sentry sur `main`. Le job Go n'est pas modifié ; la CI échoue si le front échoue.

--------------------------------------------------------------------------------

## Alternatives Considered

### Alternative 1 : Next.js / Remix (rendu serveur)
* **Pros** : SSR/pré-rendu prêts, routing et data-fetching intégrés, écosystème riche.
* **Cons** : runtime Node en production, servage incompatible avec le single-origin servi par le binaire Go, complexité SSR inutile sans SEO.
* **Reason for Rejection** : Rejected by author : aucun SEO (outil intégralement authentifié), le backend Go doit rester le seul serveur — un SSR imposerait un second runtime sans bénéfice produit.

### Alternative 2 : HTMX / Alpine.js + templating Gin
* **Pros** : outillage front quasi nul, réutilise `LoadHTMLGlob` et `r.Static` existants, bundle minimal.
* **Cons** : pas d'architecture SPA (PWA future impossible), pas d'écosystème composants, l'avantage Lead FE de l'auteur est perdu.
* **Reason for Rejection** : Rejected by author : l'écosystème React est une décision d'auteur et la PWA différée exige une vraie SPA ; HTMX ferme la direction produit.

### Alternative 3 : TanStack Router
* **Pros** : routes typées de bout en bout, search params typés.
* **Cons** : écosystème plus jeune, le typage des routes apporte peu sur une SPA à deux sous-arbres.
* **Reason for Rejection** : Rejected by author : bénéfice marginal face à la maturité de React Router ; le lazy `/admin` fonctionne avec les deux.

### Alternative 4 : Redux Toolkit (et Zustand en v1)
* **Pros** : DevTools éprouvés, conventions d'équipe, middlewares.
* **Cons** : boilerplate pour un dev solo, surdimensionné face à TanStack Query ; Zustand lui-même est superflu tant que la session tient dans un Context.
* **Reason for Rejection** : Rejected by author : Context + TanStack Query couvrent v1 ; Redux et Zustand sont reportés à un besoin réel.

### Alternative 5 : MUI / Ant Design
* **Pros** : bibliothèque complète, thème prêt, AntD éprouvé pour les backoffices.
* **Cons** : verrouillage design, styles opaques difficiles à personnaliser, dépendance versionnée.
* **Reason for Rejection** : Rejected by author : modèle shadcn (composants possédés, primitives Radix accessibles) préféré ; le fardeau de maintenance des composants copiés est assumé.

### Alternative 6 : Headless UI + CSS maison
* **Pros** : bundle minimal, contrôle total du rendu.
* **Cons** : tout le styling et l'accessibilité restent à écrire et maintenir seul.
* **Reason for Rejection** : Rejected by author : shadcn fournit des primitives Radix accessibles déjà stylées Tailwind — coût d'écriture disproportionné pour un solo lead.

### Alternative 7 : client API généré depuis la doc swag
* **Pros** : régénération au fil de l'API, types dérivés du contrat, zéro drift manuel.
* **Cons** : schémas swag bruités pour ~15 endpoints, étape de génération en CI, code généré à maintenir.
* **Reason for Rejection** : Rejected for v1 : un client manuel typé (~100 lignes) coûte moins que la chaîne de génération ; à réévaluer si la surface API croît.

### Alternative 8 : Biome (lint + format en un outil)
* **Pros** : un seul outil, rapide, pas de config croisée.
* **Cons** : écosystème de règles plus mince, typescript-eslint et react-hooks plus profonds.
* **Reason for Rejection** : Rejected by author : la profondeur de l'écosystème ESLint (typescript-eslint, `react-hooks/exhaustive-deps`) prime sur la légèreté.

### Alternative 9 : `go:embed` du build dans le binaire
* **Pros** : déploiement en un seul artefact, assets garantis cohérents avec le binaire.
* **Cons** : chaque release UI redéploie le binaire Go, la CI Go dépend du build Node, la taille du binaire croît avec les assets.
* **Reason for Rejection** : Rejected by author (après revue Goldfish) : servage en fichiers statiques via le fallback `NoRoute` existant — déploiements front/back découplés, CI Go intact ; embed réévalué si un besoin de binaire unique apparaît.

### Alternative 10 : build vers `./public/assets` (réutiliser `r.Static`)
* **Pros** : zéro modification de `main.go` au-delà de la cible de build.
* **Cons** : mélange le build SPA avec les gabarits de démo servis par `/public`, brouille la frontière démo/produit, headers de cache moins contrôlables.
* **Reason for Rejection** : Rejected by author : `app/dist` garde le build isolé de `public/` et permet des headers dédiés (hashés `immutable`, `index.html` `no-cache`).

### Alternative 11 : 100 % cookies httpOnly (access tenant inclus)
* **Pros** : zéro token lisible par le JS, posture de sécurité maximale.
* **Cons** : `TokenAuth` ET `AdminAuth` à modifier, le client ne peut plus décoder les claims → `GET /user/me` ET `/admin/me` requis en v1, CSRF à couvrir sur toutes les routes.
* **Reason for Rejection** : Rejected by author : portée backend disproportionnée pour un gain marginal — l'access hybride (15 min, mémoire) borne déjà la fenêtre XSS ; les endpoints `/me` arrivent en v1.5 avec le reste. Trade-off assumé explicitement : l'hybride déplace la complexité vers le client (boucle 401→refresh single-flight, retry ×1, purge de session, CUJ 2 et 5) — l'empreinte backend minimale prime en v1.

### Alternative 12 : refresh seul en cookie (admin en Bearer JS)
* **Pros** : changements backend minimaux (`AdminAuth` intact).
* **Cons** : laisse un access admin de 8 h lisible par le JS — la fenêtre d'exposition XSS la plus longue du design (access tenant borné à 15 min en mémoire).
* **Reason for Rejection** : Rejected by author : le principe httpOnly doit couvrir toutes les créances de durée longue ; le coût additionnel (lire le cookie dans `AdminAuth`) est faible.

### Alternative 13 : jetons CSRF double-submit
* **Pros** : défense CSRF indépendante du comportement navigateur.
* **Cons** : état supplémentaire par session, complexité client/serveur, redondant avec `SameSite=Strict`.
* **Reason for Rejection** : Rejected by author : `SameSite=Strict` + scoping `Path` suffisent (le cookie refresh n'est envoyé qu'à `/v1/token/refresh`) ; pas de jeton CSRF en v1.

### Alternative 14 : tokens en localStorage (conception initiale pré-pivot)
* **Pros** : simplicité maximale, zéro changement backend, survie au reload directe.
* **Cons** : un refresh de 7 j lisible par le JS — vol possible par XSS, contradiction frontale avec le principe httpOnly.
* **Reason for Rejection** : Rejected by author (pivot du 2026-11-21) : le principe httpOnly est installé dès la v1 ; seul le profil d'affichage est caché (`sessionStorage`, non sensible, jamais de token).

### Alternative 15 : repo séparé pour la SPA
* **Pros** : CI et versioning front indépendants, pas de couplage au repo Go.
* **Cons** : deux pipelines, deux clones, coordination de versions front/back, contrainte single-origin mal servie.
* **Reason for Rejection** : Rejected by author : monorepo `/app` décidé — un seul clone/CI, la séparation vient des jobs CI distincts.

--------------------------------------------------------------------------------

## Non-Functional Requirements

* **Sécurité** : cookies `httpOnly` + `SameSite=Strict` + `Secure` en prod, scopés par `Path` (`/v1/token` pour le refresh, `/v1/admin` pour l'admin) ; `refresh_token` absent de tout JSON ; aucun token en localStorage ; profil d'affichage seul en `sessionStorage` ; CSRF couvert par `SameSite` + `Path` sans jeton dédié.
* **Cache/perf** : assets hashés en `max-age=31536000, immutable`, `index.html` en `no-cache` ; chunk lazy `/admin` hors bundle tenant ; TanStack Query pour la déduplication.
* **Rollout/rollback** : le servage SPA garde un fallback vers le template démo si `app/dist` est absent — un binaire sans build front reste fonctionnel (le smoke test Go CI repose sur ce fallback) ; rollback front = resservir le `dist` précédent sans toucher au binaire.
* **Déploiement** : binaire unique servant l'API + les fichiers statiques ; pas de proxy inverse ni CDN dans l'infra visée en v1 (aucune config de déploiement dans le dépôt). Charge statique assumée pour le trafic visé ; alternative (proxy/CDN dédié pour décharger le Go) réévaluée si la charge ou la réplication l'exige.

--------------------------------------------------------------------------------

## Detailed Implementation Plan

Ordre d'exécution recommandé : la couche auth backend est **livrée** (items 1-11 : vérification uniquement), puis le scaffolding `/app` (items 12-18), les features front (items 19-32), enfin l'intégration CI et les docs (items 33-38).

### Target Files to Create/Modify

#### Backend — couche auth cookies (LIVRÉE — vérification uniquement)

1. `./models/auth.go` (LIVRÉ)
   * Type / Scope : session/tokens Go + helpers cookies.
   * Proposed Changes : sans action — en place : `RefreshCookieName`/`AdminCookieName` (`models/auth.go:21,25`), `sessionCookie` (httpOnly, `SameSite=Strict`, `Path` scopé, `Secure` via `cookieSecure()` = `ENV=PRODUCTION` **ou** `SSL=TRUE`, `models/auth.go:36-51`), `SetRefreshCookie`/`ClearRefreshCookie`/`SetAdminCookie`/`ClearAdminCookie` (`models/auth.go:55-74`), `ExtractAdminTokenMetadata` lit le cookie (`models/auth.go:365-394`). `CreateToken`/`CreateAuth`/rotation Redis inchangés.

2. `./controllers/user.go` (LIVRÉ)
   * Proposed Changes : sans action — en place : `Login` pose le cookie refresh (`user.go:68`) et répond `{message, user, tenant, token: {access_token}}` via `models.LoginToken` (`user.go:69-74`) ; `Register` inchangé (user+tenant, sans token — le front chaîne login) ; `Logout` retire le cookie (`user.go:135`).

3. `./controllers/auth.go` (LIVRÉ)
   * Proposed Changes : sans action — en place : `Refresh` lit le refresh token **dans le cookie** (`auth.go:31-35`), garde la rotation atomique `ConsumeRefreshAuth` + relecture `OneWithTenant` (401 si suspendu — inchangé), répond `{access_token}` (`auth.go:102-105`) et ré-émet le cookie (`auth.go:104`).

4. `./controllers/admin.go` (LIVRÉ)
   * Proposed Changes : sans action — en place : `Login` pose le cookie admin et répond `{message}` (`admin.go:55-56`) ; `Logout` retire le cookie (`admin.go:142`). `Tenants`/`SetTenantStatus` inchangés.

5. `./middleware/auth.go` (LIVRÉ)
   * Proposed Changes : sans action — en place : `AdminAuth` lit le token depuis le cookie (implémentation `models/auth.go:365-394`) ; rejet croisé tenant/admin inchangé (secrets distincts). `TokenAuth` inchangé (Bearer 15 min).

6. ~~`./forms/auth.go`~~ (aucune action) — **cible corrigée** : `forms/` ne contient pas de `auth.go` (`validator.go`, `errors.go`, `admin.go`, `user.go`, `article.go`) ; le struct `Token` (refresh en body, tag `refresh_token`) vit dans `models/auth.go:114-118` et reste retourné par `UserModel.Login` (`models/user.go:67`) : **aucune suppression**.

7. `./main.go` (LIVRÉ)
   * Proposed Changes : sans action — en place (`main.go:76-214`) : `serveSPAIndex` (`os.Stat`, fallback template démo), servage `/assets` → `./app/dist/assets` (`max-age=31536000, immutable`), `r.GET("/")`, `NoRoute` (JSON 404 `/v1`/non-GET), `index.html` en `Cache-Control: no-cache`, `LoadHTMLGlob` et `r.Static("/public")` conservés (démo invoice).

8. `./tests/article_test.go` (LIVRÉ) — helpers cookies et assertions en place (`article_test.go:144,364,789,834`).
9. `./controllers/auth_test.go` (LIVRÉ) — refresh par cookie : rotation, rejet sans cookie, re-use, suspendu → 401 (`auth_test.go:96,140,148-181`).
10. `./middleware/auth_test.go` (LIVRÉ) — `AdminAuth` par cookie, rejet croisé, logout admin retire le cookie (`middleware/auth_test.go:214,243`).
11. `./docs/` (déjà à jour) — annotations swagger en place (`AccessTokenResponse`, `MessageResponse`, `UserLoginResponse`) ; `make generate_docs` + vérif `git diff` uniquement si des annotations sont modifiées.

#### Frontend — nouveau répertoire `/app` (Create)

12. `./app/package.json` + `./app/pnpm-lock.yaml` (Create) — deps : react, react-dom, react-router, @tanstack/react-query, react-hook-form, @hookform/resolvers, zod, tailwindcss, shadcn/ui (radix), @sentry/react, web-vitals ; devDeps : vite, @vitejs/plugin-react, typescript, vitest, @testing-library/react, msw, @playwright/test, eslint (flat) + plugins react-hooks/typescript, prettier. Fichier-lock commité ; versions fixées : Node LTS + `packageManager` pnpm (Corepack) dans `package.json`.
13. `./app/vite.config.ts` (Create) — plugin React ; proxy serve-side `/v1` → `http://localhost:8080` (les `Set-Cookie` traversent sans CORS) ; `build.outDir dist`, sourcemaps en prod, `define` pour `__APP_VERSION__` (SHA git).
14. `./app/tsconfig.json` (Create) — strict, alias `@/*` → `src/*`, `noUncheckedIndexedAccess`.
15. `./app/index.html` (Create) — shell SPA, `lang="fr"`, montage `#root`.
16. `./app/eslint.config.js` + `./app/.prettierrc` (Create) — flat config, règles react-hooks/exhaustive-deps actives.
17. `./app/playwright.config.ts` (Create) — `webServer` : binaire Go avec `app/dist` servi (base URL env) ; projets desktop + viewport mobile ; échecs tracés.
18. `./app/src/main.tsx` (Create) — init Sentry (`VITE_SENTRY_DSN`), `QueryClientProvider`, `RouterProvider`, toaster, ErrorBoundary global.
19. `./app/src/App.tsx` (Create) — routes : `/login`, `/register`, `/` (layout tenant protégé), `/admin/*` en `lazy()` (chunk séparé), `*` → NotFound.
20. `./app/src/lib/api/client.ts` (Create) — `ApiError` typée ; `createTenantClient` : Bearer depuis le magasin mémoire, 401 → refresh par cookie **single-flight** → retry ×1 → sinon purge+login ; `createAdminClient` : sans Authorization (cookie `Path=/v1/admin` attaché par le navigateur), 401 → re-login admin, jamais de boucle ; bootstrap : refresh silencieux au démarrage si mémoire vide.
21. `./app/src/lib/api/endpoints.ts` + `./app/src/lib/api/types.ts` (Create) — endpoints typés (register/login/logout/refresh, articles CRUD, `public/:slug`, admin login/logout/tenants/status) ; types alignés sur le **contrat backend livré** (figé ici d'après les contrôleurs actuels) : login `{message, user, tenant, token: {access_token}}`, register `{user, tenant}`, refresh `{access_token}`, articles create `{message, id}` (id **string** UUID), listes `{results}` / détail `{data}`, update/delete `{message}`, admin login/logout `{message}` ; tenants list/statut : figer depuis `controllers/admin.go`.
22. `./app/src/lib/auth/claims.ts` (Create) — décodage base64 des claims access (`user_id`, `tenant_id`, `role`) sans vérification crypto (le serveur reste seul juge).
23. `./app/src/lib/auth/session.tsx` (Create) — `SessionContext` + bootstrap (refresh silencieux → claims → profil `sessionStorage`) + `ProtectedRoute` / `ProtectedAdminRoute` (garde admin = sonde `GET /v1/admin/tenants`, 401 → login).
24. `./app/src/lib/validation/schemas.ts` + `./app/src/lib/validation/messages.ts` (Create) — zod mirroir des tags `binding` : login (email, password 3-50), register (name 3-20 sans chiffres/caractères spéciaux, email, password 3-50, company 3-60), article (title 3-100, content 3-1000), admin status ; messages **FR possédés** — jamais de relais des chaînes backend.
25. `./app/src/features/auth/` (Create) — `LoginForm`, `RegisterForm` (chaînage register→login), logout (appel API + purge mémoire + `sessionStorage`).
26. `./app/src/features/articles/` (Create) — liste/création/édition/suppression via TanStack Query (queryKeys `['tenant','articles']`, invalidation ciblée), formulaires RHF+zod.
27. `./app/src/admin/` (Create) — `AdminLogin`, console tenants (liste + toggle status avec confirmation), sources isolées du bundle tenant.
28. `./app/src/lib/observability/sentry.ts` (Create) — init, `ErrorBoundary`, capture web-vitals, tag release/tenant_id non-invasive.
29. `./app/src/components/` (Create) — composants shadcn générés + composants app (layout responsive 3 facteurs, EmptyState, ErrorState, mapping 401/403/406 (violation d'unicité côté backend)/5xx).
30. `./app/src/test/` (Create) — setup (jest-dom, MSW server), handlers MSW couvrant login/register/refresh/articles/admin + scénarios 401/403, fixtures.
31. `./app/src/lib/api/client.test.ts` + tests composants (Create) — single-flight du refresh (N appels 401 concurrents → 1 refresh), retry ×1, admin 401 → re-login, guards, validation FR.
32. `./app/tests/e2e/` (Create) — specs Playwright par CUJ : register→login→app ; CRUD article + refresh silencieux (cookie httpOnly présent dans DevTools, zéro localStorage) ; tenant suspendu (console admin → 403/401 attendus) ; admin login→tenants→toggle ; reload sans re-login ; isolation : tenant dédié par run (suffixe unique) + purge en teardown.

#### Intégration monorepo

33. `./Makefile` (Modify) — cibles `dev` (Go + Vite en parallèle, arrêt groupé), `build_spa` (pnpm install --frozen-lockfile + build → `app/dist`), `test_spa` (vitest run), `e2e_spa` (playwright test) ; cibles existantes inchangées.
34. `./.gitignore` (Modify) — `/app/node_modules/`, `/app/dist/`, `/app/coverage/`, `playwright-report/`, `test-results/`.
35. `./.github/workflows/ci.yml` (Modify) — nouveau job `frontend` (setup pnpm + Node LTS, cache pnpm + navigateurs Playwright, install --frozen-lockfile, lint, `tsc --noEmit`, vitest, build, e2e sur backend réel en réutilisant les services PostgreSQL/Redis et la génération de secrets existantes, upload sourcemaps Sentry sur `main` via `SENTRY_AUTH_TOKEN`) ; job Go **inchangé** — le fallback template permet d'exécuter le smoke test Go sans build Node ; échec du front = échec de la CI.
36. `./memory-bank/projectbrief.md`, `./memory-bank/productContext.md`, `./memory-bank/activeContext.md` (Modify) — cookie httpOnly livré en v1 (inversion du report), SPA monorepo `/app` (« repo séparé » obsolète), état de session courant.
37. `./README.md` (Modify) — section front : prérequis (Node LTS, pnpm), `make dev`, `make build_spa`, servage single-origin, cookies httpOnly ; **corriger aussi la section Authentication (README L220-236)** qui décrit encore des tokens en JSON (refresh en body, `stored in Redis`).
38. `./AGENTS.md` (Modify) — commandes front (`make dev/build_spa/test_spa/e2e_spa`), structure `/app`, conventions front (messages FR possédés, jamais de relais backend ; SQL inchangé).

### Verification Plan

* **Backend** : `go build -mod=readonly .` ; `go vet -mod=readonly ./...` ; `make test` complet (`-race -tags=all`, PG+Redis, `.env` exporté) — les tests auth/cookies doivent être verts, y compris rejet croisé et suspendu.
* **Front — étape 0 (préalable)** : items 12-18 réalisés (scaffolding), versions fixées (Node LTS, pnpm `packageManager`). Puis : `pnpm --dir app lint` ; `tsc --noEmit` ; `pnpm --dir app test` (Vitest) ; `pnpm --dir app build` ; `make e2e_spa` sur backend réel.
* **CI** : job Go vert **sans** Node (smoke test racine = fallback template) ; job frontend vert de bout en bout.
* **Vérification manuelle de sécurité** : `make dev` → DevTools → Application > Cookies : cookie `refresh` présent et `httpOnly` (`Secure` en prod, `Path=/v1/token`), cookie `admin` présent et `httpOnly` (`Path=/v1/admin`) ; aucune clé de token dans `localStorage` ; suspension d'un tenant via la console admin → CUJ 3 vérifié dans l'onglet tenant.

--------------------------------------------------------------------------------

### Future Improvements

* **v1.5 (prérequis backend)** : `GET /user/me` (profil complet au bootstrap), câblage `RequirePermission` + invitations/membres, réévaluation de la génération de types client si la surface API croît.
* **v2** : PWA/Workbox (service workers, offline), impersonation superadmin (API backend dédiée + traçabilité `admin_audit`), réglages tenant, UI de lecture d'audit.

--------------------------------------------------------------------------------

*Drafted with EGM*
