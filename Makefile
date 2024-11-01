include .env.example
export

.PHONY: run
run: ### run app in dev mode
	go run cmd/shortener/main.go

.PHONY: lint
lint: ### run linter
	golangci-lint run ./...
