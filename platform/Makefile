.PHONY: build run dev test clean tidy

build:
	go build -o bin/platform ./cmd/server

run: build
	./bin/platform

dev:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy
