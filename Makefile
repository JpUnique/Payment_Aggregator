include $(if $(wildcard .env), .env)
export
GOPRIVATE = github.com/subdialia/*
.DEFAULT_GOAL := start

.PHONY: fmt build clean start

fmt:
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@v0.29.0 -w .

build:
	go build

test:
	go test -v ./... -short

start: build
	./fiat-ramp-service --log-level info

clean:
	go clean
	go clean -cache

tidy:
	go mod tidy

run:
	go run main.go