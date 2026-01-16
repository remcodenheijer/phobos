.PHONY: build run dev migrate-up migrate-down migrate-create templ clean

# Build the application
build: templ
	go build -o bin/server ./cmd/server

# Run the application
run: build
	./bin/server

# Development mode with hot reload
dev:
	@echo "Starting development server..."
	@templ generate --watch &
	@go run ./cmd/server

# Generate templ files
templ:
	templ generate

# Run database migrations
migrate-up:
	go run ./cmd/server -migrate

# Create a new migration
migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql

# Clean build artifacts
clean:
	rm -rf bin/
	find . -name "*_templ.go" -delete

# Download dependencies
deps:
	go mod download
	go mod tidy

# Vendor dependencies
vendor:
	go mod vendor
