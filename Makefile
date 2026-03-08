.PHONY: build test lint clean

build:
	go build -o bin/doozer-scaffold ./cmd/doozer-scaffold

test:
	go test ./... -v

test-unit:
	go test ./internal/... -v

test-acceptance:
	go test ./tests/acceptance/ -v

lint:
	golangci-lint run

clean:
	rm -rf bin/
