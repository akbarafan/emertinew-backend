run:
	go run main.go

# run seed file in backend/seeds without seeding tracks
run-seed:
	go run main.go -seed

# run database migration sql file in backend/migrations
run-migrations:
	go run main.go -migrate

run-seedtrack:
	go run main.go -seedtrack

tidy:
	go mod tidy

download:
	go mod download

build:
	go build -o main .

build-and-seed:
	go build -o main . && ./main -seed

# docker commands (legacy stack — project emerti-fms)
up:
	docker compose up -d

down-v:
	docker compose down -v

stop:
	docker compose stop

# ─── Command Center stack (project emerti-cc, see deploy/README.md) ─────────
# Kept separate from the targets above so both stacks can run side by side.

DC := docker compose --env-file deploy/.env -f deploy/docker-compose.yml

cc-build:
	$(DC) build api

cc-up:
	$(DC) up -d --build

cc-down:
	$(DC) down

cc-destroy:
	$(DC) down -v

cc-logs:
	$(DC) logs -f api

cc-ps:
	$(DC) ps -a

cc-seed:
	$(DC) run --rm seed

cc-migrate:
	$(DC) run --rm migrate

cc-shell:
	$(DC) exec api sh
