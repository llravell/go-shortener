include .env.example
export

run: ### run app in dev mode
	go run cmd/shortener/main.go
.PHONY: run

lint: ### run linter
	golangci-lint run ./...
.PHONY: lint
