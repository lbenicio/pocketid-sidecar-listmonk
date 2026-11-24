.PHONY: build run clean lint docker-build docker-push

BINARY := pocketid-sidecar-listmonk
DOCKERHUB_IMAGE ?= docker.io/lbenicio/$(BINARY)
GHCR_IMAGE ?= ghcr.io/lbenicio/$(BINARY)
VERSION ?= 0.1.1
PLATFORMS ?= linux/amd64,linux/arm64
IMAGE_TAGS = -t $(DOCKERHUB_IMAGE):$(VERSION) -t $(DOCKERHUB_IMAGE):latest \
	-t $(GHCR_IMAGE):$(VERSION) -t $(GHCR_IMAGE):latest

build:
	go build -o bin/$(BINARY) ./cmd/sync

run: build
	./bin/$(BINARY)

clean:
	rm -rf bin/

lint:
	go vet ./...

# Builds a local image for the current Docker daemon platform.
docker-build:
	docker build $(IMAGE_TAGS) .

# Builds and publishes both tags to Docker Hub and GHCR for the configured platforms.
docker-push:
	docker buildx build --platform $(PLATFORMS) --push $(IMAGE_TAGS) .
