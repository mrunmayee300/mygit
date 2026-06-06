.PHONY: build test test-cover lint clean

BINARY := mygit
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/mygit

test:
	go test ./...

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out
