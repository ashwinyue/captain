.PHONY: build run test clean tidy

build:
	go build -o bin/rag ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy
