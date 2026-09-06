.PHONY: all build test clean dev api frontend worker db-up db-down docker-up docker-down

all: build

# Build
build: build-api build-worker build-frontend

build-api:
	cd backend && go build -o ../bin/api ./cmd/api

build-worker:
	cd worker && go build -o ../bin/collector ./cmd/collector

build-frontend:
	cd frontend && npm ci && npm run build

# Test
test: test-api

test-api:
	cd backend && go test -v ./...

# Development
dev: db-up
	@echo "Starting development servers..."
	@echo "Run in separate terminals:"
	@echo "  make api"
	@echo "  make frontend"
	@echo "  make worker (when needed)"

api:
	cd backend && DATABASE_URL="postgres://spendfy:spendfy@localhost:5432/spendfy?sslmode=disable" go run ./cmd/api

frontend:
	cd frontend && npm run dev

worker:
	cd worker && API_URL="http://localhost:8080" go run ./cmd/collector

# Database
db-up:
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@docker-compose exec -T postgres psql -U spendfy -d spendfy -f /docker-entrypoint-initdb.d/001_init.sql 2>/dev/null || true

db-down:
	docker-compose down postgres

db-migrate:
	docker-compose exec -T postgres psql -U spendfy -d spendfy -f /docker-entrypoint-initdb.d/001_init.sql

# Docker
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Clean
clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
