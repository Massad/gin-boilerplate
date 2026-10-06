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
