include .env
export

LOCAL_BIN ?= $(CURDIR)/bin

.PHONY: run
run: ### run app in dev mode
	go run cmd/shortener/main.go

.PHONY: debugg
debugg: ### run app in debugg mode
	go build -gcflags=all="-N -l" cmd/shortener/main.go && ./main

.PHONY: lint
lint: ### run linter
	golangci-lint run ./...

.PHONY: test
test: ### run tests
	go test ./...

.PHONY: migrate-up
migrate-up: ### run migrations
	bin/goose up

.PHONY: reqs
reqs: ### install binary deps to bin/
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
