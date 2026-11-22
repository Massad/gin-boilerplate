# Plan — Fix intégral de l'erreur « accès public » (principe : rien de public)

## Contexte et décisions

- Principe retenu par l'utilisateur : **aucun accès non identifié aux outils du SaaS**. Seuls restent non authentifiés : les endpoints d'identification (`POST /v1/user/login`, `POST /v1/user/register`, `POST /v1/token/refresh`, `POST /v1/admin/login`) et le shell SPA (`/public` statique, `/assets`, `/` fallback `r.NoRoute` — l'app doit se charger avant auth ; le smoke CI repose sur ce fallback).
- Décision de périmètre (utilisateur) : **principe intégral** — le plan couvre les trois écarts audités :
  1. `GET /v1/public/:slug` — lookup public du répertoire tenants (l'erreur d'origine) ;
  2. `GET /v1/invoice` + `GET /v1/invoice/download` (main.go:168-169) — démo maroto non authentifiée (données statiques `SampleInvoice()`, aucune DB, mais outil accessible sans identification) ;
  3. `GET /swagger/*any` (main.go:188) — UI Swagger servie sans condition, même en `ENV=PRODUCTION`.
- La spec multi-tenant reste portée par le plan existant `.kilo/plans/1795282638021-spec-multi-tenant-rls.md` (à exécuter après ce fix).

## Tâches (ordre)

### T1 — Suppression du lookup public `GET /v1/public/:slug`

1. `main.go:182-185` — retirer le bloc `START PUBLIC` (`pub := new(controllers.PublicController)` + `v1.GET("/public/:slug", pub.One)`).
2. Supprimer `controllers/public.go` (controller + `publicTenantModel`).
3. `models/tenant.go:152-159` — retirer `OneBySlug` (seul appelant = public.go). **Garder** `Slugify`, `ValidSlug`, `reservedSlugs` (utilisés à l'enregistrement ; défensifs contre les collisions de routes).
4. `db/queries/tenants.sql:6-7` — retirer la requête `GetTenantBySlug`.
5. `tests/article_test.go` — retirer l'enregistrement de la route (~lignes 84-86) et le bloc `TestPublicTenant` (~lignes 937-962). **Garder** `tenantSlugB` (utilisé par le test admin tenants, ligne 848) et `LoadHTMLGlob`.
6. `README.md:178-181` — supprimer la section `### Public`.

### T2 — Suppression des routes invoice démo (l'outil redeviendra une vraie feature tenant-scoped)

1. `main.go:165-169` — retirer le bloc `START Invoice` (routes + `inv := new(invoice.InvoiceController)`).
2. Supprimer `invoice/handler.go` (code mort une fois les routes retirées ; ses annotations Swagger disparaissent).
3. `invoice/invoice_test.go:115-130` — supprimer `TestDownloadPDFContent` (teste le handler). **Garder** `TestGeneratePDFContent` + helpers `pdfLines`/`assertPDFLines`/`samplePDFLines` : le paquet devient une bibliothèque de génération PDF (conservée pour la future persistance invoices).
4. Supprimer `public/html/invoice.html` (seul consommateur = `Preview`). **Garder** `index.html` et `404.html` (fallback SPA).
5. `tests/article_test.go` — retirer l'import `invoice` (ligne 24), l'enregistrement des routes (~lignes 66-70) et les blocs `TestInvoicePreview` (~964-979) + `TestInvoiceDownload` (~981-996).
6. `README.md` — supprimer la ligne 21 (What's Included), la section `### Invoice (Example)` (~183-187) et la section `## Invoice Demo` (~200-208) ; ajuster la ligne ~263 : `invoice/` = bibliothèque de génération PDF (plus de routes).
7. `AGENTS.md` — structure : ajuster `invoice/ — exemple de facture HTML/PDF` → bibliothèque de génération HTML/PDF sans routes (la mention test pdftotext/poppler reste valable).

### T3 — Swagger servi uniquement hors production

1. `main.go:188` — encadrer `r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))` par `if os.Getenv("ENV") != "PRODUCTION" { ... }`.
2. `README.md` — section `## Swagger Docs` (~189-198) : préciser que l'UI n'est pas servie en `ENV=PRODUCTION`.

### T4 — Régénérations (ordre : sqlc puis docs)

1. `make generate_sqlc` après T1.4 (généré commité ; la CI vérifie `git diff --exit-code internal/db`).
2. `make generate_docs` après T1-T2 (annotations des handlers supprimés disparaissent de `docs/`).

### T5 — Corrections documentaires (affirmations devenues fausses)

1. `AGENTS.md` — section Structure : retirer `public.go (lookup tenant public)` de la liste des controllers.
2. `memory-bank/productContext.md:12` — retirer la ligne « Routes publiques : `GET /v1/public/:slug` … ».
3. `memory-bank/progress.md` — ligne 14 : retirer `/v1/public/:slug` de la liste des routes ; ligne 15 : retirer « public » de la liste des tests couverts ; ligne 32 : retirer « throttling `/v1/public/:slug` » de la roadmap.
4. `memory-bank/activeContext.md` — une ligne au « Récent » : retrait intégral de l'accès public (lookup, routes invoice démo, Swagger hors prod), principe « rien de public ».

## Validation (toujours `-mod=readonly`)

1. `go build -mod=readonly .` puis `go vet -mod=readonly ./...`.
2. `go test -mod=readonly -count=1 -race ./models ./invoice` (invoice requiert `pdftotext` ; s'il manque localement, le signaler — la CI l'installe).
3. `git diff --exit-code internal/db` propre après regen.
4. Suite complète : `set -a; . ./.env; set +a` puis `make test` (PostgreSQL + Redis actifs requis — si indisponibles, le signaler explicitement au lieu de conclure en silence).
5. Grep résiduel sur `*.go` : `PublicController|GetTenantBySlug|OneBySlug|/v1/public|invoice/handler|TestPublicTenant|TestInvoicePreview|TestInvoiceDownload` → 0 match (le shell statique `/public` de `r.Static` n'est PAS concerné).

## Hors périmètre

- Rédaction de la spec `specs/multi-tenant.md` (plan dédié, après ce fix).
- Vraie feature invoice tenant-scoped, câblage `RequirePermission`, invitations/membres, SPA, commit/git (aucun demandé).

## Risques

- Référence résiduelle oubliée → mitigé par le grep final + build + vet + suite complète.
- Supprimer `SampleInvoice` par erreur → ne pas y toucher (utilisé par `TestGeneratePDFContent`).
- `make test` dépend de l'environnement (PG/Redis/`.env`) → règle AGENTS.md : jamais terminer avec des tests en échec ou sautés silencieusement.
