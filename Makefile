## RUN APPLICATION
run:
	@echo -e "🚀 Running the application..."
	@go run *.go

## RUN TESTS
test:
	@echo -e "🔍 Running tests..."
	@go test -v ./tests/*

## INSTALL SWAG CLI TOOL & PACKAGES
install_swag:
	@echo -e "📥 Installing Swag CLI and dependencies..."
	@which swag >/dev/null 2>&1 || (echo -e "❌ Swag CLI not found! Installing now..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@echo -e "🔄 Updating project dependencies for Swag..."
	@go mod tidy
	@go mod download
	@echo -e "✅ Swag installation complete!"

## GENERATE API DOCUMENTATION
generate_docs: install_swag
	@echo -e "📜 Generating API documentation using Swag..."
	@swag init
	@echo -e "✅ API documentation generated successfully!"