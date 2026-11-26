include .env.docker
export

CONFIG ?= ./config/main.yaml

run:
	go run ./cmd -config=$(CONFIG)

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

install-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/*

db-up:
	docker compose up -d postgres

db-down:
	docker compose down postgres

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Missing seq parameter. Example: make migrate-create seq=init"; \
		exit 1; \
	fi;

	docker compose run --rm postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-action:
	@docker compose run --rm postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/main.go \
		-o docs \
		--parseInternal \
		--parseDependency