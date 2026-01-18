.PHONY: build run dev migrate-up migrate-down migrate-create templ clean docker-build docker-run docker-stop docker-destroy docker-logs

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

# Docker targets
DOCKER_IMAGE := phobos
DOCKER_CONTAINER := phobos-dev

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container with persistent database volume
docker-run: docker-destroy
	docker run -d --name $(DOCKER_CONTAINER) \
		-p 3000:3000 \
		-v $(PWD)/phobos.db:/app/data/phobos.db \
		$(DOCKER_IMAGE)
	@echo "Container started at http://localhost:3000"

# Stop Docker container (keeps container and database)
docker-stop:
	-docker stop $(DOCKER_CONTAINER) 2>/dev/null

# Destroy Docker container (keeps database volume)
docker-destroy: docker-stop
	-docker rm $(DOCKER_CONTAINER) 2>/dev/null

# View Docker logs
docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

# Rebuild and restart Docker container
docker-restart: docker-build docker-run
