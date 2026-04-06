.PHONY: build test run help

BIN_DIR = bin
BIN_NAME = hexlet-go-crawler
BIN_PATH = $(BIN_DIR)/$(BIN_NAME)


test:
	go mod tidy
	go test -v ./... -race

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) ./cmd/hexlet-go-crawler


run:
	go run ./cmd/hexlet-go-crawler $(URL)


install:
	go install

lint:
	golangci-lint run ./...

