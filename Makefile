
MIGRATIONS=migrations

docker-test-up:
	docker compose --profile test up -d

docker-test-down:
	docker compose --profile test down -v

docker-prod:
	docker compose --profile prod up -d

migrate-create:
	goose -dir $(MIGRATIONS) create $(name) sql

test-integration:
	go test ./... \
		-tags=integration \
		-v \
		-count=1 \
		-timeout=5m

test:
	go test ./... -v -count=1