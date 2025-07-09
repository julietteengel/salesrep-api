# Declare phony targets
.PHONY: all build run test swag generate-docs migrate-up migrate-down migrate-create migrate-force docker-up docker-down logs

# Go and tools
GO        := go
SWAG      := swag
MIGRATE   := migrate
BINARY    := main

# Database config (update as needed)
DB_URL := postgres://postgres:password@localhost:5432/salesrep-api?sslmode=disable

# Default target
all: build

# Run the app locally (go build then run)
run:
	$(GO) run ./cmd/server

# Build the binary
build:
	$(GO) build -o $(BINARY) ./cmd/server

# Run tests
test:
	$(GO) test ./... -v

# Generate Swagger docs
generate-docs:
	$(SWAG) init --generalInfo cmd/server/main.go --output docs --parseDependency --parseInternal

# Migrations (use https://github.com/golang-migrate/migrate)
migrate-up:
	$(MIGRATE) -database "$(DB_URL)" -path db/migrations up

migrate-down:
	$(MIGRATE) -database "$(DB_URL)" -path db/migrations down

migrate-create:
	$(MIGRATE) create -ext sql -dir db/migrations -seq $(name)

migrate-force:
	$(MIGRATE) -database "$(DB_URL)" -path db/migrations force $(version)

# Docker helpers
docker-up:
	docker compose --profile dev up --build -d

docker-down:
	docker compose down

logs:
	docker compose logs -f backend-api
