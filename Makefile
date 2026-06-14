include .env
export 

run:
	go run cmd/talon/main.go 

build:
	go build ./...

test:
	go test ./...

db-up:
	docker compose up -d db

