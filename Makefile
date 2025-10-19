.PHONY: help run build test clean docker-up docker-down migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o reelscore cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Remove build artifacts
	rm -f reelscore
	go clean

docker-up: ## Start Docker services (PostgreSQL, Redis)
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker services logs
	docker-compose logs -f

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run linters

dev: docker-up run ## Start Docker services and run the application

migrate: ## Run database migrations
	go run cmd/server/main.go migrate

db-reset: ## Reset database (down and up)
	docker-compose down -v
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 3
	$(MAKE) migrate
