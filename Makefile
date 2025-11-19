.PHONY: build run test clean docker-build docker-up docker-down

# Build the application
build:
	go build -o bin/gubernator ./cmd/server

# Run the application locally
run:
	go run ./cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	docker build -t gubernator:latest .

# Start services with Docker Compose
docker-up:
	docker-compose up -d

# Stop services
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f gubernator

# Run the visualization script
visualize:
	python3 scripts/flood_test.py

# Install dependencies
deps:
	go mod download
	go mod tidy

