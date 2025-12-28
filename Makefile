.PHONY: build test test-coverage lint fmt clean run docker-up docker-down tidy

# Build the application
build:
	go build -o bin/golwarc ./main.go

# Run the application
run:
	go run main.go

# Run all tests
test:
	go test -v -coverpkg=./cache/...,./configs/...,./crawlers/...,./database/...,./inject/...,./libs/...,./message-queue/...,./models/...,./services/... ./tests/...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -coverpkg=./cache/...,./configs/...,./crawlers/...,./database/...,./inject/...,./libs/...,./message-queue/...,./models/...,./services/... ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "Coverage Summary:"
	@go tool cover -func=coverage.out | grep total

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Start Docker services
docker-up:
	docker-compose -f docker/docker-compose.yaml up -d

# Stop Docker services
docker-down:
	docker-compose -f docker/docker-compose.yaml down

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy go.mod"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-up     - Start Docker services"
	@echo "  docker-down   - Stop Docker services"
