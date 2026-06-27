.PHONY: build run clean lint

BINARY := pocketid-sidecar-listmonk

build:
	go build -o bin/$(BINARY) ./cmd/sync

run: build
	./bin/$(BINARY)

clean:
	rm -rf bin/

lint:
	go vet ./...
