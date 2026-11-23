## WINDOWS: make (3.81) falls back to cmd.exe and cannot run POSIX recipes
## (mkdir -p, GOBIN= prefix, quoted paths). Route recipes through Git for
## Windows' sh unless the caller overrides SHELL. No-op on Unix.
ifeq ($(OS),Windows_NT)
ifeq ($(origin SHELL),default)
SHELL := C:/Program Files/Git/bin/sh.exe
endif
endif

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

SQLC_VERSION := v1.27.0
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


