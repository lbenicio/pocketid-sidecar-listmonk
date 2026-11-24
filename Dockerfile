FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pocketid-sidecar-listmonk ./cmd/sync

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/pocketid-sidecar-listmonk /usr/local/bin/pocketid-sidecar-listmonk

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/pocketid-sidecar-listmonk"]
