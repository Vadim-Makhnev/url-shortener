.PHONY: migrate-up migrate-down docker-run docker-down

docker-run:
	docker compose up --build -d

docker-down:
	docker compose down

DB_URL=postgres://postgres:password@localhost:5432/url_shortener?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down