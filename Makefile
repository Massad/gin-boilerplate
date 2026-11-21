## RUN APPLICATION
run:
	@echo -e "🚀 Running the application..."
	@go run *.go

## RUN TESTS
test:
	@echo -e "🔍 Running tests..."
	@go test -mod=readonly -count=1 -v -tags=all ./...

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
