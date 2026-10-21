.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application locally
run: build
	./bin/server

# Run all tests
test:
	go test ./... -v

# Run unit tests only
test-unit:
	go test ./internal/... -v

# Run database migrations up
migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

# Start all services with Docker Compose
docker-up:
	docker compose up --build -d

# Stop all services
docker-down:
	docker compose down

# Stop all services and remove volumes
docker-clean:
	docker compose down -v

# View logs
docker-logs:
	docker compose logs -f app

# Run smoke tests
smoke-test:
	go test ./tests/smoke/... -v -tags=smoke

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
