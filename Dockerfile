FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pocketid-sidecar-listmonk ./cmd/sync

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/pocketid-sidecar-listmonk /usr/local/bin/pocketid-sidecar-listmonk

ENTRYPOINT ["/usr/local/bin/pocketid-sidecar-listmonk"]
