include .env.example
export

LOCAL_BIN:=$(CURDIR)/bin

.PHONY: run
run: ### run app in dev mode
	go run cmd/shortener/main.go

.PHONY: lint
lint: ### run linter
	golangci-lint run ./...

.PHONY: bin-deps
bin-deps: ### install binary deps to bin/
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
