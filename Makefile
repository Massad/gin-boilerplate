## WINDOWS: make (3.81) falls back to cmd.exe and cannot run POSIX recipes
## (mkdir -p, GOBIN= prefix, quoted paths). Route recipes through Git for
## Windows' sh unless the caller overrides SHELL. No-op on Unix.
ifeq ($(OS),Windows_NT)
ifeq ($(origin SHELL),default)
SHELL := C:/Program Files/Git/bin/sh.exe
endif
endif

## The targets never create a file named after themselves: they must always run.
.PHONY: run test e2e migrate_up migrate_down generate_docs install_swag \
	generate_sqlc install_sqlc generate_mocks install_mockery install_goose

## RUN APPLICATION
run:
	@echo -e "🚀 Running the application..."
	@go run .

## RUN TESTS
## - Sources .env first (set -a exports the sourced assignments) so the tagged
##   integration tests see DB_*/REDIS_* without a manual shell export.
## - Drops Git-for-Windows' bundled xpdf pdftotext (v4.00, layout incompatible
##   with the poppler CI uses) from PATH so the invoice PDF content tests
##   resolve a real poppler; the filter is a no-op elsewhere.
## - -p 1 serializes the packages: the suites share one PostgreSQL/Redis and
##   the fixtures purge at suite start must not race the isolated-service
##   packages seeding their own rows (convention AGENTS.md : suite séquentielle).
test:
	@echo -e "🔍 Running tests..."
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	case "$$(uname -s)" in *MSYS*|*MINGW*) \
		PATH=$$(printf '%s' "$$PATH" | tr ':' '\n' | awk '!/^\/(mingw64|usr)\/bin/ && !/\/[Gg]it\/(mingw64|usr)\/bin/' | paste -sd ':' -) ;; \
	esac; \
	go test -mod=readonly -count=1 -p 1 -v -tags=all ./...

SWAG_VERSION := v1.16.6
SWAG_BIN := $(CURDIR)/.cache/tools/swag/$(SWAG_VERSION)/swag

## INSTALL THE PINNED SWAG CLI
install_swag:
	@echo "Installing Swag $(SWAG_VERSION)..."
	@mkdir -p "$(dir $(SWAG_BIN))"
	@GOBIN="$(dir $(SWAG_BIN))" go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

## GENERATE API DOCUMENTATION
generate_docs: install_swag
	@echo -e "📜 Generating API documentation using Swag..."
	@"$(SWAG_BIN)" init
	@echo -e "✅ API documentation generated successfully!"

## v1.27.0 crashed at startup in this environment (old wazero wasm runtime
## incompatible with the current Go toolchain); v1.30.0 vendors a fixed one.
SQLC_VERSION := v1.30.0
SQLC_BIN := $(CURDIR)/.cache/tools/sqlc/$(SQLC_VERSION)/sqlc

## INSTALL THE PINNED SQLC CLI
install_sqlc:
	@echo "Installing sqlc $(SQLC_VERSION)..."
	@mkdir -p "$(dir $(SQLC_BIN))"
	@GOBIN="$(dir $(SQLC_BIN))" go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

## GENERATE SQLC CODE
generate_sqlc: install_sqlc
	@echo "Generating sqlc code..."
	@"$(SQLC_BIN)" generate

MOCKERY_VERSION := v2.53.8
MOCKERY_BIN := $(CURDIR)/.cache/tools/mockery/$(MOCKERY_VERSION)/mockery

## INSTALL THE PINNED MOCKERY CLI
install_mockery:
	@echo "Installing mockery $(MOCKERY_VERSION)..."
	@mkdir -p "$(dir $(MOCKERY_BIN))"
	@GOBIN="$(dir $(MOCKERY_BIN))" go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)

## GENERATE MOCKS FROM .mockery.yaml
generate_mocks: install_mockery
	@echo "Generating mockery mocks..."
	@"$(MOCKERY_BIN)" --config=.mockery.yaml

GOOSE_VERSION := v3.20.0
GOOSE_BIN := $(CURDIR)/.cache/tools/goose/$(GOOSE_VERSION)/goose

## INSTALL THE PINNED GOOSE CLI
install_goose:
	@echo "Installing goose $(GOOSE_VERSION)..."
	@mkdir -p "$(dir $(GOOSE_BIN))"
	@GOBIN="$(dir $(GOOSE_BIN))" go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

## BUILD THE POSTGRES DSN FROM THE DB_* ENVIRONMENT VARIABLES
DB_SSLMODE := $(if $(filter TRUE,$(SSL)),require,disable)
DB_DSN := host='$(DB_HOST)' port='$(DB_PORT)' user='$(DB_USER)' password='$(DB_PASS)' dbname='$(DB_NAME)' sslmode=$(DB_SSLMODE)

## APPLY DATABASE MIGRATIONS
migrate_up: install_goose
	@echo "Applying database migrations..."
	@"$(GOOSE_BIN)" -dir db/migrations postgres "$(DB_DSN)" up

## REVERT THE LAST DATABASE MIGRATION
migrate_down: install_goose
	@echo "Reverting the last database migration..."
	@"$(GOOSE_BIN)" -dir db/migrations postgres "$(DB_DSN)" down

## E2E DATABASE (recreated from scratch at every run of the e2e cycle)
E2E_DB_NAME ?= artisan_e2e

## THE FRONTEND COMMANDS RUN INSIDE app/ SO THAT COREPACK RESOLVES THE PINNED
## packageManager VERSION (corepack reads the nearest package.json). From the
## repository root, corepack resolves its default version instead, which the
## pinned project refuses (ERR_PNPM_BAD_PM_VERSION). Override PNPM to use
## another binary (PNPM=pnpm make e2e).
PNPM ?= corepack pnpm

## RUN THE END-TO-END CYCLE
## - Recreates the e2e database and flushes the Redis database of the sessions.
## - Applies the migrations as the migration superuser, then seeds the superadmin.
## - Builds the frontend and runs the Playwright suite against the Go binary
##   serving app/dist (single origin, port 9100).
e2e:
	@echo -e "🧪 Preparing the e2e database ($(E2E_DB_NAME))..."
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	DB_USER=postgres DB_PASS=postgres E2E_DB_NAME=$(E2E_DB_NAME) go run -mod=readonly ./cmd/e2edb
	@echo -e "🗄️  Applying the migrations on $(E2E_DB_NAME)..."
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	"$(MAKE)" --no-print-directory SHELL="$(SHELL)" migrate_up DB_NAME=$(E2E_DB_NAME) DB_USER=postgres DB_PASS=postgres DB_HOST="$$DB_HOST" DB_PORT="$$DB_PORT" SSL="$$SSL"
	@echo -e "👤 Seeding the e2e superadmin..."
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	DB_NAME=$(E2E_DB_NAME) go run -mod=readonly . -create-admin -email e2e-admin@atrium.dev -password e2e-admin-password-0 -name "Prénom Nom"
	@echo -e "🏗️  Building the frontend..."
	@cd app && $(PNPM) build
	@echo -e "🎭 Running the Playwright suite..."
	@cd app && E2E_DB_NAME=$(E2E_DB_NAME) $(PNPM) exec playwright test


