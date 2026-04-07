BINARY := bucket-store-server
PORT ?= 8080
DATA_DIR ?= ./data

.PHONY: build run test lint clean image

build:
	go build -o $(BINARY) .

run: build
	PORT=$(PORT) DATA_DIR=$(DATA_DIR) ./$(BINARY)

test:
	go test -v -race ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -rf $(DATA_DIR)

image:
	podman build -t bucket-store-server -f Containerfile .
