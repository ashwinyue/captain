.PHONY: build run test clean dev

APP_NAME=aicenter
BUILD_DIR=build

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

tidy:
	go mod tidy

lint:
	golangci-lint run

.DEFAULT_GOAL := build
