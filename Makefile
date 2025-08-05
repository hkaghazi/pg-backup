.PHONY: build run clean docker-build docker-run docker-stop deps test help

# Build the application
build:
	go build -o pg-backup .

# Run the application
run: build
	./pg-backup

# Run once for testing
test-run: build
	./pg-backup -once

# List configured databases
list: build
	./pg-backup -list

# Clean build artifacts and logs
clean:
	rm -f pg-backup
	rm -rf backups/
	rm -f backup.log

# Build Docker image
docker-build:
	docker build -t pg-backup .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# Download dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test ./...

# Validate configuration
validate-config: build
	./pg-backup -config config.example.yaml -list

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  test-run      - Run a one-time backup test"
	@echo "  list          - List configured databases"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker containers"
	@echo "  deps          - Download dependencies"
	@echo "  test          - Run tests"
	@echo "  validate-config - Validate configuration files"
