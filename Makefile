include .env

run:
	sudo docker compose up


migrate-up:
	migrate -path ./migration -database "postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" up

migrate-down:
	migrate -path ./migration -database "postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" down

migrate-force-drop:
	migrate -path ./migration -database "postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" drop -f