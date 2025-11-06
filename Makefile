.PHONY: build install test testacc vet fmt lint clean

# Build the provider
build:
	@echo "==> Building the provider..."
	@go build -o terraform-provider-n8n

# Install the provider locally
install: build
	@echo "==> Installing the provider..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/artus-engineering/n8n/1.0.0/darwin_amd64
	@cp terraform-provider-n8n ~/.terraform.d/plugins/registry.terraform.io/artus-engineering/n8n/1.0.0/darwin_amd64/

# Run tests
test:
	@echo "==> Running tests..."
	@go test -v ./...

# Run acceptance tests
testacc:
	@echo "==> Running acceptance tests..."
	@TF_ACC=1 go test -v ./...

# Run go vet
vet:
	@echo "==> Running go vet..."
	@go vet ./...

# Format code
fmt:
	@echo "==> Formatting code..."
	@go fmt ./...
	@terraform fmt -recursive ./examples/

# Run linter
lint:
	@echo "==> Running linter..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "==> Cleaning..."
	@rm -f terraform-provider-n8n
	@go clean -testcache

# Generate documentation
docs:
	@echo "==> Generating documentation..."
	@go generate ./...

# Run pre-commit hooks on all files
pre-commit:
	@echo "==> Running pre-commit hooks..."
	@pre-commit run --all-files

# Run all checks
check: fmt vet lint test
