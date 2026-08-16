.PHONY: test race vet run frontend-build compose-up compose-down

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

run:
	go run ./cmd/server

frontend-build:
	npm --prefix web run build

compose-up:
	docker compose up --build

compose-down:
	docker compose down
